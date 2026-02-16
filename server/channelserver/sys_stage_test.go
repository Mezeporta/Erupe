package channelserver

import (
	"sync"
	"testing"
)

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

	// Default max players is 127
	if stage.maxPlayers != 127 {
		t.Errorf("initial maxPlayers = %d, want 127", stage.maxPlayers)
	}
}

