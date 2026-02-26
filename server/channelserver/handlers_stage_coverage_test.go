package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgSysReserveStage_NewSlot(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: make(map[uint32]bool),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
		maxPlayers:          4,
	}
	server.stages.Store("test_stage", stage)

	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "test_stage", Ready: 1}
	handleMsgSysReserveStage(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("expected response")
	}

	if _, exists := stage.reservedClientSlots[100]; !exists {
		t.Error("charID should be in reserved slots")
	}
}

func TestHandleMsgSysReserveStage_AlreadyReservedReady1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: map[uint32]bool{100: true},
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
		maxPlayers:          4,
	}
	server.stages.Store("test_stage", stage)

	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "test_stage", Ready: 1}
	handleMsgSysReserveStage(session, pkt)
	<-session.sendPackets

	if stage.reservedClientSlots[100] != false {
		t.Error("ready=1 should set slot to false")
	}
}

func TestHandleMsgSysReserveStage_AlreadyReservedReady17(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: map[uint32]bool{100: false},
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
		maxPlayers:          4,
	}
	server.stages.Store("test_stage", stage)

	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "test_stage", Ready: 17}
	handleMsgSysReserveStage(session, pkt)
	<-session.sendPackets

	if stage.reservedClientSlots[100] != true {
		t.Error("ready=17 should set slot to true")
	}
}

func TestHandleMsgSysReserveStage_Locked(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: make(map[uint32]bool),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
		maxPlayers:          4,
		locked:              true,
	}
	server.stages.Store("test_stage", stage)

	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "test_stage", Ready: 1}
	handleMsgSysReserveStage(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgSysReserveStage_PasswordMismatch(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: make(map[uint32]bool),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
		maxPlayers:          4,
		password:            "secret",
	}
	server.stages.Store("test_stage", stage)

	session.stagePass = "wrong"
	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "test_stage", Ready: 1}
	handleMsgSysReserveStage(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgSysReserveStage_Full(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: map[uint32]bool{200: false, 300: false},
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
		maxPlayers:          2,
	}
	server.stages.Store("test_stage", stage)

	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "test_stage", Ready: 1}
	handleMsgSysReserveStage(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgSysReserveStage_StageNotFound(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgSysReserveStage{AckHandle: 1, StageID: "nonexistent", Ready: 1}
	handleMsgSysReserveStage(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgSysUnreserveStage_WithReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: map[uint32]bool{100: false},
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
	}
	session.reservationStage = stage

	pkt := &mhfpacket.MsgSysUnreserveStage{}
	handleMsgSysUnreserveStage(session, pkt)

	if session.reservationStage != nil {
		t.Error("reservation should be cleared")
	}
	if _, exists := stage.reservedClientSlots[100]; exists {
		t.Error("charID should be removed from reserved slots")
	}
}

func TestHandleMsgSysUnreserveStage_NoReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgSysUnreserveStage{}
	handleMsgSysUnreserveStage(session, pkt)
	// Should not panic
}

func TestHandleMsgSysSetStagePass_Host(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: map[uint32]bool{100: false},
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
	}
	session.reservationStage = stage

	pkt := &mhfpacket.MsgSysSetStagePass{Password: "mypass"}
	handleMsgSysSetStagePass(session, pkt)

	if stage.password != "mypass" {
		t.Errorf("stage password = %q, want %q", stage.password, "mypass")
	}
}

func TestHandleMsgSysSetStagePass_NonHost(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgSysSetStagePass{Password: "mypass"}
	handleMsgSysSetStagePass(session, pkt)

	if session.stagePass != "mypass" {
		t.Errorf("session stagePass = %q, want %q", session.stagePass, "mypass")
	}
}

func TestHandleMsgSysSetAndGetStageBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: make(map[uint32]bool),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
	}
	server.stages.Store("test_stage", stage)

	// Set binary
	setPkt := &mhfpacket.MsgSysSetStageBinary{
		BinaryType0:    1,
		BinaryType1:    2,
		StageID:        "test_stage",
		RawDataPayload: []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}
	handleMsgSysSetStageBinary(session, setPkt)

	// Get binary
	getPkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   1,
		BinaryType0: 1,
		BinaryType1: 2,
		StageID:     "test_stage",
	}
	handleMsgSysGetStageBinary(session, getPkt)
	<-session.sendPackets
}

func TestHandleMsgSysGetStageBinary_Type1Equals4Fallback(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: make(map[uint32]bool),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
	}
	server.stages.Store("test_stage", stage)

	getPkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   1,
		BinaryType0: 0,
		BinaryType1: 4,
		StageID:     "test_stage",
	}
	handleMsgSysGetStageBinary(session, getPkt)
	<-session.sendPackets
}

func TestHandleMsgSysGetStageBinary_MissingBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: make(map[uint32]bool),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
	}
	server.stages.Store("test_stage", stage)

	getPkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   1,
		BinaryType0: 9,
		BinaryType1: 9,
		StageID:     "test_stage",
	}
	handleMsgSysGetStageBinary(session, getPkt)
	<-session.sendPackets
}

func TestHandleMsgSysGetStageBinary_MissingStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	getPkt := &mhfpacket.MsgSysGetStageBinary{
		AckHandle:   1,
		BinaryType0: 0,
		BinaryType1: 0,
		StageID:     "nonexistent",
	}
	handleMsgSysGetStageBinary(session, getPkt)
	<-session.sendPackets
}

func TestHandleMsgSysSetStageBinary_MissingStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgSysSetStageBinary{
		BinaryType0:    1,
		BinaryType1:    2,
		StageID:        "nonexistent",
		RawDataPayload: []byte{1, 2, 3},
	}
	handleMsgSysSetStageBinary(session, pkt)
	// Should not panic, just logs warning
}

func TestHandleMsgSysUnlockStage_WithReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	stage := &Stage{
		id:                  "test_stage",
		reservedClientSlots: map[uint32]bool{100: false},
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		clients:             make(map[*Session]uint32),
	}
	server.stages.Store("test_stage", stage)
	session.reservationStage = stage

	pkt := &mhfpacket.MsgSysUnlockStage{}
	handleMsgSysUnlockStage(session, pkt)

	if _, exists := server.stages.Get("test_stage"); exists {
		t.Error("stage should have been deleted")
	}
}

func TestHandleMsgSysUnlockStage_NoReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgSysUnlockStage{}
	handleMsgSysUnlockStage(session, pkt)
	// Should not panic
}
