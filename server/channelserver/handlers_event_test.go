package channelserver

import (
	"math/bits"
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"time"
)

func TestHandleMsgMhfRegisterEvent(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegisterEvent{
		AckHandle: 12345,
		WorldID:   1,
		LandID:    2,
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
			result := generateFeatureWeapons(tt.count, cfg.ZZ)

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
		result := generateFeatureWeapons(5, cfg.ZZ)
		results[result.ActiveFeatures]++
	}

	// Should have some variation (not all the same)
	if len(results) == 1 {
		t.Error("Expected some variation in generated weapons")
	}
}

func TestGenerateFeatureWeapons_ZeroCount(t *testing.T) {
	result := generateFeatureWeapons(0, cfg.ZZ)

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
			result := generateFeatureWeapons(tt.count, cfg.ZZ)
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
		result := generateFeatureWeapons(7, cfg.ZZ)
		// Bits 14+ should never be set
		if result.ActiveFeatures&^uint32(0x3FFF) != 0 {
			t.Errorf("Bits above 13 are set: 0x%08X", result.ActiveFeatures)
		}
	}
}

// TestGenerateFeatureWeapons_MaxYieldsAllBits verifies that requesting 14
// weapons sets exactly bits 0-13 (the value 16383 = 0x3FFF).
func TestGenerateFeatureWeapons_MaxYieldsAllBits(t *testing.T) {
	result := generateFeatureWeapons(14, cfg.ZZ)
	if result.ActiveFeatures != 0x3FFF {
		t.Errorf("ActiveFeatures = 0x%04X, want 0x3FFF (all 14 bits set)", result.ActiveFeatures)
	}
}

// TestGenerateFeatureWeapons_StartTimeZero verifies that the returned
// activeFeature has a zero StartTime (not set by generateFeatureWeapons).
func TestGenerateFeatureWeapons_StartTimeZero(t *testing.T) {
	result := generateFeatureWeapons(5, cfg.ZZ)
	if !result.StartTime.IsZero() {
		t.Errorf("StartTime should be zero, got %v", result.StartTime)
	}
}

func TestHandleMsgMhfReleaseEvent_ErrorCode(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReleaseEvent{
		AckHandle: 88888,
	}

	handleMsgMhfReleaseEvent(session, pkt)

	// This handler manually sends a response with error code 0x41
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateEvent_Stub(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateEvent{
		AckHandle: 77777,
	}

	handleMsgMhfEnumerateEvent(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSetRestrictionEvent_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetRestrictionEvent{
		AckHandle: 11111,
	}

	handleMsgMhfSetRestrictionEvent(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetRestrictionEvent_Empty(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetRestrictionEvent panicked: %v", r)
		}
	}()

	handleMsgMhfGetRestrictionEvent(session, nil)
}

// TestHandleMsgMhfRegisterEvent_DifferentValues tests with various Unk2/Unk4 values.
func TestHandleMsgMhfRegisterEvent_DifferentValues(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name    string
		worldID uint16
		landID  uint16
	}{
		{"zeros", 0, 0},
		{"max values", 65535, 65535},
		{"typical", 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			pkt := &mhfpacket.MsgMhfRegisterEvent{
				AckHandle: 99999,
				WorldID:   tt.worldID,
				LandID:    tt.landID,
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

func TestHandleMsgMhfGetWeeklySchedule(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{}
	srv.erupeConfig.GameplayOptions.MinFeatureWeapons = 1
	srv.erupeConfig.GameplayOptions.MaxFeatureWeapons = 3
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetWeeklySchedule{AckHandle: 1}
	handleMsgMhfGetWeeklySchedule(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetKeepLoginBoostStatus_Disabled(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{}
	srv.erupeConfig.GameplayOptions.DisableLoginBoost = true
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetKeepLoginBoostStatus{AckHandle: 1}
	handleMsgMhfGetKeepLoginBoostStatus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetKeepLoginBoostStatus_EmptyBoosts(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetKeepLoginBoostStatus{AckHandle: 1}
	handleMsgMhfGetKeepLoginBoostStatus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetKeepLoginBoostStatus_WithBoosts(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{
		loginBoosts: []loginBoost{
			{WeekReq: 1, Expiration: time.Now().Add(-7 * 24 * time.Hour)},
			{WeekReq: 2, Expiration: time.Now().Add(-14 * 24 * time.Hour)},
			{WeekReq: 3, Expiration: time.Now().Add(-21 * 24 * time.Hour)},
			{WeekReq: 4, Expiration: time.Now().Add(-28 * 24 * time.Hour)},
			{WeekReq: 5, Expiration: time.Now().Add(-35 * 24 * time.Hour)},
		},
	}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetKeepLoginBoostStatus{AckHandle: 1}
	handleMsgMhfGetKeepLoginBoostStatus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfUseKeepLoginBoost(t *testing.T) {
	tests := []struct {
		name          string
		boostWeekUsed uint8
	}{
		{"week1", 1},
		{"week2", 2},
		{"week3", 3},
		{"week4", 4},
		{"week5", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := createMockServer()
			srv.eventRepo = &mockEventRepo{}
			s := createMockSession(100, srv)

			pkt := &mhfpacket.MsgMhfUseKeepLoginBoost{AckHandle: 1, BoostWeekUsed: tt.boostWeekUsed}
			handleMsgMhfUseKeepLoginBoost(s, pkt)

			select {
			case p := <-s.sendPackets:
				if len(p.data) == 0 {
					t.Fatal("Expected non-empty response")
				}
			default:
				t.Fatal("No response packet queued")
			}
		})
	}
}

func TestHandleMsgMhfLoadScenarioData(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfLoadScenarioData{AckHandle: 1}
	handleMsgMhfLoadScenarioData(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfSaveScenarioData(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfSaveScenarioData{AckHandle: 1, RawDataPayload: []byte{0x01, 0x02, 0x03}}
	handleMsgMhfSaveScenarioData(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfListMember(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["blocked"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfListMember{AckHandle: 1}
	handleMsgMhfListMember(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfOprMember_AddBlacklist(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["blocked"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOprMember{AckHandle: 1, Blacklist: true, Operation: false, CharIDs: []uint32{42}}
	handleMsgMhfOprMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfOprMember_AddFriend(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["friends"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOprMember{AckHandle: 1, Blacklist: false, Operation: false, CharIDs: []uint32{42}}
	handleMsgMhfOprMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfOprMember_RemoveBlacklist(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["blocked"] = "42"
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOprMember{AckHandle: 1, Blacklist: true, Operation: true, CharIDs: []uint32{42}}
	handleMsgMhfOprMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}
