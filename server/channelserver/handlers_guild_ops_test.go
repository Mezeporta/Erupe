package channelserver

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfOperateGuild tests ---

func TestOperateGuild_Disband_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildDisband,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.disbandedID != 10 {
		t.Errorf("Disband called with guild %d, want 10", guildMock.disbandedID)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("No response data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_Disband_NotLeader(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 5},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999 // different from session charID
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildDisband,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.disbandedID != 0 {
		t.Error("Disband should not be called for non-leader")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_Disband_RepoError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
		disbandErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildDisband,
	}

	handleMsgMhfOperateGuild(session, pkt)

	// response=0 when disband fails
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_Resign_TransferLeadership(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	guildMock.members = []*GuildMember{
		{CharID: 1, OrderIndex: 1, IsLeader: true},
		{CharID: 2, OrderIndex: 2, AvoidLeadership: false},
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildResign,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.guild.LeaderCharID != 2 {
		t.Errorf("Leader should transfer to charID 2, got %d", guildMock.guild.LeaderCharID)
	}
	if len(guildMock.savedMembers) < 2 {
		t.Fatalf("Expected 2 saved members, got %d", len(guildMock.savedMembers))
	}
	if guildMock.savedGuild == nil {
		t.Error("Guild should be saved after resign")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_Resign_SkipsAvoidLeadership(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	guildMock.members = []*GuildMember{
		{CharID: 1, OrderIndex: 1, IsLeader: true},
		{CharID: 2, OrderIndex: 2, AvoidLeadership: true},
		{CharID: 3, OrderIndex: 3, AvoidLeadership: false},
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildResign,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.guild.LeaderCharID != 3 {
		t.Errorf("Leader should transfer to charID 3 (skipping 2), got %d", guildMock.guild.LeaderCharID)
	}
}

func TestOperateGuild_Apply_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 5},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildApply,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.createdAppArgs == nil {
		t.Fatal("CreateApplication should be called")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_Apply_RepoError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership:   &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 5},
		createAppErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildApply,
	}

	handleMsgMhfOperateGuild(session, pkt)

	// Should still succeed with 0 leader ID
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_Leave_AsApplicant(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsApplicant: true, OrderIndex: 5},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildLeave,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.rejectedCharID != 1 {
		t.Errorf("RejectApplication should be called for applicant, got rejectedCharID=%d", guildMock.rejectedCharID)
	}
	if guildMock.removedCharID != 0 {
		t.Error("RemoveCharacter should not be called for applicant")
	}
}

func TestOperateGuild_Leave_AsMember(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsApplicant: false, OrderIndex: 5},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildLeave,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.removedCharID != 1 {
		t.Errorf("RemoveCharacter should be called with charID 1, got %d", guildMock.removedCharID)
	}
	if len(mailMock.sentMails) != 1 {
		t.Fatalf("Expected 1 withdrawal mail, got %d", len(mailMock.sentMails))
	}
	if mailMock.sentMails[0].recipientID != 1 {
		t.Errorf("Mail recipientID = %d, want 1", mailMock.sentMails[0].recipientID)
	}
}

func TestOperateGuild_Leave_MailError(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{sendErr: errNotFound}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsApplicant: false, OrderIndex: 5},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildLeave,
	}

	// Should not panic; mail error is logged as warning
	handleMsgMhfOperateGuild(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_UpdateComment_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildUpdateComment,
		Data2:     newNullTermBF([]byte("Test\x00")),
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.savedGuild == nil {
		t.Error("Guild should be saved after comment update")
	}
}

func TestOperateGuild_UpdateComment_NotLeader(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 10}, // not leader, not sub-leader
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildUpdateComment,
	}

	handleMsgMhfOperateGuild(session, pkt)

	// Should return fail ack
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected fail response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuild_UpdateMotto_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildUpdateMotto,
		Data1:     newMottoBF(5, 3),
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.savedGuild == nil {
		t.Error("Guild should be saved after motto update")
	}
	if guildMock.savedGuild.MainMotto != 3 {
		t.Errorf("MainMotto = %d, want 3", guildMock.savedGuild.MainMotto)
	}
	if guildMock.savedGuild.SubMotto != 5 {
		t.Errorf("SubMotto = %d, want 5", guildMock.savedGuild.SubMotto)
	}
}

func TestOperateGuild_UpdateMotto_NotLeader(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 10},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildUpdateMotto,
	}

	handleMsgMhfOperateGuild(session, pkt)

	if guildMock.savedGuild != nil {
		t.Error("Guild should not be saved when not leader")
	}
}

func TestOperateGuild_GuildNotFound(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuild{
		AckHandle: 100,
		GuildID:   10,
		Action:    mhfpacket.OperateGuildDisband,
	}

	handleMsgMhfOperateGuild(session, pkt)

	// Should return fail ack
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfOperateGuildMember tests ---

func TestOperateGuildMember_Accept(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildMember{
		AckHandle: 100,
		GuildID:   10,
		CharID:    42,
		Action:    mhfpacket.OPERATE_GUILD_MEMBER_ACTION_ACCEPT,
	}

	handleMsgMhfOperateGuildMember(session, pkt)

	if guildMock.acceptedCharID != 42 {
		t.Errorf("AcceptApplication charID = %d, want 42", guildMock.acceptedCharID)
	}
	if len(mailMock.sentMails) != 1 {
		t.Fatalf("Expected 1 mail, got %d", len(mailMock.sentMails))
	}
	if mailMock.sentMails[0].recipientID != 42 {
		t.Errorf("Mail recipientID = %d, want 42", mailMock.sentMails[0].recipientID)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuildMember_Reject(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildMember{
		AckHandle: 100,
		GuildID:   10,
		CharID:    42,
		Action:    mhfpacket.OPERATE_GUILD_MEMBER_ACTION_REJECT,
	}

	handleMsgMhfOperateGuildMember(session, pkt)

	if guildMock.rejectedCharID != 42 {
		t.Errorf("RejectApplication charID = %d, want 42", guildMock.rejectedCharID)
	}
	if len(mailMock.sentMails) != 1 {
		t.Fatalf("Expected 1 mail, got %d", len(mailMock.sentMails))
	}
}

func TestOperateGuildMember_Kick(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildMember{
		AckHandle: 100,
		GuildID:   10,
		CharID:    42,
		Action:    mhfpacket.OPERATE_GUILD_MEMBER_ACTION_KICK,
	}

	handleMsgMhfOperateGuildMember(session, pkt)

	if guildMock.removedCharID != 42 {
		t.Errorf("RemoveCharacter charID = %d, want 42", guildMock.removedCharID)
	}
	if len(mailMock.sentMails) != 1 {
		t.Fatalf("Expected 1 mail, got %d", len(mailMock.sentMails))
	}
}

func TestOperateGuildMember_MailError(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{sendErr: errNotFound}
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, IsLeader: true, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.mailRepo = mailMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildMember{
		AckHandle: 100,
		GuildID:   10,
		CharID:    42,
		Action:    mhfpacket.OPERATE_GUILD_MEMBER_ACTION_ACCEPT,
	}

	// Should not panic; mail error logged as warning
	handleMsgMhfOperateGuildMember(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateGuildMember_NotLeaderOrSub(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 10}, // not sub-leader
	}
	guildMock.guild = &Guild{ID: 10, Name: "TestGuild"}
	guildMock.guild.LeaderCharID = 999 // not the session char
	server.guildRepo = guildMock
	server.mailRepo = &mockMailRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildMember{
		AckHandle: 100,
		GuildID:   10,
		CharID:    42,
		Action:    mhfpacket.OPERATE_GUILD_MEMBER_ACTION_ACCEPT,
	}

	handleMsgMhfOperateGuildMember(session, pkt)

	if guildMock.acceptedCharID != 0 {
		t.Error("Should not accept when actor lacks permission")
	}
}

// --- byteframe helpers for packet Data fields ---

func newNullTermBF(data []byte) *byteframe.ByteFrame {
	bf := byteframe.NewByteFrame()
	bf.WriteBytes(data)
	_, _ = bf.Seek(0, 0)
	return bf
}

func newMottoBF(sub, main uint8) *byteframe.ByteFrame {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)    // skipped
	bf.WriteUint8(sub)   // SubMotto
	bf.WriteUint8(main)  // MainMotto
	_, _ = bf.Seek(0, 0)
	return bf
}
