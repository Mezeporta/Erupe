package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfAnswerGuildScout tests ---

func TestAnswerGuildScout_Accept(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepo{
		application: &GuildApplication{GuildID: 10, CharID: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAnswerGuildScout{
		AckHandle: 100,
		LeaderID:  50,
		Answer:    true,
	}

	handleMsgMhfAnswerGuildScout(session, pkt)

	if guildMock.acceptedCharID != 1 {
		t.Errorf("AcceptApplication charID = %d, want 1", guildMock.acceptedCharID)
	}
	if len(mailMock.sentMails) != 2 {
		t.Fatalf("Expected 2 mails (self + leader), got %d", len(mailMock.sentMails))
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestAnswerGuildScout_Decline(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepo{
		application: &GuildApplication{GuildID: 10, CharID: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAnswerGuildScout{
		AckHandle: 100,
		LeaderID:  50,
		Answer:    false,
	}

	handleMsgMhfAnswerGuildScout(session, pkt)

	if guildMock.rejectedCharID != 1 {
		t.Errorf("RejectApplication charID = %d, want 1", guildMock.rejectedCharID)
	}
	if len(mailMock.sentMails) != 2 {
		t.Fatalf("Expected 2 mails (self + leader), got %d", len(mailMock.sentMails))
	}
}

func TestAnswerGuildScout_GuildNotFound(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	server.mailRepo = &mockMailRepo{}
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAnswerGuildScout{
		AckHandle: 100,
		LeaderID:  50,
		Answer:    true,
	}

	handleMsgMhfAnswerGuildScout(session, pkt)

	// Should return fail ack
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestAnswerGuildScout_ApplicationMissing(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepo{
		application: nil, // no application found
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAnswerGuildScout{
		AckHandle: 100,
		LeaderID:  50,
		Answer:    true,
	}

	handleMsgMhfAnswerGuildScout(session, pkt)

	// No mails should be sent when application is missing
	if len(mailMock.sentMails) != 0 {
		t.Errorf("Expected 0 mails for missing application, got %d", len(mailMock.sentMails))
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestAnswerGuildScout_MailError(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{sendErr: errNotFound}
	guildMock := &mockGuildRepo{
		application: &GuildApplication{GuildID: 10, CharID: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAnswerGuildScout{
		AckHandle: 100,
		LeaderID:  50,
		Answer:    true,
	}

	// Should not panic; mail errors logged as warnings
	handleMsgMhfAnswerGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfGetRejectGuildScout tests ---

func TestGetRejectGuildScout_Restricted(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.bools["restrict_guild_scout"] = true
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRejectGuildScout{AckHandle: 100}

	handleMsgMhfGetRejectGuildScout(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestGetRejectGuildScout_Open(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.bools["restrict_guild_scout"] = false
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRejectGuildScout{AckHandle: 100}

	handleMsgMhfGetRejectGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetRejectGuildScout_DBError(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.readErr = errNotFound
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRejectGuildScout{AckHandle: 100}

	handleMsgMhfGetRejectGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfSetRejectGuildScout tests ---

func TestSetRejectGuildScout_Success(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetRejectGuildScout{
		AckHandle: 100,
		Reject:    true,
	}

	handleMsgMhfSetRejectGuildScout(session, pkt)

	if !charMock.bools["restrict_guild_scout"] {
		t.Error("restrict_guild_scout should be true")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestSetRejectGuildScout_DBError(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.saveErr = errNotFound
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetRejectGuildScout{
		AckHandle: 100,
		Reject:    true,
	}

	handleMsgMhfSetRejectGuildScout(session, pkt)

	// Should return fail ack
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfPostGuildScout tests ---

func TestPostGuildScout_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		membership: &GuildMember{GuildID: 10, Recruiter: true},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.mailRepo = &mockMailRepo{}
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostGuildScout{
		AckHandle: 100,
		CharID:    42,
	}

	handleMsgMhfPostGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestPostGuildScout_AlreadyInvited(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		membership:  &GuildMember{GuildID: 10, Recruiter: true},
		createAppErr: ErrAlreadyInvited,
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.mailRepo = &mockMailRepo{}
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostGuildScout{
		AckHandle: 100,
		CharID:    42,
	}

	handleMsgMhfPostGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestPostGuildScout_Error(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		getMemberErr: errNotFound,
	}
	server.guildRepo = guildMock
	server.mailRepo = &mockMailRepo{}
	ensureGuildService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostGuildScout{
		AckHandle: 100,
		CharID:    42,
	}

	handleMsgMhfPostGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfCancelGuildScout tests ---

func TestCancelGuildScout_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		membership: &GuildMember{GuildID: 10, Recruiter: true},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCancelGuildScout{
		AckHandle:    100,
		InvitationID: 42,
	}

	handleMsgMhfCancelGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCancelGuildScout_NoMembership(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		getMemberErr: errNotFound,
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCancelGuildScout{
		AckHandle:    100,
		InvitationID: 42,
	}

	handleMsgMhfCancelGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCancelGuildScout_NilMembership(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		membership: nil,
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCancelGuildScout{
		AckHandle:    100,
		InvitationID: 42,
	}

	handleMsgMhfCancelGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCancelGuildScout_GuildNotFound(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		membership: &GuildMember{GuildID: 99, Recruiter: true},
		getErr:     errNotFound,
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCancelGuildScout{
		AckHandle:    100,
		InvitationID: 42,
	}

	handleMsgMhfCancelGuildScout(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfGetGuildScoutList tests ---

func TestGetGuildScoutList_NoGuildNoPrevID(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{} // GetByCharID returns nil
	server.guildRepo = guildMock
	session := createMockSession(1, server)
	session.prevGuildID = 0

	pkt := &mhfpacket.MsgMhfGetGuildScoutList{AckHandle: 100}

	handleMsgMhfGetGuildScoutList(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetGuildScoutList_NilGuildWithPrevID_GetByIDFails(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{} // GetByCharID returns nil, GetByID for prevGuildID returns not found
	server.guildRepo = guildMock
	session := createMockSession(1, server)
	session.prevGuildID = 99 // non-zero triggers else branch

	pkt := &mhfpacket.MsgMhfGetGuildScoutList{AckHandle: 100}

	handleMsgMhfGetGuildScoutList(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetGuildScoutList_WithGuild(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 10, Name: "TestGuild"}
	guildMock := &mockGuildRepo{}
	guildMock.guild = guild
	server.guildRepo = guildMock
	session := createMockSession(1, server)
	session.prevGuildID = 10

	pkt := &mhfpacket.MsgMhfGetGuildScoutList{AckHandle: 100}

	handleMsgMhfGetGuildScoutList(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
