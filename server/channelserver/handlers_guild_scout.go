package channelserver

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfPostGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGuildScout)

	err := s.server.guildService.PostScout(s.charID, pkt.CharID, ScoutInviteStrings{
		Title: s.server.i18n.guild.invite.title,
		Body:  s.server.i18n.guild.invite.body,
	})

	if errors.Is(err, ErrAlreadyInvited) {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x04})
		return
	}
	if err != nil {
		s.logger.Error("Failed to post guild scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfCancelGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildScout)

	guildCharData, err := s.server.guildRepo.GetCharacterMembership(s.charID)

	if err != nil {
		s.logger.Error("Failed to get character guild data for cancel scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if guildCharData == nil || !guildCharData.CanRecruit() {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	guild, err := s.server.guildRepo.GetByID(guildCharData.GuildID)

	if err != nil || guild == nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	err = s.server.guildRepo.CancelInvitation(guild.ID, pkt.InvitationID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAnswerGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAnswerGuildScout)

	i := s.server.i18n.guild.invite
	result, err := s.server.guildService.AnswerScout(s.charID, pkt.LeaderID, pkt.Answer, AnswerScoutStrings{
		SuccessTitle:  i.success.title,
		SuccessBody:   i.success.body,
		AcceptedTitle: i.accepted.title,
		AcceptedBody:  i.accepted.body,
		RejectedTitle: i.rejected.title,
		RejectedBody:  i.rejected.body,
		DeclinedTitle: i.declined.title,
		DeclinedBody:  i.declined.body,
	})

	if err != nil && !errors.Is(err, ErrApplicationMissing) {
		s.logger.Error("Failed to answer guild scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, nil)
		return
	}

	bf := byteframe.NewByteFrame()
	if result != nil && result.Success {
		bf.WriteUint32(0)
		bf.WriteUint32(result.GuildID)
	} else {
		if errors.Is(err, ErrApplicationMissing) {
			s.logger.Warn("Guild invite missing, deleted?",
				zap.Uint32("charID", s.charID))
		}
		bf.WriteUint32(7)
		if result != nil {
			bf.WriteUint32(result.GuildID)
		} else {
			bf.WriteUint32(0)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildScoutList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildScoutList)

	guildInfo, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil {
		s.logger.Warn("Failed to get guild for scout list", zap.Error(err))
	}

	if guildInfo == nil {
		if s.prevGuildID == 0 {
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		guildInfo, err = s.server.guildRepo.GetByID(s.prevGuildID)
		if guildInfo == nil || err != nil {
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	chars, err := s.server.guildRepo.ListInvitedCharacters(guildInfo.ID)
	if err != nil {
		s.logger.Error("failed to retrieve scouted characters", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.SetBE()
	bf.WriteUint32(uint32(len(chars)))

	for _, sc := range chars {
		// This seems to be used as a unique ID for the invitation sent
		// we can just use the charID and then filter on guild_id+charID when performing operations
		// this might be a problem later with mails sent referencing IDs but we'll see.
		bf.WriteUint32(sc.CharID)
		bf.WriteUint32(sc.ActorID)
		bf.WriteUint32(sc.CharID)
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
		bf.WriteUint16(sc.HR)
		bf.WriteUint16(sc.GR)
		bf.WriteBytes(stringsupport.PaddedString(sc.Name, 32, true))
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRejectGuildScout)

	currentStatus, err := s.server.charRepo.ReadBool(s.charID, "restrict_guild_scout")

	if err != nil {
		s.logger.Error(
			"failed to retrieve character guild scout status",
			zap.Error(err),
			zap.Uint32("charID", s.charID),
		)
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}

	response := uint8(0x00)

	if currentStatus {
		response = 0x01
	}

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, response})
}

func handleMsgMhfSetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetRejectGuildScout)

	err := s.server.charRepo.SaveBool(s.charID, "restrict_guild_scout", pkt.Reject)

	if err != nil {
		s.logger.Error(
			"failed to update character guild scout status",
			zap.Error(err),
			zap.Uint32("charID", s.charID),
		)
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, nil)
}
