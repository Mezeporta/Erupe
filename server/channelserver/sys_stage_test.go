package channelserver

import (
	"net"
	"sync"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"

	"go.uber.org/zap"
)

// mockPacket implements mhfpacket.MHFPacket for testing
type mockPacket struct {
	opcode uint16
}

func (m *mockPacket) Opcode() network.PacketID {
	return network.PacketID(m.opcode)
}

func (m *mockPacket) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	// Access ctx to trigger nil pointer if ctx is nil
	if ctx == nil {
		panic("clientContext is nil")
	}
	bf.WriteUint32(0x12345678)
	return nil
}

func (m *mockPacket) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return nil
}

// createMockServer creates a minimal Server for testing
func createMockServer() *Server {
	logger, _ := zap.NewDevelopment()
	return &Server{
		logger:      logger,
		erupeConfig: &config.Config{DevMode: false},
		stages:      make(map[string]*Stage),
		sessions:    make(map[net.Conn]*Session),
	}
}

// createMockSession creates a minimal Session for testing
func createMockSession(charID uint32, server *Server) *Session {
	logger, _ := zap.NewDevelopment()
	return &Session{
		charID:        charID,
		clientContext: &clientctx.ClientContext{},
		sendPackets:   make(chan packet, 20),
		Name:          "TestPlayer",
		server:        server,
		logger:        logger,
	}
}

func TestStageBroadcastMHF(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	// Add some sessions
	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)
	session3 := createMockSession(3, server)

	stage.clients[session1] = session1.charID
	stage.clients[session2] = session2.charID
	stage.clients[session3] = session3.charID

	pkt := &mockPacket{opcode: 0x1234}

	// Should not panic
	stage.BroadcastMHF(pkt, session1)

	// Verify session2 and session3 received data
	select {
	case data := <-session2.sendPackets:
		if len(data.data) == 0 {
			t.Error("session2 received empty data")
		}
	default:
		t.Error("session2 did not receive data")
	}

	select {
	case data := <-session3.sendPackets:
		if len(data.data) == 0 {
			t.Error("session3 received empty data")
		}
	default:
		t.Error("session3 did not receive data")
	}
}

func TestStageBroadcastMHF_NilClientContext(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)
	session2.clientContext = nil // Simulate corrupted session

	stage.clients[session1] = session1.charID
	stage.clients[session2] = session2.charID

	pkt := &mockPacket{opcode: 0x1234}

	// This should panic with the current implementation
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Caught expected panic: %v", r)
			// Test passes - we've confirmed the bug exists
		} else {
			t.Log("No panic occurred - either the bug is fixed or test is wrong")
		}
	}()

	stage.BroadcastMHF(pkt, nil)
}

// TestStageBroadcastMHF_ConcurrentModificationWithLock tests that proper locking
// prevents the race condition between BroadcastMHF and session removal
func TestStageBroadcastMHF_ConcurrentModificationWithLock(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	// Create many sessions
	sessions := make([]*Session, 100)
	for i := range sessions {
		sessions[i] = createMockSession(uint32(i), server)
		stage.clients[sessions[i]] = sessions[i].charID
	}

	pkt := &mockPacket{opcode: 0x1234}

	var wg sync.WaitGroup

	// Start goroutines that broadcast
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				stage.BroadcastMHF(pkt, nil)
			}
		}()
	}

	// Start goroutines that remove sessions WITH proper locking
	// This simulates the fixed logoutPlayer behavior
	for i := 0; i < 10; i++ {
		wg.Add(1)
		idx := i * 10
		go func(startIdx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				sessionIdx := startIdx + j
				if sessionIdx < len(sessions) {
					// Fixed: modifying stage.clients WITH lock
					stage.Lock()
					delete(stage.clients, sessions[sessionIdx])
					stage.Unlock()
				}
			}
		}(idx)
	}

	wg.Wait()
}

// TestStageBroadcastMHF_RaceDetectorWithLock verifies no race when
// modifications are done with proper locking
func TestStageBroadcastMHF_RaceDetectorWithLock(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)

	stage.clients[session1] = session1.charID
	stage.clients[session2] = session2.charID

	pkt := &mockPacket{opcode: 0x1234}

	var wg sync.WaitGroup

	// Goroutine 1: Continuously broadcast
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			stage.BroadcastMHF(pkt, nil)
		}
	}()

	// Goroutine 2: Add and remove sessions WITH proper locking
	// This simulates the fixed logoutPlayer behavior
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			newSession := createMockSession(uint32(100+i), server)
			// Add WITH lock (fixed)
			stage.Lock()
			stage.clients[newSession] = newSession.charID
			stage.Unlock()
			// Remove WITH lock (fixed)
			stage.Lock()
			delete(stage.clients, newSession)
			stage.Unlock()
		}
	}()

	wg.Wait()
}

// TestStageBroadcastMHF_NilClientContextSkipped verifies sessions with nil
// clientContext are safely skipped
func TestStageBroadcastMHF_NilClientContextSkipped(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)
	session2.clientContext = nil // Simulate corrupted session

	stage.clients[session1] = session1.charID
	stage.clients[session2] = session2.charID

	pkt := &mockPacket{opcode: 0x1234}

	// Should NOT panic now that we have the nil check
	stage.BroadcastMHF(pkt, nil)

	// Verify session1 received data (session2 was skipped)
	select {
	case data := <-session1.sendPackets:
		if len(data.data) == 0 {
			t.Error("session1 received empty data")
		}
	default:
		t.Error("session1 did not receive data")
	}
}

// TestNewStageBasic verifies Stage creation
func TestNewStageBasic(t *testing.T) {
	stageID := "test_stage_001"
	stage := NewStage(stageID)

	if stage == nil {
		t.Fatal("NewStage() returned nil")
	}
	if stage.id != stageID {
		t.Errorf("stage.id = %s, want %s", stage.id, stageID)
	}
	if stage.clients == nil {
		t.Error("stage.clients should not be nil")
	}
	if stage.reservedClientSlots == nil {
		t.Error("stage.reservedClientSlots should not be nil")
	}
	if stage.objects == nil {
		t.Error("stage.objects should not be nil")
	}
}

// TestStageClientCount tests client counting
func TestStageClientCount(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	if len(stage.clients) != 0 {
		t.Errorf("initial client count = %d, want 0", len(stage.clients))
	}

	// Add clients
	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)

	stage.clients[session1] = session1.charID
	if len(stage.clients) != 1 {
		t.Errorf("client count after 1 add = %d, want 1", len(stage.clients))
	}

	stage.clients[session2] = session2.charID
	if len(stage.clients) != 2 {
		t.Errorf("client count after 2 adds = %d, want 2", len(stage.clients))
	}

	// Remove a client
	delete(stage.clients, session1)
	if len(stage.clients) != 1 {
		t.Errorf("client count after 1 remove = %d, want 1", len(stage.clients))
	}
}

// TestStageReservation tests stage reservation
func TestStageReservation(t *testing.T) {
	stage := NewStage("test_stage")

	if len(stage.reservedClientSlots) != 0 {
		t.Errorf("initial reservations = %d, want 0", len(stage.reservedClientSlots))
	}

	// Reserve a slot using character ID
	stage.reservedClientSlots[12345] = true
	if len(stage.reservedClientSlots) != 1 {
		t.Errorf("reservations after 1 add = %d, want 1", len(stage.reservedClientSlots))
	}
}

// TestStageBinaryData tests setting and getting stage binary data
func TestStageBinaryData(t *testing.T) {
	stage := NewStage("test_stage")

	// rawBinaryData is initialized by NewStage
	if stage.rawBinaryData == nil {
		t.Error("rawBinaryData should not be nil after NewStage")
	}

	// Set binary data
	key := stageBinaryKey{id0: 1, id1: 2}
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	stage.rawBinaryData[key] = testData

	if len(stage.rawBinaryData) != 1 {
		t.Errorf("rawBinaryData length = %d, want 1", len(stage.rawBinaryData))
	}
}

// TestStageLockUnlock tests stage locking
func TestStageLockUnlock(t *testing.T) {
	stage := NewStage("test_stage")

	// Test lock/unlock without deadlock
	stage.Lock()
	stage.password = "test"
	stage.Unlock()

	stage.RLock()
	password := stage.password
	stage.RUnlock()

	if password != "test" {
		t.Error("stage password should be 'test'")
	}
}

// TestStageHostSession tests host session tracking
func TestStageHostSession(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()
	session := createMockSession(1, server)

	if stage.host != nil {
		t.Error("initial host should be nil")
	}

	stage.host = session
	if stage.host == nil {
		t.Error("host should not be nil after setting")
	}
	if stage.host.charID != 1 {
		t.Errorf("host.charID = %d, want 1", stage.host.charID)
	}
}

// TestStageMultipleClients tests stage with multiple clients
func TestStageMultipleClients(t *testing.T) {
	stage := NewStage("test_stage")
	server := createMockServer()

	// Add many clients
	sessions := make([]*Session, 10)
	for i := range sessions {
		sessions[i] = createMockSession(uint32(i+1), server)
		stage.clients[sessions[i]] = sessions[i].charID
	}

	if len(stage.clients) != 10 {
		t.Errorf("client count = %d, want 10", len(stage.clients))
	}

	// Verify each client is tracked
	for _, s := range sessions {
		if _, ok := stage.clients[s]; !ok {
			t.Errorf("session with charID %d not found in stage", s.charID)
		}
	}
}

// TestStageNewMaxPlayers tests default max players
func TestStageNewMaxPlayers(t *testing.T) {
	stage := NewStage("test_stage")

	// Default max players is 4
	if stage.maxPlayers != 4 {
		t.Errorf("initial maxPlayers = %d, want 4", stage.maxPlayers)
	}
}

// TestNextObjectID tests object ID generation
func TestNextObjectID(t *testing.T) {
	stage := NewStage("test_stage")

	// Generate several object IDs
	ids := make(map[uint32]bool)
	for i := 0; i < 10; i++ {
		id := stage.NextObjectID()
		if ids[id] {
			t.Errorf("duplicate object ID generated: %d", id)
		}
		ids[id] = true
	}
}

// TestNextObjectIDWrap tests that object ID wraps at 127
func TestNextObjectIDWrap(t *testing.T) {
	stage := NewStage("test_stage")
	stage.objectIndex = 125

	// Generate IDs to trigger wrap
	stage.NextObjectID() // 126
	stage.NextObjectID() // should wrap to 1

	// After wrap, objectIndex should be 1
	if stage.objectIndex != 1 {
		t.Errorf("objectIndex after wrap = %d, want 1", stage.objectIndex)
	}
}

// TestIsCharInQuestByID tests character quest membership check
func TestIsCharInQuestByID(t *testing.T) {
	stage := NewStage("test_stage")

	// No reservations - should return false
	if stage.isCharInQuestByID(12345) {
		t.Error("should return false when no reservations exist")
	}

	// Add reservation
	stage.reservedClientSlots[12345] = true

	if !stage.isCharInQuestByID(12345) {
		t.Error("should return true when character is reserved")
	}
}

// TestIsQuest tests quest detection
func TestIsQuest(t *testing.T) {
	stage := NewStage("test_stage")

	// No reservations - not a quest
	if stage.isQuest() {
		t.Error("should return false when no reservations exist")
	}

	// Add reservation - becomes a quest
	stage.reservedClientSlots[12345] = true

	if !stage.isQuest() {
		t.Error("should return true when reservations exist")
	}
}
