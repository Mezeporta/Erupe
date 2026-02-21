package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
)

func handleMsgMhfPostGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGuildScout)

	actorCharGuildData, err := s.server.guildRepo.GetCharacterMembership(s.charID)

	if err != nil {
		s.logger.Error("Failed to get character guild data for scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if actorCharGuildData == nil || !actorCharGuildData.CanRecruit() {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	guildInfo, err := s.server.guildRepo.GetByID(actorCharGuildData.GuildID)

	if err != nil {
		s.logger.Error("Failed to get guild info for scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	hasApplication, err := s.server.guildRepo.HasApplication(guildInfo.ID, pkt.CharID)

	if err != nil {
		s.logger.Error("Failed to check application for scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if hasApplication {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x04})
		return
	}

	err = s.server.guildRepo.CreateApplicationWithMail(
		guildInfo.ID, pkt.CharID, s.charID, GuildApplicationTypeInvited,
		s.charID, pkt.CharID,
		s.server.i18n.guild.invite.title,
		fmt.Sprintf(s.server.i18n.guild.invite.body, guildInfo.Name))

	if err != nil {
		s.logger.Error("Failed to create guild scout application with mail", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, nil)
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

	if err != nil {
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
	bf := byteframe.NewByteFrame()
	guild, err := s.server.guildRepo.GetByCharID(pkt.LeaderID)

	if err != nil {
		s.logger.Error("Failed to get guild info for answer scout", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, nil)
		return
	}

	app, err := s.server.guildRepo.GetApplication(guild.ID, s.charID, GuildApplicationTypeInvited)

	if app == nil || err != nil {
		s.logger.Warn(
			"Guild invite missing, deleted?",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", s.charID),
		)
		bf.WriteUint32(7)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}

	type mailMsg struct {
		senderID    uint32
		recipientID uint32
		subject     string
		body        string
	}
	var msgs []mailMsg
	if pkt.Answer {
		err = s.server.guildRepo.AcceptApplication(guild.ID, s.charID)
		msgs = append(msgs,
			mailMsg{0, s.charID, s.server.i18n.guild.invite.success.title, fmt.Sprintf(s.server.i18n.guild.invite.success.body, guild.Name)},
			mailMsg{s.charID, pkt.LeaderID, s.server.i18n.guild.invite.accepted.title, fmt.Sprintf(s.server.i18n.guild.invite.accepted.body, guild.Name)},
		)
	} else {
		err = s.server.guildRepo.RejectApplication(guild.ID, s.charID)
		msgs = append(msgs,
			mailMsg{0, s.charID, s.server.i18n.guild.invite.rejected.title, fmt.Sprintf(s.server.i18n.guild.invite.rejected.body, guild.Name)},
			mailMsg{s.charID, pkt.LeaderID, s.server.i18n.guild.invite.declined.title, fmt.Sprintf(s.server.i18n.guild.invite.declined.body, guild.Name)},
		)
	}
	if err != nil {
		bf.WriteUint32(7)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		bf.WriteUint32(0)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		for _, m := range msgs {
			_ = s.server.mailRepo.SendMail(m.senderID, m.recipientID, m.subject, m.body, 0, 0, false, true)
		}
	}
}

func handleMsgMhfGetGuildScoutList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildScoutList)

	guildInfo, _ := s.server.guildRepo.GetByCharID(s.charID)

	if guildInfo == nil && s.prevGuildID == 0 {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	} else {
		guildInfo, err := s.server.guildRepo.GetByID(s.prevGuildID)
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
