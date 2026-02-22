package channelserver

import (
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

	if err != nil {
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
		isApplicant, _ := s.server.guildRepo.HasApplication(guild.ID, s.charID)
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
		s.logger.Error("Failed to get alliance data")
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

	if guild.AllianceID > 0 {
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
		guild, err := s.server.guildRepo.GetByID(s.prevGuildID)
		s.prevGuildID = 0
		if guild == nil || err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(guild.MemberCount))
	members, _ := s.server.guildRepo.GetMembers(guild.ID, false)
	for _, member := range members {
		bf.WriteUint32(member.CharID)
		bf.WriteBool(member.Recruiter)
		bf.WriteBytes(make([]byte, 3))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdGuildMapInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdGuildMapInfo)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
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

	if err != nil {
		s.logger.Error("Failed to get guild info for icon update", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	characterInfo, err := s.server.guildRepo.GetCharacterMembership(s.charID)

	if err != nil {
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
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateForceGuildRank(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGenerateUdGuildMap(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGenerateUdGuildMap)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuild(s *Session, p mhfpacket.MHFPacket) {}

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

func handleMsgMhfUpdateGuildcard(s *Session, p mhfpacket.MHFPacket) {}

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
