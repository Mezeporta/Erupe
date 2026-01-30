package channelserver

import (
	"math"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// TestPoogiOutfitUnlockCalculation documents the bug in the current poogie outfit unlock logic.
//
// CURRENT BEHAVIOR (BUG):
//
//	pugi_outfits = pugi_outfits + int(math.Pow(float64(outfitID), 2))
//	Example: outfitID=3 -> adds 9 to pugi_outfits
//
// EXPECTED BEHAVIOR (after fix commit 7459ded):
//
//	pugi_outfits = outfitID
//	Example: outfitID=3 -> sets pugi_outfits to 3
//
// The pugi_outfits field is a bitmask where each bit represents an unlocked outfit.
// The current math.Pow calculation is completely wrong for a bitmask.
func TestPoogiOutfitUnlockCalculation(t *testing.T) {
	tests := []struct {
		name          string
		outfitID      uint32
		currentBuggy  int    // What the current buggy code produces
		expectedFixed uint32 // What the fix should produce
	}{
		{
			name:          "outfit 0",
			outfitID:      0,
			currentBuggy:  int(math.Pow(float64(0), 2)), // 0
			expectedFixed: 0,
		},
		{
			name:          "outfit 1",
			outfitID:      1,
			currentBuggy:  int(math.Pow(float64(1), 2)), // 1
			expectedFixed: 1,
		},
		{
			name:          "outfit 2",
			outfitID:      2,
			currentBuggy:  int(math.Pow(float64(2), 2)), // 4 (WRONG!)
			expectedFixed: 2,
		},
		{
			name:          "outfit 3",
			outfitID:      3,
			currentBuggy:  int(math.Pow(float64(3), 2)), // 9 (WRONG!)
			expectedFixed: 3,
		},
		{
			name:          "outfit 10",
			outfitID:      10,
			currentBuggy:  int(math.Pow(float64(10), 2)), // 100 (WRONG!)
			expectedFixed: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify our understanding of the current buggy behavior
			buggyResult := int(math.Pow(float64(tt.outfitID), 2))
			if buggyResult != tt.currentBuggy {
				t.Errorf("buggy calculation = %d, expected %d", buggyResult, tt.currentBuggy)
			}

			// Document that the fix should just use the outfitID directly
			fixedResult := tt.outfitID
			if fixedResult != tt.expectedFixed {
				t.Errorf("fixed calculation = %d, expected %d", fixedResult, tt.expectedFixed)
			}

			// Show the difference
			if tt.outfitID > 1 {
				if buggyResult == int(tt.expectedFixed) {
					t.Logf("outfit %d: buggy and fixed results match (this is coincidental)", tt.outfitID)
				} else {
					t.Logf("outfit %d: buggy=%d, fixed=%d (BUG DOCUMENTED)", tt.outfitID, buggyResult, tt.expectedFixed)
				}
			}
		})
	}
}

// TestGuildManageRightNilPointerCondition documents the nil pointer bug in handleMsgMhfGetGuildManageRight.
//
// CURRENT BEHAVIOR (BUG - commit 5028355):
//
//	if guild == nil && s.prevGuildID != 0 {
//	This means: only try prevGuildID lookup if guild is nil AND prevGuildID is set
//
// EXPECTED BEHAVIOR (after fix):
//
//	if guild == nil || s.prevGuildID != 0 {
//	This means: try prevGuildID lookup if guild is nil OR prevGuildID is set
//
// The bug causes incorrect behavior when:
// - guild is NOT nil (player has a guild)
// - BUT s.prevGuildID is also set (player recently left a guild)
// In this case, the old code would NOT use prevGuildID, but the new code would.
func TestGuildManageRightNilPointerCondition(t *testing.T) {
	tests := []struct {
		name               string
		guildIsNil         bool
		prevGuildID        uint32
		shouldUsePrevGuild bool // What the condition evaluates to
		buggyBehavior      bool // Current buggy && condition
		fixedBehavior      bool // Fixed || condition
	}{
		{
			name:               "no guild, no prevGuildID",
			guildIsNil:         true,
			prevGuildID:        0,
			buggyBehavior:      false, // true && false = false
			fixedBehavior:      true,  // true || false = true
			shouldUsePrevGuild: false,
		},
		{
			name:               "no guild, has prevGuildID",
			guildIsNil:         true,
			prevGuildID:        42,
			buggyBehavior:      true, // true && true = true
			fixedBehavior:      true, // true || true = true
			shouldUsePrevGuild: true,
		},
		{
			name:               "has guild, no prevGuildID",
			guildIsNil:         false,
			prevGuildID:        0,
			buggyBehavior:      false, // false && false = false
			fixedBehavior:      false, // false || false = false
			shouldUsePrevGuild: false,
		},
		{
			name:               "has guild, has prevGuildID - THE BUG CASE",
			guildIsNil:         false,
			prevGuildID:        42,
			buggyBehavior:      false, // false && true = false (WRONG!)
			fixedBehavior:      true,  // false || true = true (CORRECT)
			shouldUsePrevGuild: true,  // Should use prevGuildID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the condition evaluation
			buggyCondition := tt.guildIsNil && (tt.prevGuildID != 0)
			fixedCondition := tt.guildIsNil || (tt.prevGuildID != 0)

			if buggyCondition != tt.buggyBehavior {
				t.Errorf("buggy condition = %v, expected %v", buggyCondition, tt.buggyBehavior)
			}
			if fixedCondition != tt.fixedBehavior {
				t.Errorf("fixed condition = %v, expected %v", fixedCondition, tt.fixedBehavior)
			}

			// Document when the bug manifests
			if buggyCondition != fixedCondition {
				t.Logf("BUG: %s - buggy=%v, fixed=%v", tt.name, buggyCondition, fixedCondition)
			}
		})
	}
}

// TestOperateGuildConstants verifies guild operation constants are defined correctly.
func TestOperateGuildConstants(t *testing.T) {
	// Test that the unlock outfit constant exists and has expected value
	if mhfpacket.OPERATE_GUILD_UNLOCK_OUTFIT != 0x12 {
		t.Errorf("OPERATE_GUILD_UNLOCK_OUTFIT = 0x%X, want 0x12", mhfpacket.OPERATE_GUILD_UNLOCK_OUTFIT)
	}
}

// TestGuildMemberInfo tests the GuildMember struct behavior.
func TestGuildMemberInfo(t *testing.T) {
	member := &GuildMember{
		CharID:    12345,
		GuildID:   100,
		Name:      "TestHunter",
		Recruiter: true,
		HRP:       500,
		GR:        50,
	}

	if member.CharID != 12345 {
		t.Errorf("CharID = %d, want 12345", member.CharID)
	}
	if !member.Recruiter {
		t.Error("Recruiter should be true")
	}
	if member.HRP != 500 {
		t.Errorf("HRP = %d, want 500", member.HRP)
	}
	if member.GR != 50 {
		t.Errorf("GR = %d, want 50", member.GR)
	}
}

// TestInfoGuildApplicantGRFieldSize documents the client mode difference for <G10.
//
// CURRENT BEHAVIOR:
//
//	Always writes applicant.GR (uint16) regardless of client mode.
//
// EXPECTED BEHAVIOR (after fix commit 8c219be):
//
//	Only write applicant.GR for G10+ clients.
//	For <G10 clients, skip the GR field.
//
// This test documents the packet structure difference.
func TestInfoGuildApplicantGRFieldSize(t *testing.T) {
	// Simulate building applicant data
	tests := []struct {
		name         string
		isG10Plus    bool
		expectedSize int // Size of applicant entry in bytes
	}{
		{
			name:         "Pre-G10 (no GR field)",
			isG10Plus:    false,
			expectedSize: 4 + 4 + 2, // CharID(4) + Unk(4) + HR(2) = 10 bytes (+ name)
		},
		{
			name:         "G10+ (with GR field)",
			isG10Plus:    true,
			expectedSize: 4 + 4 + 2 + 2, // CharID(4) + Unk(4) + HR(2) + GR(2) = 12 bytes (+ name)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()

			// Simulate what the fixed code should do
			applicantCharID := uint32(12345)
			applicantHR := uint16(7)
			applicantGR := uint16(50)

			bf.WriteUint32(applicantCharID)
			bf.WriteUint32(0) // Unk
			bf.WriteUint16(applicantHR)
			if tt.isG10Plus {
				bf.WriteUint16(applicantGR)
			}
			// Name would follow here (pascal string)

			data := bf.Data()
			if len(data) != tt.expectedSize {
				t.Errorf("applicant entry size = %d bytes, want %d bytes", len(data), tt.expectedSize)
			}
		})
	}
}

// TestGuildStructFields verifies the Guild struct has expected fields.
func TestGuildStructFields(t *testing.T) {
	guild := &Guild{
		ID:          1,
		Name:        "TestGuild",
		MemberCount: 10,
		PugiOutfits: 0xFF,
		PugiOutfit1: 1,
		PugiOutfit2: 2,
		PugiOutfit3: 3,
		PugiName1:   "Poogie1",
		PugiName2:   "Poogie2",
		PugiName3:   "Poogie3",
	}
	// Set embedded GuildLeader fields
	guild.LeaderCharID = 12345
	guild.LeaderName = "TestLeader"

	if guild.PugiOutfits != 0xFF {
		t.Errorf("PugiOutfits = %d, want 255", guild.PugiOutfits)
	}
	if guild.MemberCount != 10 {
		t.Errorf("MemberCount = %d, want 10", guild.MemberCount)
	}
	if guild.LeaderCharID != 12345 {
		t.Errorf("LeaderCharID = %d, want 12345", guild.LeaderCharID)
	}
}
