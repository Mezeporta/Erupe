package channelserver

import (
	"sync"
	"testing"
)

func TestNewSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphoreIndex = 6 // Start index (IDs 0-6 are reserved)

	sema := NewSemaphore(server, "test_semaphore", 16)

	if sema == nil {
		t.Fatal("NewSemaphore() returned nil")
	}
	if sema.id_semaphore != "test_semaphore" {
		t.Errorf("id_semaphore = %s, want test_semaphore", sema.id_semaphore)
	}
	if sema.maxPlayers != 16 {
		t.Errorf("maxPlayers = %d, want 16", sema.maxPlayers)
	}
	if sema.clients == nil {
		t.Error("clients map should be initialized")
	}
	if sema.reservedClientSlots == nil {
		t.Error("reservedClientSlots map should be initialized")
	}
}

func TestNewSemaphoreIDIncrement(t *testing.T) {
	server := createMockServer()
	server.semaphoreIndex = 6

	sema1 := NewSemaphore(server, "sema1", 4)
	sema2 := NewSemaphore(server, "sema2", 4)
	sema3 := NewSemaphore(server, "sema3", 4)

	// IDs should increment
	if sema1.id == sema2.id {
		t.Error("semaphore IDs should be unique")
	}
	if sema2.id == sema3.id {
		t.Error("semaphore IDs should be unique")
	}
}

func TestSemaphoreClients(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

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

func TestSemaphoreReservedSlots(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

	// Reserve slots
	sema.reservedClientSlots[100] = nil
	sema.reservedClientSlots[200] = nil

	if len(sema.reservedClientSlots) != 2 {
		t.Errorf("reservedClientSlots count = %d, want 2", len(sema.reservedClientSlots))
	}

	// Check existence
	if _, ok := sema.reservedClientSlots[100]; !ok {
		t.Error("charID 100 should be reserved")
	}
	if _, ok := sema.reservedClientSlots[200]; !ok {
		t.Error("charID 200 should be reserved")
	}
	if _, ok := sema.reservedClientSlots[300]; ok {
		t.Error("charID 300 should not be reserved")
	}
}

func TestSemaphoreRemoveClient(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

	session := createMockSession(100, server)
	sema.clients[session] = session.charID

	// Remove client
	delete(sema.clients, session)

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
			sema := NewSemaphore(server, tt.name, tt.maxPlayers)

			if sema.maxPlayers != tt.maxPlayers {
				t.Errorf("maxPlayers = %d, want %d", sema.maxPlayers, tt.maxPlayers)
			}
		})
	}
}

func TestSemaphoreBroadcastMHF(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

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

func TestSemaphoreBroadcastRavi(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "raviente", 32)

	session1 := createMockSession(100, server)
	session2 := createMockSession(200, server)

	sema.clients[session1] = session1.charID
	sema.clients[session2] = session2.charID

	pkt := &mockPacket{opcode: 0x5678}

	// Broadcast to all (no ignored session)
	sema.BroadcastRavi(pkt)

	// Both should receive
	select {
	case data := <-session1.sendPackets:
		if len(data.data) == 0 {
			t.Error("session1 received empty data")
		}
	default:
		t.Error("session1 did not receive Ravi broadcast")
	}

	select {
	case data := <-session2.sendPackets:
		if len(data.data) == 0 {
			t.Error("session2 received empty data")
		}
	default:
		t.Error("session2 did not receive Ravi broadcast")
	}
}

func TestSemaphoreBroadcastToAll(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

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
	sema := NewSemaphore(server, "test", 4)

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
	sema := NewSemaphore(server, "test", 100)

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				session := createMockSession(uint32(id*100+j), server)
				sema.Lock()
				sema.clients[session] = session.charID
				sema.Unlock()

				sema.Lock()
				delete(sema.clients, session)
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
	sema := NewSemaphore(server, "test", 4)

	pkt := &mockPacket{opcode: 0x1234}

	// Should not panic with no clients
	sema.BroadcastMHF(pkt, nil)
	sema.BroadcastRavi(pkt)
}

func TestSemaphoreIDString(t *testing.T) {
	server := createMockServer()

	tests := []string{
		"quest_001",
		"raviente_phase1",
		"tournament_round3",
		"diva_defense",
	}

	for _, id := range tests {
		sema := NewSemaphore(server, id, 4)
		if sema.id_semaphore != id {
			t.Errorf("id_semaphore = %s, want %s", sema.id_semaphore, id)
		}
	}
}

func TestSemaphoreNumericID(t *testing.T) {
	server := createMockServer()
	server.semaphoreIndex = 6 // IDs 0-6 reserved

	sema := NewSemaphore(server, "test", 4)

	// First semaphore should get ID 7
	if sema.id < 7 {
		t.Errorf("semaphore id = %d, should be >= 7", sema.id)
	}
}

func TestSemaphoreReserveAndRelease(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

	// Reserve
	sema.reservedClientSlots[100] = nil
	if _, ok := sema.reservedClientSlots[100]; !ok {
		t.Error("slot 100 should be reserved")
	}

	// Release
	delete(sema.reservedClientSlots, 100)
	if _, ok := sema.reservedClientSlots[100]; ok {
		t.Error("slot 100 should be released")
	}
}

func TestSemaphoreClientAndReservedSeparate(t *testing.T) {
	server := createMockServer()
	sema := NewSemaphore(server, "test", 4)

	session := createMockSession(100, server)

	// Client in active clients
	sema.clients[session] = 100

	// Same charID reserved
	sema.reservedClientSlots[100] = nil

	// Both should exist independently
	if _, ok := sema.clients[session]; !ok {
		t.Error("session should be in active clients")
	}
	if _, ok := sema.reservedClientSlots[100]; !ok {
		t.Error("charID 100 should be reserved")
	}

	// Remove from one doesn't affect other
	delete(sema.clients, session)
	if _, ok := sema.reservedClientSlots[100]; !ok {
		t.Error("charID 100 should still be reserved after removing from clients")
	}
}
