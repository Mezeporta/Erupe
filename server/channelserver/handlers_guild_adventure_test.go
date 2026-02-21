package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfLoadGuildAdventure tests ---

func TestLoadGuildAdventure_NoAdventures(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		adventures: []*GuildAdventure{},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadGuildAdventure{AckHandle: 100}

	handleMsgMhfLoadGuildAdventure(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestLoadGuildAdventure_WithAdventures(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		adventures: []*GuildAdventure{
			{ID: 1, Destination: 5, Charge: 0, Depart: 1000, Return: 2000, CollectedBy: ""},
			{ID: 2, Destination: 8, Charge: 100, Depart: 1000, Return: 2000, CollectedBy: "1"},
		},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadGuildAdventure{AckHandle: 100}

	handleMsgMhfLoadGuildAdventure(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 10 {
			t.Errorf("Response too short for 2 adventures: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestLoadGuildAdventure_DBError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		listAdvErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadGuildAdventure{AckHandle: 100}

	handleMsgMhfLoadGuildAdventure(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfRegistGuildAdventure tests ---

func TestRegistGuildAdventure_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegistGuildAdventure{
		AckHandle:   100,
		Destination: 5,
	}

	handleMsgMhfRegistGuildAdventure(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestRegistGuildAdventure_Error(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{createAdvErr: errNotFound}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegistGuildAdventure{
		AckHandle:   100,
		Destination: 5,
	}

	// Should not panic; error is logged
	handleMsgMhfRegistGuildAdventure(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfAcquireGuildAdventure tests ---

func TestAcquireGuildAdventure_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireGuildAdventure{
		AckHandle: 100,
		ID:        42,
	}

	handleMsgMhfAcquireGuildAdventure(session, pkt)

	if guildMock.collectAdvID != 42 {
		t.Errorf("CollectAdventure ID = %d, want 42", guildMock.collectAdvID)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfChargeGuildAdventure tests ---

func TestChargeGuildAdventure_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfChargeGuildAdventure{
		AckHandle: 100,
		ID:        42,
		Amount:    500,
	}

	handleMsgMhfChargeGuildAdventure(session, pkt)

	if guildMock.chargeAdvID != 42 {
		t.Errorf("ChargeAdventure ID = %d, want 42", guildMock.chargeAdvID)
	}
	if guildMock.chargeAdvAmount != 500 {
		t.Errorf("ChargeAdventure Amount = %d, want 500", guildMock.chargeAdvAmount)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfRegistGuildAdventureDiva tests ---

func TestRegistGuildAdventureDiva_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegistGuildAdventureDiva{
		AckHandle:   100,
		Destination: 3,
		Charge:      200,
	}

	handleMsgMhfRegistGuildAdventureDiva(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
