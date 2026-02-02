package channelserver

import (
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
