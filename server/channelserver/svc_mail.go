package channelserver

import (
	"fmt"

	"go.uber.org/zap"
)

// MailService encapsulates mail-sending business logic, sitting between
// handlers/services and the MailRepo. It provides convenient methods for
// common mail patterns (system notifications, guild broadcasts, player mail)
// so callers don't need to specify boolean flags directly.
type MailService struct {
	mailRepo  MailRepo
	guildRepo GuildRepo
	logger    *zap.Logger
}

// NewMailService creates a new MailService.
func NewMailService(mr MailRepo, gr GuildRepo, log *zap.Logger) *MailService {
	return &MailService{
		mailRepo:  mr,
		guildRepo: gr,
		logger:    log,
	}
}

// Send sends a player-to-player mail with an optional item attachment.
func (svc *MailService) Send(senderID, recipientID uint32, subject, body string, itemID, quantity uint16) error {
	return svc.mailRepo.SendMail(senderID, recipientID, subject, body, itemID, quantity, false, false)
}

// SendSystem sends a system notification mail (no item, flagged as system message).
func (svc *MailService) SendSystem(recipientID uint32, subject, body string) error {
	return svc.mailRepo.SendMail(0, recipientID, subject, body, 0, 0, false, true)
}

// SendGuildInvite sends a guild invitation mail (flagged as guild invite).
func (svc *MailService) SendGuildInvite(senderID, recipientID uint32, subject, body string) error {
	return svc.mailRepo.SendMail(senderID, recipientID, subject, body, 0, 0, true, false)
}

// BroadcastToGuild sends a mail from senderID to all members of the specified guild.
func (svc *MailService) BroadcastToGuild(senderID, guildID uint32, subject, body string) error {
	members, err := svc.guildRepo.GetMembers(guildID, false)
	if err != nil {
		return fmt.Errorf("get guild members for broadcast: %w", err)
	}
	for _, m := range members {
		if err := svc.mailRepo.SendMail(senderID, m.CharID, subject, body, 0, 0, false, false); err != nil {
			return fmt.Errorf("send guild broadcast to char %d: %w", m.CharID, err)
		}
	}
	return nil
}
