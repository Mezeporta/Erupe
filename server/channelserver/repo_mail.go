package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// MailRepository centralizes all database access for the mail table.
type MailRepository struct {
	db *sqlx.DB
}

// NewMailRepository creates a new MailRepository.
func NewMailRepository(db *sqlx.DB) *MailRepository {
	return &MailRepository{db: db}
}

const mailInsertQuery = `
	INSERT INTO mail (sender_id, recipient_id, subject, body, attached_item, attached_item_amount, is_guild_invite, is_sys_message)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

// SendMail inserts a new mail row.
func (r *MailRepository) SendMail(senderID, recipientID uint32, subject, body string, itemID, itemAmount uint16, isGuildInvite, isSystemMessage bool) error {
	_, err := r.db.Exec(mailInsertQuery, senderID, recipientID, subject, body, itemID, itemAmount, isGuildInvite, isSystemMessage)
	return err
}

// GetListForCharacter loads all non-deleted mail for a character (max 32).
func (r *MailRepository) GetListForCharacter(charID uint32) ([]Mail, error) {
	rows, err := r.db.Queryx(`
		SELECT
			m.id,
			m.sender_id,
			m.recipient_id,
			m.subject,
			m.read,
			m.attached_item_received,
			m.attached_item,
			m.attached_item_amount,
			m.created_at,
			m.is_guild_invite,
			m.is_sys_message,
			m.deleted,
			m.locked,
			c.name as sender_name
		FROM mail m
			JOIN characters c ON c.id = m.sender_id
		WHERE recipient_id = $1 AND m.deleted = false
		ORDER BY m.created_at DESC, id DESC
		LIMIT 32
	`, charID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var allMail []Mail
	for rows.Next() {
		var mail Mail
		if err := rows.StructScan(&mail); err != nil {
			return nil, err
		}
		allMail = append(allMail, mail)
	}
	return allMail, nil
}

// GetByID loads a single mail by ID.
func (r *MailRepository) GetByID(id int) (*Mail, error) {
	row := r.db.QueryRowx(`
		SELECT
			m.id,
			m.sender_id,
			m.recipient_id,
			m.subject,
			m.read,
			m.body,
			m.attached_item_received,
			m.attached_item,
			m.attached_item_amount,
			m.created_at,
			m.is_guild_invite,
			m.is_sys_message,
			m.deleted,
			m.locked,
			c.name as sender_name
		FROM mail m
			JOIN characters c ON c.id = m.sender_id
		WHERE m.id = $1
		LIMIT 1
	`, id)

	mail := &Mail{}
	if err := row.StructScan(mail); err != nil {
		return nil, err
	}
	return mail, nil
}

// MarkRead marks a mail as read.
func (r *MailRepository) MarkRead(id int) error {
	_, err := r.db.Exec(`UPDATE mail SET read = true WHERE id = $1`, id)
	return err
}

// MarkDeleted marks a mail as deleted.
func (r *MailRepository) MarkDeleted(id int) error {
	_, err := r.db.Exec(`UPDATE mail SET deleted = true WHERE id = $1`, id)
	return err
}

// SetLocked sets the locked state of a mail.
func (r *MailRepository) SetLocked(id int, locked bool) error {
	_, err := r.db.Exec(`UPDATE mail SET locked = $1 WHERE id = $2`, locked, id)
	return err
}

// MarkItemReceived marks a mail's attached item as received.
func (r *MailRepository) MarkItemReceived(id int) error {
	_, err := r.db.Exec(`UPDATE mail SET attached_item_received = TRUE WHERE id = $1`, id)
	return err
}
