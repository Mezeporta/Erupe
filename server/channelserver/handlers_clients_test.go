package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgSysEnumerateClient_StageNotExists(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysEnumerateClient{
		AckHandle: 12345,
		StageID:   "nonexistent_stage",
		Get:       0,
	}

	handleMsgSysEnumerateClient(session, pkt)

	// Verify response packet was queued (failure expected)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysEnumerateClient_AllClients(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create stage with clients
	stage := NewStage("test_stage")
	server.stages["test_stage"] = stage

	client1 := createMockSession(100, server)
	client2 := createMockSession(200, server)
	stage.clients[client1] = client1.charID
	stage.clients[client2] = client2.charID

	pkt := &mhfpacket.MsgSysEnumerateClient{
		AckHandle: 12345,
		StageID:   "test_stage",
		Get:       0, // All clients
	}

	handleMsgSysEnumerateClient(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysEnumerateClient_NotReady(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create stage with reserved slots
	stage := NewStage("test_stage")
	server.stages["test_stage"] = stage

	stage.reservedClientSlots[100] = false // Not ready
	stage.reservedClientSlots[200] = true  // Ready

	pkt := &mhfpacket.MsgSysEnumerateClient{
		AckHandle: 12345,
		StageID:   "test_stage",
		Get:       1, // Not ready
	}

	handleMsgSysEnumerateClient(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysEnumerateClient_Ready(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Create stage with reserved slots
	stage := NewStage("test_stage")
	server.stages["test_stage"] = stage

	stage.reservedClientSlots[100] = false // Not ready
	stage.reservedClientSlots[200] = true  // Ready

	pkt := &mhfpacket.MsgSysEnumerateClient{
		AckHandle: 12345,
		StageID:   "test_stage",
		Get:       2, // Ready
	}

	handleMsgSysEnumerateClient(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfShutClient(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfShutClient panicked: %v", r)
		}
	}()

	handleMsgMhfShutClient(session, nil)
}

func TestHandleMsgSysHideClient(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysHideClient panicked: %v", r)
		}
	}()

	handleMsgSysHideClient(session, nil)
}
