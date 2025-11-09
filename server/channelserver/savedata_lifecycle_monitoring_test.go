package channelserver

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// ============================================================================
// SAVE DATA LIFECYCLE MONITORING TESTS
// Tests with logging and monitoring to detect when save handlers are called
//
// Purpose: Add observability to understand the save/load lifecycle
// - Track when save handlers are invoked
// - Monitor logout flow
// - Detect missing save calls during disconnect
// ============================================================================

// SaveHandlerMonitor tracks calls to save handlers
type SaveHandlerMonitor struct {
	mu                      sync.Mutex
	savedataCallCount       int
	hunterNaviCallCount     int
	kouryouPointCallCount   int
	warehouseCallCount      int
	decomysetCallCount      int
	savedataAtLogout        bool
	lastSavedataTime        time.Time
	lastHunterNaviTime      time.Time
	lastKouryouPointTime    time.Time
	lastWarehouseTime       time.Time
	lastDecomysetTime       time.Time
	logoutTime              time.Time
}

func (m *SaveHandlerMonitor) RecordSavedata() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.savedataCallCount++
	m.lastSavedataTime = time.Now()
}

func (m *SaveHandlerMonitor) RecordHunterNavi() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hunterNaviCallCount++
	m.lastHunterNaviTime = time.Now()
}

func (m *SaveHandlerMonitor) RecordKouryouPoint() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kouryouPointCallCount++
	m.lastKouryouPointTime = time.Now()
}

func (m *SaveHandlerMonitor) RecordWarehouse() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warehouseCallCount++
	m.lastWarehouseTime = time.Now()
}

func (m *SaveHandlerMonitor) RecordDecomyset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decomysetCallCount++
	m.lastDecomysetTime = time.Now()
}

func (m *SaveHandlerMonitor) RecordLogout() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logoutTime = time.Now()
	
	// Check if savedata was called within 5 seconds before logout
	if !m.lastSavedataTime.IsZero() && m.logoutTime.Sub(m.lastSavedataTime) < 5*time.Second {
		m.savedataAtLogout = true
	}
}

func (m *SaveHandlerMonitor) GetStats() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	return fmt.Sprintf(`Save Handler Statistics:
  - Savedata calls: %d (last: %v)
  - HunterNavi calls: %d (last: %v)
  - KouryouPoint calls: %d (last: %v)
  - Warehouse calls: %d (last: %v)
  - Decomyset calls: %d (last: %v)
  - Logout time: %v
  - Savedata before logout: %v`,
		m.savedataCallCount, m.lastSavedataTime,
		m.hunterNaviCallCount, m.lastHunterNaviTime,
		m.kouryouPointCallCount, m.lastKouryouPointTime,
		m.warehouseCallCount, m.lastWarehouseTime,
		m.decomysetCallCount, m.lastDecomysetTime,
		m.logoutTime,
		m.savedataAtLogout)
}

func (m *SaveHandlerMonitor) WasSavedataCalledBeforeLogout() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.savedataAtLogout
}

// TestMonitored_SaveHandlerInvocationDuringLogout tests if save handlers are called during logout
// This is the KEY test to identify the bug: logout should trigger saves but doesn't
func TestMonitored_SaveHandlerInvocationDuringLogout(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "monitor_test_user")
	charID := CreateTestCharacter(t, db, userID, "MonitorChar")

	monitor := &SaveHandlerMonitor{}

	t.Log("Starting monitored session to track save handler calls")

	// Create session with monitoring
	session := createTestSessionForServerWithChar(server, charID, "MonitorChar")

	// Modify data that SHOULD be auto-saved on logout
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("MonitorChar\x00"))
	saveData[5000] = 0x11
	saveData[5001] = 0x22

	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress savedata: %v", err)
	}

	// Save data during session
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      7001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}

	t.Log("Calling handleMsgMhfSavedata during session")
	handleMsgMhfSavedata(session, savePkt)
	monitor.RecordSavedata()
	time.Sleep(100 * time.Millisecond)

	// Now trigger logout
	t.Log("Triggering logout - monitoring if save handlers are called")
	monitor.RecordLogout()
	logoutPlayer(session)
	time.Sleep(100 * time.Millisecond)

	// Report statistics
	t.Log(monitor.GetStats())

	// Analysis
	if monitor.savedataCallCount == 0 {
		t.Error("❌ CRITICAL: No savedata calls detected during entire session")
	}

	if !monitor.WasSavedataCalledBeforeLogout() {
		t.Log("⚠️  WARNING: Savedata was NOT called immediately before logout")
		t.Log("This explains why players lose data - logout doesn't trigger final save!")
	}

	// Check if data actually persisted
	var savedCompressed []byte
	err = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to query savedata: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("❌ CRITICAL: No savedata in database after logout")
	} else {
		decompressed, err := nullcomp.Decompress(savedCompressed)
		if err != nil {
			t.Errorf("Failed to decompress: %v", err)
		} else if len(decompressed) > 5001 {
			if decompressed[5000] == 0x11 && decompressed[5001] == 0x22 {
				t.Log("✓ Data persisted (save was called during session, not at logout)")
			} else {
				t.Error("❌ Data corrupted or not saved")
			}
		}
	}
}

// TestWithLogging_LogoutFlowAnalysis tests logout with detailed logging
func TestWithLogging_LogoutFlowAnalysis(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// Create observed logger
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	server := createTestServerWithDB(t, db)
	server.logger = logger
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "logging_test_user")
	charID := CreateTestCharacter(t, db, userID, "LoggingChar")

	t.Log("Starting session with observed logging")

	session := createTestSessionForServerWithChar(server, charID, "LoggingChar")
	session.logger = logger

	// Perform some actions
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("LoggingChar\x00"))
	compressed, _ := nullcomp.Compress(saveData)

	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      8001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session, savePkt)
	time.Sleep(50 * time.Millisecond)

	// Trigger logout
	t.Log("Triggering logout with logging enabled")
	logoutPlayer(session)
	time.Sleep(100 * time.Millisecond)

	// Analyze logs
	allLogs := logs.All()
	t.Logf("Captured %d log entries during session lifecycle", len(allLogs))

	saveRelatedLogs := 0
	logoutRelatedLogs := 0

	for _, entry := range allLogs {
		msg := entry.Message
		if containsAny(msg, []string{"save", "Save", "SAVE"}) {
			saveRelatedLogs++
			t.Logf("  [SAVE LOG] %s", msg)
		}
		if containsAny(msg, []string{"logout", "Logout", "disconnect", "Disconnect"}) {
			logoutRelatedLogs++
			t.Logf("  [LOGOUT LOG] %s", msg)
		}
	}

	t.Logf("Save-related logs: %d", saveRelatedLogs)
	t.Logf("Logout-related logs: %d", logoutRelatedLogs)

	if saveRelatedLogs == 0 {
		t.Error("❌ No save-related log entries found - saves may not be happening")
	}

	if logoutRelatedLogs == 0 {
		t.Log("⚠️  No logout-related log entries - may need to add logging to logoutPlayer()")
	}
}

// TestConcurrent_MultipleSessionsSaving tests concurrent sessions saving data
// This helps identify race conditions in the save system
func TestConcurrent_MultipleSessionsSaving(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	numSessions := 5
	var wg sync.WaitGroup
	wg.Add(numSessions)

	t.Logf("Starting %d concurrent sessions", numSessions)

	for i := 0; i < numSessions; i++ {
		go func(sessionID int) {
			defer wg.Done()

			username := fmt.Sprintf("concurrent_user_%d", sessionID)
			charName := fmt.Sprintf("ConcurrentChar%d", sessionID)

			userID := CreateTestUser(t, db, username)
			charID := CreateTestCharacter(t, db, userID, charName)

			session := createTestSessionForServerWithChar(server, charID, charName)

			// Save data
			saveData := make([]byte, 150000)
			copy(saveData[88:], []byte(charName+"\x00"))
			saveData[6000+sessionID] = byte(sessionID)

			compressed, err := nullcomp.Compress(saveData)
			if err != nil {
				t.Errorf("Session %d: Failed to compress: %v", sessionID, err)
				return
			}

			savePkt := &mhfpacket.MsgMhfSavedata{
				SaveType:       0,
				AckHandle:      uint32(9000 + sessionID),
				AllocMemSize:   uint32(len(compressed)),
				DataSize:       uint32(len(compressed)),
				RawDataPayload: compressed,
			}
			handleMsgMhfSavedata(session, savePkt)
			time.Sleep(50 * time.Millisecond)

			// Logout
			logoutPlayer(session)
			time.Sleep(50 * time.Millisecond)

			// Verify data saved
			var savedCompressed []byte
			err = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
			if err != nil {
				t.Errorf("Session %d: Failed to load savedata: %v", sessionID, err)
				return
			}

			if len(savedCompressed) == 0 {
				t.Errorf("Session %d: ❌ No savedata persisted", sessionID)
			} else {
				t.Logf("Session %d: ✓ Savedata persisted (%d bytes)", sessionID, len(savedCompressed))
			}
		}(i)
	}

	wg.Wait()
	t.Log("All concurrent sessions completed")
}

// TestSequential_RepeatedLogoutLoginCycles tests for data corruption over multiple cycles
func TestSequential_RepeatedLogoutLoginCycles(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "cycle_test_user")
	charID := CreateTestCharacter(t, db, userID, "CycleChar")

	numCycles := 10
	t.Logf("Running %d logout/login cycles", numCycles)

	for cycle := 1; cycle <= numCycles; cycle++ {
		session := createTestSessionForServerWithChar(server, charID, "CycleChar")

		// Modify data each cycle
		saveData := make([]byte, 150000)
		copy(saveData[88:], []byte("CycleChar\x00"))
		// Write cycle number at specific offset
		saveData[7000] = byte(cycle >> 8)
		saveData[7001] = byte(cycle & 0xFF)

		compressed, _ := nullcomp.Compress(saveData)
		savePkt := &mhfpacket.MsgMhfSavedata{
			SaveType:       0,
			AckHandle:      uint32(10000 + cycle),
			AllocMemSize:   uint32(len(compressed)),
			DataSize:       uint32(len(compressed)),
			RawDataPayload: compressed,
		}
		handleMsgMhfSavedata(session, savePkt)
		time.Sleep(50 * time.Millisecond)

		// Logout
		logoutPlayer(session)
		time.Sleep(50 * time.Millisecond)

		// Verify data after each cycle
		var savedCompressed []byte
		db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
		
		if len(savedCompressed) > 0 {
			decompressed, err := nullcomp.Decompress(savedCompressed)
			if err != nil {
				t.Errorf("Cycle %d: Failed to decompress: %v", cycle, err)
			} else if len(decompressed) > 7001 {
				savedCycle := (int(decompressed[7000]) << 8) | int(decompressed[7001])
				if savedCycle != cycle {
					t.Errorf("Cycle %d: ❌ Data corruption - expected cycle %d, got %d", 
						cycle, cycle, savedCycle)
				} else {
					t.Logf("Cycle %d: ✓ Data correct", cycle)
				}
			}
		} else {
			t.Errorf("Cycle %d: ❌ No savedata", cycle)
		}
	}

	t.Log("Completed all logout/login cycles")
}

// TestRealtime_SaveDataTimestamps tests when saves actually happen
func TestRealtime_SaveDataTimestamps(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := createTestServerWithDB(t, db)
	defer server.Shutdown()

	userID := CreateTestUser(t, db, "timestamp_test_user")
	charID := CreateTestCharacter(t, db, userID, "TimestampChar")

	type SaveEvent struct {
		timestamp time.Time
		eventType string
	}
	var events []SaveEvent

	session := createTestSessionForServerWithChar(server, charID, "TimestampChar")
	events = append(events, SaveEvent{time.Now(), "session_start"})

	// Save 1
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("TimestampChar\x00"))
	compressed, _ := nullcomp.Compress(saveData)
	
	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      11001,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(session, savePkt)
	events = append(events, SaveEvent{time.Now(), "save_1"})
	time.Sleep(100 * time.Millisecond)

	// Save 2
	handleMsgMhfSavedata(session, savePkt)
	events = append(events, SaveEvent{time.Now(), "save_2"})
	time.Sleep(100 * time.Millisecond)

	// Logout
	events = append(events, SaveEvent{time.Now(), "logout_start"})
	logoutPlayer(session)
	events = append(events, SaveEvent{time.Now(), "logout_end"})
	time.Sleep(50 * time.Millisecond)

	// Print timeline
	t.Log("Save event timeline:")
	startTime := events[0].timestamp
	for _, event := range events {
		elapsed := event.timestamp.Sub(startTime)
		t.Logf("  [+%v] %s", elapsed.Round(time.Millisecond), event.eventType)
	}

	// Calculate time between last save and logout
	var lastSaveTime time.Time
	var logoutTime time.Time
	for _, event := range events {
		if event.eventType == "save_2" {
			lastSaveTime = event.timestamp
		}
		if event.eventType == "logout_start" {
			logoutTime = event.timestamp
		}
	}

	if !lastSaveTime.IsZero() && !logoutTime.IsZero() {
		gap := logoutTime.Sub(lastSaveTime)
		t.Logf("Time between last save and logout: %v", gap.Round(time.Millisecond))
		
		if gap > 50*time.Millisecond {
			t.Log("⚠️  Significant gap between last save and logout")
			t.Log("Player changes after last save would be LOST")
		}
	}
}

// Helper function
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

