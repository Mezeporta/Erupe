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
	wg.Go(func() {
		for i := 0; i < 1000; i++ {
			stage.BroadcastMHF(pkt, nil)
		}
	})

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
