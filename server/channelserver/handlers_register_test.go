package channelserver

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// createMockServerWithRaviente creates a mock server with raviente and semaphore
// initialized, which the base createMockServer() does not do.
func createMockServerWithRaviente() *Server {
	s := createMockServer()
	s.raviente = NewRaviente()
	s.semaphore = make(map[string]*Semaphore)
	return s
}

// --- NewRaviente ---

func TestNewRaviente_FullValidation(t *testing.T) {
	r := NewRaviente()
	if r == nil {
		t.Fatal("NewRaviente returned nil")
	}
	if r.register == nil {
		t.Fatal("register is nil")
	}
	if r.state == nil {
		t.Fatal("state is nil")
	}
	if r.support == nil {
		t.Fatal("support is nil")
	}
	if len(r.register.register) != 5 {
		t.Errorf("register length = %d, want 5", len(r.register.register))
	}
	if len(r.state.stateData) != 29 {
		t.Errorf("stateData length = %d, want 29", len(r.state.stateData))
	}
	if len(r.support.supportData) != 25 {
		t.Errorf("supportData length = %d, want 25", len(r.support.supportData))
	}
	// All values should be zero-initialized
	for i, v := range r.register.register {
		if v != 0 {
			t.Errorf("register[%d] = %d, want 0", i, v)
		}
	}
	for i, v := range r.state.stateData {
		if v != 0 {
			t.Errorf("stateData[%d] = %d, want 0", i, v)
		}
	}
	for i, v := range r.support.supportData {
		if v != 0 {
			t.Errorf("supportData[%d] = %d, want 0", i, v)
		}
	}
	if r.register.nextTime != 0 {
		t.Errorf("nextTime = %d, want 0", r.register.nextTime)
	}
	if r.register.startTime != 0 {
		t.Errorf("startTime = %d, want 0", r.register.startTime)
	}
	if r.register.killedTime != 0 {
		t.Errorf("killedTime = %d, want 0", r.register.killedTime)
	}
	if r.register.postTime != 0 {
		t.Errorf("postTime = %d, want 0", r.register.postTime)
	}
	if r.register.ravienteType != 0 {
		t.Errorf("ravienteType = %d, want 0", r.register.ravienteType)
	}
	if r.register.maxPlayers != 0 {
		t.Errorf("maxPlayers = %d, want 0", r.register.maxPlayers)
	}
	if r.register.carveQuest != 0 {
		t.Errorf("carveQuest = %d, want 0", r.register.carveQuest)
	}
}

// --- handleMsgSysLoadRegister ---

func TestHandleMsgSysLoadRegister_Case12(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.nextTime = 100
	server.raviente.register.startTime = 200
	server.raviente.register.killedTime = 300
	server.raviente.register.postTime = 400
	server.raviente.register.register[0] = 10
	server.raviente.register.register[1] = 20
	server.raviente.register.register[2] = 30
	server.raviente.register.register[3] = 40
	server.raviente.register.register[4] = 50
	server.raviente.register.carveQuest = 500
	server.raviente.register.maxPlayers = 32
	server.raviente.register.ravienteType = 2
	session := createMockSession(1, server)

	handleMsgSysLoadRegister(session, &mhfpacket.MsgSysLoadRegister{
		AckHandle: 1, RegisterID: 0, Unk1: 12,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysLoadRegister_Case29(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[0] = 111
	server.raviente.state.stateData[14] = 222
	server.raviente.state.stateData[28] = 333
	session := createMockSession(1, server)

	handleMsgSysLoadRegister(session, &mhfpacket.MsgSysLoadRegister{
		AckHandle: 2, RegisterID: 0, Unk1: 29,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysLoadRegister_Case25(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.support.supportData[0] = 777
	server.raviente.support.supportData[12] = 888
	server.raviente.support.supportData[24] = 999
	session := createMockSession(1, server)

	handleMsgSysLoadRegister(session, &mhfpacket.MsgSysLoadRegister{
		AckHandle: 3, RegisterID: 0, Unk1: 25,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysLoadRegister_UnknownCase(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	// Unk1=99 doesn't match any case, so no response should be sent
	handleMsgSysLoadRegister(session, &mhfpacket.MsgSysLoadRegister{
		AckHandle: 4, RegisterID: 0, Unk1: 99,
	})

	select {
	case <-session.sendPackets:
		t.Error("no response expected for unknown Unk1 value")
	default:
		// Expected: no packet queued
	}
}

// --- handleMsgSysOperateRegister ---

type opEntry struct {
	op   uint8
	dest uint8
	data uint32
}

func buildPayload(entries ...opEntry) []byte {
	bf := byteframe.NewByteFrame()
	for _, e := range entries {
		bf.WriteUint8(e.op)
		bf.WriteUint8(e.dest)
		bf.WriteUint32(e.data)
	}
	bf.WriteUint8(0) // terminator
	return bf.Data()
}

// --- SemaphoreID=4 (stateData) ---

func TestHandleMsgSysOperateRegister_State_Op2_Normal(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[0] = 100
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 0, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// With no ravi semaphore, GetRaviMultiplier returns 0, so data becomes 0
	// *ref += 0 => stateData[0] stays 100
	if server.raviente.state.stateData[0] != 100 {
		t.Errorf("stateData[0] = %d, want 100 (multiplier=0 makes data=0)", server.raviente.state.stateData[0])
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_Op2_Dest28(t *testing.T) {
	// dest=28 is the Berserk resurrection tracker, no multiplier applied
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[28] = 100
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 28, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// dest=28 adds data directly without multiplier
	if server.raviente.state.stateData[28] != 150 {
		t.Errorf("stateData[28] = %d, want 150", server.raviente.state.stateData[28])
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_Op2_Dest17_MultiplierIs1(t *testing.T) {
	// dest=17 is Berserk poison tracker, only adds when damageMultiplier==1
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[17] = 100
	server.raviente.register.maxPlayers = 4 // small ravi, minPlayers=4

	// Create a ravi semaphore with enough clients for multiplier=1
	sema := &Semaphore{
		id_semaphore: "hs_l0u3B51234_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// Need > 4 clients (minPlayers) for multiplier=1
	for i := 0; i < 5; i++ {
		s := createMockSession(uint32(100+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	session := createMockSession(1, server)
	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 17, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// multiplier=1, so dest=17 adds data
	if server.raviente.state.stateData[17] != 150 {
		t.Errorf("stateData[17] = %d, want 150", server.raviente.state.stateData[17])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_Op2_Dest17_MultiplierNot1(t *testing.T) {
	// dest=17 with multiplier != 1 should NOT add data
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[17] = 100
	server.raviente.register.maxPlayers = 4 // small ravi, minPlayers=4

	// Create a ravi semaphore with fewer clients than minPlayers for multiplier > 1
	sema := &Semaphore{
		id_semaphore: "hs_l0u3B51234_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// Need <= 4 clients so multiplier = 4/len(clients) != 1
	for i := 0; i < 2; i++ {
		s := createMockSession(uint32(100+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	session := createMockSession(1, server)
	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 17, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// multiplier=4/2=2 != 1, so dest=17 does NOT add data
	if server.raviente.state.stateData[17] != 100 {
		t.Errorf("stateData[17] = %d, want 100 (should not change)", server.raviente.state.stateData[17])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[5] = 999
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 5, data: 42}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.state.stateData[5] != 42 {
		t.Errorf("stateData[5] = %d, want 42", server.raviente.state.stateData[5])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[3] = 888
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 3, data: 77}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// op=13 falls through to op=14 behavior: sets value
	if server.raviente.state.stateData[3] != 77 {
		t.Errorf("stateData[3] = %d, want 77", server.raviente.state.stateData[3])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_MultipleEntries(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:   1,
		SemaphoreID: 4,
		RawDataPayload: buildPayload(
			opEntry{op: 14, dest: 0, data: 10},
			opEntry{op: 14, dest: 1, data: 20},
			opEntry{op: 14, dest: 2, data: 30},
		),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.state.stateData[0] != 10 {
		t.Errorf("stateData[0] = %d, want 10", server.raviente.state.stateData[0])
	}
	if server.raviente.state.stateData[1] != 20 {
		t.Errorf("stateData[1] = %d, want 20", server.raviente.state.stateData[1])
	}
	if server.raviente.state.stateData[2] != 30 {
		t.Errorf("stateData[2] = %d, want 30", server.raviente.state.stateData[2])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

// --- SemaphoreID=5 (supportData) ---

func TestHandleMsgSysOperateRegister_Support_Op2(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.support.supportData[0] = 100
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    5,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 0, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.support.supportData[0] != 150 {
		t.Errorf("supportData[0] = %d, want 150", server.raviente.support.supportData[0])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Support_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.support.supportData[10] = 999
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    5,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 10, data: 42}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.support.supportData[10] != 42 {
		t.Errorf("supportData[10] = %d, want 42", server.raviente.support.supportData[10])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Support_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.support.supportData[5] = 888
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    5,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 5, data: 77}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.support.supportData[5] != 77 {
		t.Errorf("supportData[5] = %d, want 77", server.raviente.support.supportData[5])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Support_MultipleEntries(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:   1,
		SemaphoreID: 5,
		RawDataPayload: buildPayload(
			opEntry{op: 2, dest: 0, data: 10},
			opEntry{op: 14, dest: 1, data: 20},
			opEntry{op: 13, dest: 2, data: 30},
		),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.support.supportData[0] != 10 {
		t.Errorf("supportData[0] = %d, want 10", server.raviente.support.supportData[0])
	}
	if server.raviente.support.supportData[1] != 20 {
		t.Errorf("supportData[1] = %d, want 20", server.raviente.support.supportData[1])
	}
	if server.raviente.support.supportData[2] != 30 {
		t.Errorf("supportData[2] = %d, want 30", server.raviente.support.supportData[2])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

// --- SemaphoreID=6 (register fields) ---

func TestHandleMsgSysOperateRegister_Register_Dest0_NextTime(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 0, data: 12345}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.nextTime != 12345 {
		t.Errorf("nextTime = %d, want 12345", server.raviente.register.nextTime)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest1_StartTime(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 1, data: 67890}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.startTime != 67890 {
		t.Errorf("startTime = %d, want 67890", server.raviente.register.startTime)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest2_KilledTime(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 2, data: 11111}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.killedTime != 11111 {
		t.Errorf("killedTime = %d, want 11111", server.raviente.register.killedTime)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest3_PostTime(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 3, data: 22222}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.postTime != 22222 {
		t.Errorf("postTime = %d, want 22222", server.raviente.register.postTime)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest4_Register0_Op2(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[0] = 100
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 4, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[0] != 150 {
		t.Errorf("register[0] = %d, want 150", server.raviente.register.register[0])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest4_Register0_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[0] = 999
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 4, data: 42}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[0] != 42 {
		t.Errorf("register[0] = %d, want 42", server.raviente.register.register[0])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest4_Register0_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[0] = 999
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 4, data: 77}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// op=14 for dest=4 writes response data but does NOT set *ref (unlike op=13)
	if server.raviente.register.register[0] != 999 {
		t.Errorf("register[0] = %d, want 999 (op=14 doesn't set ref)", server.raviente.register.register[0])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest5_CarveQuest(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 5, data: 33333}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.carveQuest != 33333 {
		t.Errorf("carveQuest = %d, want 33333", server.raviente.register.carveQuest)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest6_Register1_Op2(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[1] = 200
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 6, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[1] != 250 {
		t.Errorf("register[1] = %d, want 250", server.raviente.register.register[1])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest6_Register1_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 6, data: 55}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[1] != 55 {
		t.Errorf("register[1] = %d, want 55", server.raviente.register.register[1])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest6_Register1_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[1] = 999
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 6, data: 77}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// op=14 for register dests doesn't set *ref
	if server.raviente.register.register[1] != 999 {
		t.Errorf("register[1] = %d, want 999", server.raviente.register.register[1])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest7_Register2_Op2(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[2] = 300
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 7, data: 25}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[2] != 325 {
		t.Errorf("register[2] = %d, want 325", server.raviente.register.register[2])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest7_Register2_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 7, data: 66}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[2] != 66 {
		t.Errorf("register[2] = %d, want 66", server.raviente.register.register[2])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest7_Register2_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[2] = 500
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 7, data: 88}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[2] != 500 {
		t.Errorf("register[2] = %d, want 500", server.raviente.register.register[2])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest8_Register3_Op2(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[3] = 400
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 8, data: 10}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[3] != 410 {
		t.Errorf("register[3] = %d, want 410", server.raviente.register.register[3])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest8_Register3_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 8, data: 99}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[3] != 99 {
		t.Errorf("register[3] = %d, want 99", server.raviente.register.register[3])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest8_Register3_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[3] = 777
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 8, data: 11}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[3] != 777 {
		t.Errorf("register[3] = %d, want 777", server.raviente.register.register[3])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest9_MaxPlayers(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 9, data: 32}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.maxPlayers != 32 {
		t.Errorf("maxPlayers = %d, want 32", server.raviente.register.maxPlayers)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest10_RavienteType(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 10, data: 3}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.ravienteType != 3 {
		t.Errorf("ravienteType = %d, want 3", server.raviente.register.ravienteType)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest11_Register4_Op2(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[4] = 500
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 11, data: 100}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[4] != 600 {
		t.Errorf("register[4] = %d, want 600", server.raviente.register.register[4])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest11_Register4_Op13(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 13, dest: 11, data: 44}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[4] != 44 {
		t.Errorf("register[4] = %d, want 44", server.raviente.register.register[4])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_Dest11_Register4_Op14(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.register[4] = 888
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 11, data: 55}),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.register[4] != 888 {
		t.Errorf("register[4] = %d, want 888", server.raviente.register.register[4])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_DefaultDest(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	// dest=99 doesn't match any case, hits default branch
	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    6,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 99, data: 123}),
	}
	handleMsgSysOperateRegister(session, pkt)

	select {
	case <-session.sendPackets:
		// Default case writes zeros and sends response
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_Register_AllDests(t *testing.T) {
	// Exercise all dest cases in a single operation to test full branch coverage
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:   1,
		SemaphoreID: 6,
		RawDataPayload: buildPayload(
			opEntry{op: 14, dest: 0, data: 1},   // nextTime
			opEntry{op: 14, dest: 1, data: 2},   // startTime
			opEntry{op: 14, dest: 2, data: 3},   // killedTime
			opEntry{op: 14, dest: 3, data: 4},   // postTime
			opEntry{op: 2, dest: 4, data: 5},    // register[0] op2
			opEntry{op: 14, dest: 5, data: 6},   // carveQuest
			opEntry{op: 2, dest: 6, data: 7},    // register[1] op2
			opEntry{op: 2, dest: 7, data: 8},    // register[2] op2
			opEntry{op: 2, dest: 8, data: 9},    // register[3] op2
			opEntry{op: 14, dest: 9, data: 10},  // maxPlayers
			opEntry{op: 14, dest: 10, data: 11}, // ravienteType
			opEntry{op: 2, dest: 11, data: 12},  // register[4] op2
			opEntry{op: 14, dest: 99, data: 0},  // default
		),
	}
	handleMsgSysOperateRegister(session, pkt)

	if server.raviente.register.nextTime != 1 {
		t.Errorf("nextTime = %d, want 1", server.raviente.register.nextTime)
	}
	if server.raviente.register.startTime != 2 {
		t.Errorf("startTime = %d, want 2", server.raviente.register.startTime)
	}
	if server.raviente.register.killedTime != 3 {
		t.Errorf("killedTime = %d, want 3", server.raviente.register.killedTime)
	}
	if server.raviente.register.postTime != 4 {
		t.Errorf("postTime = %d, want 4", server.raviente.register.postTime)
	}
	if server.raviente.register.register[0] != 5 {
		t.Errorf("register[0] = %d, want 5", server.raviente.register.register[0])
	}
	if server.raviente.register.carveQuest != 6 {
		t.Errorf("carveQuest = %d, want 6", server.raviente.register.carveQuest)
	}
	if server.raviente.register.register[1] != 7 {
		t.Errorf("register[1] = %d, want 7", server.raviente.register.register[1])
	}
	if server.raviente.register.register[2] != 8 {
		t.Errorf("register[2] = %d, want 8", server.raviente.register.register[2])
	}
	if server.raviente.register.register[3] != 9 {
		t.Errorf("register[3] = %d, want 9", server.raviente.register.register[3])
	}
	if server.raviente.register.maxPlayers != 10 {
		t.Errorf("maxPlayers = %d, want 10", server.raviente.register.maxPlayers)
	}
	if server.raviente.register.ravienteType != 11 {
		t.Errorf("ravienteType = %d, want 11", server.raviente.register.ravienteType)
	}
	if server.raviente.register.register[4] != 12 {
		t.Errorf("register[4] = %d, want 12", server.raviente.register.register[4])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

// --- getRaviSemaphore ---

func TestGetRaviSemaphore_NoMatch(t *testing.T) {
	server := createMockServerWithRaviente()
	result := getRaviSemaphore(server)
	if result != nil {
		t.Error("should return nil when no semaphore matches")
	}
}

func TestGetRaviSemaphore_WrongPrefix(t *testing.T) {
	server := createMockServerWithRaviente()
	server.semaphore["wrong"] = &Semaphore{
		id_semaphore: "wrong_prefix_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	result := getRaviSemaphore(server)
	if result != nil {
		t.Error("should return nil when no semaphore has correct prefix")
	}
}

func TestGetRaviSemaphore_WrongSuffix(t *testing.T) {
	server := createMockServerWithRaviente()
	server.semaphore["wrong"] = &Semaphore{
		id_semaphore: "hs_l0u3B5test_4",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	result := getRaviSemaphore(server)
	if result != nil {
		t.Error("should return nil when semaphore has wrong suffix")
	}
}

func TestGetRaviSemaphore_Match(t *testing.T) {
	server := createMockServerWithRaviente()
	expected := &Semaphore{
		id_semaphore: "hs_l0u3B5some_data_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	server.semaphore["ravi"] = expected
	result := getRaviSemaphore(server)
	if result != expected {
		t.Error("should return matching semaphore")
	}
}

func TestGetRaviSemaphore_ExactMinimal(t *testing.T) {
	server := createMockServerWithRaviente()
	expected := &Semaphore{
		id_semaphore: "hs_l0u3B53",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	server.semaphore["ravi"] = expected
	result := getRaviSemaphore(server)
	if result != expected {
		t.Error("should match when prefix immediately followed by suffix '3'")
	}
}

func TestGetRaviSemaphore_MultipleOnlyOneMatches(t *testing.T) {
	server := createMockServerWithRaviente()
	server.semaphore["wrong1"] = &Semaphore{
		id_semaphore: "something_else",
		id:           8,
		clients:      make(map[*Session]uint32),
	}
	expected := &Semaphore{
		id_semaphore: "hs_l0u3B5ravi_3",
		id:           9,
		clients:      make(map[*Session]uint32),
	}
	server.semaphore["ravi"] = expected
	server.semaphore["wrong2"] = &Semaphore{
		id_semaphore: "hs_l0u3B5_no_suffix",
		id:           10,
		clients:      make(map[*Session]uint32),
	}
	result := getRaviSemaphore(server)
	if result != expected {
		t.Error("should return the one matching semaphore")
	}
}

// --- notifyRavi ---

func TestNotifyRavi_NoSemaphore(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	// Should not panic when no ravi semaphore exists
	session.notifyRavi()

	// No clients to receive notifications, so nothing should be queued
	select {
	case <-session.sendPackets:
		t.Error("no packet should be queued on the calling session")
	default:
		// Expected
	}
}

func TestNotifyRavi_WithSemaphoreAndClients(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	client1 := createMockSession(10, server)
	client2 := createMockSession(20, server)

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	sema.clients[client1] = client1.charID
	sema.clients[client2] = client2.charID
	server.semaphore["ravi"] = sema

	session.notifyRavi()

	// Both clients on the semaphore should receive notification packets
	receivedCount := 0
	select {
	case p := <-client1.sendPackets:
		if len(p.data) > 0 {
			receivedCount++
		}
	default:
		t.Error("client1 should have received notification")
	}
	select {
	case p := <-client2.sendPackets:
		if len(p.data) > 0 {
			receivedCount++
		}
	default:
		t.Error("client2 should have received notification")
	}
	if receivedCount != 2 {
		t.Errorf("received %d notifications, want 2", receivedCount)
	}
}

// --- resetRavi ---

func TestResetRavi(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	// Set various values
	server.raviente.register.nextTime = 12345
	server.raviente.register.startTime = 67890
	server.raviente.register.killedTime = 11111
	server.raviente.register.postTime = 22222
	server.raviente.register.ravienteType = 3
	server.raviente.register.maxPlayers = 32
	server.raviente.register.carveQuest = 44444
	server.raviente.register.register[0] = 100
	server.raviente.register.register[1] = 200
	server.raviente.register.register[2] = 300
	server.raviente.register.register[3] = 400
	server.raviente.register.register[4] = 500
	server.raviente.state.stateData[0] = 999
	server.raviente.state.stateData[14] = 888
	server.raviente.state.stateData[28] = 777
	server.raviente.support.supportData[0] = 666
	server.raviente.support.supportData[12] = 555
	server.raviente.support.supportData[24] = 444

	resetRavi(session)

	// Verify all register fields reset
	if server.raviente.register.nextTime != 0 {
		t.Errorf("nextTime = %d, want 0", server.raviente.register.nextTime)
	}
	if server.raviente.register.startTime != 0 {
		t.Errorf("startTime = %d, want 0", server.raviente.register.startTime)
	}
	if server.raviente.register.killedTime != 0 {
		t.Errorf("killedTime = %d, want 0", server.raviente.register.killedTime)
	}
	if server.raviente.register.postTime != 0 {
		t.Errorf("postTime = %d, want 0", server.raviente.register.postTime)
	}
	if server.raviente.register.ravienteType != 0 {
		t.Errorf("ravienteType = %d, want 0", server.raviente.register.ravienteType)
	}
	if server.raviente.register.maxPlayers != 0 {
		t.Errorf("maxPlayers = %d, want 0", server.raviente.register.maxPlayers)
	}
	if server.raviente.register.carveQuest != 0 {
		t.Errorf("carveQuest = %d, want 0", server.raviente.register.carveQuest)
	}

	// Verify register array reset
	for i, v := range server.raviente.register.register {
		if v != 0 {
			t.Errorf("register[%d] = %d, want 0", i, v)
		}
	}
	if len(server.raviente.register.register) != 5 {
		t.Errorf("register length = %d, want 5", len(server.raviente.register.register))
	}

	// Verify stateData reset
	for i, v := range server.raviente.state.stateData {
		if v != 0 {
			t.Errorf("stateData[%d] = %d, want 0", i, v)
		}
	}
	if len(server.raviente.state.stateData) != 29 {
		t.Errorf("stateData length = %d, want 29", len(server.raviente.state.stateData))
	}

	// Verify supportData reset
	for i, v := range server.raviente.support.supportData {
		if v != 0 {
			t.Errorf("supportData[%d] = %d, want 0", i, v)
		}
	}
	if len(server.raviente.support.supportData) != 25 {
		t.Errorf("supportData length = %d, want 25", len(server.raviente.support.supportData))
	}
}

// --- GetRaviMultiplier ---

func TestGetRaviMultiplier_NoSemaphore(t *testing.T) {
	server := createMockServerWithRaviente()
	result := server.raviente.GetRaviMultiplier(server)
	if result != 0 {
		t.Errorf("expected 0, got %f", result)
	}
}

func TestGetRaviMultiplier_LargeRavi_EnoughPlayers(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.maxPlayers = 32 // > 8, so minPlayers=24

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// Need > 24 clients
	for i := 0; i < 25; i++ {
		s := createMockSession(uint32(100+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	result := server.raviente.GetRaviMultiplier(server)
	if result != 1 {
		t.Errorf("expected 1, got %f", result)
	}
}

func TestGetRaviMultiplier_LargeRavi_NotEnoughPlayers(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.maxPlayers = 32 // > 8, so minPlayers=24

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// 12 clients < 24 minPlayers => multiplier = 24/12 = 2
	for i := 0; i < 12; i++ {
		s := createMockSession(uint32(100+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	result := server.raviente.GetRaviMultiplier(server)
	expected := float64(24 / 12) // integer division: 2
	if result != expected {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestGetRaviMultiplier_SmallRavi_EnoughPlayers(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.maxPlayers = 4 // <= 8, so minPlayers=4

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// Need > 4 clients
	for i := 0; i < 5; i++ {
		s := createMockSession(uint32(100+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	result := server.raviente.GetRaviMultiplier(server)
	if result != 1 {
		t.Errorf("expected 1, got %f", result)
	}
}

func TestGetRaviMultiplier_SmallRavi_NotEnoughPlayers(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.register.maxPlayers = 4 // <= 8, so minPlayers=4

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// 2 clients < 4 minPlayers => multiplier = 4/2 = 2
	for i := 0; i < 2; i++ {
		s := createMockSession(uint32(100+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	result := server.raviente.GetRaviMultiplier(server)
	expected := float64(4 / 2) // integer division: 2
	if result != expected {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

// --- handleMsgSysNotifyRegister (empty handler) ---

func TestHandleMsgSysNotifyRegister(t *testing.T) {
	server := createMockServerWithRaviente()
	session := createMockSession(1, server)

	// Should not panic - handler is empty
	handleMsgSysNotifyRegister(session, &mhfpacket.MsgSysNotifyRegister{
		RegisterID: 4,
	})

	// No response expected from empty handler
	select {
	case <-session.sendPackets:
		t.Error("empty handler should not queue packets")
	default:
		// Expected
	}
}

// --- State op2 with multiplier applied (normal dest, not 17 or 28) ---

func TestHandleMsgSysOperateRegister_State_Op2_WithMultiplier(t *testing.T) {
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[5] = 100
	server.raviente.register.maxPlayers = 32 // large ravi, minPlayers=24

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	// 12 clients < 24 minPlayers => multiplier = 24/12 = 2
	for i := 0; i < 12; i++ {
		s := createMockSession(uint32(200+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	session := createMockSession(1, server)
	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 5, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// data = uint32(float64(50) * 2.0) = 100
	// stateData[5] = 100 + 100 = 200
	if server.raviente.state.stateData[5] != 200 {
		t.Errorf("stateData[5] = %d, want 200", server.raviente.state.stateData[5])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgSysOperateRegister_State_Op2_Dest28_WithMultiplier(t *testing.T) {
	// dest=28 should ignore multiplier regardless
	server := createMockServerWithRaviente()
	server.raviente.state.stateData[28] = 100
	server.raviente.register.maxPlayers = 32

	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	for i := 0; i < 12; i++ {
		s := createMockSession(uint32(200+i), server)
		sema.clients[s] = s.charID
	}
	server.semaphore["ravi"] = sema

	session := createMockSession(1, server)
	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    4,
		RawDataPayload: buildPayload(opEntry{op: 2, dest: 28, data: 50}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// dest=28 always adds raw data, ignoring multiplier
	if server.raviente.state.stateData[28] != 150 {
		t.Errorf("stateData[28] = %d, want 150", server.raviente.state.stateData[28])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("no response queued")
	}
}

// Test that notifyRavi is called as part of handleMsgSysOperateRegister
// by verifying that clients on the ravi semaphore get notifications.
func TestHandleMsgSysOperateRegister_NotifiesRaviClients(t *testing.T) {
	server := createMockServerWithRaviente()

	raviClient := createMockSession(50, server)
	sema := &Semaphore{
		id_semaphore: "hs_l0u3B5test_3",
		id:           7,
		clients:      make(map[*Session]uint32),
	}
	sema.clients[raviClient] = raviClient.charID
	server.semaphore["ravi"] = sema

	session := createMockSession(1, server)
	pkt := &mhfpacket.MsgSysOperateRegister{
		AckHandle:      1,
		SemaphoreID:    5,
		RawDataPayload: buildPayload(opEntry{op: 14, dest: 0, data: 1}),
	}
	handleMsgSysOperateRegister(session, pkt)

	// The calling session should receive the ack response
	select {
	case <-session.sendPackets:
	default:
		t.Error("calling session should receive ack response")
	}

	// The ravi client should receive a notification
	select {
	case p := <-raviClient.sendPackets:
		if len(p.data) == 0 {
			t.Error("ravi client should receive non-empty notification")
		}
	default:
		t.Error("ravi client should receive notification")
	}
}
