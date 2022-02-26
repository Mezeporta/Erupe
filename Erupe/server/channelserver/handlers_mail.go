package channelserver

import (
	"database/sql"
	"time"

	"github.com/Solenataris/Erupe/common/stringsupport"
	"github.com/Solenataris/Erupe/network/binpacket"
	"github.com/Solenataris/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
)

type Mail struct {
	ID                   int       `db:"id"`
	SenderID             uint32    `db:"sender_id"`
	RecipientID          uint32    `db:"recipient_id"`
	Subject              string    `db:"subject"`
	Body                 string    `db:"body"`
	Read                 bool      `db:"read"`
	Deleted              bool      `db:"deleted"`
	AttachedItemReceived bool      `db:"attached_item_received"`
	AttachedItemID       *uint16   `db:"attached_item"`
	AttachedItemAmount   int16     `db:"attached_item_amount"`
	CreatedAt            time.Time `db:"created_at"`
	IsGuildInvite        bool      `db:"is_guild_invite"`
	SenderName           string    `db:"sender_name"`
}

func (m *Mail) Send(s *Session, transaction *sql.Tx) error {
	query := `
		INSERT INTO mail (sender_id, recipient_id, subject, body, attached_item, attached_item_amount, is_guild_invite)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var err error

	if transaction == nil {
		_, err = s.server.db.Exec(query, m.SenderID, m.RecipientID, m.Subject, m.Body, m.AttachedItemID, m.AttachedItemAmount, m.IsGuildInvite)
	} else {
		_, err = transaction.Exec(query, m.SenderID, m.RecipientID, m.Subject, m.Body, m.AttachedItemID, m.AttachedItemAmount, m.IsGuildInvite)
	}

	if err != nil {
		s.logger.Error(
			"failed to send mail",
			zap.Error(err),
			zap.Uint32("senderID", m.SenderID),
			zap.Uint32("recipientID", m.RecipientID),
			zap.String("subject", m.Subject),
			zap.String("body", m.Body),
			zap.Uint16p("itemID", m.AttachedItemID),
			zap.Int16("itemAmount", m.AttachedItemAmount),
			zap.Bool("isGuildInvite", m.IsGuildInvite),
		)
		return err
	}

	return nil
}

func (m *Mail) MarkRead(s *Session) error {
	_, err := s.server.db.Exec(`
		UPDATE mail SET read = true WHERE id = $1 
	`, m.ID)

	if err != nil {
		s.logger.Error(
			"failed to mark mail as read",
			zap.Error(err),
			zap.Int("mailID", m.ID),
		)
		return err
	}

	return nil
}

func (m *Mail) MarkDeleted(s *Session) error {
	_, err := s.server.db.Exec(`
		UPDATE mail SET deleted = true WHERE id = $1 
	`, m.ID)

	if err != nil {
		s.logger.Error(
			"failed to mark mail as deleted",
			zap.Error(err),
			zap.Int("mailID", m.ID),
		)
		return err
	}

	return nil
}

func GetMailListForCharacter(s *Session, charID uint32) ([]Mail, error) {
	rows, err := s.server.db.Queryx(`
		SELECT 
			m.id,
			m.sender_id,
			m.recipient_id,
			m.subject,
			m.read,
			m.attached_item,
			m.attached_item_amount,
			m.created_at,
			m.is_guild_invite,
			m.deleted,
			c.name as sender_name
		FROM mail m 
			JOIN characters c ON c.id = m.sender_id 
		WHERE recipient_id = $1 AND deleted = false
		ORDER BY m.created_at DESC, id DESC
		LIMIT 32
	`, charID)

	if err != nil {
		s.logger.Error("failed to get mail for character", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	defer rows.Close()

	allMail := make([]Mail, 0)

	for rows.Next() {
		mail := Mail{}

		err := rows.StructScan(&mail)

		if err != nil {
			return nil, err
		}

		allMail = append(allMail, mail)
	}

	return allMail, nil
}

func GetMailByID(s *Session, ID int) (*Mail, error) {
	row := s.server.db.QueryRowx(`
		SELECT 
			m.id,
			m.sender_id,
			m.recipient_id,
			m.subject,
			m.read,
			m.body,
			m.attached_item,
			m.attached_item_amount,
			m.created_at,
			m.is_guild_invite,
			m.deleted,
			c.name as sender_name
		FROM mail m 
			JOIN characters c ON c.id = m.sender_id 
		WHERE m.id = $1
		LIMIT 1
	`, ID)

	mail := &Mail{}

	err := row.StructScan(mail)

	if err != nil {
		s.logger.Error(
			"failed to retrieve mail",
			zap.Error(err),
			zap.Int("mailID", ID),
		)
		return nil, err
	}

	return mail, nil
}

func SendMailNotification(s *Session, m *Mail, recipient *Session) {
	senderName, err := getCharacterName(s, m.SenderID)

	if err != nil {
		panic(err)
	}

	bf := byteframe.NewByteFrame()

	notification := &binpacket.MsgBinMailNotify{
		SenderName: senderName,
	}

	notification.Build(bf)

	castedBinary := &mhfpacket.MsgSysCastedBinary{
		CharID:         m.SenderID,
		BroadcastType:  0x00,
		MessageType:    BinaryMessageTypeMailNotify,
		RawDataPayload: bf.Data(),
	}

	castedBinary.Build(bf, s.clientContext)

	recipient.QueueSendMHF(castedBinary)
}

func getCharacterName(s *Session, charID uint32) (string, error) {
	row := s.server.db.QueryRow("SELECT name FROM characters WHERE id = $1", charID)

	charName := ""

	err := row.Scan(&charName)

	if err != nil {
		return "", err
	}

	return charName, nil
}

func handleMsgMhfReadMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMail)

	mailId := s.mailList[pkt.AccIndex]

	if mailId == 0 {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		panic("attempting to read mail that doesn't exist in session map")
	}

	mail, err := GetMailByID(s, mailId)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	_ = mail.MarkRead(s)

	bodyBytes, _ := stringsupport.ConvertUTF8ToShiftJIS(mail.Body)

	doAckBufSucceed(s, pkt.AckHandle, bodyBytes)
}

func handleMsgMhfListMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMail)

	mail, err := GetMailListForCharacter(s, s.charID)

	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		panic(err)
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

		itemAttached := m.AttachedItemID != nil
		subjectBytes, _ := stringsupport.ConvertUTF8ToShiftJIS(m.Subject)
		senderNameBytes, _ := stringsupport.ConvertUTF8ToShiftJIS(m.SenderName)

		msg.WriteUint32(m.SenderID)
		msg.WriteUint32(uint32(m.CreatedAt.Unix()))

		msg.WriteUint8(uint8(accIndex))
		msg.WriteUint8(uint8(i))

		flags := uint8(0x00)

		if m.Read {
			flags |= 0x01
		}

		if m.AttachedItemReceived {
			flags |= 0x08
		}

		if m.IsGuildInvite {
			// Guild Invite
			flags |= 0x10

			// System message?
			flags |= 0x04
		}

		msg.WriteUint8(flags)
		msg.WriteBool(itemAttached)
		msg.WriteUint8(uint8(len(subjectBytes)))
		msg.WriteUint8(uint8(len(senderNameBytes)))
		msg.WriteBytes(subjectBytes)
		msg.WriteBytes(senderNameBytes)

		if itemAttached {
			msg.WriteInt16(m.AttachedItemAmount)
			msg.WriteUint16(*m.AttachedItemID)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, msg.Data())
}

func handleMsgMhfOprtMail(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprtMail)

	mail, err := GetMailByID(s, s.mailList[pkt.AccIndex])

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, nil)
		panic(err)
	}

	switch mhfpacket.OperateMailOperation(pkt.Operation) {
	case mhfpacket.OperateMailOperationDelete:
		err = mail.MarkDeleted(s)

		if err != nil {
			doAckSimpleFail(s, pkt.AckHandle, nil)
			panic(err)
		}
	}

	doAckSimpleSucceed(s, pkt.AckHandle, nil)
}

func handleMsgMhfSendMail(s *Session, p mhfpacket.MHFPacket) {}
