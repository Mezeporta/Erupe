package channelserver

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func newTestMailService(mr MailRepo, gr GuildRepo) *MailService {
	logger, _ := zap.NewDevelopment()
	return NewMailService(mr, gr, logger)
}

func newTestGuildService(gr GuildRepo, mr MailRepo) *GuildService {
	logger, _ := zap.NewDevelopment()
	ms := newTestMailService(mr, gr)
	return NewGuildService(gr, ms, nil, logger)
}

func TestGuildService_OperateMember(t *testing.T) {
	tests := []struct {
		name          string
		actorCharID   uint32
		targetCharID  uint32
		action        GuildMemberAction
		guild         *Guild
		membership    *GuildMember
		acceptErr     error
		rejectErr     error
		removeErr     error
		sendErr       error
		wantErr       bool
		wantErrIs     error
		wantAccepted  uint32
		wantRejected  uint32
		wantRemoved   uint32
		wantMailCount int
		wantRecipient uint32
		wantMailSubj  string
	}{
		{
			name:          "accept application as leader",
			actorCharID:   1,
			targetCharID:  42,
			action:        GuildMemberActionAccept,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:    &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			wantAccepted:  42,
			wantMailCount: 1,
			wantRecipient: 42,
			wantMailSubj:  "Accepted!",
		},
		{
			name:          "reject application as sub-leader",
			actorCharID:   2,
			targetCharID:  42,
			action:        GuildMemberActionReject,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:    &GuildMember{GuildID: 10, CharID: 2, OrderIndex: 2}, // sub-leader
			wantRejected:  42,
			wantMailCount: 1,
			wantRecipient: 42,
			wantMailSubj:  "Rejected",
		},
		{
			name:          "kick member as leader",
			actorCharID:   1,
			targetCharID:  42,
			action:        GuildMemberActionKick,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:    &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			wantRemoved:   42,
			wantMailCount: 1,
			wantRecipient: 42,
			wantMailSubj:  "Kicked",
		},
		{
			name:         "unauthorized - not leader or sub",
			actorCharID:  5,
			targetCharID: 42,
			action:       GuildMemberActionAccept,
			guild:        &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:   &GuildMember{GuildID: 10, CharID: 5, OrderIndex: 10},
			wantErr:      true,
			wantErrIs:    ErrUnauthorized,
		},
		{
			name:         "repo error on accept",
			actorCharID:  1,
			targetCharID: 42,
			action:       GuildMemberActionAccept,
			guild:        &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:   &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			acceptErr:    errors.New("db error"),
			wantErr:      true,
		},
		{
			name:          "mail error is best-effort",
			actorCharID:   1,
			targetCharID:  42,
			action:        GuildMemberActionAccept,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:    &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			sendErr:       errors.New("mail failed"),
			wantAccepted:  42,
			wantMailCount: 1,
			wantRecipient: 42,
			wantMailSubj:  "Accepted!",
		},
		{
			name:         "unknown action",
			actorCharID:  1,
			targetCharID: 42,
			action:       GuildMemberAction(99),
			guild:        &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 1}},
			membership:   &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			wantErr:      true,
			wantErrIs:    ErrUnknownAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildMock := &mockGuildRepo{
				membership: tt.membership,
				acceptErr:  tt.acceptErr,
				rejectErr:  tt.rejectErr,
				removeErr:  tt.removeErr,
			}
			guildMock.guild = tt.guild
			mailMock := &mockMailRepo{sendErr: tt.sendErr}

			svc := newTestGuildService(guildMock, mailMock)

			result, err := svc.OperateMember(tt.actorCharID, tt.targetCharID, tt.action)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.wantAccepted != 0 && guildMock.acceptedCharID != tt.wantAccepted {
				t.Errorf("acceptedCharID = %d, want %d", guildMock.acceptedCharID, tt.wantAccepted)
			}
			if tt.wantRejected != 0 && guildMock.rejectedCharID != tt.wantRejected {
				t.Errorf("rejectedCharID = %d, want %d", guildMock.rejectedCharID, tt.wantRejected)
			}
			if tt.wantRemoved != 0 && guildMock.removedCharID != tt.wantRemoved {
				t.Errorf("removedCharID = %d, want %d", guildMock.removedCharID, tt.wantRemoved)
			}
			if len(mailMock.sentMails) != tt.wantMailCount {
				t.Fatalf("sentMails count = %d, want %d", len(mailMock.sentMails), tt.wantMailCount)
			}
			if tt.wantMailCount > 0 {
				if mailMock.sentMails[0].recipientID != tt.wantRecipient {
					t.Errorf("mail recipientID = %d, want %d", mailMock.sentMails[0].recipientID, tt.wantRecipient)
				}
				if mailMock.sentMails[0].subject != tt.wantMailSubj {
					t.Errorf("mail subject = %q, want %q", mailMock.sentMails[0].subject, tt.wantMailSubj)
				}
			}
			if result.MailRecipientID != tt.targetCharID {
				t.Errorf("result.MailRecipientID = %d, want %d", result.MailRecipientID, tt.targetCharID)
			}
		})
	}
}

func TestGuildService_Disband(t *testing.T) {
	tests := []struct {
		name        string
		actorCharID uint32
		guild       *Guild
		disbandErr  error
		wantSuccess bool
		wantDisbID  uint32
	}{
		{
			name:        "leader disbands successfully",
			actorCharID: 1,
			guild:       &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			wantSuccess: true,
			wantDisbID:  10,
		},
		{
			name:        "non-leader cannot disband",
			actorCharID: 5,
			guild:       &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			wantSuccess: false,
		},
		{
			name:        "repo error returns failure",
			actorCharID: 1,
			guild:       &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			disbandErr:  errors.New("db error"),
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildMock := &mockGuildRepo{disbandErr: tt.disbandErr}
			guildMock.guild = tt.guild
			svc := newTestGuildService(guildMock, &mockMailRepo{})

			result, err := svc.Disband(tt.actorCharID, tt.guild.ID)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if tt.wantDisbID != 0 && guildMock.disbandedID != tt.wantDisbID {
				t.Errorf("disbandedID = %d, want %d", guildMock.disbandedID, tt.wantDisbID)
			}
		})
	}
}

func TestGuildService_ResignLeadership(t *testing.T) {
	tests := []struct {
		name           string
		actorCharID    uint32
		guild          *Guild
		members        []*GuildMember
		getMembersErr  error
		wantNewLeader  uint32
		wantErr        bool
		wantSavedCount int
		wantGuildSaved bool
	}{
		{
			name:        "transfers to next eligible member",
			actorCharID: 1,
			guild:       &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			members: []*GuildMember{
				{CharID: 1, OrderIndex: 1, IsLeader: true},
				{CharID: 2, OrderIndex: 2, AvoidLeadership: false},
			},
			wantNewLeader:  2,
			wantSavedCount: 2,
			wantGuildSaved: true,
		},
		{
			name:        "skips members avoiding leadership",
			actorCharID: 1,
			guild:       &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			members: []*GuildMember{
				{CharID: 1, OrderIndex: 1, IsLeader: true},
				{CharID: 2, OrderIndex: 2, AvoidLeadership: true},
				{CharID: 3, OrderIndex: 3, AvoidLeadership: false},
			},
			wantNewLeader:  3,
			wantSavedCount: 2,
			wantGuildSaved: true,
		},
		{
			name:        "no eligible successor returns zero",
			actorCharID: 1,
			guild:       &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			members: []*GuildMember{
				{CharID: 1, OrderIndex: 1, IsLeader: true},
				{CharID: 2, OrderIndex: 2, AvoidLeadership: true},
			},
			wantNewLeader: 0,
		},
		{
			name:          "get members error",
			actorCharID:   1,
			guild:         &Guild{ID: 10, GuildLeader: GuildLeader{LeaderCharID: 1}},
			getMembersErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildMock := &mockGuildRepo{getMembersErr: tt.getMembersErr}
			guildMock.guild = tt.guild
			guildMock.members = tt.members
			svc := newTestGuildService(guildMock, &mockMailRepo{})

			result, err := svc.ResignLeadership(tt.actorCharID, tt.guild.ID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.NewLeaderCharID != tt.wantNewLeader {
				t.Errorf("NewLeaderCharID = %d, want %d", result.NewLeaderCharID, tt.wantNewLeader)
			}
			if tt.wantSavedCount > 0 && len(guildMock.savedMembers) != tt.wantSavedCount {
				t.Errorf("savedMembers count = %d, want %d", len(guildMock.savedMembers), tt.wantSavedCount)
			}
			if tt.wantGuildSaved && guildMock.savedGuild == nil {
				t.Error("Guild should be saved")
			}
		})
	}
}

func TestGuildService_Leave(t *testing.T) {
	tests := []struct {
		name          string
		isApplicant   bool
		rejectErr     error
		removeErr     error
		sendErr       error
		wantSuccess   bool
		wantRejected  uint32
		wantRemoved   uint32
		wantMailCount int
	}{
		{
			name:          "member leaves successfully",
			isApplicant:   false,
			wantSuccess:   true,
			wantRemoved:   1,
			wantMailCount: 1,
		},
		{
			name:          "applicant withdraws via reject",
			isApplicant:   true,
			wantSuccess:   true,
			wantRejected:  1,
			wantMailCount: 1,
		},
		{
			name:        "remove error returns failure",
			isApplicant: false,
			removeErr:   errors.New("db error"),
			wantSuccess: false,
		},
		{
			name:          "mail error is best-effort",
			isApplicant:   false,
			sendErr:       errors.New("mail failed"),
			wantSuccess:   true,
			wantRemoved:   1,
			wantMailCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildMock := &mockGuildRepo{
				rejectErr: tt.rejectErr,
				removeErr: tt.removeErr,
			}
			guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
			mailMock := &mockMailRepo{sendErr: tt.sendErr}
			svc := newTestGuildService(guildMock, mailMock)

			result, _ := svc.Leave(1, 10, tt.isApplicant, "TestGuild")
			if result.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if tt.wantRejected != 0 && guildMock.rejectedCharID != tt.wantRejected {
				t.Errorf("rejectedCharID = %d, want %d", guildMock.rejectedCharID, tt.wantRejected)
			}
			if tt.wantRemoved != 0 && guildMock.removedCharID != tt.wantRemoved {
				t.Errorf("removedCharID = %d, want %d", guildMock.removedCharID, tt.wantRemoved)
			}
			if len(mailMock.sentMails) != tt.wantMailCount {
				t.Errorf("sentMails count = %d, want %d", len(mailMock.sentMails), tt.wantMailCount)
			}
		})
	}
}

func TestGuildService_PostScout(t *testing.T) {
	strings := ScoutInviteStrings{Title: "Invite", Body: "Join 「%s」"}

	tests := []struct {
		name          string
		membership    *GuildMember
		guild         *Guild
		hasInvite     bool
		hasInviteErr  error
		createAppErr  error
		getMemberErr  error
		wantErr       error
	}{
		{
			name:       "successful scout",
			membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			guild:      &Guild{ID: 10, Name: "TestGuild"},
		},
		{
			name:       "already invited",
			membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
			guild:      &Guild{ID: 10, Name: "TestGuild"},
			hasInvite:  true,
			wantErr:    ErrAlreadyInvited,
		},
		{
			name:       "cannot recruit",
			membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 10}, // not recruiter, not sub-leader
			guild:      &Guild{ID: 10, Name: "TestGuild"},
			wantErr:    ErrCannotRecruit,
		},
		{
			name:         "nil membership",
			getMemberErr: errors.New("not found"),
			guild:        &Guild{ID: 10, Name: "TestGuild"},
			wantErr:      errors.New("any"), // just check err != nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildMock := &mockGuildRepo{
				membership:      tt.membership,
				hasInviteResult: tt.hasInvite,
				hasInviteErr:    tt.hasInviteErr,
				createAppErr:    tt.createAppErr,
				getMemberErr:    tt.getMemberErr,
			}
			guildMock.guild = tt.guild
			svc := newTestGuildService(guildMock, &mockMailRepo{})

			err := svc.PostScout(1, 42, strings)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if errors.Is(tt.wantErr, ErrAlreadyInvited) || errors.Is(tt.wantErr, ErrCannotRecruit) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("Expected %v, got %v", tt.wantErr, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGuildService_AnswerScout(t *testing.T) {
	strings := AnswerScoutStrings{
		SuccessTitle:  "Success!",
		SuccessBody:   "Joined 「%s」.",
		AcceptedTitle: "Accepted",
		AcceptedBody:  "Accepted invite to 「%s」.",
		RejectedTitle: "Rejected",
		RejectedBody:  "Rejected invite to 「%s」.",
		DeclinedTitle: "Declined",
		DeclinedBody:  "Declined invite to 「%s」.",
	}

	tests := []struct {
		name          string
		accept        bool
		guild         *Guild
		hasInvite     bool
		acceptErr     error
		rejectErr     error
		sendErr       error
		getErr        error
		wantSuccess   bool
		wantErr       error
		wantMailCount int
		wantAccepted  uint32
		wantDeclined  uint32
	}{
		{
			name:          "accept invitation",
			accept:        true,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 50}},
			hasInvite:     true,
			wantSuccess:   true,
			wantMailCount: 2,
			wantAccepted:  1,
		},
		{
			name:          "decline invitation",
			accept:        false,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 50}},
			hasInvite:     true,
			wantSuccess:   true,
			wantMailCount: 2,
			wantDeclined:  1,
		},
		{
			name:        "application missing",
			accept:      true,
			guild:       &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 50}},
			hasInvite:   false,
			wantSuccess: false,
			wantErr:     ErrApplicationMissing,
		},
		{
			name:    "guild not found",
			accept:  true,
			guild:   &Guild{ID: 10, Name: "TestGuild"},
			getErr:  errors.New("not found"),
			wantErr: errors.New("any"),
		},
		{
			name:          "mail error is best-effort",
			accept:        true,
			guild:         &Guild{ID: 10, Name: "TestGuild", GuildLeader: GuildLeader{LeaderCharID: 50}},
			hasInvite:     true,
			sendErr:       errors.New("mail failed"),
			wantSuccess:   true,
			wantMailCount: 2,
			wantAccepted:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildMock := &mockGuildRepo{
				hasInviteResult: tt.hasInvite,
				acceptErr:       tt.acceptErr,
				rejectErr:       tt.rejectErr,
			}
			guildMock.guild = tt.guild
			guildMock.getErr = tt.getErr
			mailMock := &mockMailRepo{sendErr: tt.sendErr}
			svc := newTestGuildService(guildMock, mailMock)

			result, err := svc.AnswerScout(1, 50, tt.accept, strings)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if errors.Is(tt.wantErr, ErrApplicationMissing) && !errors.Is(err, ErrApplicationMissing) {
					t.Errorf("Expected ErrApplicationMissing, got %v", err)
				}
				if result != nil && result.Success {
					t.Error("Result should not be successful")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if len(mailMock.sentMails) != tt.wantMailCount {
				t.Errorf("sentMails count = %d, want %d", len(mailMock.sentMails), tt.wantMailCount)
			}
			if tt.wantAccepted != 0 && guildMock.acceptInviteCharID != tt.wantAccepted {
				t.Errorf("acceptInviteCharID = %d, want %d", guildMock.acceptInviteCharID, tt.wantAccepted)
			}
			if tt.wantDeclined != 0 && guildMock.declineInviteCharID != tt.wantDeclined {
				t.Errorf("declineInviteCharID = %d, want %d", guildMock.declineInviteCharID, tt.wantDeclined)
			}
		})
	}
}
