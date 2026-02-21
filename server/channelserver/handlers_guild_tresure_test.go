package channelserver

import (
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfEnumerateGuildTresure tests ---

func TestEnumerateGuildTresure_NoGuild(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildTresure{AckHandle: 100, MaxHunts: 30}

	handleMsgMhfEnumerateGuildTresure(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateGuildTresure_PendingHunt(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		pendingHunt: &TreasureHunt{
			HuntID:      1,
			Destination: 5,
			Level:       3,
			Start:       time.Now(),
			HuntData:    make([]byte, 10),
		},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildTresure{AckHandle: 100, MaxHunts: 1}

	handleMsgMhfEnumerateGuildTresure(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 8 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateGuildTresure_GuildHunts(t *testing.T) {
	server := createMockServer()
	// Set a large expiry so hunts are considered active
	server.erupeConfig.GameplayOptions.TreasureHuntExpiry = 86400
	guildMock := &mockGuildRepoOps{
		guildHunts: []*TreasureHunt{
			{HuntID: 1, Destination: 5, Level: 2, Start: TimeAdjusted(), HuntData: make([]byte, 10)},
			{HuntID: 2, Destination: 8, Level: 3, Start: TimeAdjusted(), HuntData: make([]byte, 10)},
		},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildTresure{AckHandle: 100, MaxHunts: 30}

	handleMsgMhfEnumerateGuildTresure(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 8 {
			t.Errorf("Response too short for 2 hunts: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateGuildTresure_ListError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		listHuntsErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildTresure{AckHandle: 100, MaxHunts: 30}

	handleMsgMhfEnumerateGuildTresure(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfAcquireGuildTresure tests ---

func TestAcquireGuildTresure_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireGuildTresure{AckHandle: 100, HuntID: 42}

	handleMsgMhfAcquireGuildTresure(session, pkt)

	if guildMock.acquireHuntID != 42 {
		t.Errorf("AcquireHunt ID = %d, want 42", guildMock.acquireHuntID)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfOperateGuildTresureReport tests ---

func TestOperateGuildTresureReport_Register(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildTresureReport{
		AckHandle: 100,
		HuntID:    42,
		State:     0, // Register
	}

	handleMsgMhfOperateGuildTresureReport(session, pkt)

	if guildMock.reportHuntID != 42 {
		t.Errorf("RegisterHuntReport ID = %d, want 42", guildMock.reportHuntID)
	}
}

func TestOperateGuildTresureReport_Collect(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildTresureReport{
		AckHandle: 100,
		HuntID:    42,
		State:     1, // Collect
	}

	handleMsgMhfOperateGuildTresureReport(session, pkt)

	if guildMock.collectHuntID != 42 {
		t.Errorf("CollectHunt ID = %d, want 42", guildMock.collectHuntID)
	}
}

func TestOperateGuildTresureReport_Claim(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperateGuildTresureReport{
		AckHandle: 100,
		HuntID:    42,
		State:     2, // Claim
	}

	handleMsgMhfOperateGuildTresureReport(session, pkt)

	if guildMock.claimHuntID != 42 {
		t.Errorf("ClaimHuntReward ID = %d, want 42", guildMock.claimHuntID)
	}
}

// --- handleMsgMhfGetGuildTresureSouvenir tests ---

func TestGetGuildTresureSouvenir_Empty(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetGuildTresureSouvenir{AckHandle: 100}

	handleMsgMhfGetGuildTresureSouvenir(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
