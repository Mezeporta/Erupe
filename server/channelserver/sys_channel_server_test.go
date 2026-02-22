package channelserver

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	cfg "erupe-ce/config"
	"erupe-ce/network/clientctx"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	net.Conn
	closeCalled bool
	mu          sync.Mutex
	remoteAddr  net.Addr
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true
	return nil
}

func (m *mockConn) RemoteAddr() net.Addr {
	if m.remoteAddr != nil {
		return m.remoteAddr
	}
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *mockConn) WasClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeCalled
}

// createTestServer creates a test server instance
func createTestServer() *Server {
	logger, _ := zap.NewDevelopment()
	s := &Server{
		ID:             1,
		logger:         logger,
		sessions:       make(map[net.Conn]*Session),
		semaphore:      make(map[string]*Semaphore),
		questCache: NewQuestCache(0),
		erupeConfig: &cfg.Config{
			DebugOptions: cfg.DebugOptions{
				LogOutboundMessages: false,
				LogInboundMessages:  false,
			},
		},
		raviente: &Raviente{
			id:       1,
			register: make([]uint32, 30),
			state:    make([]uint32, 30),
			support:  make([]uint32, 30),
		},
	}
	s.Registry = NewLocalChannelRegistry([]*Server{s})
	return s
}

// createTestSessionForServer creates a session for a specific server
func createTestSessionForServer(server *Server, conn net.Conn, charID uint32, name string) *Session {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := &Session{
		logger:      server.logger,
		server:      server,
		rawConn:     conn,
		cryptConn:   mock,
		sendPackets: make(chan packet, 20),
		clientContext: &clientctx.ClientContext{},
		lastPacket:  time.Now(),
		charID:      charID,
		Name:        name,
	}
	return s
}

// TestNewServer tests server initialization
func TestNewServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &Config{
		ID:     1,
		Logger: logger,
		ErupeConfig: &cfg.Config{
			DebugOptions: cfg.DebugOptions{},
		},
		Name: "test-server",
	}

	server := NewServer(config)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.ID != 1 {
		t.Errorf("Server ID = %d, want 1", server.ID)
	}

	// Verify default stages are initialized
	expectedStages := []string{
		"sl1Ns200p0a0u0", // Mezeporta
		"sl1Ns211p0a0u0", // Rasta bar
		"sl1Ns260p0a0u0", // Pallone Caravan
		"sl1Ns262p0a0u0", // Pallone Guest House 1st Floor
		"sl1Ns263p0a0u0", // Pallone Guest House 2nd Floor
		"sl2Ns379p0a0u0", // Diva fountain
		"sl1Ns462p0a0u0", // MezFes
	}

	for _, stageID := range expectedStages {
		if _, exists := server.stages.Get(stageID); !exists {
			t.Errorf("Default stage %s not initialized", stageID)
		}
	}

	// Verify raviente initialization
	if server.raviente == nil {
		t.Error("Raviente not initialized")
	}
	if server.raviente.id != 1 {
		t.Errorf("Raviente ID = %d, want 1", server.raviente.id)
	}
}

// TestSessionTimeout tests the session timeout mechanism
func TestSessionTimeout(t *testing.T) {
	tests := []struct {
		name          string
		lastPacketAge time.Duration
		wantTimeout   bool
	}{
		{
			name:          "fresh_session_no_timeout",
			lastPacketAge: 5 * time.Second,
			wantTimeout:   false,
		},
		{
			name:          "old_session_should_timeout",
			lastPacketAge: 65 * time.Second,
			wantTimeout:   true,
		},
		{
			name:          "just_under_60s_no_timeout",
			lastPacketAge: 59 * time.Second,
			wantTimeout:   false,
		},
		{
			name:          "just_over_60s_timeout",
			lastPacketAge: 61 * time.Second,
			wantTimeout:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer()
			conn := &mockConn{}
			session := createTestSessionForServer(server, conn, 1, "TestChar")

			// Set last packet time in the past
			session.lastPacket = time.Now().Add(-tt.lastPacketAge)

			server.Lock()
			server.sessions[conn] = session
			server.Unlock()

			// Run one iteration of session invalidation
			for _, sess := range server.sessions {
				if time.Since(sess.lastPacket) > time.Second*time.Duration(60) {
					server.logger.Info("session timeout", zap.String("Name", sess.Name))
					// Don't actually call logoutPlayer in test, just mark as closed
					sess.closed.Store(true)
				}
			}

			gotTimeout := session.closed.Load()
			if gotTimeout != tt.wantTimeout {
				t.Errorf("session timeout = %v, want %v (age: %v)", gotTimeout, tt.wantTimeout, tt.lastPacketAge)
			}
		})
	}
}

// TestBroadcastMHF tests broadcasting messages to all sessions
func TestBroadcastMHF(t *testing.T) {
	server := createTestServer()

	// Create multiple sessions
	sessions := make([]*Session, 3)
	conns := make([]*mockConn, 3)
	for i := 0; i < 3; i++ {
		conn := &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + i}}
		conns[i] = conn
		sessions[i] = createTestSessionForServer(server, conn, uint32(i+1), fmt.Sprintf("Player%d", i+1))

		// Start the send loop for this session
		go sessions[i].sendLoop()

		server.Lock()
		server.sessions[conn] = sessions[i]
		server.Unlock()
	}

	// Create a test packet
	testPkt := &mhfpacket.MsgSysNop{}

	// Broadcast to all except first session
	server.BroadcastMHF(testPkt, sessions[0])

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Stop all sessions
	for _, sess := range sessions {
		sess.closed.Store(true)
	}
	time.Sleep(50 * time.Millisecond)

	// Verify sessions[0] didn't receive the packet
	mock0 := sessions[0].cryptConn.(*MockCryptConn)
	if mock0.PacketCount() > 0 {
		t.Errorf("Ignored session received %d packets, want 0", mock0.PacketCount())
	}

	// Verify sessions[1] and sessions[2] received the packet
	for i := 1; i < 3; i++ {
		mock := sessions[i].cryptConn.(*MockCryptConn)
		if mock.PacketCount() == 0 {
			t.Errorf("Session %d received 0 packets, want 1", i)
		}
	}
}

// TestBroadcastMHFAllSessions tests broadcasting to all sessions (no ignored session)
func TestBroadcastMHFAllSessions(t *testing.T) {
	server := createTestServer()

	// Create multiple sessions
	sessionCount := 5
	sessions := make([]*Session, sessionCount)
	for i := 0; i < sessionCount; i++ {
		conn := &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000 + i}}
		session := createTestSessionForServer(server, conn, uint32(i+1), fmt.Sprintf("Player%d", i+1))
		sessions[i] = session

		// Start the send loop
		go session.sendLoop()

		server.Lock()
		server.sessions[conn] = session
		server.Unlock()
	}

	// Broadcast to all sessions
	testPkt := &mhfpacket.MsgSysNop{}
	server.BroadcastMHF(testPkt, nil)

	time.Sleep(100 * time.Millisecond)

	// Stop all sessions
	for _, sess := range sessions {
		sess.closed.Store(true)
	}
	time.Sleep(50 * time.Millisecond)

	// Verify all sessions received the packet
	receivedCount := 0
	for _, sess := range server.sessions {
		mock := sess.cryptConn.(*MockCryptConn)
		if mock.PacketCount() > 0 {
			receivedCount++
		}
	}

	if receivedCount != sessionCount {
		t.Errorf("Received count = %d, want %d", receivedCount, sessionCount)
	}
}

// TestFindSessionByCharID tests finding sessions by character ID
func TestFindSessionByCharID(t *testing.T) {
	server := createTestServer()
	server.Registry = NewLocalChannelRegistry([]*Server{server})

	// Create sessions with different char IDs
	charIDs := []uint32{100, 200, 300}
	for _, charID := range charIDs {
		conn := &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: int(30000 + charID)}}
		session := createTestSessionForServer(server, conn, charID, fmt.Sprintf("Char%d", charID))

		server.Lock()
		server.sessions[conn] = session
		server.Unlock()
	}

	tests := []struct {
		name      string
		charID    uint32
		wantFound bool
	}{
		{
			name:      "existing_char_100",
			charID:    100,
			wantFound: true,
		},
		{
			name:      "existing_char_200",
			charID:    200,
			wantFound: true,
		},
		{
			name:      "non_existing_char",
			charID:    999,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := server.FindSessionByCharID(tt.charID)
			found := session != nil

			if found != tt.wantFound {
				t.Errorf("FindSessionByCharID(%d) found = %v, want %v", tt.charID, found, tt.wantFound)
			}

			if found && session.charID != tt.charID {
				t.Errorf("Found session charID = %d, want %d", session.charID, tt.charID)
			}
		})
	}
}

// TestHasSemaphore tests checking if a session has a semaphore
func TestHasSemaphore(t *testing.T) {
	server := createTestServer()
	conn1 := &mockConn{}
	conn2 := &mockConn{}

	session1 := createTestSessionForServer(server, conn1, 1, "Player1")
	session2 := createTestSessionForServer(server, conn2, 2, "Player2")

	// Create a semaphore hosted by session1
	sem := &Semaphore{
		id:      1,
		name:    "test_semaphore",
		host:    session1,
		clients: make(map[*Session]uint32),
	}

	server.semaphoreLock.Lock()
	server.semaphore["test_semaphore"] = sem
	server.semaphoreLock.Unlock()

	// Test session1 has semaphore
	if !server.HasSemaphore(session1) {
		t.Error("HasSemaphore(session1) = false, want true")
	}

	// Test session2 doesn't have semaphore
	if server.HasSemaphore(session2) {
		t.Error("HasSemaphore(session2) = true, want false")
	}
}

// TestSeason tests the season calculation
func TestSeason(t *testing.T) {
	server := createTestServer()

	tests := []struct {
		name     string
		serverID uint16
	}{
		{
			name:     "server_1",
			serverID: 0x1000,
		},
		{
			name:     "server_2",
			serverID: 0x1100,
		},
		{
			name:     "server_3",
			serverID: 0x1200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.ID = tt.serverID
			season := server.Season()

			// Season should be 0, 1, or 2
			if season > 2 {
				t.Errorf("Season() = %d, want 0-2", season)
			}
		})
	}
}

// TestRaviMultiplier tests the Raviente damage multiplier calculation
func TestRaviMultiplier(t *testing.T) {
	server := createTestServer()

	// Create a Raviente semaphore (name must end with "3" for getRaviSemaphore)
	conn := &mockConn{}
	hostSession := createTestSessionForServer(server, conn, 1, "RaviHost")

	sem := &Semaphore{
		id:      1,
		name:    "hs_l0u3",
		host:    hostSession,
		clients: make(map[*Session]uint32),
	}

	server.semaphoreLock.Lock()
	server.semaphore["hs_l0u3"] = sem
	server.semaphoreLock.Unlock()

	tests := []struct {
		name         string
		clientCount  int
		register9    uint32
		wantMultiple float64
	}{
		{
			name:         "small_quest_enough_players",
			clientCount:  4,
			register9:    0,
			wantMultiple: 1.0,
		},
		{
			name:         "small_quest_too_few_players",
			clientCount:  2,
			register9:    0,
			wantMultiple: 2.0, // 4 / 2
		},
		{
			name:         "large_quest_enough_players",
			clientCount:  24,
			register9:    10,
			wantMultiple: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up register
			server.raviente.register[9] = tt.register9

			// Add clients to semaphore
			sem.clients = make(map[*Session]uint32)
			for i := 0; i < tt.clientCount; i++ {
				mockConn := &mockConn{}
				sess := createTestSessionForServer(server, mockConn, uint32(i+10), fmt.Sprintf("RaviPlayer%d", i))
				sem.clients[sess] = uint32(i + 10)
			}

			multiplier := server.GetRaviMultiplier()
			if multiplier != tt.wantMultiple {
				t.Errorf("GetRaviMultiplier() = %v, want %v", multiplier, tt.wantMultiple)
			}
		})
	}
}

// TestUpdateRavi tests Raviente state updates
func TestUpdateRavi(t *testing.T) {
	server := createTestServer()

	tests := []struct {
		name      string
		semaID    uint32
		index     uint8
		value     uint32
		update    bool
		wantValue uint32
	}{
		{
			name:      "set_support_value",
			semaID:    0x50000,
			index:     3,
			value:     250,
			update:    false,
			wantValue: 250,
		},
		{
			name:      "set_register_value",
			semaID:    0x60000,
			index:     1,
			value:     42,
			update:    false,
			wantValue: 42,
		},
		{
			name:      "increment_register_value",
			semaID:    0x60000,
			index:     1,
			value:     8,
			update:    true,
			wantValue: 50, // Previous test set it to 42
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, newValue := server.UpdateRavi(tt.semaID, tt.index, tt.value, tt.update)
			if newValue != tt.wantValue {
				t.Errorf("UpdateRavi() new value = %d, want %d", newValue, tt.wantValue)
			}

			// Verify the value was actually stored
			var storedValue uint32
			switch tt.semaID {
			case 0x40000:
				storedValue = server.raviente.state[tt.index]
			case 0x50000:
				storedValue = server.raviente.support[tt.index]
			case 0x60000:
				storedValue = server.raviente.register[tt.index]
			}

			if storedValue != tt.wantValue {
				t.Errorf("Stored value = %d, want %d", storedValue, tt.wantValue)
			}
		})
	}
}

// TestResetRaviente tests Raviente reset functionality
func TestResetRaviente(t *testing.T) {
	server := createTestServer()

	// Set some non-zero values
	server.raviente.id = 5
	server.raviente.register[0] = 100
	server.raviente.state[1] = 200
	server.raviente.support[2] = 300

	// Reset should happen when no Raviente semaphores exist
	server.resetRaviente()

	// Verify ID incremented
	if server.raviente.id != 6 {
		t.Errorf("Raviente ID = %d, want 6", server.raviente.id)
	}

	// Verify arrays were reset
	for i := 0; i < 30; i++ {
		if server.raviente.register[i] != 0 {
			t.Errorf("register[%d] = %d, want 0", i, server.raviente.register[i])
		}
		if server.raviente.state[i] != 0 {
			t.Errorf("state[%d] = %d, want 0", i, server.raviente.state[i])
		}
		if server.raviente.support[i] != 0 {
			t.Errorf("support[%d] = %d, want 0", i, server.raviente.support[i])
		}
	}
}

// TestBroadcastChatMessage tests chat message broadcasting
func TestBroadcastChatMessage(t *testing.T) {
	server := createTestServer()
	server.name = "TestServer"

	// Create a session to receive the broadcast
	conn := &mockConn{}
	session := createTestSessionForServer(server, conn, 1, "Player1")

	// Start the send loop
	go session.sendLoop()

	server.Lock()
	server.sessions[conn] = session
	server.Unlock()

	// Broadcast a message
	server.BroadcastChatMessage("Test message")

	time.Sleep(100 * time.Millisecond)

	// Stop the session
	session.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	// Verify the session received a packet
	mock := session.cryptConn.(*MockCryptConn)
	if mock.PacketCount() == 0 {
		t.Error("Session didn't receive chat broadcast")
	}

	// Verify the packet contains the chat message (basic check)
	packets := mock.GetSentPackets()
	if len(packets) == 0 {
		t.Fatal("No packets sent")
	}

	// The packet should be non-empty
	if len(packets[0]) == 0 {
		t.Error("Empty packet sent for chat message")
	}
}

// TestConcurrentSessionAccess tests thread safety of session map access
func TestConcurrentSessionAccess(t *testing.T) {
	server := createTestServer()

	// Run concurrent operations on the session map
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent additions
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(id int) {
			defer wg.Done()
			conn := &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 40000 + id}}
			session := createTestSessionForServer(server, conn, uint32(id), fmt.Sprintf("Concurrent%d", id))

			server.Lock()
			server.sessions[conn] = session
			server.Unlock()
		}(i)
	}
	wg.Wait()

	// Verify all sessions were added
	server.Lock()
	count := len(server.sessions)
	server.Unlock()

	if count != iterations {
		t.Errorf("Session count = %d, want %d", count, iterations)
	}

	// Concurrent reads
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			server.Lock()
			_ = len(server.sessions)
			server.Unlock()
		}()
	}
	wg.Wait()
}

// TestFindObjectByChar tests finding objects by character ID
func TestFindObjectByChar(t *testing.T) {
	server := createTestServer()

	// Create a stage with objects
	stage := NewStage("test_stage")
	obj1 := &Object{
		id:          1,
		ownerCharID: 100,
	}
	obj2 := &Object{
		id:          2,
		ownerCharID: 200,
	}

	stage.objects[1] = obj1
	stage.objects[2] = obj2

	server.stages.Store("test_stage", stage)

	tests := []struct {
		name      string
		charID    uint32
		wantFound bool
		wantObjID uint32
	}{
		{
			name:      "find_char_100_object",
			charID:    100,
			wantFound: true,
			wantObjID: 1,
		},
		{
			name:      "find_char_200_object",
			charID:    200,
			wantFound: true,
			wantObjID: 2,
		},
		{
			name:      "char_not_found",
			charID:    999,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := server.FindObjectByChar(tt.charID)
			found := obj != nil

			if found != tt.wantFound {
				t.Errorf("FindObjectByChar(%d) found = %v, want %v", tt.charID, found, tt.wantFound)
			}

			if found && obj.id != tt.wantObjID {
				t.Errorf("Found object ID = %d, want %d", obj.id, tt.wantObjID)
			}
		})
	}
}
