package channelserver

import (
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfOperateGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuild)

	guild, err := s.server.guildRepo.GetByID(pkt.GuildID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	characterGuildInfo, err := s.server.guildRepo.GetCharacterMembership(s.charID)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()

	switch pkt.Action {
	case mhfpacket.OperateGuildDisband:
		result, err := s.server.guildService.Disband(s.charID, guild.ID)
		if err != nil {
			s.logger.Error("Failed to disband guild", zap.Error(err))
		}
		response := 0
		if result != nil && result.Success {
			response = 1
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildResign:
		result, err := s.server.guildService.ResignLeadership(s.charID, guild.ID)
		if err == nil && result.NewLeaderCharID != 0 {
			bf.WriteUint32(result.NewLeaderCharID)
		}
	case mhfpacket.OperateGuildApply:
		err = s.server.guildRepo.CreateApplication(guild.ID, s.charID, s.charID, GuildApplicationTypeApplied)
		if err == nil {
			bf.WriteUint32(guild.LeaderCharID)
		} else {
			bf.WriteUint32(0)
		}
	case mhfpacket.OperateGuildLeave:
		isApplicant := characterGuildInfo != nil && characterGuildInfo.IsApplicant
		result, err := s.server.guildService.Leave(s.charID, guild.ID, isApplicant, guild.Name)
		if err != nil {
			s.logger.Error("Failed to leave guild", zap.Error(err))
		}
		response := 0
		if result != nil && result.Success {
			response = 1
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildDonateRank:
		bf.WriteBytes(handleDonateRP(s, uint16(pkt.Data1.ReadUint32()), guild, 0))
	case mhfpacket.OperateGuildSetApplicationDeny:
		if err := s.server.guildRepo.SetRecruiting(guild.ID, false); err != nil {
			s.logger.Error("Failed to deny guild applications", zap.Error(err))
		}
	case mhfpacket.OperateGuildSetApplicationAllow:
		if err := s.server.guildRepo.SetRecruiting(guild.ID, true); err != nil {
			s.logger.Error("Failed to allow guild applications", zap.Error(err))
		}
	case mhfpacket.OperateGuildSetAvoidLeadershipTrue:
		handleAvoidLeadershipUpdate(s, pkt, true)
	case mhfpacket.OperateGuildSetAvoidLeadershipFalse:
		handleAvoidLeadershipUpdate(s, pkt, false)
	case mhfpacket.OperateGuildUpdateComment:
		if characterGuildInfo == nil || (!characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader()) {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		guild.Comment = stringsupport.SJISToUTF8Lossy(pkt.Data2.ReadNullTerminatedBytes())
		if err := s.server.guildRepo.Save(guild); err != nil {
			s.logger.Error("Failed to save guild comment", zap.Error(err))
		}
	case mhfpacket.OperateGuildUpdateMotto:
		if characterGuildInfo == nil || (!characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader()) {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		_ = pkt.Data1.ReadUint16()
		guild.SubMotto = pkt.Data1.ReadUint8()
		guild.MainMotto = pkt.Data1.ReadUint8()
		if err := s.server.guildRepo.Save(guild); err != nil {
			s.logger.Error("Failed to save guild motto", zap.Error(err))
		}
	case mhfpacket.OperateGuildRenamePugi1:
		handleRenamePugi(s, pkt.Data2, guild, 1)
	case mhfpacket.OperateGuildRenamePugi2:
		handleRenamePugi(s, pkt.Data2, guild, 2)
	case mhfpacket.OperateGuildRenamePugi3:
		handleRenamePugi(s, pkt.Data2, guild, 3)
	case mhfpacket.OperateGuildChangePugi1:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 1)
	case mhfpacket.OperateGuildChangePugi2:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 2)
	case mhfpacket.OperateGuildChangePugi3:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 3)
	case mhfpacket.OperateGuildUnlockOutfit:
		if err := s.server.guildRepo.SetPugiOutfits(guild.ID, pkt.Data1.ReadUint32()); err != nil {
			s.logger.Error("Failed to unlock guild pugi outfit", zap.Error(err))
		}
	case mhfpacket.OperateGuildDonateRoom:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 2))
	case mhfpacket.OperateGuildDonateEvent:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 1))
		if err := s.server.guildRepo.AddMemberDailyRP(s.charID, quantity); err != nil {
			s.logger.Error("Failed to update guild character daily RP", zap.Error(err))
		}
	case mhfpacket.OperateGuildEventExchange:
		rp := uint16(pkt.Data1.ReadUint32())
		balance, err := s.server.guildRepo.ExchangeEventRP(guild.ID, rp)
		if err != nil {
			s.logger.Error("Failed to exchange guild event RP", zap.Error(err))
		}
		bf.WriteUint32(balance)
	case mhfpacket.OperateGuildGraduateRookie, mhfpacket.OperateGuildGraduateReturn:
		// Player graduates (leaves) a temporary return/rookie guild.
		// No extra packet data — just remove and succeed.
		isApplicant := characterGuildInfo != nil && characterGuildInfo.IsApplicant
		if _, err := s.server.guildService.Leave(s.charID, guild.ID, isApplicant, guild.Name); err != nil {
			s.logger.Error("Failed to graduate from return guild", zap.Error(err))
		}
	default:
		s.logger.Error("unhandled operate guild action", zap.Uint8("action", uint8(pkt.Action)))
	}

	if len(bf.Data()) > 0 {
		doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleRenamePugi(s *Session, bf *byteframe.ByteFrame, guild *Guild, num int) {
	name := stringsupport.SJISToUTF8Lossy(bf.ReadNullTerminatedBytes())
	switch num {
	case 1:
		guild.PugiName1 = name
	case 2:
		guild.PugiName2 = name
	default:
		guild.PugiName3 = name
	}
	if err := s.server.guildRepo.Save(guild); err != nil {
		s.logger.Error("Failed to save guild pugi name", zap.Error(err))
	}
}

func handleChangePugi(s *Session, outfit uint8, guild *Guild, num int) {
	switch num {
	case 1:
		guild.PugiOutfit1 = outfit
	case 2:
		guild.PugiOutfit2 = outfit
	case 3:
		guild.PugiOutfit3 = outfit
	}
	if err := s.server.guildRepo.Save(guild); err != nil {
		s.logger.Error("Failed to save guild pugi outfit", zap.Error(err))
	}
}

func handleDonateRP(s *Session, amount uint16, guild *Guild, _type int) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	saveData, err := GetCharacterSaveData(s, s.charID)
	if err != nil {
		return bf.Data()
	}
	var resetRoom bool
	if _type == 2 {
		currentRP, err := s.server.guildRepo.GetRoomRP(guild.ID)
		if err != nil {
			s.logger.Error("Failed to get guild room RP", zap.Error(err))
		}
		if currentRP+amount >= 30 {
			amount = 30 - currentRP
			resetRoom = true
		}
	}
	saveData.RP -= amount
	if err := saveData.Save(s); err != nil {
		s.logger.Error("Failed to save RP after guild donation", zap.Error(err))
	}
	switch _type {
	case 0:
		if err := s.server.guildRepo.AddRankRP(guild.ID, amount); err != nil {
			s.logger.Error("Failed to update guild rank RP", zap.Error(err))
		}
	case 1:
		if err := s.server.guildRepo.AddEventRP(guild.ID, amount); err != nil {
			s.logger.Error("Failed to update guild event RP", zap.Error(err))
		}
	case 2:
		if resetRoom {
			if err := s.server.guildRepo.SetRoomRP(guild.ID, 0); err != nil {
				s.logger.Error("Failed to reset guild room RP", zap.Error(err))
			}
			if err := s.server.guildRepo.SetRoomExpiry(guild.ID, TimeAdjusted().Add(time.Hour*24*7)); err != nil {
				s.logger.Error("Failed to update guild room expiry", zap.Error(err))
			}
		} else {
			if err := s.server.guildRepo.AddRoomRP(guild.ID, amount); err != nil {
				s.logger.Error("Failed to update guild room RP", zap.Error(err))
			}
		}
	}
	_, _ = bf.Seek(0, 0)
	bf.WriteUint32(uint32(saveData.RP))
	return bf.Data()
}

func handleAvoidLeadershipUpdate(s *Session, pkt *mhfpacket.MsgMhfOperateGuild, avoidLeadership bool) {
	characterGuildData, err := s.server.guildRepo.GetCharacterMembership(s.charID)

	if err != nil || characterGuildData == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	characterGuildData.AvoidLeadership = avoidLeadership

	err = s.server.guildRepo.SaveMember(characterGuildData)

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfOperateGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildMember)

	action, ok := mapMemberAction(pkt.Action)
	if !ok {
		s.logger.Warn("Unhandled operateGuildMember action", zap.Uint8("action", pkt.Action))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	result, err := s.server.guildService.OperateMember(s.charID, pkt.CharID, action)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	s.server.Registry.NotifyMailToCharID(result.MailRecipientID, s, &result.Mail)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func mapMemberAction(proto uint8) (GuildMemberAction, bool) {
	switch proto {
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_ACCEPT:
		return GuildMemberActionAccept, true
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_REJECT:
		return GuildMemberActionReject, true
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_KICK:
		return GuildMemberActionKick, true
	default:
		return 0, false
	}
}
