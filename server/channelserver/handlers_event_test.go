package channelserver

import (
	"math/bits"
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfRegisterEvent(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegisterEvent{
		AckHandle: 12345,
		Unk2:      1,
		Unk4:      2,
	}

	handleMsgMhfRegisterEvent(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReleaseEvent(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReleaseEvent{
		AckHandle: 12345,
	}

	handleMsgMhfReleaseEvent(session, pkt)

	// Verify response packet was queued (with special error code 0x41)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateEvent(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateEvent{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateEvent(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetRestrictionEvent(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetRestrictionEvent panicked: %v", r)
		}
	}()

	handleMsgMhfGetRestrictionEvent(session, nil)
}

func TestHandleMsgMhfSetRestrictionEvent(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetRestrictionEvent{
		AckHandle: 12345,
	}

	handleMsgMhfSetRestrictionEvent(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestGenerateFeatureWeapons(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"single weapon", 1},
		{"few weapons", 3},
		{"normal count", 7},
		{"max weapons", 14},
		{"over max", 20}, // Should cap at 14
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFeatureWeapons(tt.count)

			// Result should be non-zero for positive counts
			if tt.count > 0 && result.ActiveFeatures == 0 {
				t.Error("Expected non-zero ActiveFeatures")
			}

			// Should not exceed max value (2^14 - 1 = 16383)
			if result.ActiveFeatures > 16383 {
				t.Errorf("ActiveFeatures = %d, exceeds max of 16383", result.ActiveFeatures)
			}
		})
	}
}

func TestGenerateFeatureWeapons_Randomness(t *testing.T) {
	// Generate multiple times and verify some variation
	results := make(map[uint32]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		result := generateFeatureWeapons(5)
		results[result.ActiveFeatures]++
	}

	// Should have some variation (not all the same)
	if len(results) == 1 {
		t.Error("Expected some variation in generated weapons")
	}
}

func TestGenerateFeatureWeapons_ZeroCount(t *testing.T) {
	result := generateFeatureWeapons(0)

	// Should return 0 for no weapons
	if result.ActiveFeatures != 0 {
		t.Errorf("Expected 0 for zero count, got %d", result.ActiveFeatures)
	}
}

// --- NEW TESTS ---

// TestGenerateFeatureWeapons_BitCount verifies that the number of set bits
// in ActiveFeatures matches the requested count (capped at 14).
func TestGenerateFeatureWeapons_BitCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		wantBits int
	}{
		{"1 weapon", 1, 1},
		{"5 weapons", 5, 5},
		{"10 weapons", 10, 10},
		{"14 weapons", 14, 14},
		{"20 capped to 14", 20, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFeatureWeapons(tt.count)
			setBits := bits.OnesCount32(result.ActiveFeatures)
			if setBits != tt.wantBits {
				t.Errorf("Set bits = %d, want %d (ActiveFeatures=0b%032b)",
					setBits, tt.wantBits, result.ActiveFeatures)
			}
		})
	}
}

// TestGenerateFeatureWeapons_BitsInRange verifies that all set bits are within
// bits 0-13 (no bits above bit 13 should be set).
func TestGenerateFeatureWeapons_BitsInRange(t *testing.T) {
	for i := 0; i < 50; i++ {
		result := generateFeatureWeapons(7)
		// Bits 14+ should never be set
		if result.ActiveFeatures&^uint32(0x3FFF) != 0 {
			t.Errorf("Bits above 13 are set: 0x%08X", result.ActiveFeatures)
		}
	}
}

// TestGenerateFeatureWeapons_MaxYieldsAllBits verifies that requesting 14
// weapons sets exactly bits 0-13 (the value 16383 = 0x3FFF).
func TestGenerateFeatureWeapons_MaxYieldsAllBits(t *testing.T) {
	result := generateFeatureWeapons(14)
	if result.ActiveFeatures != 0x3FFF {
		t.Errorf("ActiveFeatures = 0x%04X, want 0x3FFF (all 14 bits set)", result.ActiveFeatures)
	}
}

// TestGenerateFeatureWeapons_StartTimeZero verifies that the returned
// activeFeature has a zero StartTime (not set by generateFeatureWeapons).
func TestGenerateFeatureWeapons_StartTimeZero(t *testing.T) {
	result := generateFeatureWeapons(5)
	if !result.StartTime.IsZero() {
		t.Errorf("StartTime should be zero, got %v", result.StartTime)
	}
}

// TestHandleMsgMhfRegisterEvent_DifferentValues tests with various Unk2/Unk4 values.
func TestHandleMsgMhfRegisterEvent_DifferentValues(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		unk2 uint8
		unk4 uint8
	}{
		{"zeros", 0, 0},
		{"max values", 255, 255},
		{"typical", 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			pkt := &mhfpacket.MsgMhfRegisterEvent{
				AckHandle: 99999,
				Unk2:      tt.unk2,
				Unk4:      tt.unk4,
			}

			handleMsgMhfRegisterEvent(session, pkt)

			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Error("Response packet should have data")
				}
			default:
				t.Error("No response packet queued")
			}
		})
	}
}
