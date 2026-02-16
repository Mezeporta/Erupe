package channelserver

import (
	"sync"
	"testing"
)

func TestNewSemaphore(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	sema := NewSemaphore(session, "test_semaphore", 16)

	if sema == nil {
		t.Fatal("NewSemaphore() returned nil")
	}
	if sema.name != "test_semaphore" {
		t.Errorf("name = %s, want test_semaphore", sema.name)
	}
	if sema.maxPlayers != 16 {
		t.Errorf("maxPlayers = %d, want 16", sema.maxPlayers)
	}
	if sema.clients == nil {
		t.Error("clients map should be initialized")
	}
	if sema.host != session {
		t.Error("host should be set to the creating session")
	}
}

func TestNewSemaphoreIDIncrement(t *testing.T) {
	server := createMockServer()
	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)
	session3 := createMockSession(3, server)

	sema1 := NewSemaphore(session1, "sema1", 4)
	sema2 := NewSemaphore(session2, "sema2", 4)
	sema3 := NewSemaphore(session3, "sema3", 4)

	// IDs should be set (may or may not be unique depending on session state)
	if sema1.id == 0 && sema2.id == 0 && sema3.id == 0 {
		t.Error("at least some semaphore IDs should be non-zero")
	}
}

func TestSemaphoreClients(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)
	sema := NewSemaphore(session, "test", 4)

	session1 := createMockSession(100, server)
	session2 := createMockSession(200, server)

	// Add clients
	sema.clients[session1] = session1.charID
	sema.clients[session2] = session2.charID

	if len(sema.clients) != 2 {
		t.Errorf("clients count = %d, want 2", len(sema.clients))
	}

	// Verify client IDs
	if sema.clients[session1] != 100 {
		t.Errorf("clients[session1] = %d, want 100", sema.clients[session1])
	}
	if sema.clients[session2] != 200 {
		t.Errorf("clients[session2] = %d, want 200", sema.clients[session2])
	}
}

func TestSemaphoreRemoveClient(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)
	sema := NewSemaphore(session, "test", 4)

	clientSession := createMockSession(100, server)
	sema.clients[clientSession] = clientSession.charID

	// Remove client
	delete(sema.clients, clientSession)

	if len(sema.clients) != 0 {
		t.Errorf("clients count = %d, want 0 after delete", len(sema.clients))
	}
}

func TestSemaphoreMaxPlayers(t *testing.T) {
	tests := []struct {
		name       string
		maxPlayers uint16
	}{
		{"quest party", 4},
		{"small event", 16},
		{"raviente", 32},
		{"large event", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer()
			session := createMockSession(1, server)
			sema := NewSemaphore(session, tt.name, tt.maxPlayers)

			if sema.maxPlayers != tt.maxPlayers {
				t.Errorf("maxPlayers = %d, want %d", sema.maxPlayers, tt.maxPlayers)
			}
		})
	}
}

func TestSemaphoreBroadcastMHF(t *testing.T) {
	server := createMockServer()
	hostSession := createMockSession(1, server)
	sema := NewSemaphore(hostSession, "test", 4)

	session1 := createMockSession(100, server)
	session2 := createMockSession(200, server)
	session3 := createMockSession(300, server)

	sema.clients[session1] = session1.charID
	sema.clients[session2] = session2.charID
	sema.clients[session3] = session3.charID

	pkt := &mockPacket{opcode: 0x1234}

	// Broadcast excluding session1
	sema.BroadcastMHF(pkt, session1)

	// session2 and session3 should receive
	select {
	case data := <-session2.sendPackets:
		if len(data.data) == 0 {
			t.Error("session2 received empty data")
		}
	default:
		t.Error("session2 did not receive broadcast")
	}

	select {
	case data := <-session3.sendPackets:
		if len(data.data) == 0 {
			t.Error("session3 received empty data")
		}
	default:
		t.Error("session3 did not receive broadcast")
	}

	// session1 should NOT receive (it was ignored)
	select {
	case <-session1.sendPackets:
		t.Error("session1 should not receive broadcast (it was ignored)")
	default:
		// Expected - no data for session1
	}
}

func TestSemaphoreBroadcastToAll(t *testing.T) {
	server := createMockServer()
	hostSession := createMockSession(1, server)
	sema := NewSemaphore(hostSession, "test", 4)

	session1 := createMockSession(100, server)
	session2 := createMockSession(200, server)

	sema.clients[session1] = session1.charID
	sema.clients[session2] = session2.charID

	pkt := &mockPacket{opcode: 0x1234}

	// Broadcast to all (nil ignored session)
	sema.BroadcastMHF(pkt, nil)

	// Both should receive
	count := 0
	select {
	case <-session1.sendPackets:
		count++
	default:
	}
	select {
	case <-session2.sendPackets:
		count++
	default:
	}

	if count != 2 {
		t.Errorf("expected 2 broadcasts, got %d", count)
	}
}

func TestSemaphoreRWMutex(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)
	sema := NewSemaphore(session, "test", 4)

	// Test that RWMutex works
	sema.RLock()
	_ = len(sema.clients) // Read operation
	sema.RUnlock()

	sema.Lock()
	sema.clients[createMockSession(100, server)] = 100 // Write operation
	sema.Unlock()
}

func TestSemaphoreConcurrentAccess(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)
	sema := NewSemaphore(session, "test", 100)

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s := createMockSession(uint32(id*100+j), server)
				sema.Lock()
				sema.clients[s] = s.charID
				sema.Unlock()

				sema.Lock()
				delete(sema.clients, s)
				sema.Unlock()
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				sema.RLock()
				_ = len(sema.clients)
				sema.RUnlock()
			}
		}()
	}

	wg.Wait()
}

func TestSemaphoreEmptyBroadcast(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)
	sema := NewSemaphore(session, "test", 4)

	pkt := &mockPacket{opcode: 0x1234}

	// Should not panic with no clients
	sema.BroadcastMHF(pkt, nil)
}

func TestSemaphoreNameString(t *testing.T) {
	server := createMockServer()

	tests := []string{
		"quest_001",
		"raviente_phase1",
		"tournament_round3",
		"diva_defense",
	}

	for _, id := range tests {
		session := createMockSession(1, server)
		sema := NewSemaphore(session, id, 4)
		if sema.name != id {
			t.Errorf("name = %s, want %s", sema.name, id)
		}
	}
}
