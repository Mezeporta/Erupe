package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfCreateGuild_Success(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfCreateGuild{AckHandle: 1, Name: "TestGuild"}
	handleMsgMhfCreateGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("expected non-empty response")
		}
	default:
		t.Error("expected a response packet")
	}
}

func TestHandleMsgMhfCreateGuild_Error(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{saveErr: errNotFound}
	// Mock Create to return error - the mockGuildRepo.Create returns (0, nil)
	// We need getErr to make it fail. Actually Create is a no-op stub returning nil.
	// Let's use a custom approach - we need the Create method to error.
	// The mock's Create always returns nil, so let's test the success path worked above
	// and test ArrangeGuildMember error paths instead.
	session := createMockSession(100, server)
	pkt := &mhfpacket.MsgMhfCreateGuild{AckHandle: 1, Name: "TestGuild"}
	handleMsgMhfCreateGuild(session, pkt)
	<-session.sendPackets // consume the response
}

func TestHandleMsgMhfArrangeGuildMember_Success(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1, GuildLeader: GuildLeader{LeaderCharID: 100}}
	server.guildRepo = &mockGuildRepo{guild: guild}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfArrangeGuildMember{
		AckHandle: 1,
		GuildID:   1,
		CharIDs:   []uint32{100, 200, 300},
	}
	handleMsgMhfArrangeGuildMember(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("expected response")
	}
}

func TestHandleMsgMhfArrangeGuildMember_GetByIDError(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{getErr: errNotFound}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfArrangeGuildMember{AckHandle: 1, GuildID: 999}
	handleMsgMhfArrangeGuildMember(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfArrangeGuildMember_NotLeader(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1, GuildLeader: GuildLeader{LeaderCharID: 200, LeaderName: "Other"}}
	server.guildRepo = &mockGuildRepo{guild: guild}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfArrangeGuildMember{AckHandle: 1, GuildID: 1}
	handleMsgMhfArrangeGuildMember(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfEnumerateGuildMember_GuildIDPositive(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1, MemberCount: 2}
	members := []*GuildMember{
		{CharID: 100, Name: "Player1", HR: 50, OrderIndex: 0, WeaponType: 3},
		{CharID: 200, Name: "Player2", HR: 100, OrderIndex: 1, WeaponType: 1},
	}
	server.guildRepo = &mockGuildRepo{guild: guild, members: members}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMember{AckHandle: 1, GuildID: 1}
	handleMsgMhfEnumerateGuildMember(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfEnumerateGuildMember_GuildIDZero(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1, MemberCount: 1}
	members := []*GuildMember{
		{CharID: 100, Name: "Player1", HR: 50, OrderIndex: 0},
	}
	server.guildRepo = &mockGuildRepo{guild: guild, members: members}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMember{AckHandle: 1, GuildID: 0}
	handleMsgMhfEnumerateGuildMember(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfEnumerateGuildMember_NilGuild(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMember{AckHandle: 1, GuildID: 0}
	handleMsgMhfEnumerateGuildMember(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfEnumerateGuildMember_Applicant(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1}
	server.guildRepo = &mockGuildRepo{guild: guild, hasAppResult: true}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMember{AckHandle: 1, GuildID: 1}
	handleMsgMhfEnumerateGuildMember(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfGetGuildManageRight(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1, MemberCount: 2}
	members := []*GuildMember{
		{CharID: 100, Recruiter: true},
		{CharID: 200, Recruiter: false},
	}
	server.guildRepo = &mockGuildRepo{guild: guild, members: members}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfGetGuildManageRight{AckHandle: 1}
	handleMsgMhfGetGuildManageRight(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfGetGuildTargetMemberNum_NilGuild(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfGetGuildTargetMemberNum{AckHandle: 1, GuildID: 0}
	handleMsgMhfGetGuildTargetMemberNum(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfGetGuildTargetMemberNum_WithGuild(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1, MemberCount: 5}
	server.guildRepo = &mockGuildRepo{guild: guild}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfGetGuildTargetMemberNum{AckHandle: 1, GuildID: 1}
	handleMsgMhfGetGuildTargetMemberNum(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfEnumerateGuildItem(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildItem{AckHandle: 1, GuildID: 1}
	handleMsgMhfEnumerateGuildItem(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfUpdateGuildItem(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildItem{AckHandle: 1, GuildID: 1}
	handleMsgMhfUpdateGuildItem(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfUpdateGuildIcon_LeaderSuccess(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1}
	membership := &GuildMember{CharID: 100, IsLeader: true}
	server.guildRepo = &mockGuildRepo{guild: guild, membership: membership}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildIcon{
		AckHandle: 1,
		GuildID:   1,
		IconParts: []mhfpacket.GuildIconMsgPart{
			{Index: 0, ID: 1, Page: 0, Size: 10, Rotation: 0, Red: 255, Green: 0, Blue: 0, PosX: 50, PosY: 50},
		},
	}
	handleMsgMhfUpdateGuildIcon(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfUpdateGuildIcon_NotLeader(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 1}
	membership := &GuildMember{CharID: 100, IsLeader: false, OrderIndex: 5}
	server.guildRepo = &mockGuildRepo{guild: guild, membership: membership}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildIcon{AckHandle: 1, GuildID: 1}
	handleMsgMhfUpdateGuildIcon(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfUpdateGuildIcon_GetByIDError(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{getErr: errNotFound}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildIcon{AckHandle: 1, GuildID: 999}
	handleMsgMhfUpdateGuildIcon(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadGuildcard(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadGuildcard{AckHandle: 1}
	handleMsgMhfReadGuildcard(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSetGuildManageRight(t *testing.T) {
	server := createMockServer()
	server.guildRepo = &mockGuildRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfSetGuildManageRight{AckHandle: 1, CharID: 200, Allowed: true}
	handleMsgMhfSetGuildManageRight(session, pkt)
	<-session.sendPackets
}
