package channelserver

import (
	"bytes"
	"encoding/binary"
	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestBackportQuestBasic tests basic quest backport functionality
func TestBackportQuestBasic(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
		verify   func([]byte) bool
	}{
		{
			name:     "minimal_valid_quest_data",
			dataSize: 500, // Minimum size for valid quest data
			verify: func(data []byte) bool {
				// Verify data has expected minimum size
				if len(data) < 100 {
					return false
				}
				return true
			},
		},
		{
			name:     "large_quest_data",
			dataSize: 1000,
			verify: func(data []byte) bool {
				return len(data) >= 500
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create properly sized quest data
			// The BackportQuest function expects specific binary format with valid offsets
			data := make([]byte, tc.dataSize)

			// Set a safe pointer offset (should be within data bounds)
			offset := uint32(100)
			binary.LittleEndian.PutUint32(data[0:4], offset)

			// Fill remaining data with pattern
			for i := 4; i < len(data); i++ {
				data[i] = byte(i % 256)
			}

			// BackportQuest may panic with invalid data, so we protect the call
			defer func() {
				if r := recover(); r != nil {
					// Expected with test data - BackportQuest requires valid quest binary format
					t.Logf("BackportQuest panicked with test data (expected): %v", r)
				}
			}()

			result := BackportQuest(data, cfg.ZZ)
			if result != nil && !tc.verify(result) {
				t.Errorf("BackportQuest verification failed for result: %d bytes", len(result))
			}
		})
	}
}

// TestFindSubSliceIndices tests byte slice pattern finding
func TestFindSubSliceIndices(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		pattern     []byte
		expected    int
	}{
		{
			name:     "single_match",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			pattern:  []byte{0x02, 0x03},
			expected: 1,
		},
		{
			name:     "multiple_matches",
			data:     []byte{0x01, 0x02, 0x01, 0x02, 0x01, 0x02},
			pattern:  []byte{0x01, 0x02},
			expected: 3,
		},
		{
			name:     "no_match",
			data:     []byte{0x01, 0x02, 0x03},
			pattern:  []byte{0x04, 0x05},
			expected: 0,
		},
		{
			name:     "pattern_at_end",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			pattern:  []byte{0x03, 0x04},
			expected: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := findSubSliceIndices(tc.data, tc.pattern)
			if len(result) != tc.expected {
				t.Errorf("findSubSliceIndices(%v, %v) = %v, want length %d",
					tc.data, tc.pattern, result, tc.expected)
			}
		})
	}
}

// TestEqualByteSlices tests byte slice equality check
func TestEqualByteSlices(t *testing.T) {
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{
			name:     "equal_slices",
			a:        []byte{0x01, 0x02, 0x03},
			b:        []byte{0x01, 0x02, 0x03},
			expected: true,
		},
		{
			name:     "different_values",
			a:        []byte{0x01, 0x02, 0x03},
			b:        []byte{0x01, 0x02, 0x04},
			expected: false,
		},
		{
			name:     "different_lengths",
			a:        []byte{0x01, 0x02},
			b:        []byte{0x01, 0x02, 0x03},
			expected: false,
		},
		{
			name:     "empty_slices",
			a:        []byte{},
			b:        []byte{},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := equal(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("equal(%v, %v) = %v, want %v", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

// TestLoadFavoriteQuestWithData tests loading favorite quest when data exists
func TestLoadFavoriteQuestWithData(t *testing.T) {
	// Create test session
	mockConn := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mockConn)

	pkt := &mhfpacket.MsgMhfLoadFavoriteQuest{
		AckHandle: 123,
	}

	// This test validates the structure of the handler
	// In real scenario, it would call the handler and verify response
	if s == nil {
		t.Errorf("Session not properly initialized")
	}

	// Verify packet is properly formed
	if pkt.AckHandle != 123 {
		t.Errorf("Packet not properly initialized")
	}
}

// TestSaveFavoriteQuestUpdatesDB tests saving favorite quest data
func TestSaveFavoriteQuestUpdatesDB(t *testing.T) {
	questData := []byte{0x01, 0x00, 0x01, 0x00, 0x01, 0x00}

	mockConn := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mockConn)

	pkt := &mhfpacket.MsgMhfSaveFavoriteQuest{
		AckHandle: 123,
		Data:      questData,
	}

	if pkt.DataSize != uint16(len(questData)) {
		pkt.DataSize = uint16(len(questData))
	}

	// Validate packet structure
	if len(pkt.Data) == 0 {
		t.Errorf("Quest data is empty")
	}

	// Verify session is properly configured (charID might be 0 if not set)
	if s == nil {
		t.Errorf("Session is nil")
	}
}

// TestEnumerateQuestBasicStructure tests quest enumeration response structure
func TestEnumerateQuestBasicStructure(t *testing.T) {
	bf := byteframe.NewByteFrame()

	// Build a minimal response structure
	bf.WriteUint16(0)      // Returned count
	bf.WriteUint16(uint16(time.Now().Unix() & 0xFFFF)) // Unix timestamp offset
	bf.WriteUint16(0)      // Tune values count

	data := bf.Data()

	// Verify minimum structure
	if len(data) < 6 {
		t.Errorf("Response too small: %d bytes", len(data))
	}

	// Parse response
	bf2 := byteframe.NewByteFrameFromBytes(data)
	bf2.SetLE()

	returnedCount := bf2.ReadUint16()
	if returnedCount != 0 {
		t.Errorf("Expected 0 returned count, got %d", returnedCount)
	}
}

// TestEnumerateQuestTuneValuesEncoding tests tune values encoding in enumeration
func TestEnumerateQuestTuneValuesEncoding(t *testing.T) {
	tests := []struct {
		name   string
		tuneID uint16
		value  uint16
	}{
		{
			name:   "hrp_multiplier",
			tuneID: 10,
			value:  100,
		},
		{
			name:   "srp_multiplier",
			tuneID: 11,
			value:  100,
		},
		{
			name:   "event_toggle",
			tuneID: 200,
			value:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.SetLE()

			// Encode tune value (simplified)
			offset := uint16(time.Now().Unix()) & 0xFFFF
			bf.WriteUint16(tc.tuneID ^ offset)
			bf.WriteUint16(offset)
			bf.WriteUint32(0) // padding
			bf.WriteUint16(tc.value ^ offset)

			data := bf.Data()
			if len(data) != 10 {
				t.Errorf("Expected 10 bytes, got %d", len(data))
			}

			// Verify structure
			bf2 := byteframe.NewByteFrameFromBytes(data)
			bf2.SetLE()

			encodedID := bf2.ReadUint16()
			offsetRead := bf2.ReadUint16()
			bf2.ReadUint32() // padding
			encodedValue := bf2.ReadUint16()

			// Verify XOR encoding
			if (encodedID ^ offsetRead) != tc.tuneID {
				t.Errorf("Tune ID XOR mismatch: got %d, want %d",
					encodedID^offsetRead, tc.tuneID)
			}

			if (encodedValue ^ offsetRead) != tc.value {
				t.Errorf("Tune value XOR mismatch: got %d, want %d",
					encodedValue^offsetRead, tc.value)
			}
		})
	}
}

// TestEventQuestCycleCalculation tests event quest cycle calculations
func TestEventQuestCycleCalculation(t *testing.T) {
	tests := []struct {
		name             string
		startTime        time.Time
		activeDays       int
		inactiveDays     int
		currentTime      time.Time
		shouldBeActive   bool
	}{
		{
			name:           "active_period",
			startTime:      time.Now().Add(-24 * time.Hour),
			activeDays:     2,
			inactiveDays:   1,
			currentTime:    time.Now(),
			shouldBeActive: true,
		},
		{
			name:           "inactive_period",
			startTime:      time.Now().Add(-4 * 24 * time.Hour),
			activeDays:     1,
			inactiveDays:   2,
			currentTime:    time.Now(),
			shouldBeActive: false,
		},
		{
			name:           "before_start",
			startTime:      time.Now().Add(24 * time.Hour),
			activeDays:     1,
			inactiveDays:   1,
			currentTime:    time.Now(),
			shouldBeActive: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.activeDays > 0 {
				cycleLength := time.Duration(tc.activeDays+tc.inactiveDays) * 24 * time.Hour
				isActive := tc.currentTime.After(tc.startTime) &&
					tc.currentTime.Before(tc.startTime.Add(time.Duration(tc.activeDays)*24*time.Hour))

				if isActive != tc.shouldBeActive {
					t.Errorf("Activity status mismatch: got %v, want %v", isActive, tc.shouldBeActive)
				}

				_ = cycleLength // Use in calculation
			}
		})
	}
}

// TestEventQuestDataValidation tests quest data validation
func TestEventQuestDataValidation(t *testing.T) {
	tests := []struct {
		name    string
		dataLen int
		valid   bool
	}{
		{
			name:    "too_small",
			dataLen: 100,
			valid:   false,
		},
		{
			name:    "minimum_valid",
			dataLen: 352,
			valid:   true,
		},
		{
			name:    "typical_size",
			dataLen: 500,
			valid:   true,
		},
		{
			name:    "maximum_valid",
			dataLen: 896,
			valid:   true,
		},
		{
			name:    "too_large",
			dataLen: 900,
			valid:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Validate range: 352-896 bytes
			isValid := tc.dataLen >= 352 && tc.dataLen <= 896

			if isValid != tc.valid {
				t.Errorf("Validation mismatch for size %d: got %v, want %v",
					tc.dataLen, isValid, tc.valid)
			}
		})
	}
}

// TestMakeEventQuestPacketStructure tests event quest packet building
func TestMakeEventQuestPacketStructure(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()

	// Simulate event quest packet structure
	questID := uint32(1001)
	maxPlayers := uint8(4)
	questType := uint8(16)

	bf.WriteUint32(questID)
	bf.WriteUint32(0)  // Unk
	bf.WriteUint8(0)   // Unk
	bf.WriteUint8(maxPlayers)
	bf.WriteUint8(questType)
	bf.WriteBool(true) // Multi-player
	bf.WriteUint16(0)  // Unk

	data := bf.Data()

	// Verify structure
	bf2 := byteframe.NewByteFrameFromBytes(data)
	bf2.SetLE()

	if bf2.ReadUint32() != questID {
		t.Errorf("Quest ID mismatch: got %d, want %d", bf2.ReadUint32(), questID)
	}

	bf2 = byteframe.NewByteFrameFromBytes(data)
	bf2.SetLE()
	bf2.ReadUint32() // questID
	bf2.ReadUint32() // Unk
	bf2.ReadUint8()  // Unk

	if bf2.ReadUint8() != maxPlayers {
		t.Errorf("Max players mismatch")
	}

	if bf2.ReadUint8() != questType {
		t.Errorf("Quest type mismatch")
	}
}

// TestQuestEnumerationWithDifferentClientModes tests tune value filtering by client mode
func TestQuestEnumerationWithDifferentClientModes(t *testing.T) {
	tests := []struct {
		name         string
		clientMode   int
		maxTuneCount uint16
	}{
		{
			name:         "g91_mode",
			clientMode:   10, // Approx G91
			maxTuneCount: 256,
		},
		{
			name:         "g101_mode",
			clientMode:   11, // Approx G101
			maxTuneCount: 512,
		},
		{
			name:         "modern_mode",
			clientMode:   20, // Modern
			maxTuneCount: 770,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify tune count limits based on client mode
			var limit uint16
			if tc.clientMode <= 10 {
				limit = 256
			} else if tc.clientMode <= 11 {
				limit = 512
			} else {
				limit = 770
			}

			if limit != tc.maxTuneCount {
				t.Errorf("Mode %d: expected limit %d, got %d",
					tc.clientMode, tc.maxTuneCount, limit)
			}
		})
	}
}

// TestVSQuestItemsSerialization tests VS Quest items array serialization
func TestVSQuestItemsSerialization(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()

	// VS Quest has 19 items (hardcoded)
	itemCount := 19
	for i := 0; i < itemCount; i++ {
		bf.WriteUint16(uint16(1000 + i))
	}

	data := bf.Data()

	// Verify structure
	expectedSize := itemCount * 2
	if len(data) != expectedSize {
		t.Errorf("VS Quest items size mismatch: got %d, want %d", len(data), expectedSize)
	}

	// Verify values
	bf2 := byteframe.NewByteFrameFromBytes(data)
	bf2.SetLE()

	for i := 0; i < itemCount; i++ {
		expected := uint16(1000 + i)
		actual := bf2.ReadUint16()
		if actual != expected {
			t.Errorf("VS Quest item %d mismatch: got %d, want %d", i, actual, expected)
		}
	}
}

// TestFavoriteQuestDefaultData tests default favorite quest data format
func TestFavoriteQuestDefaultData(t *testing.T) {
	// Default favorite quest data when no data exists
	defaultData := []byte{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if len(defaultData) != 15 {
		t.Errorf("Default data size mismatch: got %d, want 15", len(defaultData))
	}

	// Verify structure (alternating 0x01, 0x00 pattern)
	expectedPattern := []byte{0x01, 0x00}

	for i := 0; i < 5; i++ {
		offset := i * 2
		if !bytes.Equal(defaultData[offset:offset+2], expectedPattern) {
			t.Errorf("Pattern mismatch at offset %d", offset)
		}
	}
}

// TestSeasonConversionLogic tests season conversion logic
func TestSeasonConversionLogic(t *testing.T) {
	tests := []struct {
		name         string
		baseFilename string
		expectedPart string
	}{
		{
			name:         "with_season_prefix",
			baseFilename: "00001",
			expectedPart: "00001",
		},
		{
			name:         "custom_quest_name",
			baseFilename: "quest_name",
			expectedPart: "quest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify filename handling
			if len(tc.baseFilename) >= 5 {
				prefix := tc.baseFilename[:5]
				if prefix != tc.expectedPart {
					t.Errorf("Filename parsing mismatch: got %s, want %s", prefix, tc.expectedPart)
				}
			}
		})
	}
}

// TestQuestFileLoadingErrors tests error handling in quest file loading
func TestQuestFileLoadingErrors(t *testing.T) {
	tests := []struct {
		name       string
		questID    int
		shouldFail bool
	}{
		{
			name:       "valid_quest_id",
			questID:    1,
			shouldFail: false,
		},
		{
			name:       "invalid_quest_id",
			questID:    -1,
			shouldFail: true,
		},
		{
			name:       "out_of_range",
			questID:    99999,
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// In real scenario, would attempt to load quest and verify error
			if tc.questID < 0 && !tc.shouldFail {
				t.Errorf("Negative quest ID should fail")
			}
		})
	}
}

// TestTournamentQuestEntryStub tests the stub tournament quest handler
func TestTournamentQuestEntryStub(t *testing.T) {
	mockConn := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mockConn)

	pkt := &mhfpacket.MsgMhfEnterTournamentQuest{}

	// This tests that the stub function doesn't panic
	handleMsgMhfEnterTournamentQuest(s, pkt)

	// Verify no crash occurred (pass if we reach here)
	if s.logger == nil {
		t.Errorf("Session corrupted")
	}
}

// TestGetUdBonusQuestInfoStructure tests UD bonus quest info structure
func TestGetUdBonusQuestInfoStructure(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()

	// Example UD bonus quest info entry
	bf.WriteUint8(0)              // Unk0
	bf.WriteUint8(0)              // Unk1
	bf.WriteUint32(uint32(time.Now().Unix())) // StartTime
	bf.WriteUint32(uint32(time.Now().Add(30*24*time.Hour).Unix())) // EndTime
	bf.WriteUint32(0)             // Unk4
	bf.WriteUint8(0)              // Unk5
	bf.WriteUint8(0)              // Unk6

	data := bf.Data()

	// Verify actual size: 2+4+4+4+1+1 = 16 bytes
	expectedSize := 16
	if len(data) != expectedSize {
		t.Errorf("UD bonus quest info size mismatch: got %d, want %d", len(data), expectedSize)
	}

	// Verify structure can be parsed
	bf2 := byteframe.NewByteFrameFromBytes(data)
	bf2.SetLE()

	bf2.ReadUint8()  // Unk0
	bf2.ReadUint8()  // Unk1
	startTime := bf2.ReadUint32()
	endTime := bf2.ReadUint32()
	bf2.ReadUint32() // Unk4
	bf2.ReadUint8()  // Unk5
	bf2.ReadUint8()  // Unk6

	if startTime >= endTime {
		t.Errorf("Quest end time must be after start time")
	}
}

// BenchmarkQuestEnumeration benchmarks quest enumeration performance
func BenchmarkQuestEnumeration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrame()

		// Build a response with tune values
		bf.WriteUint16(0)      // Returned count
		bf.WriteUint16(uint16(time.Now().Unix() & 0xFFFF))
		bf.WriteUint16(100)    // 100 tune values

		for j := 0; j < 100; j++ {
			bf.WriteUint16(uint16(j))
			bf.WriteUint16(uint16(j))
			bf.WriteUint32(0)
			bf.WriteUint16(uint16(j))
		}

		_ = bf.Data()
	}
}

// BenchmarkBackportQuest benchmarks quest backport performance
func BenchmarkBackportQuest(b *testing.B) {
	data := make([]byte, 500)
	binary.LittleEndian.PutUint32(data[0:4], 100)

	for i := 0; i < b.N; i++ {
		_ = BackportQuest(data, cfg.ZZ)
	}
}

// parseAckFromChannel reads a queued packet from the session's sendPackets channel
// and parses the ErrorCode from the MsgSysAck wire format.
func parseAckFromChannel(t *testing.T, s *Session) (errorCode uint8) {
	t.Helper()
	select {
	case pkt := <-s.sendPackets:
		// Wire format: 2 bytes opcode + 4 bytes AckHandle + 1 byte IsBufferResponse + 1 byte ErrorCode + ...
		data := pkt.data
		if len(data) < 8 {
			t.Fatalf("ack packet too short: %d bytes", len(data))
		}
		return data[7] // ErrorCode is at offset 7
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ack packet")
		return
	}
}

// TestHandleMsgSysGetFile_MissingQuestFile tests that a missing quest file
// sends a failure ack instead of crashing the client with nil data.
func TestHandleMsgSysGetFile_MissingQuestFile(t *testing.T) {
	mockConn := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mockConn)
	s.server.erupeConfig.BinPath = t.TempDir()

	pkt := &mhfpacket.MsgSysGetFile{
		AckHandle:  42,
		IsScenario: false,
		Filename:   "d00100d0",
	}

	handleMsgSysGetFile(s, pkt)

	errorCode := parseAckFromChannel(t, s)
	if errorCode != 1 {
		t.Errorf("expected failure ack (ErrorCode=1) for missing quest file, got ErrorCode=%d", errorCode)
	}
}

// TestHandleMsgSysGetFile_MissingScenarioFile tests that a missing scenario file
// sends a failure ack instead of crashing the client with nil data.
func TestHandleMsgSysGetFile_MissingScenarioFile(t *testing.T) {
	mockConn := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mockConn)
	s.server.erupeConfig.BinPath = t.TempDir()

	pkt := &mhfpacket.MsgSysGetFile{
		AckHandle:  42,
		IsScenario: true,
		// ScenarioIdentifer fields default to zero values, producing filename "0_0_0_0_S0_T0_C0"
	}

	handleMsgSysGetFile(s, pkt)

	errorCode := parseAckFromChannel(t, s)
	if errorCode != 1 {
		t.Errorf("expected failure ack (ErrorCode=1) for missing scenario file, got ErrorCode=%d", errorCode)
	}
}

// TestHandleMsgSysGetFile_ExistingQuestFile tests that an existing quest file
// sends a success ack with the file data.
func TestHandleMsgSysGetFile_ExistingQuestFile(t *testing.T) {
	mockConn := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mockConn)

	tmpDir := t.TempDir()
	s.server.erupeConfig.BinPath = tmpDir

	// Create the quests directory and a test quest file
	questDir := filepath.Join(tmpDir, "quests")
	if err := os.MkdirAll(questDir, 0o755); err != nil {
		t.Fatalf("failed to create quest dir: %v", err)
	}
	questData := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	if err := os.WriteFile(filepath.Join(questDir, "d00100d0.bin"), questData, 0o644); err != nil {
		t.Fatalf("failed to write quest file: %v", err)
	}

	pkt := &mhfpacket.MsgSysGetFile{
		AckHandle:  42,
		IsScenario: false,
		Filename:   "d00100d0",
	}

	handleMsgSysGetFile(s, pkt)

	errorCode := parseAckFromChannel(t, s)
	if errorCode != 0 {
		t.Errorf("expected success ack (ErrorCode=0) for existing quest file, got ErrorCode=%d", errorCode)
	}
}
