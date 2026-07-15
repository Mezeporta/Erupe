package channelserver

import (
	"encoding/json"
	"sort"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	cfg "erupe-ce/config"

	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfCreateGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateGuild)

	guildId, err := s.server.guildRepo.Create(s.charID, pkt.Name)

	if err != nil {
		bf := byteframe.NewByteFrame()

		// No reasoning behind these values other than they cause a 'failed to create'
		// style message, it's better than nothing for now.
		bf.WriteUint32(0x01010101)

		doAckSimpleFail(s, pkt.AckHandle, bf.Data())
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint32(uint32(guildId))

	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfArrangeGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfArrangeGuildMember)

	guild, err := s.server.guildRepo.GetByID(pkt.GuildID)

	if err != nil || guild == nil {
		s.logger.Error(
			"failed to respond to ArrangeGuildMember message",
			zap.Uint32("charID", s.charID),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if guild.LeaderCharID != s.charID {
		s.logger.Error("non leader attempting to rearrange guild members!",
			zap.Uint32("charID", s.charID),
			zap.Uint32("guildID", guild.ID),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	err = s.server.guildRepo.ArrangeCharacters(pkt.CharIDs)

	if err != nil {
		s.logger.Error(
			"failed to respond to ArrangeGuildMember message",
			zap.Uint32("charID", s.charID),
			zap.Uint32("guildID", guild.ID),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMember)

	var guild *Guild
	var err error

	if pkt.GuildID > 0 {
		guild, err = s.server.guildRepo.GetByID(pkt.GuildID)
	} else {
		guild, err = s.server.guildRepo.GetByCharID(s.charID)
	}

	if guild != nil {
		isApplicant, appErr := s.server.guildRepo.HasApplication(guild.ID, s.charID)
		if appErr != nil {
			s.logger.Warn("Failed to check guild application status", zap.Error(appErr))
		}
		if isApplicant {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
			return
		}
	}

	if guild == nil && s.prevGuildID > 0 {
		guild, err = s.server.guildRepo.GetByID(s.prevGuildID)
	}

	if err != nil {
		s.logger.Warn("failed to retrieve guild sending no result message")
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	} else if guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	}

	// Lazy daily RP rollover: move rp_today → rp_yesterday at noon
	midday := TimeMidnight().Add(12 * time.Hour)
	if TimeAdjusted().Before(midday) {
		midday = midday.Add(-24 * time.Hour)
	}
	if guild.RPResetAt.Before(midday) {
		if err := s.server.guildRepo.RolloverDailyRP(guild.ID, midday); err != nil {
			s.logger.Error("Failed to rollover guild daily RP", zap.Error(err))
		}
	}

	guildMembers, err := s.server.guildRepo.GetMembers(guild.ID, false)

	if err != nil {
		s.logger.Error("failed to retrieve guild")
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	alliance, err := s.server.guildRepo.GetAllianceByID(guild.AllianceID)
	if err != nil {
		s.logger.Error("Failed to get alliance data", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(uint16(len(guildMembers)))

	sort.Slice(guildMembers[:], func(i, j int) bool {
		return guildMembers[i].OrderIndex < guildMembers[j].OrderIndex
	})

	for _, member := range guildMembers {
		bf.WriteUint32(member.CharID)
		bf.WriteUint16(member.HR)
		if s.server.erupeConfig.RealClientMode >= cfg.G10 {
			bf.WriteUint16(member.GR)
		}
		if s.server.erupeConfig.RealClientMode < cfg.ZZ {
			// Magnet Spike crash workaround
			bf.WriteUint16(0)
		} else {
			bf.WriteUint16(member.WeaponID)
		}
		if member.WeaponType == 1 || member.WeaponType == 5 || member.WeaponType == 10 { // If weapon is ranged
			bf.WriteUint8(7)
		} else {
			bf.WriteUint8(6)
		}
		bf.WriteUint16(member.OrderIndex)
		bf.WriteBool(member.AvoidLeadership)
		ps.Uint8(bf, member.Name, true)
	}

	for _, member := range guildMembers {
		bf.WriteUint32(member.LastLogin)
	}

	if guild.AllianceID > 0 && alliance != nil {
		bf.WriteUint16(alliance.TotalMembers - uint16(len(guildMembers)))
		if guild.ID != alliance.ParentGuildID {
			mems, err := s.server.guildRepo.GetMembers(alliance.ParentGuildID, false)
			if err != nil {
				s.logger.Error("Failed to get parent guild members for alliance", zap.Error(err))
				doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
				return
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
		if guild.ID != alliance.SubGuild1ID {
			mems, err := s.server.guildRepo.GetMembers(alliance.SubGuild1ID, false)
			if err != nil {
				s.logger.Error("Failed to get sub guild 1 members for alliance", zap.Error(err))
				doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
				return
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
		if guild.ID != alliance.SubGuild2ID {
			mems, err := s.server.guildRepo.GetMembers(alliance.SubGuild2ID, false)
			if err != nil {
				s.logger.Error("Failed to get sub guild 2 members for alliance", zap.Error(err))
				doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
				return
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
	} else {
		bf.WriteUint16(0)
	}

	for _, member := range guildMembers {
		bf.WriteUint16(member.RPToday)
		bf.WriteUint16(member.RPYesterday)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildManageRight)

	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	if guild == nil || s.prevGuildID != 0 {
		var err error
		guild, err = s.server.guildRepo.GetByID(s.prevGuildID)
		s.prevGuildID = 0
		if guild == nil || err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(guild.MemberCount))
	members, err := s.server.guildRepo.GetMembers(guild.ID, false)
	if err != nil {
		s.logger.Error("Failed to get guild members for manage right", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	for _, member := range members {
		bf.WriteUint32(member.CharID)
		bf.WriteBool(member.Recruiter)
		bf.WriteBytes(make([]byte, 3))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdGuildMapInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdGuildMapInfo)

	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil || guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0xFF})
		return
	}
	isApplicant, appErr := s.server.guildRepo.HasApplication(guild.ID, s.charID)
	if appErr != nil {
		s.logger.Warn("Failed to check guild application status", zap.Error(appErr))
	}
	if isApplicant {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0xFF})
		return
	}

	data, err := s.server.guildRepo.GetInterceptionMaps(guild.ID)
	if err != nil {
		s.logger.Error("Failed to load interception map data", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, []byte{0xFF})
		return
	}
	var interceptionMaps InterceptionMaps
	if len(data) > 0 {
		if err := json.Unmarshal(data, &interceptionMaps); err != nil {
			s.logger.Error("Failed to parse interception map data", zap.Error(err))
			doAckBufSucceed(s, pkt.AckHandle, []byte{0xFF})
			return
		}
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // No error
	var tilesClaimed uint32
	currentMapID, prevMapID := interceptionMaps.CurrPrevID()
	currProg := byteframe.NewByteFrame()
	prevProg := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(interceptionMaps.Maps)))
	for _, _map := range interceptionMaps.Maps {
		bf.WriteUint32(_map.ID)
		bf.WriteUint32(_map.NextID)
		for _, tile := range _map.Tiles {
			bf.WriteUint16(tile.ID)
			bf.WriteUint16(tile.NextID)
			bf.WriteUint16(tile.BranchID)
			bf.WriteUint16(tile.QuestFile1)
			bf.WriteUint16(tile.QuestFile2)
			bf.WriteUint16(tile.QuestFile3)
			bf.WriteUint8(tile.BranchIndex)
			bf.WriteUint8(tile.Type)
			bf.WriteInt32(tile.PointsReq)

			bf.WriteUint8(tile.Unk1)
			bf.WriteUint32(tile.Unk2)
		}
		bf.WriteBytes(make([]byte, 23*(64-len(_map.Tiles)))) // Fill out 64 tiles

		if _map.Completed() && _map.ID != prevMapID {
			tilesClaimed += _map.GetClaimed()
		}

		if _map.ID == currentMapID {
			currProg.WriteUint32(_map.ID)
			currProg.WriteUint16(1)
			currProg.WriteUint8(uint8(len(_map.Tiles)))
			for _, tile := range _map.Tiles {
				if tile.Type != 1 {
					if _map.Points[tile.QuestFile1]-tile.PointsReq > 0 {
						tile.Claimed = true
						tilesClaimed++
						_map.Points[tile.QuestFile1] -= tile.PointsReq
						currProg.WriteInt32(tile.PointsReq)
					} else {
						currProg.WriteInt32(_map.Points[tile.QuestFile1])
						_map.Points[tile.QuestFile1] = 0
					}
				} else {
					currProg.WriteInt32(0)
				}
				currProg.WriteInt32(tile.PointsReq)
				currProg.WriteUint16(tile.ID)
				currProg.WriteUint16(tile.NextID)
				currProg.WriteUint16(tile.BranchID)
				currProg.WriteUint16(tile.QuestFile1)
				currProg.WriteUint16(tile.QuestFile2)
				currProg.WriteUint16(tile.QuestFile3)
				currProg.WriteUint8(tile.BranchIndex)
				currProg.WriteUint8(tile.Type)
				if tile.Claimed || tile.Type == 1 {
					currProg.WriteBool(true)
				} else {
					currProg.WriteBool(false)
				}
			}
		}
		if _map.ID == prevMapID {
			prevProg.WriteUint32(_map.ID)
			prevProg.WriteUint16(1)
			prevProg.WriteUint8(uint8(len(_map.Tiles)))
			for _, tile := range _map.Tiles {
				if tile.Type != 1 {
					if _map.Points[tile.QuestFile1]-tile.PointsReq > 0 {
						tile.Claimed = true
						tilesClaimed++
						_map.Points[tile.QuestFile1] -= tile.PointsReq
						prevProg.WriteInt32(tile.PointsReq)
					} else {
						prevProg.WriteInt32(_map.Points[tile.QuestFile1])
						_map.Points[tile.QuestFile1] = 0
					}
				} else {
					prevProg.WriteInt32(0)
				}
				prevProg.WriteInt32(tile.PointsReq)
				prevProg.WriteUint16(tile.ID)
				prevProg.WriteUint16(tile.NextID)
				prevProg.WriteUint16(tile.BranchID)
				prevProg.WriteUint16(tile.QuestFile1)
				prevProg.WriteUint16(tile.QuestFile2)
				prevProg.WriteUint16(tile.QuestFile3)
				prevProg.WriteUint8(tile.BranchIndex)
				prevProg.WriteUint8(tile.Type)
				if tile.Claimed || tile.Type == 1 {
					prevProg.WriteBool(true)
				} else {
					prevProg.WriteBool(false)
				}
			}
		}
	}

	bf.WriteUint16(uint16(len(interceptionMaps.Branches)))
	for _, branch := range interceptionMaps.Branches {
		bf.WriteUint32(branch.MapIndex)
		bf.WriteUint8(branch.ItemType)
		bf.WriteUint16(branch.ItemID)
		bf.WriteUint16(branch.Quantity)
		bf.WriteUint16(branch.TileIndex1)
		bf.WriteUint16(branch.TileIndex2)
		bf.WriteUint8(branch.ChestType)
	}

	if prevMapID > 0 {
		bf.WriteUint8(2)
	} else {
		bf.WriteUint8(1)
	}
	bf.WriteBytes(currProg.Data())
	if prevMapID > 0 {
		bf.WriteBytes(prevProg.Data())
	}
	bf.WriteUint32(tilesClaimed)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildTargetMemberNum(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTargetMemberNum)

	var guild *Guild
	var err error

	if pkt.GuildID == 0x0 {
		guild, err = s.server.guildRepo.GetByCharID(s.charID)
	} else {
		guild, err = s.server.guildRepo.GetByID(pkt.GuildID)
	}

	if err != nil || guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x02})
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(0x0)
	bf.WriteUint16(guild.MemberCount - 1)

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateGuildItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildItem)
	items := guildGetItems(s, pkt.GuildID)
	bf := byteframe.NewByteFrame()
	bf.WriteBytes(mhfitem.SerializeWarehouseItems(items))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateGuildItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildItem)
	newStacks := mhfitem.DiffItemStacks(guildGetItems(s, pkt.GuildID), pkt.UpdatedItems)
	if err := s.server.guildRepo.SaveItemBox(pkt.GuildID, mhfitem.SerializeWarehouseItems(newStacks)); err != nil {
		s.logger.Error("Failed to update guild item box", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuildIcon(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildIcon)

	guild, err := s.server.guildRepo.GetByID(pkt.GuildID)

	if err != nil || guild == nil {
		s.logger.Error("Failed to get guild info for icon update", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	characterInfo, err := s.server.guildRepo.GetCharacterMembership(s.charID)

	if err != nil || characterInfo == nil {
		s.logger.Error("Failed to get character guild data for icon update", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if !characterInfo.IsSubLeader() && !characterInfo.IsLeader {
		s.logger.Warn(
			"character without leadership attempting to update guild icon",
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", s.charID),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	icon := &GuildIcon{}

	icon.Parts = make([]GuildIconPart, len(pkt.IconParts))

	for i, p := range pkt.IconParts {
		icon.Parts[i] = GuildIconPart{
			Index:    p.Index,
			ID:       p.ID,
			Page:     p.Page,
			Size:     p.Size,
			Rotation: p.Rotation,
			Red:      p.Red,
			Green:    p.Green,
			Blue:     p.Blue,
			PosX:     p.PosX,
			PosY:     p.PosY,
		}
	}

	guild.Icon = icon

	err = s.server.guildRepo.Save(guild)

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfReadGuildcard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadGuildcard)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEntryRookieGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryRookieGuild)

	// pkt.Unk==0: fresh rookie entering a rookie guild (return_type=1).
	// pkt.Unk>=1: returning player entering a comeback/return guild (return_type=2).
	returnType := uint8(1)
	nameTemplate := s.I18n().guild.rookieGuildName
	if pkt.Unk >= 1 {
		returnType = 2
		nameTemplate = s.I18n().guild.returnGuildName
	}

	guildID, err := s.server.guildRepo.FindOrCreateReturnGuild(returnType, nameTemplate)
	if err != nil {
		s.logger.Error("failed to find/create return guild",
			zap.Uint32("charID", s.charID),
			zap.Error(err),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if err := s.server.guildRepo.AddMember(guildID, s.charID); err != nil {
		s.logger.Error("failed to add character to return guild",
			zap.Uint32("charID", s.charID),
			zap.Uint32("guildID", guildID),
			zap.Error(err),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(guildID)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateForceGuildRank(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgMhfGenerateUdGuildMap(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGenerateUdGuildMap)

	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil || guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0xFF})
		return
	}
	isApplicant, appErr := s.server.guildRepo.HasApplication(guild.ID, s.charID)
	if appErr != nil {
		s.logger.Warn("Failed to check guild application status", zap.Error(appErr))
	}
	if isApplicant {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0xFF})
		return
	}

	interceptionMaps := InterceptionMaps{}
	interceptionMaps.Maps, interceptionMaps.Branches = GenerateUdGuildMaps()

	data, err := json.Marshal(interceptionMaps)
	if err != nil {
		s.logger.Error("Failed to serialize interception map data", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if err := s.server.guildRepo.SaveInterceptionMaps(guild.ID, data); err != nil {
		s.logger.Error("Failed to save interception map data", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckBufSucceed(s, pkt.AckHandle, []byte{0})
}

func handleMsgMhfUpdateGuild(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgMhfSetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildManageRight)
	if err := s.server.guildRepo.SetRecruiter(pkt.CharID, pkt.Allowed); err != nil {
		s.logger.Error("Failed to update guild manage right", zap.Error(err))
	}
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// monthlyTypeString maps the packet's Type field to the DB column prefix.
func monthlyTypeString(t uint8) string {
	switch t {
	case 0:
		return "monthly"
	case 1:
		return "monthly_hl"
	case 2:
		return "monthly_ex"
	default:
		return ""
	}
}

func handleMsgMhfCheckMonthlyItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckMonthlyItem)

	typeStr := monthlyTypeString(pkt.Type)
	if typeStr == "" {
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		return
	}

	claimed, err := s.server.stampRepo.GetMonthlyClaimed(s.charID, typeStr)
	if err != nil || claimed.Before(TimeMonthStart()) {
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
}

func handleMsgMhfAcquireMonthlyItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyItem)

	typeStr := monthlyTypeString(pkt.Unk0)
	if typeStr != "" {
		if err := s.server.stampRepo.SetMonthlyClaimed(s.charID, typeStr, TimeAdjusted()); err != nil {
			s.logger.Error("Failed to set monthly item claimed", zap.Error(err))
		}
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateInvGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateInvGuild)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfOperationInvGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperationInvGuild)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuildcard(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

// guildGetItems reads and parses the guild item box.
func guildGetItems(s *Session, guildID uint32) []mhfitem.MHFItemStack {
	data, err := s.server.guildRepo.GetItemBox(guildID)
	if err != nil {
		s.logger.Error("Failed to get guild item box", zap.Error(err))
		return nil
	}
	var items []mhfitem.MHFItemStack
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			items = append(items, mhfitem.ReadWarehouseItem(box))
		}
	}
	return items
}
