package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
	"io"
)

func handleMsgMhfPostGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGuildScout)

	actorCharGuildData, err := GetCharacterGuildData(s, s.charID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	if actorCharGuildData == nil || !actorCharGuildData.CanRecruit() {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	guildInfo, err := GetGuildInfoByID(s, actorCharGuildData.GuildID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	hasApplication, err := guildInfo.HasApplicationForCharID(s, pkt.CharID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	if hasApplication {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x04})
		return
	}

	transaction, err := s.server.db.Begin()

	if err != nil {
		panic(err)
	}

	err = guildInfo.CreateApplication(s, pkt.CharID, GuildApplicationTypeInvited, transaction)

	if err != nil {
		rollbackTransaction(s, transaction)
		doAckBufFail(s, pkt.AckHandle, nil)
		panic(err)
	}

	mail := &Mail{
		SenderID:    s.charID,
		RecipientID: pkt.CharID,
		Subject:     s.server.i18n.guild.invite.title,
		Body: fmt.Sprintf(
			s.server.i18n.guild.invite.body,
			guildInfo.Name,
		),
		IsGuildInvite: true,
	}

	err = mail.Send(s, transaction)

	if err != nil {
		rollbackTransaction(s, transaction)
		doAckBufFail(s, pkt.AckHandle, nil)
		return
	}

	err = transaction.Commit()

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, nil)
		panic(err)
	}

	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfCancelGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildScout)

	guildCharData, err := GetCharacterGuildData(s, s.charID)

	if err != nil {
		panic(err)
	}

	if guildCharData == nil || !guildCharData.CanRecruit() {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	guild, err := GetGuildInfoByID(s, guildCharData.GuildID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	err = guild.CancelInvitation(s, pkt.InvitationID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAnswerGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAnswerGuildScout)
	bf := byteframe.NewByteFrame()
	guild, err := GetGuildInfoByCharacterId(s, pkt.LeaderID)

	if err != nil {
		panic(err)
	}

	app, err := guild.GetApplicationForCharID(s, s.charID, GuildApplicationTypeInvited)

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

	var mail []Mail
	if pkt.Answer {
		err = guild.AcceptApplication(s, s.charID)
		mail = append(mail, Mail{
			RecipientID:     s.charID,
			Subject:         s.server.i18n.guild.invite.success.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.success.body, guild.Name),
			IsSystemMessage: true,
		})
		mail = append(mail, Mail{
			SenderID:        s.charID,
			RecipientID:     pkt.LeaderID,
			Subject:         s.server.i18n.guild.invite.accepted.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.accepted.body, guild.Name),
			IsSystemMessage: true,
		})
	} else {
		err = guild.RejectApplication(s, s.charID)
		mail = append(mail, Mail{
			RecipientID:     s.charID,
			Subject:         s.server.i18n.guild.invite.rejected.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.rejected.body, guild.Name),
			IsSystemMessage: true,
		})
		mail = append(mail, Mail{
			SenderID:        s.charID,
			RecipientID:     pkt.LeaderID,
			Subject:         s.server.i18n.guild.invite.declined.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.declined.body, guild.Name),
			IsSystemMessage: true,
		})
	}
	if err != nil {
		bf.WriteUint32(7)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		bf.WriteUint32(0)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		for _, m := range mail {
			_ = m.Send(s, nil)
		}
	}
}

func handleMsgMhfGetGuildScoutList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildScoutList)

	guildInfo, _ := GetGuildInfoByCharacterId(s, s.charID)

	if guildInfo == nil && s.prevGuildID == 0 {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	} else {
		guildInfo, err := GetGuildInfoByID(s, s.prevGuildID)
		if guildInfo == nil || err != nil {
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	rows, err := s.server.db.Queryx(`
		SELECT c.id, c.name, c.hr, c.gr, ga.actor_id
			FROM guild_applications ga 
			JOIN characters c ON c.id = ga.character_id
		WHERE ga.guild_id = $1 AND ga.application_type = 'invited'
	`, guildInfo.ID)

	if err != nil {
		s.logger.Error("failed to retrieve scouted characters", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	defer func() { _ = rows.Close() }()

	bf := byteframe.NewByteFrame()

	bf.SetBE()

	// Result count, we will overwrite this later
	bf.WriteUint32(0x00)

	count := uint32(0)

	for rows.Next() {
		var charName string
		var charID, actorID uint32
		var HR, GR uint16

		err = rows.Scan(&charID, &charName, &HR, &GR, &actorID)

		if err != nil {
			doAckSimpleFail(s, pkt.AckHandle, nil)
			continue
		}

		// This seems to be used as a unique ID for the invitation sent
		// we can just use the charID and then filter on guild_id+charID when performing operations
		// this might be a problem later with mails sent referencing IDs but we'll see.
		bf.WriteUint32(charID)
		bf.WriteUint32(actorID)
		bf.WriteUint32(charID)
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
		bf.WriteUint16(HR) // HR?
		bf.WriteUint16(GR) // GR?
		bf.WriteBytes(stringsupport.PaddedString(charName, 32, true))
		count++
	}

	_, err = bf.Seek(0, io.SeekStart)

	if err != nil {
		panic(err)
	}

	bf.WriteUint32(count)

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRejectGuildScout)

	row := s.server.db.QueryRow("SELECT restrict_guild_scout FROM characters WHERE id=$1", s.charID)

	var currentStatus bool

	err := row.Scan(&currentStatus)

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

	_, err := s.server.db.Exec("UPDATE characters SET restrict_guild_scout=$1 WHERE id=$2", pkt.Reject, s.charID)

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
