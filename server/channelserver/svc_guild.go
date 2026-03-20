package channelserver

import (
	"errors"
	"fmt"
	"sort"

	"go.uber.org/zap"
)

// GuildMemberAction is a domain enum for guild member operations.
type GuildMemberAction uint8

const (
	GuildMemberActionAccept GuildMemberAction = iota + 1
	GuildMemberActionReject
	GuildMemberActionKick
)

// ErrUnauthorized is returned when the actor lacks permission for the operation.
var ErrUnauthorized = errors.New("unauthorized")

// ErrUnknownAction is returned for unrecognized guild member actions.
var ErrUnknownAction = errors.New("unknown guild member action")

// ErrNoEligibleLeader is returned when no member can accept leadership.
var ErrNoEligibleLeader = errors.New("no eligible leader")

// ErrAlreadyInvited is returned when a scout target already has a pending application.
var ErrAlreadyInvited = errors.New("already invited")

// ErrCannotRecruit is returned when the actor lacks recruit permission.
var ErrCannotRecruit = errors.New("cannot recruit")

// ErrApplicationMissing is returned when the expected guild application is not found.
var ErrApplicationMissing = errors.New("application missing")

// OperateMemberResult holds the outcome of a guild member operation.
type OperateMemberResult struct {
	MailRecipientID uint32
	Mail            Mail
}

// DisbandResult holds the outcome of a guild disband operation.
type DisbandResult struct {
	Success bool
}

// ResignResult holds the outcome of a leadership resignation.
type ResignResult struct {
	NewLeaderCharID uint32
}

// LeaveResult holds the outcome of a guild leave operation.
type LeaveResult struct {
	Success bool
}

// ScoutInviteStrings holds i18n strings needed for scout invitation mails.
type ScoutInviteStrings struct {
	Title string
	Body  string // must contain %s for guild name
}

// AnswerScoutStrings holds i18n strings needed for scout answer mails.
type AnswerScoutStrings struct {
	SuccessTitle  string
	SuccessBody   string // %s for guild name
	AcceptedTitle string
	AcceptedBody  string // %s for guild name
	RejectedTitle string
	RejectedBody  string // %s for guild name
	DeclinedTitle string
	DeclinedBody  string // %s for guild name
}

// AnswerScoutResult holds the outcome of answering a guild scout invitation.
type AnswerScoutResult struct {
	GuildID uint32
	Success bool
	Mails   []Mail
}

// GuildService encapsulates guild business logic, sitting between handlers and repos.
type GuildService struct {
	guildRepo GuildRepo
	mailSvc   *MailService
	charRepo  CharacterRepo
	logger    *zap.Logger
}

// NewGuildService creates a new GuildService.
func NewGuildService(gr GuildRepo, ms *MailService, cr CharacterRepo, log *zap.Logger) *GuildService {
	return &GuildService{
		guildRepo: gr,
		mailSvc:   ms,
		charRepo:  cr,
		logger:    log,
	}
}

// OperateMember performs a guild member management action (accept/reject/kick).
// The actor must be the guild leader or a sub-leader. On success, a notification
// mail is sent (best-effort) and the result is returned for protocol-level notification.
func (svc *GuildService) OperateMember(actorCharID, targetCharID uint32, action GuildMemberAction) (*OperateMemberResult, error) {
	guild, err := svc.guildRepo.GetByCharID(targetCharID)
	if err != nil || guild == nil {
		return nil, fmt.Errorf("guild lookup for char %d: %w", targetCharID, err)
	}

	actorMember, err := svc.guildRepo.GetCharacterMembership(actorCharID)
	if err != nil || (!actorMember.IsSubLeader() && guild.LeaderCharID != actorCharID) {
		return nil, ErrUnauthorized
	}

	var mail Mail
	switch action {
	case GuildMemberActionAccept:
		err = svc.guildRepo.AcceptApplication(guild.ID, targetCharID)
		mail = Mail{
			RecipientID:     targetCharID,
			Subject:         "Accepted!",
			Body:            fmt.Sprintf("Your application to join 「%s」 was accepted.", guild.Name),
			IsSystemMessage: true,
		}
	case GuildMemberActionReject:
		err = svc.guildRepo.RejectApplication(guild.ID, targetCharID)
		mail = Mail{
			RecipientID:     targetCharID,
			Subject:         "Rejected",
			Body:            fmt.Sprintf("Your application to join 「%s」 was rejected.", guild.Name),
			IsSystemMessage: true,
		}
	case GuildMemberActionKick:
		err = svc.guildRepo.RemoveCharacter(targetCharID)
		mail = Mail{
			RecipientID:     targetCharID,
			Subject:         "Kicked",
			Body:            fmt.Sprintf("You were kicked from 「%s」.", guild.Name),
			IsSystemMessage: true,
		}
	default:
		return nil, ErrUnknownAction
	}

	if err != nil {
		return nil, fmt.Errorf("guild member action %d: %w", action, err)
	}

	// Send mail best-effort
	if mailErr := svc.mailSvc.SendSystem(mail.RecipientID, mail.Subject, mail.Body); mailErr != nil {
		svc.logger.Warn("Failed to send guild member operation mail", zap.Error(mailErr))
	}

	return &OperateMemberResult{
		MailRecipientID: targetCharID,
		Mail:            mail,
	}, nil
}

// Disband disbands a guild. Only the guild leader may disband.
func (svc *GuildService) Disband(actorCharID, guildID uint32) (*DisbandResult, error) {
	guild, err := svc.guildRepo.GetByID(guildID)
	if err != nil {
		return nil, fmt.Errorf("guild lookup: %w", err)
	}

	if guild.LeaderCharID != actorCharID {
		svc.logger.Warn("Unauthorized guild disband attempt",
			zap.Uint32("charID", actorCharID), zap.Uint32("guildID", guildID))
		return &DisbandResult{Success: false}, nil
	}

	if err := svc.guildRepo.Disband(guildID); err != nil {
		return &DisbandResult{Success: false}, nil
	}

	return &DisbandResult{Success: true}, nil
}

// ResignLeadership transfers guild leadership to the next eligible member.
// Members are sorted by order index; those with AvoidLeadership set are skipped.
func (svc *GuildService) ResignLeadership(actorCharID, guildID uint32) (*ResignResult, error) {
	guild, err := svc.guildRepo.GetByID(guildID)
	if err != nil {
		return nil, fmt.Errorf("guild lookup: %w", err)
	}

	members, err := svc.guildRepo.GetMembers(guildID, false)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].OrderIndex < members[j].OrderIndex
	})

	// Find current leader in sorted list (should be index 0)
	var leaderIdx int
	for i, m := range members {
		if m.CharID == actorCharID {
			leaderIdx = i
			break
		}
	}

	// Find first eligible successor (skip leader and anyone avoiding leadership)
	var newLeaderIdx int
	found := false
	for i := 1; i < len(members); i++ {
		if i == leaderIdx {
			continue
		}
		if !members[i].AvoidLeadership {
			newLeaderIdx = i
			found = true
			break
		}
	}

	if !found {
		return &ResignResult{NewLeaderCharID: 0}, nil
	}

	// Swap order indices
	guild.LeaderCharID = members[newLeaderIdx].CharID
	members[leaderIdx].OrderIndex, members[newLeaderIdx].OrderIndex =
		members[newLeaderIdx].OrderIndex, 1

	if err := svc.guildRepo.SaveMember(members[leaderIdx]); err != nil {
		svc.logger.Error("Failed to save former leader member data", zap.Error(err))
	}
	if err := svc.guildRepo.SaveMember(members[newLeaderIdx]); err != nil {
		svc.logger.Error("Failed to save new leader member data", zap.Error(err))
	}
	if err := svc.guildRepo.Save(guild); err != nil {
		svc.logger.Error("Failed to save guild after leadership resign", zap.Error(err))
	}

	return &ResignResult{NewLeaderCharID: members[newLeaderIdx].CharID}, nil
}

// Leave removes a character from their guild. If the character is an applicant,
// their application is rejected; otherwise they are removed as a member.
// A withdrawal notification mail is sent on success.
func (svc *GuildService) Leave(charID, guildID uint32, isApplicant bool, guildName string) (*LeaveResult, error) {
	if isApplicant {
		if err := svc.guildRepo.RejectApplication(guildID, charID); err != nil {
			return &LeaveResult{Success: false}, nil
		}
	} else {
		if err := svc.guildRepo.RemoveCharacter(charID); err != nil {
			return &LeaveResult{Success: false}, nil
		}
	}

	// Best-effort withdrawal notification
	if err := svc.mailSvc.SendSystem(charID, "Withdrawal",
		fmt.Sprintf("You have withdrawn from 「%s」.", guildName)); err != nil {
		svc.logger.Warn("Failed to send guild withdrawal notification", zap.Error(err))
	}

	return &LeaveResult{Success: true}, nil
}

// PostScout sends a guild scout invitation to a target character.
// The actor must have recruit permission. Returns ErrAlreadyInvited if the target
// already has a pending application.
func (svc *GuildService) PostScout(actorCharID, targetCharID uint32, strings ScoutInviteStrings) error {
	actorMember, err := svc.guildRepo.GetCharacterMembership(actorCharID)
	if err != nil {
		return fmt.Errorf("actor membership lookup: %w", err)
	}
	if actorMember == nil || !actorMember.CanRecruit() {
		return ErrCannotRecruit
	}

	guild, err := svc.guildRepo.GetByID(actorMember.GuildID)
	if err != nil {
		return fmt.Errorf("guild lookup: %w", err)
	}

	hasApp, err := svc.guildRepo.HasApplication(guild.ID, targetCharID)
	if err != nil {
		return fmt.Errorf("check application: %w", err)
	}
	if hasApp {
		return ErrAlreadyInvited
	}

	err = svc.guildRepo.CreateApplicationWithMail(
		guild.ID, targetCharID, actorCharID, GuildApplicationTypeInvited,
		actorCharID, targetCharID,
		strings.Title,
		fmt.Sprintf(strings.Body, guild.Name))
	if err != nil {
		return fmt.Errorf("create scout application: %w", err)
	}

	return nil
}

// AnswerScout processes a character's response to a guild scout invitation.
// If accept is true, the character joins the guild; otherwise the invitation is rejected.
// Notification mails are sent to both the character and the leader.
func (svc *GuildService) AnswerScout(charID, leaderID uint32, accept bool, strings AnswerScoutStrings) (*AnswerScoutResult, error) {
	guild, err := svc.guildRepo.GetByCharID(leaderID)
	if err != nil {
		return nil, fmt.Errorf("guild lookup for leader %d: %w", leaderID, err)
	}

	app, err := svc.guildRepo.GetApplication(guild.ID, charID, GuildApplicationTypeInvited)
	if app == nil || err != nil {
		return &AnswerScoutResult{
			GuildID: guild.ID,
			Success: false,
		}, ErrApplicationMissing
	}

	var mails []Mail
	if accept {
		err = svc.guildRepo.AcceptApplication(guild.ID, charID)
		mails = []Mail{
			{SenderID: 0, RecipientID: charID, Subject: strings.SuccessTitle, Body: fmt.Sprintf(strings.SuccessBody, guild.Name), IsSystemMessage: true},
			{SenderID: charID, RecipientID: leaderID, Subject: strings.AcceptedTitle, Body: fmt.Sprintf(strings.AcceptedBody, guild.Name), IsSystemMessage: true},
		}
	} else {
		err = svc.guildRepo.RejectApplication(guild.ID, charID)
		mails = []Mail{
			{SenderID: 0, RecipientID: charID, Subject: strings.RejectedTitle, Body: fmt.Sprintf(strings.RejectedBody, guild.Name), IsSystemMessage: true},
			{SenderID: charID, RecipientID: leaderID, Subject: strings.DeclinedTitle, Body: fmt.Sprintf(strings.DeclinedBody, guild.Name), IsSystemMessage: true},
		}
	}

	if err != nil {
		return &AnswerScoutResult{
			GuildID: guild.ID,
			Success: false,
		}, nil
	}

	// Send mails best-effort
	for _, m := range mails {
		if mailErr := svc.mailSvc.SendSystem(m.RecipientID, m.Subject, m.Body); mailErr != nil {
			svc.logger.Warn("Failed to send guild scout response mail", zap.Error(mailErr))
		}
	}

	return &AnswerScoutResult{
		GuildID: guild.ID,
		Success: true,
		Mails:   mails,
	}, nil
}
