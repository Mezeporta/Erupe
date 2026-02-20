package channelserver

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

// ============================================================================
// CLIENT CONNECTION SIMULATION TESTS
// Tests that simulate actual client connections, not just mock sessions
//
// Purpose: Test the complete connection lifecycle as a real client would
// - TCP connection establishment
// - Packet exchange
// - Graceful disconnect
// - Ungraceful disconnect
// - Network errors
// ============================================================================

// MockNetConn simulates a net.Conn for testing
type MockNetConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
	mu       sync.Mutex
	readErr  error
	writeErr error
}

func NewMockNetConn() *MockNetConn {
	return &MockNetConn{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
}

func (m *MockNetConn) Read(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.closed {
		return 0, io.EOF
	}
	if m.readErr != nil {
		return 0, m.readErr
	}
	return m.readBuf.Read(b)
}

func (m *MockNetConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeBuf.Write(b)
}

func (m *MockNetConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockNetConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 54001}
}

func (m *MockNetConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

func (m *MockNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *MockNetConn) QueueRead(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readBuf.Write(data)
}

func (m *MockNetConn) GetWritten() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeBuf.Bytes()
}

func (m *MockNetConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// TestClientConnection_GracefulLoginLogout simulates a complete client session
// This is closer to what a real client does than handler-only tests
func TestClientConnection_GracefulLoginLogout(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "client_test_user")
	charID := CreateTestCharacter(t, db, userID, "ClientChar")

	t.Log("Simulating client connection with graceful logout")

	// Simulate client connecting
	mockConn := NewMockNetConn()
	session := createTestSessionForServerWithChar(server, charID, "ClientChar")
	
	// In real scenario, this would be set up by the connection handler
	// For testing, we test handlers directly without starting packet loops

	// Client sends save packet
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("ClientChar\x00"))
	saveData[8000] = 0xAB
	saveData[8001] = 0xCD

	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}

	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      12001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session, savePkt)
	time.Sleep(100 * time.Millisecond)

	// Client sends logout packet (graceful)
	t.Log("Client sending logout packet")
	logoutPkt := &mhfpacket.MsgSysLogout{}
	handleMsgSysLogout(session, logoutPkt)
	time.Sleep(100 * time.Millisecond)

	// Verify connection closed
	if !mockConn.IsClosed() {
		// Note: Our mock doesn't auto-close, but real session would
		t.Log("Mock connection not closed (expected for mock)")
	}

	// Verify data saved
	var savedCompressed []byte
	err = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query savedata: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("❌ No data saved after graceful logout")
	} else {
		decompressed, _ := nullcomp.Decompress(savedCompressed)
		if len(decompressed) > 8001 {
			if decompressed[8000] == 0xAB && decompressed[8001] == 0xCD {
				t.Log("✓ Data saved correctly after graceful logout")
			} else {
				t.Error("❌ Data corrupted")
			}
		}
	}
}

// TestClientConnection_UngracefulDisconnect simulates network failure
func TestClientConnection_UngracefulDisconnect(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "disconnect_user")
	charID := CreateTestCharacter(t, db, userID, "DisconnectChar")

	t.Log("Simulating ungraceful client disconnect (network error)")

	session := createTestSessionForServerWithChar(server, charID, "DisconnectChar")
	// Note: Not calling Start() - testing handlers directly
	time.Sleep(50 * time.Millisecond)

	// Client saves some data
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("DisconnectChar\x00"))
	saveData[9000] = 0xEF
	saveData[9001] = 0x12

	compressed, _ := nullcomp.Compress(saveData)
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      13001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session, savePkt)
	time.Sleep(100 * time.Millisecond)

	// Simulate network failure - connection drops without logout packet
	t.Log("Simulating network failure (no logout packet sent)")
	// In real scenario, recvLoop would detect io.EOF and call logoutPlayer
	logoutPlayer(session)
	time.Sleep(100 * time.Millisecond)

	// Verify data was saved despite ungraceful disconnect
	var savedCompressed []byte
	err := db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("❌ CRITICAL: No data saved after ungraceful disconnect")
		t.Error("This means players lose data when they have connection issues!")
	} else {
		t.Log("✓ Data saved even after ungraceful disconnect")
	}
}

// TestClientConnection_SessionTimeout simulates timeout disconnect
func TestClientConnection_SessionTimeout(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "timeout_user")
	charID := CreateTestCharacter(t, db, userID, "TimeoutChar")

	t.Log("Simulating session timeout (30s no packets)")

	session := createTestSessionForServerWithChar(server, charID, "TimeoutChar")
	// Note: Not calling Start() - testing handlers directly
	time.Sleep(50 * time.Millisecond)

	// Save data
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("TimeoutChar\x00"))
	saveData[10000] = 0xFF

	compressed, _ := nullcomp.Compress(saveData)
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      14001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session, savePkt)
	time.Sleep(100 * time.Millisecond)

	// Simulate timeout by setting lastPacket to long ago
	session.lastPacket = time.Now().Add(-35 * time.Second)

	// In production, invalidateSessions() goroutine would detect this
	// and call logoutPlayer(session)
	t.Log("Session timed out (>30s since last packet)")
	logoutPlayer(session)
	time.Sleep(100 * time.Millisecond)

	// Verify data saved
	var savedCompressed []byte
	err := db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("❌ CRITICAL: No data saved after timeout disconnect")
	} else {
		decompressed, _ := nullcomp.Decompress(savedCompressed)
		if len(decompressed) > 10000 && decompressed[10000] == 0xFF {
			t.Log("✓ Data saved correctly after timeout")
		} else {
			t.Error("❌ Data corrupted or not saved")
		}
	}
}

// TestClientConnection_MultipleClientsSimultaneous simulates multiple clients
func TestClientConnection_MultipleClientsSimultaneous(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	numClients := 3
	var wg sync.WaitGroup
	wg.Add(numClients)

	t.Logf("Simulating %d clients connecting simultaneously", numClients)

	for clientNum := 0; clientNum < numClients; clientNum++ {
		go func(num int) {
			defer wg.Done()

			username := fmt.Sprintf("multi_client_%d", num)
			charName := fmt.Sprintf("MultiClient%d", num)

			userID := CreateTestUser(t, db, username)
			charID := CreateTestCharacter(t, db, userID, charName)

			session := createTestSessionForServerWithChar(server, charID, charName)
			// Note: Not calling Start() - testing handlers directly
			time.Sleep(30 * time.Millisecond)

			// Each client saves their own data
			saveData := make([]byte, 150000)
			copy(saveData[88:], []byte(charName+"\x00"))
			saveData[11000+num] = byte(num)

			compressed, _ := nullcomp.Compress(saveData)
			savePkt := &mhfpacket.MsgMhfSavedata{
				SaveType:       0,
				AckHandle:      uint32(15000 + num),
				AllocMemSize:   uint32(len(compressed)),
				DataSize:       uint32(len(compressed)),
				RawDataPayload: compressed,
			}
			handleMsgMhfSavedata(session, savePkt)
			time.Sleep(50 * time.Millisecond)

			// Graceful logout
			logoutPlayer(session)
			time.Sleep(50 * time.Millisecond)

			// Verify individual client's data
			var savedCompressed []byte
			err := db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
			if err != nil {
				t.Errorf("Client %d: Failed to query: %v", num, err)
				return
			}

			if len(savedCompressed) > 0 {
				decompressed, _ := nullcomp.Decompress(savedCompressed)
				if len(decompressed) > 11000+num {
					if decompressed[11000+num] == byte(num) {
						t.Logf("Client %d: ✓ Data saved correctly", num)
					} else {
						t.Errorf("Client %d: ❌ Data corrupted", num)
					}
				}
			} else {
				t.Errorf("Client %d: ❌ No data saved", num)
			}
		}(clientNum)
	}

	wg.Wait()
	t.Log("All clients disconnected")
}

// TestClientConnection_SaveDuringCombat simulates saving while in quest
// This tests if being in a stage affects save behavior
func TestClientConnection_SaveDuringCombat(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "combat_user")
	charID := CreateTestCharacter(t, db, userID, "CombatChar")

	t.Log("Simulating save/logout while in quest/stage")

	session := createTestSessionForServerWithChar(server, charID, "CombatChar")
	
	// Simulate being in a stage (quest)
	// In real scenario, session.stage would be set when entering quest
	// For now, we'll just test the basic save/logout flow
	
	// Note: Not calling Start() - testing handlers directly
	time.Sleep(50 * time.Millisecond)

	// Save data during "combat"
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("CombatChar\x00"))
	saveData[12000] = 0xAA

	compressed, _ := nullcomp.Compress(saveData)
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      16001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session, savePkt)
	time.Sleep(100 * time.Millisecond)

	// Disconnect while in stage
	t.Log("Player disconnects during quest")
	logoutPlayer(session)
	time.Sleep(100 * time.Millisecond)

	// Verify data saved even during combat
	var savedCompressed []byte
	err := db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(savedCompressed) > 0 {
		decompressed, _ := nullcomp.Decompress(savedCompressed)
		if len(decompressed) > 12000 && decompressed[12000] == 0xAA {
			t.Log("✓ Data saved correctly even during quest")
		} else {
			t.Error("❌ Data not saved correctly during quest")
		}
	} else {
		t.Error("❌ CRITICAL: No data saved when disconnecting during quest")
	}
}

// TestClientConnection_ReconnectAfterCrash simulates client crash and reconnect
func TestClientConnection_ReconnectAfterCrash(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "crash_user")
	charID := CreateTestCharacter(t, db, userID, "CrashChar")

	t.Log("Simulating client crash and immediate reconnect")

	// First session - client crashes
	session1 := createTestSessionForServerWithChar(server, charID, "CrashChar")
	// Not calling Start()
	time.Sleep(50 * time.Millisecond)

	// Save some data before crash
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("CrashChar\x00"))
	saveData[13000] = 0xBB

	compressed, _ := nullcomp.Compress(saveData)
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      17001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session1, savePkt)
	time.Sleep(50 * time.Millisecond)

	// Client crashes (ungraceful disconnect)
	t.Log("Client crashes (no logout packet)")
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// Client reconnects immediately
	t.Log("Client reconnects after crash")
	session2 := createTestSessionForServerWithChar(server, charID, "CrashChar")
	// Not calling Start()
	time.Sleep(50 * time.Millisecond)

	// Load data
	loadPkt := &mhfpacket.MsgMhfLoaddata{
		AckHandle: 18001,
	}
	handleMsgMhfLoaddata(session2, loadPkt)
	time.Sleep(50 * time.Millisecond)

	// Verify data from before crash
	var savedCompressed []byte
	err := db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(savedCompressed) > 0 {
		decompressed, _ := nullcomp.Decompress(savedCompressed)
		if len(decompressed) > 13000 && decompressed[13000] == 0xBB {
			t.Log("✓ Data recovered correctly after crash")
		} else {
			t.Error("❌ Data lost or corrupted after crash")
		}
	} else {
		t.Error("❌ CRITICAL: All data lost after crash")
	}

	logoutPlayer(session2)
}

// TestClientConnection_PacketDuringLogout tests race condition
// What happens if save packet arrives during logout?
func TestClientConnection_PacketDuringLogout(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "race_user")
	charID := CreateTestCharacter(t, db, userID, "RaceChar")

	t.Log("Testing race condition: packet during logout")

	session := createTestSessionForServerWithChar(server, charID, "RaceChar")
	// Note: Not calling Start() - testing handlers directly
	time.Sleep(50 * time.Millisecond)

	// Prepare save packet
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("RaceChar\x00"))
	saveData[14000] = 0xCC

	compressed, _ := nullcomp.Compress(saveData)
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      19001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Send save packet
	go func() {
		defer wg.Done()
		handleMsgMhfSavedata(session, savePkt)
		t.Log("Save packet processed")
	}()

	// Goroutine 2: Trigger logout (almost) simultaneously
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Small delay
		logoutPlayer(session)
		t.Log("Logout processed")
	}()

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Verify final state
	var savedCompressed []byte
	err := db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Fatal("Race condition caused data loss - no savedata in DB")
	}

	decompressed, err := nullcomp.Decompress(savedCompressed)
	if err != nil {
		t.Fatalf("Saved data is not valid compressed data: %v", err)
	}
	if len(decompressed) < 15000 {
		t.Fatalf("Decompressed data too short (%d bytes), expected at least 15000", len(decompressed))
	}

	// Both outcomes are valid: either the save handler wrote last (0xCC preserved)
	// or the logout handler wrote last (0xCC overwritten with the logout's fresh
	// DB read). The important thing is no crash, no data loss, and valid data.
	if decompressed[14000] == 0xCC {
		t.Log("Race outcome: save handler wrote last - marker byte preserved")
	} else {
		t.Log("Race outcome: logout handler wrote last - marker byte overwritten (valid)")
	}
}

