package channelserver

import (
	"erupe-ce/common/stringsupport"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// Mail represents an in-game mail message.
type Mail struct {
	ID                   int       `db:"id"`
	SenderID             uint32    `db:"sender_id"`
	RecipientID          uint32    `db:"recipient_id"`
	Subject              string    `db:"subject"`
	Body                 string    `db:"body"`
	Read                 bool      `db:"read"`
	Deleted              bool      `db:"deleted"`
	Locked               bool      `db:"locked"`
	AttachedItemReceived bool      `db:"attached_item_received"`
	AttachedItemID       uint16    `db:"attached_item"`
	AttachedItemAmount   uint16    `db:"attached_item_amount"`
	CreatedAt            time.Time `db:"created_at"`
	IsGuildInvite        bool      `db:"is_guild_invite"`
	IsSystemMessage      bool      `db:"is_sys_message"`
	SenderName           string    `db:"sender_name"`
}

// SendMailNotification sends a new mail notification to a player.
func SendMailNotification(s *Session, m *Mail, recipient *Session) {
	bf := byteframe.NewByteFrame()

	notification := &binpacket.MsgBinMailNotify{
		SenderName: getCharacterName(s, m.SenderID),
	}

	_ = notification.Build(bf)

	castedBinary := &mhfpacket.MsgSysCastedBinary{
		CharID:         m.SenderID,
		BroadcastType:  0x00,
		MessageType:    BinaryMessageTypeMailNotify,
		RawDataPayload: bf.Data(),
	}

	_ = castedBinary.Build(bf, s.clientContext)

	recipient.QueueSendMHFNonBlocking(castedBinary)
}

func getCharacterName(s *Session, charID uint32) string {
	name, err := s.server.charRepo.GetName(charID)
	if err != nil {
		return ""
	}
	return name
}

func handleMsgMhfReadMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMail)

	if int(pkt.AccIndex) >= len(s.mailList) {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0})
		return
	}
	mailId := s.mailList[pkt.AccIndex]
	if mailId == 0 {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0})
		return
	}

	mail, err := s.server.mailRepo.GetByID(mailId)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0})
		return
	}

	if err := s.server.mailRepo.MarkRead(mail.ID); err != nil {
		s.logger.Error("Failed to mark mail as read", zap.Error(err))
	}
	bf := byteframe.NewByteFrame()
	body := stringsupport.UTF8ToSJIS(mail.Body)
	bf.WriteNullTerminatedBytes(body)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfListMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMail)

	mail, err := s.server.mailRepo.GetListForCharacter(s.charID)
	if err != nil {
		s.logger.Error("failed to get mail for character", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckBufSucceed(s, pkt.AckHandle, []byte{0})
		return
	}

	if s.mailList == nil {
		s.mailList = make([]int, 256)
	}

	msg := byteframe.NewByteFrame()

	msg.WriteUint32(uint32(len(mail)))

	startIndex := s.mailAccIndex

	for i, m := range mail {
		accIndex := startIndex + uint8(i)
		s.mailList[accIndex] = m.ID
		s.mailAccIndex++

		itemAttached := m.AttachedItemID != 0

		msg.WriteUint32(m.SenderID)
		msg.WriteUint32(uint32(m.CreatedAt.Unix()))

		msg.WriteUint8(accIndex)
		msg.WriteUint8(uint8(i))

		flags := uint8(0x00)

		if m.Read {
			flags |= 0x01
		}

		if m.Locked {
			flags |= 0x02
		}

		if m.IsSystemMessage {
			flags |= 0x04
		}

		if m.AttachedItemReceived {
			flags |= 0x08
		}

		if m.IsGuildInvite {
			flags |= 0x10
		}

		msg.WriteUint8(flags)
		msg.WriteBool(itemAttached)
		msg.WriteUint8(16)
		msg.WriteUint8(21)
		msg.WriteBytes(stringsupport.PaddedString(m.Subject, 16, true))
		msg.WriteBytes(stringsupport.PaddedString(m.SenderName, 21, true))
		if itemAttached {
			msg.WriteUint16(m.AttachedItemAmount)
			msg.WriteUint16(m.AttachedItemID)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, msg.Data())
}

func handleMsgMhfOprtMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprtMail)

	if int(pkt.AccIndex) >= len(s.mailList) {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	mail, err := s.server.mailRepo.GetByID(s.mailList[pkt.AccIndex])
	if err != nil {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	switch pkt.Operation {
	case mhfpacket.OperateMailDelete:
		if err := s.server.mailRepo.MarkDeleted(mail.ID); err != nil {
			s.logger.Error("Failed to delete mail", zap.Error(err))
		}
	case mhfpacket.OperateMailLock:
		if err := s.server.mailRepo.SetLocked(mail.ID, true); err != nil {
			s.logger.Error("Failed to lock mail", zap.Error(err))
		}
	case mhfpacket.OperateMailUnlock:
		if err := s.server.mailRepo.SetLocked(mail.ID, false); err != nil {
			s.logger.Error("Failed to unlock mail", zap.Error(err))
		}
	case mhfpacket.OperateMailAcquireItem:
		if err := s.server.mailRepo.MarkItemReceived(mail.ID); err != nil {
			s.logger.Error("Failed to mark mail item received", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfSendMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSendMail)

	if pkt.RecipientID == 0 { // Guild mail
		g, err := s.server.guildRepo.GetByCharID(s.charID)
		if err != nil {
			s.logger.Error("Failed to get guild info for mail")
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		gm, err := s.server.guildRepo.GetMembers(g.ID, false)
		if err != nil {
			s.logger.Error("Failed to get guild members for mail")
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		for i := 0; i < len(gm); i++ {
			err := s.server.mailRepo.SendMail(s.charID, gm[i].CharID, pkt.Subject, pkt.Body, 0, 0, false, false)
			if err != nil {
				s.logger.Error("Failed to send mail")
				doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
				return
			}
		}
	} else {
		err := s.server.mailRepo.SendMail(s.charID, pkt.RecipientID, pkt.Subject, pkt.Body, pkt.ItemID, pkt.Quantity, false, false)
		if err != nil {
			s.logger.Error("Failed to send mail")
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
