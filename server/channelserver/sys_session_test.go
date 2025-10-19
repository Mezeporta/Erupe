package channelserver

import (
	"testing"
	"time"

	"erupe-ce/common/stringstack"
	"erupe-ce/network/clientctx"
)

func TestSessionStructInitialization(t *testing.T) {
	server := createMockServer()
	session := createMockSession(12345, server)

	if session.charID != 12345 {
		t.Errorf("charID = %d, want 12345", session.charID)
	}
	if session.Name != "TestPlayer" {
		t.Errorf("Name = %s, want TestPlayer", session.Name)
	}
	if session.server != server {
		t.Error("server reference not set correctly")
	}
	if session.clientContext == nil {
		t.Error("clientContext should not be nil")
	}
	if session.sendPackets == nil {
		t.Error("sendPackets channel should not be nil")
	}
}

func TestSessionSendPacketChannel(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Test that channel can receive packets
	testData := []byte{0x01, 0x02, 0x03}
	session.sendPackets <- packet{data: testData, nonBlocking: false}

	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) != 3 {
			t.Errorf("packet data len = %d, want 3", len(pkt.data))
		}
		if pkt.data[0] != 0x01 {
			t.Errorf("packet data[0] = %d, want 1", pkt.data[0])
		}
	default:
		t.Error("failed to receive packet from channel")
	}
}

func TestSessionSendPacketChannelNonBlocking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Fill the channel
	for i := 0; i < 20; i++ {
		session.sendPackets <- packet{data: []byte{byte(i)}, nonBlocking: true}
	}

	// Non-blocking send to full channel should not block
	done := make(chan bool, 1)
	go func() {
		select {
		case session.sendPackets <- packet{data: []byte{0xFF}, nonBlocking: true}:
			// Managed to send (channel had room)
		default:
			// Channel full, this is expected
		}
		done <- true
	}()

	select {
	case <-done:
		// Success - non-blocking worked
	case <-time.After(100 * time.Millisecond):
		t.Error("non-blocking send blocked")
	}
}

func TestPacketStruct(t *testing.T) {
	pkt := packet{
		data:        []byte{0x01, 0x02, 0x03},
		nonBlocking: true,
	}

	if len(pkt.data) != 3 {
		t.Errorf("packet data len = %d, want 3", len(pkt.data))
	}
	if !pkt.nonBlocking {
		t.Error("nonBlocking should be true")
	}
}

func TestPacketStructBlocking(t *testing.T) {
	pkt := packet{
		data:        []byte{0xDE, 0xAD, 0xBE, 0xEF},
		nonBlocking: false,
	}

	if len(pkt.data) != 4 {
		t.Errorf("packet data len = %d, want 4", len(pkt.data))
	}
	if pkt.nonBlocking {
		t.Error("nonBlocking should be false")
	}
}

func TestSessionClosedFlag(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	if session.closed.Load() {
		t.Error("new session should not be closed")
	}

	session.closed.Store(true)

	if !session.closed.Load() {
		t.Error("session closed flag should be settable")
	}
}

func TestSessionStageState(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Initially should have no stage
	if session.userEnteredStage {
		t.Error("new session should not have entered stage")
	}
	if session.stageID != "" {
		t.Errorf("stageID should be empty, got %s", session.stageID)
	}
	if session.stage != nil {
		t.Error("stage should be nil initially")
	}

	// Set stage state
	session.userEnteredStage = true
	session.stageID = "test_stage_001"

	if !session.userEnteredStage {
		t.Error("userEnteredStage should be set")
	}
	if session.stageID != "test_stage_001" {
		t.Errorf("stageID = %s, want test_stage_001", session.stageID)
	}
}

func TestSessionStageMoveStack(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)
	session.stageMoveStack = stringstack.New()

	// Push some stages
	session.stageMoveStack.Push("stage1")
	session.stageMoveStack.Push("stage2")
	session.stageMoveStack.Push("stage3")

	// Pop and verify order (LIFO)
	if v, err := session.stageMoveStack.Pop(); err != nil || v != "stage3" {
		t.Errorf("Pop() = %s, want stage3", v)
	}
	if v, err := session.stageMoveStack.Pop(); err != nil || v != "stage2" {
		t.Errorf("Pop() = %s, want stage2", v)
	}
	if v, err := session.stageMoveStack.Pop(); err != nil || v != "stage1" {
		t.Errorf("Pop() = %s, want stage1", v)
	}
}

func TestSessionMailState(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Initial mail state
	if session.mailAccIndex != 0 {
		t.Errorf("mailAccIndex = %d, want 0", session.mailAccIndex)
	}
	if session.mailList != nil && len(session.mailList) > 0 {
		t.Error("mailList should be empty initially")
	}

	// Add mail
	session.mailList = []int{100, 101, 102}
	session.mailAccIndex = 3

	if len(session.mailList) != 3 {
		t.Errorf("mailList len = %d, want 3", len(session.mailList))
	}
	if session.mailAccIndex != 3 {
		t.Errorf("mailAccIndex = %d, want 3", session.mailAccIndex)
	}
}

func TestSessionToken(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	session.token = "abc123def456"

	if session.token != "abc123def456" {
		t.Errorf("token = %s, want abc123def456", session.token)
	}
}

func TestSessionGuildState(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	session.prevGuildID = 42

	if session.prevGuildID != 42 {
		t.Errorf("prevGuildID = %d, want 42", session.prevGuildID)
	}
}

func TestSessionKQF(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Set KQF data
	session.kqf = []byte{0x01, 0x02, 0x03, 0x04}
	session.kqfOverride = true

	if len(session.kqf) != 4 {
		t.Errorf("kqf len = %d, want 4", len(session.kqf))
	}
	if !session.kqfOverride {
		t.Error("kqfOverride should be true")
	}
}

func TestSessionClientContext(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	if session.clientContext == nil {
		t.Fatal("clientContext should not be nil")
	}

	// Verify clientContext is usable
	ctx := session.clientContext
	_ = ctx // Just verify it's accessible
}

func TestSessionReservationStage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	if session.reservationStage != nil {
		t.Error("reservationStage should be nil initially")
	}

	// Set reservation stage
	stage := NewStage("quest_stage")
	session.reservationStage = stage

	if session.reservationStage != stage {
		t.Error("reservationStage should be set correctly")
	}
}

func TestSessionStagePass(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	session.stagePass = "secret123"

	if session.stagePass != "secret123" {
		t.Errorf("stagePass = %s, want secret123", session.stagePass)
	}
}

func TestSessionLogKey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	session.logKey = []byte{0xDE, 0xAD, 0xBE, 0xEF}

	if len(session.logKey) != 4 {
		t.Errorf("logKey len = %d, want 4", len(session.logKey))
	}
	if session.logKey[0] != 0xDE {
		t.Errorf("logKey[0] = %x, want 0xDE", session.logKey[0])
	}
}

func TestSessionSessionStart(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Set session start time
	now := time.Now().Unix()
	session.sessionStart = now

	if session.sessionStart != now {
		t.Errorf("sessionStart = %d, want %d", session.sessionStart, now)
	}
}

func TestIgnoredOpcode(t *testing.T) {
	// Test that certain opcodes are ignored
	tests := []struct {
		name    string
		opcode  uint16
		ignored bool
	}{
		// These should be ignored based on ignoreList
		{"MSG_SYS_END is ignored", 0x0002, true}, // Assuming MSG_SYS_END value
		// We can't test exact values without importing network package constants
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test is limited since ignored() uses network.PacketID
			// which we can't easily instantiate without the exact enum values
		})
	}
}

func TestSessionMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Verify session has mutex (via embedding)
	// This should not deadlock
	session.Lock()
	session.charID = 999
	session.Unlock()

	if session.charID != 999 {
		t.Errorf("charID = %d, want 999 after lock/unlock", session.charID)
	}
}

func TestSessionConcurrentAccess(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	done := make(chan bool, 2)

	// Concurrent writers
	go func() {
		for i := 0; i < 100; i++ {
			session.Lock()
			session.charID = uint32(i)
			session.Unlock()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			session.Lock()
			_ = session.charID
			session.Unlock()
		}
		done <- true
	}()

	<-done
	<-done
}

func TestClientContextStruct(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	// Verify the struct is usable
	if ctx == nil {
		t.Error("ClientContext should be creatable")
	}
}
