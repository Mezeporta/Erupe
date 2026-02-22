package channelserver

import (
	"net"
	"testing"
	"time"

	cfg "erupe-ce/config"

	"go.uber.org/zap"
)

// createListeningTestServer creates a channel server that binds to a real TCP port.
// Port 0 lets the OS assign a free port. The server is automatically shut down
// when the test completes.
func createListeningTestServer(t *testing.T, id uint16) *Server {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	s := NewServer(&Config{
		ID:     id,
		Logger: logger,
		ErupeConfig: &cfg.Config{
			DebugOptions: cfg.DebugOptions{
				LogOutboundMessages: false,
				LogInboundMessages:  false,
			},
		},
	})
	s.Port = 0 // Let OS pick a free port
	if err := s.Start(); err != nil {
		t.Fatalf("channel %d failed to start: %v", id, err)
	}
	t.Cleanup(func() {
		s.Shutdown()
		time.Sleep(200 * time.Millisecond) // Let background goroutines and sessions exit.
	})
	return s
}

// listenerAddr returns the address the server is listening on.
func listenerAddr(s *Server) string {
	return s.listener.Addr().String()
}

// TestChannelIsolation_ShutdownDoesNotAffectOthers verifies that shutting down
// one channel server does not prevent other channels from accepting connections.
func TestChannelIsolation_ShutdownDoesNotAffectOthers(t *testing.T) {
	ch1 := createListeningTestServer(t, 1)
	ch2 := createListeningTestServer(t, 2)
	ch3 := createListeningTestServer(t, 3)

	addr1 := listenerAddr(ch1)
	addr2 := listenerAddr(ch2)
	addr3 := listenerAddr(ch3)

	// Verify all three channels accept connections initially.
	for _, addr := range []string{addr1, addr2, addr3} {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			t.Fatalf("initial connection to %s failed: %v", addr, err)
		}
		_ = conn.Close()
	}

	// Shut down channel 1.
	ch1.Shutdown()
	time.Sleep(50 * time.Millisecond)

	// Channel 1 should refuse connections.
	_, err := net.DialTimeout("tcp", addr1, 500*time.Millisecond)
	if err == nil {
		t.Error("channel 1 should refuse connections after shutdown")
	}

	// Channels 2 and 3 must still accept connections.
	for _, tc := range []struct {
		name string
		addr string
	}{
		{"channel 2", addr2},
		{"channel 3", addr3},
	} {
		conn, err := net.DialTimeout("tcp", tc.addr, time.Second)
		if err != nil {
			t.Errorf("%s should still accept connections after channel 1 shutdown, got: %v", tc.name, err)
		} else {
			_ = conn.Close()
		}
	}
}

// TestChannelIsolation_ListenerCloseDoesNotAffectOthers simulates an unexpected
// listener failure (e.g. port conflict, OS-level error) on one channel and
// verifies other channels continue operating.
func TestChannelIsolation_ListenerCloseDoesNotAffectOthers(t *testing.T) {
	ch1 := createListeningTestServer(t, 1)
	ch2 := createListeningTestServer(t, 2)

	addr2 := listenerAddr(ch2)

	// Forcibly close channel 1's listener (simulating unexpected failure).
	_ = ch1.listener.Close()
	time.Sleep(50 * time.Millisecond)

	// Channel 2 must still work.
	conn, err := net.DialTimeout("tcp", addr2, time.Second)
	if err != nil {
		t.Fatalf("channel 2 should still accept connections after channel 1 listener closed: %v", err)
	}
	_ = conn.Close()
}

// TestChannelIsolation_SessionPanicDoesNotAffectChannel verifies that a panic
// inside a session handler is recovered and does not crash the channel server.
func TestChannelIsolation_SessionPanicDoesNotAffectChannel(t *testing.T) {
	ch := createListeningTestServer(t, 1)
	addr := listenerAddr(ch)

	// Connect a client that will trigger a session.
	conn1, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("first connection failed: %v", err)
	}

	// Send garbage data that will cause handlePacketGroup to hit the panic recovery.
	// The session's defer/recover should catch it without killing the channel.
	_, _ = conn1.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	time.Sleep(100 * time.Millisecond)
	_ = conn1.Close()
	time.Sleep(100 * time.Millisecond)

	// The channel should still accept new connections after the panic.
	conn2, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("channel should still accept connections after session panic: %v", err)
	}
	_ = conn2.Close()
}

// TestChannelIsolation_CrossChannelRegistryAfterShutdown verifies that the
// channel registry handles a shut-down channel gracefully during cross-channel
// operations (search, find, disconnect).
func TestChannelIsolation_CrossChannelRegistryAfterShutdown(t *testing.T) {
	channels := createTestChannels(3)
	reg := NewLocalChannelRegistry(channels)

	// Add sessions to all channels.
	for i, ch := range channels {
		conn := &mockConn{}
		sess := createTestSessionForServer(ch, conn, uint32(i+1), "Player")
		sess.stage = NewStage("sl1Ns200p0a0u0")
		ch.Lock()
		ch.sessions[conn] = sess
		ch.Unlock()
	}

	// Simulate channel 1 shutting down by marking it and clearing sessions.
	channels[0].Lock()
	channels[0].isShuttingDown = true
	channels[0].sessions = make(map[net.Conn]*Session)
	channels[0].Unlock()

	// Registry operations should still work for remaining channels.
	found := reg.FindSessionByCharID(2)
	if found == nil {
		t.Error("FindSessionByCharID(2) should find session on channel 2")
	}

	found = reg.FindSessionByCharID(3)
	if found == nil {
		t.Error("FindSessionByCharID(3) should find session on channel 3")
	}

	// Session from shut-down channel should not be found.
	found = reg.FindSessionByCharID(1)
	if found != nil {
		t.Error("FindSessionByCharID(1) should not find session on shut-down channel")
	}

	// SearchSessions should return only sessions from live channels.
	results := reg.SearchSessions(func(s SessionSnapshot) bool { return true }, 10)
	if len(results) != 2 {
		t.Errorf("SearchSessions should return 2 results from live channels, got %d", len(results))
	}
}

// TestChannelIsolation_IndependentStages verifies that stages are per-channel
// and one channel's stages don't leak into another.
func TestChannelIsolation_IndependentStages(t *testing.T) {
	channels := createTestChannels(2)

	stageName := "sl1Qs999p0a0u42"

	// Add stage only to channel 1.
	channels[0].stages.Store(stageName, NewStage(stageName))

	// Channel 1 should have the stage.
	_, ok1 := channels[0].stages.Get(stageName)
	if !ok1 {
		t.Error("channel 1 should have the stage")
	}

	// Channel 2 should NOT have the stage.
	_, ok2 := channels[1].stages.Get(stageName)
	if ok2 {
		t.Error("channel 2 should not have channel 1's stage")
	}
}
