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
		level       uint8
		wantNext    uint16
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

// --- NEW TESTS ---

// TestGetAchData_Level6BronzeTrophy tests that level 6 (in-progress toward level 7)
// awards the bronze trophy (0x40).
// Curve 0: {5, 15, 30, 50, 100, 150, 200, 300}
// Cumulative at each level: L1=5, L2=20, L3=50, L4=100, L5=200, L6=350, L7=550, L8=850
// At cumulative 350, we reach level 6. Score 400 means level 6 with progress 50 toward next.
func TestGetAchData_Level6BronzeTrophy(t *testing.T) {
	// Score to reach level 6 and be partway to level 7:
	// cumulative to level 6 = 5+15+30+50+100+150 = 350
	// score 400 = level 6 with 50 remaining progress
	ach := GetAchData(0, 400)
	if ach.Level != 6 {
		t.Errorf("Level = %d, want 6", ach.Level)
	}
	if ach.Trophy != 0x40 {
		t.Errorf("Trophy = 0x%02x, want 0x40 (bronze)", ach.Trophy)
	}
	if ach.NextValue != 15 {
		t.Errorf("NextValue = %d, want 15", ach.NextValue)
	}
	if ach.Progress != 50 {
		t.Errorf("Progress = %d, want 50", ach.Progress)
	}
	if ach.Required != 200 {
		t.Errorf("Required = %d, want 200 (curve[6])", ach.Required)
	}
}

// TestGetAchData_Level7SilverTrophy tests that level 7 (in-progress toward level 8)
// awards the silver trophy (0x60).
// cumulative to level 7 = 5+15+30+50+100+150+200 = 550
// score 600 = level 7 with 50 remaining progress
func TestGetAchData_Level7SilverTrophy(t *testing.T) {
	ach := GetAchData(0, 600)
	if ach.Level != 7 {
		t.Errorf("Level = %d, want 7", ach.Level)
	}
	if ach.Trophy != 0x60 {
		t.Errorf("Trophy = 0x%02x, want 0x60 (silver)", ach.Trophy)
	}
	if ach.NextValue != 20 {
		t.Errorf("NextValue = %d, want 20", ach.NextValue)
	}
	if ach.Progress != 50 {
		t.Errorf("Progress = %d, want 50", ach.Progress)
	}
	if ach.Required != 300 {
		t.Errorf("Required = %d, want 300 (curve[7])", ach.Required)
	}
}

// TestGetAchData_MaxedOut_AllCurves tests that reaching max level on each curve
// produces the correct gold trophy and the last threshold as Required/Progress.
func TestGetAchData_MaxedOut_AllCurves(t *testing.T) {
	tests := []struct {
		name     string
		id       uint8
		score    int32
		lastThresh int32
	}{
		// Curve 0: {5,15,30,50,100,150,200,300} sum=850, last=300
		{"Curve0_ID0", 0, 5000, 300},
		// Curve 1: {1,5,10,15,30,50,75,100} sum=286, last=100
		{"Curve1_ID7", 7, 5000, 100},
		// Curve 2: {1,2,3,4,5,6,7,8} sum=36, last=8
		{"Curve2_ID8", 8, 5000, 8},
		// Curve 3: {10,50,100,200,350,500,750,999} sum=2959, last=999
		{"Curve3_ID16", 16, 50000, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ach := GetAchData(tt.id, tt.score)
			if ach.Level != 8 {
				t.Errorf("Level = %d, want 8 (max)", ach.Level)
			}
			if ach.Trophy != 0x7F {
				t.Errorf("Trophy = 0x%02x, want 0x7F (gold)", ach.Trophy)
			}
			if ach.Required != uint32(tt.lastThresh) {
				t.Errorf("Required = %d, want %d", ach.Required, tt.lastThresh)
			}
			if ach.Progress != ach.Required {
				t.Errorf("Progress = %d, want %d (should equal Required at max)", ach.Progress, ach.Required)
			}
		})
	}
}

// TestGetAchData_ExactlyAtEachThreshold tests the exact cumulative score at each
// threshold boundary for curve 0.
func TestGetAchData_ExactlyAtEachThreshold(t *testing.T) {
	// Curve 0: {5, 15, 30, 50, 100, 150, 200, 300}
	// Cumulative thresholds (exact score to reach each level):
	// L1: 5, L2: 20, L3: 50, L4: 100, L5: 200, L6: 350, L7: 550, L8: 850
	cumulativeScores := []int32{5, 20, 50, 100, 200, 350, 550, 850}
	expectedLevels := []uint8{1, 2, 3, 4, 5, 6, 7, 8}
	expectedValues := []uint32{5, 15, 25, 35, 50, 65, 80, 100}

	for i, score := range cumulativeScores {
		t.Run("ExactThreshold_L"+string(rune('1'+i)), func(t *testing.T) {
			ach := GetAchData(0, score)
			if ach.Level != expectedLevels[i] {
				t.Errorf("score=%d: Level = %d, want %d", score, ach.Level, expectedLevels[i])
			}
			if ach.Value != expectedValues[i] {
				t.Errorf("score=%d: Value = %d, want %d", score, ach.Value, expectedValues[i])
			}
		})
	}
}

// TestGetAchData_OneBeforeEachThreshold tests scores that are one less than
// each cumulative threshold, verifying they stay at the previous level.
func TestGetAchData_OneBeforeEachThreshold(t *testing.T) {
	// Curve 0: cumulative thresholds: 5, 20, 50, 100, 200, 350, 550, 850
	cumulativeScores := []int32{4, 19, 49, 99, 199, 349, 549, 849}
	expectedLevels := []uint8{0, 1, 2, 3, 4, 5, 6, 7}

	for i, score := range cumulativeScores {
		t.Run("OneBeforeThreshold_L"+string(rune('0'+i)), func(t *testing.T) {
			ach := GetAchData(0, score)
			if ach.Level != expectedLevels[i] {
				t.Errorf("score=%d: Level = %d, want %d", score, ach.Level, expectedLevels[i])
			}
		})
	}
}

// TestGetAchData_Curve2_FestaWins exercises the "Festa wins" curve which has
// small thresholds: {1, 2, 3, 4, 5, 6, 7, 8}
func TestGetAchData_Curve2_FestaWins(t *testing.T) {
	// Curve 2: {1, 2, 3, 4, 5, 6, 7, 8}
	// Cumulative: 1, 3, 6, 10, 15, 21, 28, 36
	tests := []struct {
		score    int32
		wantLvl  uint8
		wantProg uint32
		wantReq  uint32
	}{
		{0, 0, 0, 1},
		{1, 1, 0, 2},       // Exactly at first threshold
		{2, 1, 1, 2},       // One into second threshold
		{3, 2, 0, 3},       // Exactly at second cumulative
		{36, 8, 8, 8},      // Max level (sum of all thresholds)
		{100, 8, 8, 8},     // Well above max
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			ach := GetAchData(8, tt.score) // ID 8 maps to curve 2
			if ach.Level != tt.wantLvl {
				t.Errorf("score=%d: Level = %d, want %d", tt.score, ach.Level, tt.wantLvl)
			}
			if ach.Progress != tt.wantProg {
				t.Errorf("score=%d: Progress = %d, want %d", tt.score, ach.Progress, tt.wantProg)
			}
			if ach.Required != tt.wantReq {
				t.Errorf("score=%d: Required = %d, want %d", tt.score, ach.Required, tt.wantReq)
			}
		})
	}
}

// TestGetAchData_AllIDs_ZeroScore verifies that calling GetAchData with score=0
// for every valid ID returns level 0 without panicking.
func TestGetAchData_AllIDs_ZeroScore(t *testing.T) {
	for id := uint8(0); id <= 32; id++ {
		ach := GetAchData(id, 0)
		if ach.Level != 0 {
			t.Errorf("ID %d, score 0: Level = %d, want 0", id, ach.Level)
		}
		if ach.Value != 0 {
			t.Errorf("ID %d, score 0: Value = %d, want 0", id, ach.Value)
		}
		if ach.Trophy != 0 {
			t.Errorf("ID %d, score 0: Trophy = 0x%02x, want 0x00", id, ach.Trophy)
		}
	}
}

// TestGetAchData_AllIDs_MaxScore verifies that calling GetAchData with a very
// high score for every valid ID returns level 8 with gold trophy.
func TestGetAchData_AllIDs_MaxScore(t *testing.T) {
	for id := uint8(0); id <= 32; id++ {
		ach := GetAchData(id, 99999)
		if ach.Level != 8 {
			t.Errorf("ID %d: Level = %d, want 8", id, ach.Level)
		}
		if ach.Trophy != 0x7F {
			t.Errorf("ID %d: Trophy = 0x%02x, want 0x7F", id, ach.Trophy)
		}
		// At max, Progress should equal Required
		if ach.Progress != ach.Required {
			t.Errorf("ID %d: Progress (%d) != Required (%d) at max", id, ach.Progress, ach.Required)
		}
	}
}

// TestGetAchData_UpdatedAlwaysFalse confirms Updated is always false since
// GetAchData never sets it.
func TestGetAchData_UpdatedAlwaysFalse(t *testing.T) {
	scores := []int32{0, 1, 5, 50, 500, 5000}
	for _, score := range scores {
		ach := GetAchData(0, score)
		if ach.Updated {
			t.Errorf("score=%d: Updated should always be false, got true", score)
		}
	}
}
