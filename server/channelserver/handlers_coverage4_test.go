package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// =============================================================================
// handleMsgMhfGetPaperData: 565-line pure data serialization function.
// Tests all switch cases: 0, 5, 6, >1000 (known & unknown), default <1000.
// =============================================================================

func TestHandleMsgMhfGetPaperData_Case0(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:      0,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("case 0: response should have data")
		}
	default:
		t.Error("case 0: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_Case5(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:      5,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("case 5: response should have data")
		}
	default:
		t.Error("case 5: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_Case6(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:      6,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("case 6: response should have data")
		}
	default:
		t.Error("case 6: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_GreaterThan1000_KnownKey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// 6001 is a known key in paperGiftData
	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:      6001,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error(">1000 known: response should have data")
		}
	default:
		t.Error(">1000 known: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_GreaterThan1000_UnknownKey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// 9999 is not a known key in paperGiftData
	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:      9999,
	})

	select {
	case p := <-session.sendPackets:
		// Even unknown keys should produce a response (empty earth succeed)
		_ = p
	default:
		t.Error(">1000 unknown: no response queued")
	}
}

func TestHandleMsgMhfGetPaperData_DefaultUnknownLessThan1000(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Unknown type < 1000, hits default case then falls to else branch
	handleMsgMhfGetPaperData(session, &mhfpacket.MsgMhfGetPaperData{
		AckHandle: 1,
		DataType:      99,
	})

	select {
	case p := <-session.sendPackets:
		_ = p
	default:
		t.Error("default <1000: no response queued")
	}
}

// =============================================================================
// handleMsgMhfGetGachaPlayHistory and handleMsgMhfPlayFreeGacha
// =============================================================================

func TestHandleMsgMhfGetGachaPlayHistory(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetGachaPlayHistory(session, &mhfpacket.MsgMhfGetGachaPlayHistory{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfPlayFreeGacha(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfPlayFreeGacha(session, &mhfpacket.MsgMhfPlayFreeGacha{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

// Seibattle handlers: GetBreakSeibatuLevelReward, GetFixedSeibatuRankingTable,
// ReadLastWeekBeatRanking, ReadBeatLevelAllRanking, ReadBeatLevelMyRanking
// are already tested in handlers_misc_test.go and handlers_tower_test.go.

// =============================================================================
// grpToGR: pure function, no dependencies
// =============================================================================

func TestGrpToGR(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected uint16
	}{
		{"zero", 0, 1},
		{"low_value", 500, 2},
		{"first_bracket", 1000, 2},
		{"mid_bracket", 208750, 51},
		{"second_bracket", 300000, 62},
		{"high_value", 593400, 100},
		{"third_bracket", 700000, 113},
		{"very_high", 993400, 150},
		{"above_993400", 1000000, 150},
		{"fourth_bracket", 1400900, 200},
		{"max_bracket", 11345900, 900},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpToGR(tt.input)
			if got != tt.expected {
				t.Errorf("grpToGR(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// dumpSaveData: test disabled path
// =============================================================================

func TestDumpSaveData_Disabled(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.SaveDumps.Enabled = false
	session := createMockSession(1, server)

	// Should return immediately without error
	dumpSaveData(session, []byte{0x01, 0x02, 0x03}, "test")
}

// =============================================================================
// TimeGameAbsolute
// =============================================================================

func TestTimeGameAbsolute(t *testing.T) {
	result := TimeGameAbsolute()

	// TimeGameAbsolute returns (adjustedUnix - 2160) % 5760
	// Result should be in range [0, 5760)
	if result >= 5760 {
		t.Errorf("TimeGameAbsolute() = %d, should be < 5760", result)
	}
}

// =============================================================================
// handleMsgSysAuthData: empty handler
// =============================================================================

func TestHandleMsgSysAuthData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysAuthData panicked: %v", r)
		}
	}()
	handleMsgSysAuthData(session, nil)
}
