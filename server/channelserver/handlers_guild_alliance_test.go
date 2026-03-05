package channelserver

import (
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfCreateJoint tests ---

func TestCreateJoint_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCreateJoint{
		AckHandle: 100,
		GuildID:   10,
		Name:      "TestAlliance",
	}

	handleMsgMhfCreateJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCreateJoint_Error(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{createAllianceErr: errNotFound}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCreateJoint{
		AckHandle: 100,
		GuildID:   10,
		Name:      "TestAlliance",
	}

	// Should not panic; error is logged
	handleMsgMhfCreateJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfOperateJoint tests ---

func TestOperateJoint_Disband_AsOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 10,
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1 // session charID
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_DISBAND,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.deletedAllianceID != 5 {
		t.Errorf("DeleteAlliance called with %d, want 5", guildMock.deletedAllianceID)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateJoint_Disband_NotOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 99, // different guild
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_DISBAND,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.deletedAllianceID != 0 {
		t.Error("Should not disband when not alliance owner")
	}
}

func TestOperateJoint_Leave_AsLeader(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 99,
			SubGuild1ID:   10,
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_LEAVE,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.removedAllyArgs == nil {
		t.Fatal("RemoveGuildFromAlliance should be called")
	}
	if guildMock.removedAllyArgs[1] != 10 {
		t.Errorf("Removed guildID = %d, want 10", guildMock.removedAllyArgs[1])
	}
}

func TestOperateJoint_Leave_NotLeader(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{ID: 5, ParentGuildID: 99},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999 // not session char
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_LEAVE,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.removedAllyArgs != nil {
		t.Error("Non-leader should not be able to leave alliance")
	}
}

func TestOperateJoint_Kick_AsAllianceOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 10,
			ParentGuild:   Guild{},
			SubGuild1ID:   20,
		},
	}
	guildMock.alliance.ParentGuild.LeaderCharID = 1 // session char owns alliance
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1

	data1 := byteframe.NewByteFrame()
	data1.WriteUint32(20) // guildID to kick
	_, _ = data1.Seek(0, 0)

	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_KICK,
		Data1:      data1,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.removedAllyArgs == nil {
		t.Fatal("RemoveGuildFromAlliance should be called for kick")
	}
	if guildMock.removedAllyArgs[1] != 20 {
		t.Errorf("Kicked guildID = %d, want 20", guildMock.removedAllyArgs[1])
	}
}

func TestOperateJoint_Kick_NotOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 99,
			ParentGuild:   Guild{},
		},
	}
	guildMock.alliance.ParentGuild.LeaderCharID = 999 // not session char
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_KICK,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.removedAllyArgs != nil {
		t.Error("Non-owner should not kick from alliance")
	}
}

func TestOperateJoint_Allow_AsOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 10,
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_ALLOW,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.allianceRecruitingSet == nil || !*guildMock.allianceRecruitingSet {
		t.Error("SetAllianceRecruiting should be called with true")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateJoint_Allow_NotOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 99, // different guild
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_ALLOW,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.allianceRecruitingSet != nil {
		t.Error("Non-owner should not toggle alliance recruiting")
	}
}

func TestOperateJoint_Deny_AsOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 10,
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_DENY,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.allianceRecruitingSet == nil || *guildMock.allianceRecruitingSet {
		t.Error("SetAllianceRecruiting should be called with false")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestOperateJoint_Deny_NotOwner(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			ParentGuildID: 99,
		},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_DENY,
	}

	handleMsgMhfOperateJoint(session, pkt)

	if guildMock.allianceRecruitingSet != nil {
		t.Error("Non-owner should not toggle alliance recruiting")
	}
}

// --- handleMsgMhfInfoJoint tests ---

func TestInfoJoint_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			Name:          "TestAlliance",
			CreatedAt:     time.Now(),
			TotalMembers:  15,
			ParentGuildID: 10,
			ParentGuild:   Guild{Name: "ParentGuild", MemberCount: 5},
		},
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoJoint{AckHandle: 100, AllianceID: 5}

	handleMsgMhfInfoJoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 10 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestInfoJoint_WithSubGuilds(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{
			ID:            5,
			Name:          "BigAlliance",
			CreatedAt:     time.Now(),
			TotalMembers:  30,
			ParentGuildID: 10,
			ParentGuild:   Guild{Name: "Parent", MemberCount: 10},
			SubGuild1ID:   20,
			SubGuild1:     Guild{Name: "Sub1", MemberCount: 10},
			SubGuild2ID:   30,
			SubGuild2:     Guild{Name: "Sub2", MemberCount: 10},
		},
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoJoint{AckHandle: 100, AllianceID: 5}

	handleMsgMhfInfoJoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 30 {
			t.Errorf("Response too short for alliance with sub guilds: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestInfoJoint_NotFound(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{getAllianceErr: errNotFound}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoJoint{AckHandle: 100, AllianceID: 999}

	handleMsgMhfInfoJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestInfoJoint_NilAlliance(t *testing.T) {
	server := createMockServer()
	// alliance is nil, no error — simulates deleted alliance
	guildMock := &mockGuildRepo{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoJoint{AckHandle: 100, AllianceID: 999}

	handleMsgMhfInfoJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued — would softlock the client")
	}
}

func TestOperateJoint_NilGuild(t *testing.T) {
	server := createMockServer()
	// guild is nil — simulates deleted guild
	guildMock := &mockGuildRepo{
		alliance: &GuildAlliance{ID: 5, ParentGuildID: 10},
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 5,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_DISBAND,
	}

	handleMsgMhfOperateJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued — would softlock the client")
	}
}

func TestOperateJoint_NilAlliance(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateJoint{
		AckHandle:  100,
		AllianceID: 999,
		GuildID:    10,
		Action:     mhfpacket.OPERATE_JOINT_DISBAND,
	}

	handleMsgMhfOperateJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued — would softlock the client")
	}
}

// --- scanAllianceWithGuilds nil guild tests (issue #171) ---

func TestInfoJoint_MissingSubGuild1(t *testing.T) {
	// Verify that GetAllianceByID returns an error when sub guild 1 references
	// a non-existent guild (nil return from GetByID). This is the scenario from
	// issue #171 — a deleted guild causes a nil dereference in scanAllianceWithGuilds.
	server := createMockServer()
	guildMock := &mockGuildRepo{
		// GetAllianceByID returns an error for missing guilds because
		// scanAllianceWithGuilds calls GetByID for each sub guild.
		// With guild=nil and SubGuild1ID > 0, GetByID returns nil,
		// and scanAllianceWithGuilds should return an error rather than panic.
		getAllianceErr: errNotFound,
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoJoint{AckHandle: 100, AllianceID: 5}
	handleMsgMhfInfoJoint(session, pkt)

	// Handler should send a response even on error (not softlock)
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued — would softlock the client")
	}
}

func TestInfoJoint_MissingSubGuild2(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{
		getAllianceErr: errNotFound,
	}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoJoint{AckHandle: 100, AllianceID: 6}
	handleMsgMhfInfoJoint(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued — would softlock the client")
	}
}
