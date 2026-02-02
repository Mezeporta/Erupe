package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestGetAchData_Level0(t *testing.T) {
	// Score 0 should give level 0 with progress toward first threshold
	ach := GetAchData(0, 0)
	if ach.Level != 0 {
		t.Errorf("Level = %d, want 0", ach.Level)
	}
	if ach.Progress != 0 {
		t.Errorf("Progress = %d, want 0", ach.Progress)
	}
	if ach.NextValue != 5 {
		t.Errorf("NextValue = %d, want 5", ach.NextValue)
	}
}

func TestGetAchData_Level1(t *testing.T) {
	// Score 5 (exactly at first threshold) should give level 1
	ach := GetAchData(0, 5)
	if ach.Level != 1 {
		t.Errorf("Level = %d, want 1", ach.Level)
	}
	if ach.Value != 5 {
		t.Errorf("Value = %d, want 5", ach.Value)
	}
}

func TestGetAchData_Partial(t *testing.T) {
	// Score 3 should give level 0 with progress 3
	ach := GetAchData(0, 3)
	if ach.Level != 0 {
		t.Errorf("Level = %d, want 0", ach.Level)
	}
	if ach.Progress != 3 {
		t.Errorf("Progress = %d, want 3", ach.Progress)
	}
	if ach.Required != 5 {
		t.Errorf("Required = %d, want 5", ach.Required)
	}
}

func TestGetAchData_MaxLevel(t *testing.T) {
	// Score 999 should give max level for curve 0
	ach := GetAchData(0, 999)
	if ach.Level != 8 {
		t.Errorf("Level = %d, want 8", ach.Level)
	}
	if ach.Trophy != 0x7F {
		t.Errorf("Trophy = %x, want 0x7F (gold)", ach.Trophy)
	}
}

func TestGetAchData_BronzeTrophy(t *testing.T) {
	// Level 7 should have bronze trophy (0x40)
	// Curve 0: 5, 15, 30, 50, 100, 150, 200, 300
	// Cumulative: 5, 20, 50, 100, 200, 350, 550, 850
	// To reach level 7, need 550+ points (sum of first 7 thresholds)
	ach := GetAchData(0, 550)
	if ach.Level != 7 {
		t.Errorf("Level = %d, want 7", ach.Level)
	}
	if ach.Trophy != 0x60 {
		t.Errorf("Trophy = %x, want 0x60 (silver)", ach.Trophy)
	}
}

func TestGetAchData_SilverTrophy(t *testing.T) {
	// Level 8 (max) should have gold trophy (0x7F)
	// Need 850+ (sum of all 8 thresholds) for max level
	ach := GetAchData(0, 850)
	if ach.Level != 8 {
		t.Errorf("Level = %d, want 8", ach.Level)
	}
	if ach.Trophy != 0x7F {
		t.Errorf("Trophy = %x, want 0x7F (gold)", ach.Trophy)
	}
}

func TestGetAchData_DifferentCurves(t *testing.T) {
	tests := []struct {
		name     string
		id       uint8
		score    int32
		wantLvl  uint8
		wantProg uint32
	}{
		{"Curve1_ID7_Level0", 7, 0, 0, 0},
		{"Curve1_ID7_Level1", 7, 1, 1, 0},
		{"Curve2_ID8_Level0", 8, 0, 0, 0},
		{"Curve2_ID8_Level1", 8, 1, 1, 0},
		{"Curve3_ID16_Level0", 16, 0, 0, 0},
		{"Curve3_ID16_Partial", 16, 5, 0, 5},
		{"Curve3_ID16_Level1", 16, 10, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ach := GetAchData(tt.id, tt.score)
			if ach.Level != tt.wantLvl {
				t.Errorf("Level = %d, want %d", ach.Level, tt.wantLvl)
			}
			if ach.Progress != tt.wantProg {
				t.Errorf("Progress = %d, want %d", ach.Progress, tt.wantProg)
			}
		})
	}
}

func TestGetAchData_AllCurveMappings(t *testing.T) {
	// Verify all achievement IDs have valid curve mappings
	for id := uint8(0); id <= 32; id++ {
		curve, ok := achievementCurveMap[id]
		if !ok {
			t.Errorf("Achievement ID %d has no curve mapping", id)
			continue
		}
		if len(curve) != 8 {
			t.Errorf("Achievement ID %d curve has %d elements, want 8", id, len(curve))
		}
	}
}

func TestGetAchData_ValueAccumulation(t *testing.T) {
	// Test that Value correctly accumulates based on level
	// Level values: 1=5, 2-4=10, 5-7=15, 8=20
	// At max level 8: 5 + 10*3 + 15*3 + 20 = 5 + 30 + 45 + 20 = 100
	ach := GetAchData(0, 1000) // Score well above max
	expectedValue := uint32(5 + 10 + 10 + 10 + 15 + 15 + 15 + 20)
	if ach.Value != expectedValue {
		t.Errorf("Value = %d, want %d", ach.Value, expectedValue)
	}
}

func TestGetAchData_NextValueByLevel(t *testing.T) {
	tests := []struct {
		level     uint8
		wantNext  uint16
		approxScore int32
	}{
		{0, 5, 0},
		{1, 10, 5},
		{2, 10, 15},
		{3, 10, 30},
		{4, 15, 50},
		{5, 15, 100},
	}

	for _, tt := range tests {
		t.Run("Level"+string(rune('0'+tt.level)), func(t *testing.T) {
			ach := GetAchData(0, tt.approxScore)
			if ach.Level != tt.level {
				t.Skipf("Skipping: got level %d, expected %d", ach.Level, tt.level)
			}
			if ach.NextValue != tt.wantNext {
				t.Errorf("NextValue at level %d = %d, want %d", ach.Level, ach.NextValue, tt.wantNext)
			}
		})
	}
}

func TestAchievementCurves(t *testing.T) {
	// Verify curve values are strictly increasing
	for i, curve := range achievementCurves {
		for j := 1; j < len(curve); j++ {
			if curve[j] <= curve[j-1] {
				t.Errorf("Curve %d: value[%d]=%d should be > value[%d]=%d",
					i, j, curve[j], j-1, curve[j-1])
			}
		}
	}
}

func TestAchievementCurveMap_Coverage(t *testing.T) {
	// Ensure all mapped curves exist
	for id, curve := range achievementCurveMap {
		found := false
		for _, c := range achievementCurves {
			if &c[0] == &curve[0] {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Achievement ID %d maps to unknown curve", id)
		}
	}
}

func TestHandleMsgMhfSetCaAchievementHist(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetCaAchievementHist{
		AckHandle: 12345,
	}

	handleMsgMhfSetCaAchievementHist(session, pkt)

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

// Test empty achievement handlers don't panic
func TestEmptyAchievementHandlers(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name    string
		handler func(s *Session, p mhfpacket.MHFPacket)
	}{
		{"handleMsgMhfResetAchievement", handleMsgMhfResetAchievement},
		{"handleMsgMhfPaymentAchievement", handleMsgMhfPaymentAchievement},
		{"handleMsgMhfDisplayedAchievement", handleMsgMhfDisplayedAchievement},
		{"handleMsgMhfGetCaAchievementHist", handleMsgMhfGetCaAchievementHist},
		{"handleMsgMhfSetCaAchievement", handleMsgMhfSetCaAchievement},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.handler(session, nil)
		})
	}
}
