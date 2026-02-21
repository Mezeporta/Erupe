package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfAnswerGuildScout tests ---

func TestAnswerGuildScout_Accept(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoOps{
		application: &GuildApplication{GuildID: 10, CharID: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
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
	guildMock := &mockGuildRepoOps{
		application: &GuildApplication{GuildID: 10, CharID: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
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
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	server.mailRepo = &mockMailRepo{}
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
	guildMock := &mockGuildRepoOps{
		application: nil, // no application found
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
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
	guildMock := &mockGuildRepoOps{
		application: &GuildApplication{GuildID: 10, CharID: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 50
	server.guildRepo = guildMock
	server.mailRepo = mailMock
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
