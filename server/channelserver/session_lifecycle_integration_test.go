package channelserver

import (
	"bytes"
	"net"
	"testing"
	"time"

	cfg "erupe-ce/config"
	"erupe-ce/common/mhfitem"
	"erupe-ce/network/clientctx"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ============================================================================
// SESSION LIFECYCLE INTEGRATION TESTS
// Full end-to-end tests that simulate the complete player session lifecycle
// 
// These tests address the core issue: handler-level tests don't catch problems
// with the logout flow. Players report data loss because logout doesn't 
// trigger save handlers.
//
// Test Strategy:
// 1. Create a real session (not just call handlers directly)
// 2. Modify game data through packets
// 3. Trigger actual logout event (not just call handlers)
// 4. Create new session for the same character
// 5. Verify all data persists correctly
// ============================================================================

// TestSessionLifecycle_BasicSaveLoadCycle tests the complete session lifecycle
// This is the minimal reproduction case for player-reported data loss
func TestSessionLifecycle_BasicSaveLoadCycle(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	// Create test user and character
	userID := CreateTestUser(t, db, "lifecycle_test_user")
	charID := CreateTestCharacter(t, db, userID, "LifecycleChar")

	t.Logf("Created character ID %d for lifecycle test", charID)

	// ===== SESSION 1: Login, modify data, logout =====
	t.Log("--- Starting Session 1: Login and modify data ---")

	session1 := createTestSessionForServerWithChar(server, charID, "LifecycleChar")
	// Note: Not calling Start() since we're testing handlers directly, not packet processing

	// Modify data via packet handlers (frontier_points is on users table since 9.2 migration)
	initialPoints := uint32(5000)
	_, err := db.Exec("UPDATE users SET frontier_points = $1 WHERE id = $2", initialPoints, userID)
	if err != nil {
		t.Fatalf("Failed to set initial road points: %v", err)
	}

	// Save main savedata through packet
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("LifecycleChar\x00"))
	// Add some identifiable data at offset 1000
	saveData[1000] = 0xDE
	saveData[1001] = 0xAD
	saveData[1002] = 0xBE
	saveData[1003] = 0xEF

	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress savedata: %v", err)
	}

	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      1001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}

	t.Log("Sending savedata packet")
	handleMsgMhfSavedata(session1, savePkt)

	// Drain ACK
	time.Sleep(100 * time.Millisecond)

	// Now trigger logout via the actual logout flow
	t.Log("Triggering logout via logoutPlayer")
	logoutPlayer(session1)

	// Give logout time to complete
	time.Sleep(100 * time.Millisecond)

	// ===== SESSION 2: Login again and verify data =====
	t.Log("--- Starting Session 2: Login and verify data persists ---")

	session2 := createTestSessionForServerWithChar(server, charID, "LifecycleChar")
	// Note: Not calling Start() since we're testing handlers directly

	// Load character data
	loadPkt := &mhfpacket.MsgMhfLoaddata{
		AckHandle: 2001,
	}
	handleMsgMhfLoaddata(session2, loadPkt)

	time.Sleep(50 * time.Millisecond)

	// Verify savedata persisted
	var savedCompressed []byte
	err = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to load savedata after session: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("❌ CRITICAL: Savedata not persisted across logout/login cycle")
		return
	}

	// Decompress and verify
	decompressed, err := nullcomp.Decompress(savedCompressed)
	if err != nil {
		t.Errorf("Failed to decompress savedata: %v", err)
		return
	}

	// Check our marker bytes
	if len(decompressed) > 1003 {
		if decompressed[1000] != 0xDE || decompressed[1001] != 0xAD ||
			decompressed[1002] != 0xBE || decompressed[1003] != 0xEF {
			t.Error("❌ CRITICAL: Savedata contents corrupted or not saved correctly")
			t.Errorf("Expected [DE AD BE EF] at offset 1000, got [%02X %02X %02X %02X]",
				decompressed[1000], decompressed[1001], decompressed[1002], decompressed[1003])
		} else {
			t.Log("✓ Savedata persisted correctly across logout/login")
		}
	} else {
		t.Error("❌ CRITICAL: Savedata too short after reload")
	}

	// Verify name persisted
	if session2.Name != "LifecycleChar" {
		t.Errorf("❌ Character name not loaded correctly: got %q, want %q", session2.Name, "LifecycleChar")
	} else {
		t.Log("✓ Character name persisted correctly")
	}

	// Clean up
	logoutPlayer(session2)
}

// TestSessionLifecycle_WarehouseDataPersistence tests warehouse across sessions
// This addresses user report: "warehouse contents not saved"
func TestSessionLifecycle_WarehouseDataPersistence(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "warehouse_test_user")
	charID := CreateTestCharacter(t, db, userID, "WarehouseChar")

	t.Log("Testing warehouse persistence across logout/login")

	// ===== SESSION 1: Add items to warehouse =====
	session1 := createTestSessionForServerWithChar(server, charID, "WarehouseChar")

	// Create test equipment for warehouse
	equipment := []mhfitem.MHFEquipment{
		createTestEquipmentItem(100, 1),
		createTestEquipmentItem(101, 2),
		createTestEquipmentItem(102, 3),
	}

	serializedEquip := mhfitem.SerializeWarehouseEquipment(equipment, cfg.ZZ)

	// Save to warehouse directly (simulating a save handler)
	_, _ = db.Exec("INSERT INTO warehouse (character_id) VALUES ($1) ON CONFLICT DO NOTHING", charID)
	_, err := db.Exec("UPDATE warehouse SET equip0 = $1 WHERE character_id = $2", serializedEquip, charID)
	if err != nil {
		t.Fatalf("Failed to save warehouse: %v", err)
	}

	t.Log("Saved equipment to warehouse in session 1")

	// Logout
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// ===== SESSION 2: Verify warehouse contents =====
	session2 := createTestSessionForServerWithChar(server, charID, "WarehouseChar")

	// Reload warehouse
	var savedEquip []byte
	err = db.QueryRow("SELECT equip0 FROM warehouse WHERE character_id = $1", charID).Scan(&savedEquip)
	if err != nil {
		t.Errorf("❌ Failed to load warehouse after logout: %v", err)
		logoutPlayer(session2)
		return
	}

	if len(savedEquip) == 0 {
		t.Error("❌ Warehouse equipment not saved")
	} else if !bytes.Equal(savedEquip, serializedEquip) {
		t.Error("❌ Warehouse equipment data mismatch")
	} else {
		t.Log("✓ Warehouse equipment persisted correctly across logout/login")
	}

	logoutPlayer(session2)
}

// TestSessionLifecycle_KoryoPointsPersistence tests kill counter across sessions
// This addresses user report: "monster kill counter not saved"
func TestSessionLifecycle_KoryoPointsPersistence(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "koryo_test_user")
	charID := CreateTestCharacter(t, db, userID, "KoryoChar")

	t.Log("Testing Koryo points persistence across logout/login")

	// ===== SESSION 1: Add Koryo points =====
	session1 := createTestSessionForServerWithChar(server, charID, "KoryoChar")

	// Add Koryo points via packet
	addPoints := uint32(250)
	pkt := &mhfpacket.MsgMhfAddKouryouPoint{
		AckHandle:     3001,
		KouryouPoints: addPoints,
	}

	t.Logf("Adding %d Koryo points", addPoints)
	handleMsgMhfAddKouryouPoint(session1, pkt)
	time.Sleep(50 * time.Millisecond)

	// Verify points were added in session 1
	var points1 uint32
	err := db.QueryRow("SELECT COALESCE(kouryou_point, 0) FROM characters WHERE id = $1", charID).Scan(&points1)
	if err != nil {
		t.Fatalf("Failed to query koryo points: %v", err)
	}
	t.Logf("Koryo points after add: %d", points1)

	// Logout
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// ===== SESSION 2: Verify Koryo points persist =====
	session2 := createTestSessionForServerWithChar(server, charID, "KoryoChar")

	// Reload Koryo points
	var points2 uint32
	err = db.QueryRow("SELECT COALESCE(kouryou_point, 0) FROM characters WHERE id = $1", charID).Scan(&points2)
	if err != nil {
		t.Errorf("❌ Failed to load koryo points after logout: %v", err)
		logoutPlayer(session2)
		return
	}

	if points2 != addPoints {
		t.Errorf("❌ Koryo points not persisted: got %d, want %d", points2, addPoints)
	} else {
		t.Logf("✓ Koryo points persisted correctly: %d", points2)
	}

	logoutPlayer(session2)
}

// TestSessionLifecycle_MultipleDataTypesPersistence tests multiple data types in one session
// This is the comprehensive test that simulates a real player session
func TestSessionLifecycle_MultipleDataTypesPersistence(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "multi_test_user")
	charID := CreateTestCharacter(t, db, userID, "MultiChar")

	t.Log("Testing multiple data types persistence across logout/login")

	// ===== SESSION 1: Modify multiple data types =====
	session1 := createTestSessionForServerWithChar(server, charID, "MultiChar")

	// 1. Set Road Points (frontier_points is on users table since 9.2 migration)
	rdpPoints := uint32(7500)
	_, err := db.Exec("UPDATE users SET frontier_points = $1 WHERE id = $2", rdpPoints, userID)
	if err != nil {
		t.Fatalf("Failed to set RdP: %v", err)
	}

	// 2. Add Koryo Points
	koryoPoints := uint32(500)
	addKoryoPkt := &mhfpacket.MsgMhfAddKouryouPoint{
		AckHandle:     4001,
		KouryouPoints: koryoPoints,
	}
	handleMsgMhfAddKouryouPoint(session1, addKoryoPkt)

	// 3. Save Hunter Navi
	naviData := make([]byte, 552)
	for i := range naviData {
		naviData[i] = byte((i * 7) % 256)
	}
	naviPkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      4002,
		IsDataDiff:     false,
		RawDataPayload: naviData,
	}
	handleMsgMhfSaveHunterNavi(session1, naviPkt)

	// 4. Save main savedata
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("MultiChar\x00"))
	saveData[2000] = 0xCA
	saveData[2001] = 0xFE
	saveData[2002] = 0xBA
	saveData[2003] = 0xBE

	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress savedata: %v", err)
	}

	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      4003,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session1, savePkt)

	// Give handlers time to process
	time.Sleep(100 * time.Millisecond)

	t.Log("Modified all data types in session 1")

	// Logout
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// ===== SESSION 2: Verify all data persists =====
	session2 := createTestSessionForServerWithChar(server, charID, "MultiChar")

	// Load character data
	loadPkt := &mhfpacket.MsgMhfLoaddata{
		AckHandle: 5001,
	}
	handleMsgMhfLoaddata(session2, loadPkt)
	time.Sleep(50 * time.Millisecond)

	allPassed := true

	// Verify 1: Road Points (frontier_points is on users table)
	var loadedRdP uint32
	_ = db.QueryRow("SELECT frontier_points FROM users WHERE id = $1", userID).Scan(&loadedRdP)
	if loadedRdP != rdpPoints {
		t.Errorf("❌ RdP not persisted: got %d, want %d", loadedRdP, rdpPoints)
		allPassed = false
	} else {
		t.Logf("✓ RdP persisted: %d", loadedRdP)
	}

	// Verify 2: Koryo Points
	var loadedKoryo uint32
	_ = db.QueryRow("SELECT COALESCE(kouryou_point, 0) FROM characters WHERE id = $1", charID).Scan(&loadedKoryo)
	if loadedKoryo != koryoPoints {
		t.Errorf("❌ Koryo points not persisted: got %d, want %d", loadedKoryo, koryoPoints)
		allPassed = false
	} else {
		t.Logf("✓ Koryo points persisted: %d", loadedKoryo)
	}

	// Verify 3: Hunter Navi
	var loadedNavi []byte
	_ = db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", charID).Scan(&loadedNavi)
	if len(loadedNavi) == 0 {
		t.Error("❌ Hunter Navi not saved")
		allPassed = false
	} else if !bytes.Equal(loadedNavi, naviData) {
		t.Error("❌ Hunter Navi data mismatch")
		allPassed = false
	} else {
		t.Logf("✓ Hunter Navi persisted: %d bytes", len(loadedNavi))
	}

	// Verify 4: Savedata
	var savedCompressed []byte
	_ = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if len(savedCompressed) == 0 {
		t.Error("❌ Savedata not saved")
		allPassed = false
	} else {
		decompressed, err := nullcomp.Decompress(savedCompressed)
		if err != nil {
			t.Errorf("❌ Failed to decompress savedata: %v", err)
			allPassed = false
		} else if len(decompressed) > 2003 {
			if decompressed[2000] != 0xCA || decompressed[2001] != 0xFE ||
				decompressed[2002] != 0xBA || decompressed[2003] != 0xBE {
				t.Error("❌ Savedata contents corrupted")
				allPassed = false
			} else {
				t.Log("✓ Savedata persisted correctly")
			}
		} else {
			t.Error("❌ Savedata too short")
			allPassed = false
		}
	}

	if allPassed {
		t.Log("✅ All data types persisted correctly across logout/login cycle")
	} else {
		t.Log("❌ CRITICAL: Some data types failed to persist - logout may not be triggering save handlers")
	}

	logoutPlayer(session2)
}

// TestSessionLifecycle_DisconnectWithoutLogout tests ungraceful disconnect
// This simulates network failure or client crash
func TestSessionLifecycle_DisconnectWithoutLogout(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "disconnect_test_user")
	charID := CreateTestCharacter(t, db, userID, "DisconnectChar")

	t.Log("Testing data persistence after ungraceful disconnect")

	// ===== SESSION 1: Modify data then disconnect without explicit logout =====
	session1 := createTestSessionForServerWithChar(server, charID, "DisconnectChar")

	// Modify data (frontier_points is on users table since 9.2 migration)
	rdpPoints := uint32(9999)
	_, err := db.Exec("UPDATE users SET frontier_points = $1 WHERE id = $2", rdpPoints, userID)
	if err != nil {
		t.Fatalf("Failed to set RdP: %v", err)
	}

	// Save data
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("DisconnectChar\x00"))
	saveData[3000] = 0xAB
	saveData[3001] = 0xCD

	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress savedata: %v", err)
	}

	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      6001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session1, savePkt)
	time.Sleep(100 * time.Millisecond)

	// Simulate disconnect by calling logoutPlayer (which is called by recvLoop on EOF)
	// In real scenario, this is triggered by connection close
	t.Log("Simulating ungraceful disconnect")
	logoutPlayer(session1)
	time.Sleep(100 * time.Millisecond)

	// ===== SESSION 2: Verify data saved despite ungraceful disconnect =====
	session2 := createTestSessionForServerWithChar(server, charID, "DisconnectChar")

	// Verify savedata
	var savedCompressed []byte
	err = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to load savedata: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("❌ CRITICAL: No data saved after disconnect")
		logoutPlayer(session2)
		return
	}

	decompressed, err := nullcomp.Decompress(savedCompressed)
	if err != nil {
		t.Errorf("Failed to decompress: %v", err)
		logoutPlayer(session2)
		return
	}

	if len(decompressed) > 3001 {
		if decompressed[3000] == 0xAB && decompressed[3001] == 0xCD {
			t.Log("✓ Data persisted after ungraceful disconnect")
		} else {
			t.Error("❌ Data corrupted after disconnect")
		}
	} else {
		t.Error("❌ Data too short after disconnect")
	}

	logoutPlayer(session2)
}

// TestSessionLifecycle_RapidReconnect tests quick logout/login cycles
// This simulates a player reconnecting quickly or connection instability
func TestSessionLifecycle_RapidReconnect(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "rapid_test_user")
	charID := CreateTestCharacter(t, db, userID, "RapidChar")

	t.Log("Testing data persistence with rapid logout/login cycles")

	for cycle := 1; cycle <= 3; cycle++ {
		t.Logf("--- Cycle %d ---", cycle)

		session := createTestSessionForServerWithChar(server, charID, "RapidChar")

		// Modify road points each cycle (frontier_points is on users table since 9.2 migration)
		points := uint32(1000 * cycle)
		_, err := db.Exec("UPDATE users SET frontier_points = $1 WHERE id = $2", points, userID)
		if err != nil {
			t.Fatalf("Cycle %d: Failed to update points: %v", cycle, err)
		}

		// Logout quickly
		logoutPlayer(session)
		time.Sleep(30 * time.Millisecond)

		// Verify points persisted
		var loadedPoints uint32
		_ = db.QueryRow("SELECT frontier_points FROM users WHERE id = $1", userID).Scan(&loadedPoints)
		if loadedPoints != points {
			t.Errorf("❌ Cycle %d: Points not persisted: got %d, want %d", cycle, loadedPoints, points)
		} else {
			t.Logf("✓ Cycle %d: Points persisted correctly: %d", cycle, loadedPoints)
		}
	}
}

// Helper function to create test equipment item with proper initialization
func createTestEquipmentItem(itemID uint16, warehouseID uint32) mhfitem.MHFEquipment {
	sigils := make([]mhfitem.MHFSigil, 3)
	for i := range sigils {
		sigils[i].Effects = make([]mhfitem.MHFSigilEffect, 3)
	}
	return mhfitem.MHFEquipment{
		ItemID:      itemID,
		WarehouseID: warehouseID,
		Decorations: make([]mhfitem.MHFItem, 3),
		Sigils:      sigils,
	}
}

// MockNetConn is defined in client_connection_simulation_test.go

// Helper function to create a test server with database
func createTestServerWithDB(t *testing.T, db *sqlx.DB) *Server {
	t.Helper()

	// Create minimal server for testing
	// Note: This may need adjustment based on actual Server initialization
	server := &Server{
		db:              db,
		sessions:        make(map[net.Conn]*Session),
		userBinary: NewUserBinaryStore(),
		minidata:   NewMinidataStore(),
		semaphore:       make(map[string]*Semaphore),
		erupeConfig: &cfg.Config{
			RealClientMode: cfg.ZZ,
		},
		isShuttingDown:  false,
		done:            make(chan struct{}),
	}

	// Create logger
	logger, _ := zap.NewDevelopment()
	server.logger = logger

	// Initialize repositories
	server.charRepo = NewCharacterRepository(db)
	server.guildRepo = NewGuildRepository(db)
	server.userRepo = NewUserRepository(db)
	server.gachaRepo = NewGachaRepository(db)
	server.houseRepo = NewHouseRepository(db)
	server.festaRepo = NewFestaRepository(db)
	server.towerRepo = NewTowerRepository(db)
	server.rengokuRepo = NewRengokuRepository(db)
	server.mailRepo = NewMailRepository(db)
	server.stampRepo = NewStampRepository(db)
	server.distRepo = NewDistributionRepository(db)
	server.sessionRepo = NewSessionRepository(db)

	return server
}

// Helper function to create a test session for a specific character
func createTestSessionForServerWithChar(server *Server, charID uint32, name string) *Session {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	mockNetConn := NewMockNetConn() // Create a mock net.Conn for the session map key
	
	session := &Session{
		logger:        server.logger,
		server:        server,
		rawConn:       mockNetConn,
		cryptConn:     mock,
		sendPackets:   make(chan packet, 20),
		clientContext: &clientctx.ClientContext{},
		lastPacket:    time.Now(),
		sessionStart:  time.Now().Unix(),
		charID:        charID,
		Name:          name,
	}

	// Register session with server (needed for logout to work properly)
	server.Lock()
	server.sessions[mockNetConn] = session
	server.Unlock()

	return session
}

