package channelserver

import (
	"fmt"
	"testing"

	_config "erupe-ce/config"
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// TestHandleMsgSysEnumerateClient tests client enumeration in stages
func TestHandleMsgSysEnumerateClient(t *testing.T) {
	tests := []struct {
		name              string
		stageID           string
		getType           uint8
		setupStage        func(*Server, string)
		wantClientCount   int
		wantFailure       bool
	}{
		{
			name:    "enumerate_all_clients",
			stageID: "test_stage_1",
			getType: 0, // All clients
			setupStage: func(server *Server, stageID string) {
				stage := NewStage(stageID)
				mock1 := &MockCryptConn{sentPackets: make([][]byte, 0)}
				mock2 := &MockCryptConn{sentPackets: make([][]byte, 0)}
				s1 := createTestSession(mock1)
				s2 := createTestSession(mock2)
				s1.charID = 100
				s2.charID = 200
				stage.clients[s1] = 100
				stage.clients[s2] = 200
				server.stagesLock.Lock()
				server.stages[stageID] = stage
				server.stagesLock.Unlock()
			},
			wantClientCount: 2,
			wantFailure:     false,
		},
		{
			name:    "enumerate_not_ready_clients",
			stageID: "test_stage_2",
			getType: 1, // Not ready
			setupStage: func(server *Server, stageID string) {
				stage := NewStage(stageID)
				stage.reservedClientSlots[100] = false // Not ready
				stage.reservedClientSlots[200] = true  // Ready
				stage.reservedClientSlots[300] = false // Not ready
				server.stagesLock.Lock()
				server.stages[stageID] = stage
				server.stagesLock.Unlock()
			},
			wantClientCount: 2, // Only not-ready clients
			wantFailure:     false,
		},
		{
			name:    "enumerate_ready_clients",
			stageID: "test_stage_3",
			getType: 2, // Ready
			setupStage: func(server *Server, stageID string) {
				stage := NewStage(stageID)
				stage.reservedClientSlots[100] = false // Not ready
				stage.reservedClientSlots[200] = true  // Ready
				stage.reservedClientSlots[300] = true  // Ready
				server.stagesLock.Lock()
				server.stages[stageID] = stage
				server.stagesLock.Unlock()
			},
			wantClientCount: 2, // Only ready clients
			wantFailure:     false,
		},
		{
			name:    "enumerate_empty_stage",
			stageID: "test_stage_empty",
			getType: 0,
			setupStage: func(server *Server, stageID string) {
				stage := NewStage(stageID)
				server.stagesLock.Lock()
				server.stages[stageID] = stage
				server.stagesLock.Unlock()
			},
			wantClientCount: 0,
			wantFailure:     false,
		},
		{
			name:    "enumerate_nonexistent_stage",
			stageID: "nonexistent_stage",
			getType: 0,
			setupStage: func(server *Server, stageID string) {
				// Don't create the stage
			},
			wantClientCount: 0,
			wantFailure:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test session (which creates a server with erupeConfig)
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)

			// Initialize stages map if needed
			if s.server.stages == nil {
				s.server.stages = make(map[string]*Stage)
			}

			// Setup stage
			tt.setupStage(s.server, tt.stageID)

			pkt := &mhfpacket.MsgSysEnumerateClient{
				AckHandle: 1234,
				StageID:   tt.stageID,
				Get:       tt.getType,
			}

			handleMsgSysEnumerateClient(s, pkt)

			// Check if ACK was sent
			if len(s.sendPackets) == 0 {
				t.Fatal("no ACK packet was sent")
			}

			// Read the ACK packet
			ackPkt := <-s.sendPackets
			if tt.wantFailure {
				// For failures, we can't easily check the exact format
				// Just verify something was sent
				return
			}

			// Parse the response to count clients
			// The ackPkt.data contains the full packet structure:
			// [opcode:2 bytes][ack_handle:4 bytes][is_buffer:1 byte][error_code:1 byte][payload_size:2 bytes][data...]
			// Total header size: 2 + 4 + 1 + 1 + 2 = 10 bytes
			if len(ackPkt.data) < 10 {
				t.Fatal("ACK packet too small")
			}

			// The response data starts after the 10-byte header
			// Response format is: [count:uint16][charID1:uint32][charID2:uint32]...
			bf := byteframe.NewByteFrameFromBytes(ackPkt.data[10:]) // Skip full ACK header
			count := bf.ReadUint16()

			if int(count) != tt.wantClientCount {
				t.Errorf("client count = %d, want %d", count, tt.wantClientCount)
			}
		})
	}
}

// TestHandleMsgMhfListMember tests listing blacklisted members
func TestHandleMsgMhfListMember_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tests := []struct {
		name           string
		blockedCSV     string
		wantBlockCount int
	}{
		{
			name:           "no_blocked_users",
			blockedCSV:     "",
			wantBlockCount: 0,
		},
		{
			name:           "single_blocked_user",
			blockedCSV:     "2",
			wantBlockCount: 1,
		},
		{
			name:           "multiple_blocked_users",
			blockedCSV:     "2,3,4",
			wantBlockCount: 3,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user and character (use short names to avoid 15 char limit)
			userID := CreateTestUser(t, db, "user_"+tt.name)
			charName := fmt.Sprintf("Char%d", i)
			charID := CreateTestCharacter(t, db, userID, charName)

			// Create blocked characters
			if tt.blockedCSV != "" {
				// Create the blocked users
				for i := 2; i <= 4; i++ {
					blockedUserID := CreateTestUser(t, db, "blocked_user_"+tt.name+"_"+string(rune(i)))
					CreateTestCharacter(t, db, blockedUserID, "BlockedChar_"+string(rune(i)))
				}
			}

			// Set blocked list
			_, err := db.Exec("UPDATE characters SET blocked = $1 WHERE id = $2", tt.blockedCSV, charID)
			if err != nil {
				t.Fatalf("Failed to update blocked list: %v", err)
			}

			// Create test session
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = charID
			s.server.db = db

			pkt := &mhfpacket.MsgMhfListMember{
				AckHandle: 5678,
			}

			handleMsgMhfListMember(s, pkt)

			// Verify ACK was sent
			if len(s.sendPackets) == 0 {
				t.Fatal("no ACK packet was sent")
			}

			// Parse response
			// The ackPkt.data contains the full packet structure:
			// [opcode:2 bytes][ack_handle:4 bytes][is_buffer:1 byte][error_code:1 byte][payload_size:2 bytes][data...]
			// Total header size: 2 + 4 + 1 + 1 + 2 = 10 bytes
			ackPkt := <-s.sendPackets
			if len(ackPkt.data) < 10 {
				t.Fatal("ACK packet too small")
			}
			bf := byteframe.NewByteFrameFromBytes(ackPkt.data[10:]) // Skip full ACK header
			count := bf.ReadUint32()

			if int(count) != tt.wantBlockCount {
				t.Errorf("blocked count = %d, want %d", count, tt.wantBlockCount)
			}
		})
	}
}

// TestHandleMsgMhfOprMember tests blacklist/friendlist operations
func TestHandleMsgMhfOprMember_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tests := []struct {
		name          string
		isBlacklist   bool
		operation     bool // true = remove, false = add
		initialList   string
		targetCharIDs []uint32
		wantList      string
	}{
		{
			name:          "add_to_blacklist",
			isBlacklist:   true,
			operation:     false,
			initialList:   "",
			targetCharIDs: []uint32{2},
			wantList:      "2",
		},
		{
			name:          "remove_from_blacklist",
			isBlacklist:   true,
			operation:     true,
			initialList:   "2,3,4",
			targetCharIDs: []uint32{3},
			wantList:      "2,4",
		},
		{
			name:          "add_to_friendlist",
			isBlacklist:   false,
			operation:     false,
			initialList:   "10",
			targetCharIDs: []uint32{20},
			wantList:      "10,20",
		},
		{
			name:          "remove_from_friendlist",
			isBlacklist:   false,
			operation:     true,
			initialList:   "10,20,30",
			targetCharIDs: []uint32{20},
			wantList:      "10,30",
		},
		{
			name:          "add_multiple_to_blacklist",
			isBlacklist:   true,
			operation:     false,
			initialList:   "1",
			targetCharIDs: []uint32{2, 3},
			wantList:      "1,2,3",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user and character (use short names to avoid 15 char limit)
			userID := CreateTestUser(t, db, "user_"+tt.name)
			charName := fmt.Sprintf("OpChar%d", i)
			charID := CreateTestCharacter(t, db, userID, charName)

			// Set initial list
			column := "blocked"
			if !tt.isBlacklist {
				column = "friends"
			}
			_, err := db.Exec("UPDATE characters SET "+column+" = $1 WHERE id = $2", tt.initialList, charID)
			if err != nil {
				t.Fatalf("Failed to set initial list: %v", err)
			}

			// Create test session
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = charID
			s.server.db = db

			pkt := &mhfpacket.MsgMhfOprMember{
				AckHandle: 9999,
				Blacklist: tt.isBlacklist,
				Operation: tt.operation,
				CharIDs:   tt.targetCharIDs,
			}

			handleMsgMhfOprMember(s, pkt)

			// Verify ACK was sent
			if len(s.sendPackets) == 0 {
				t.Fatal("no ACK packet was sent")
			}
			<-s.sendPackets

			// Verify the list was updated
			var gotList string
			err = db.QueryRow("SELECT "+column+" FROM characters WHERE id = $1", charID).Scan(&gotList)
			if err != nil {
				t.Fatalf("Failed to query updated list: %v", err)
			}

			if gotList != tt.wantList {
				t.Errorf("list = %q, want %q", gotList, tt.wantList)
			}
		})
	}
}

// TestHandleMsgMhfShutClient tests the shut client handler
func TestHandleMsgMhfShutClient(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	pkt := &mhfpacket.MsgMhfShutClient{}

	// Should not panic (handler is empty)
	handleMsgMhfShutClient(s, pkt)
}

// TestHandleMsgSysHideClient tests the hide client handler
func TestHandleMsgSysHideClient(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	tests := []struct {
		name string
		hide bool
	}{
		{
			name: "hide_client",
			hide: true,
		},
		{
			name: "show_client",
			hide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := &mhfpacket.MsgSysHideClient{
				Hide: tt.hide,
			}

			// Should not panic (handler is empty)
			handleMsgSysHideClient(s, pkt)
		})
	}
}

// TestEnumerateClient_ConcurrentAccess tests concurrent stage access
func TestEnumerateClient_ConcurrentAccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	server := &Server{
		logger: logger,
		stages: make(map[string]*Stage),
		erupeConfig: &_config.Config{
			DebugOptions: _config.DebugOptions{
				LogOutboundMessages: false,
			},
		},
	}

	stageID := "concurrent_test_stage"
	stage := NewStage(stageID)

	// Add some clients to the stage
	for i := uint32(1); i <= 10; i++ {
		mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
		sess := createTestSession(mock)
		sess.charID = i * 100
		stage.clients[sess] = i * 100
	}

	server.stagesLock.Lock()
	server.stages[stageID] = stage
	server.stagesLock.Unlock()

	// Run concurrent enumerations
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.server = server

			pkt := &mhfpacket.MsgSysEnumerateClient{
				AckHandle: 3333,
				StageID:   stageID,
				Get:       0, // All clients
			}

			handleMsgSysEnumerateClient(s, pkt)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestListMember_EmptyDatabase tests listing members when database is empty
func TestListMember_EmptyDatabase_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// Create test user and character
	userID := CreateTestUser(t, db, "emptytest")
	charID := CreateTestCharacter(t, db, userID, "EmptyChar")

	// Create test session
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	s.server.db = db

	pkt := &mhfpacket.MsgMhfListMember{
		AckHandle: 4444,
	}

	handleMsgMhfListMember(s, pkt)

	// Verify ACK was sent
	if len(s.sendPackets) == 0 {
		t.Fatal("no ACK packet was sent")
	}

	ackPkt := <-s.sendPackets
	if len(ackPkt.data) < 10 {
		t.Fatal("ACK packet too small")
	}
	bf := byteframe.NewByteFrameFromBytes(ackPkt.data[10:]) // Skip full ACK header
	count := bf.ReadUint32()

	if count != 0 {
		t.Errorf("empty blocked list should have count 0, got %d", count)
	}
}

// TestOprMember_EdgeCases tests edge cases for member operations
func TestOprMember_EdgeCases_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tests := []struct {
		name          string
		initialList   string
		operation     bool
		targetCharIDs []uint32
		wantList      string
	}{
		{
			name:          "add_duplicate_to_list",
			initialList:   "1,2,3",
			operation:     false, // add
			targetCharIDs: []uint32{2},
			wantList:      "1,2,3", // CSV helper deduplicates
		},
		{
			name:          "remove_nonexistent_from_list",
			initialList:   "1,2,3",
			operation:     true, // remove
			targetCharIDs: []uint32{99},
			wantList:      "1,2,3",
		},
		{
			name:          "operate_on_empty_list",
			initialList:   "",
			operation:     false,
			targetCharIDs: []uint32{1},
			wantList:      "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user and character
			userID := CreateTestUser(t, db, "edge_"+tt.name)
			charID := CreateTestCharacter(t, db, userID, "EdgeChar")

			// Set initial blocked list
			_, err := db.Exec("UPDATE characters SET blocked = $1 WHERE id = $2", tt.initialList, charID)
			if err != nil {
				t.Fatalf("Failed to set initial list: %v", err)
			}

			// Create test session
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = charID
			s.server.db = db

			pkt := &mhfpacket.MsgMhfOprMember{
				AckHandle: 7777,
				Blacklist: true,
				Operation: tt.operation,
				CharIDs:   tt.targetCharIDs,
			}

			handleMsgMhfOprMember(s, pkt)

			// Verify ACK was sent
			if len(s.sendPackets) == 0 {
				t.Fatal("no ACK packet was sent")
			}
			<-s.sendPackets

			// Verify the list
			var gotList string
			err = db.QueryRow("SELECT blocked FROM characters WHERE id = $1", charID).Scan(&gotList)
			if err != nil {
				t.Fatalf("Failed to query list: %v", err)
			}

			if gotList != tt.wantList {
				t.Errorf("list = %q, want %q", gotList, tt.wantList)
			}
		})
	}
}

// BenchmarkEnumerateClients benchmarks client enumeration
func BenchmarkEnumerateClients(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	server := &Server{
		logger: logger,
		stages: make(map[string]*Stage),
	}

	stageID := "bench_stage"
	stage := NewStage(stageID)

	// Add 100 clients to the stage
	for i := uint32(1); i <= 100; i++ {
		mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
		sess := createTestSession(mock)
		sess.charID = i
		stage.clients[sess] = i
	}

	server.stages[stageID] = stage

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.server = server

	pkt := &mhfpacket.MsgSysEnumerateClient{
		AckHandle: 8888,
		StageID:   stageID,
		Get:       0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear the packet channel
		select {
		case <-s.sendPackets:
		default:
		}

		handleMsgSysEnumerateClient(s, pkt)
		<-s.sendPackets
	}
}
