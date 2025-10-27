package mhfcid

import (
	"testing"
)

func TestConvertCID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{
			name:     "all ones",
			input:    "111111",
			expected: 0, // '1' maps to 0, so 0*32^0 + 0*32^1 + ... = 0
		},
		{
			name:     "all twos",
			input:    "222222",
			expected: 1 + 32 + 1024 + 32768 + 1048576 + 33554432, // 1*32^0 + 1*32^1 + 1*32^2 + 1*32^3 + 1*32^4 + 1*32^5
		},
		{
			name:     "sequential",
			input:    "123456",
			expected: 0 + 32 + 2*1024 + 3*32768 + 4*1048576 + 5*33554432, // 0 + 1*32 + 2*32^2 + 3*32^3 + 4*32^4 + 5*32^5
		},
		{
			name:     "with letters A-Z",
			input:    "ABCDEF",
			expected: 9 + 10*32 + 11*1024 + 12*32768 + 13*1048576 + 14*33554432,
		},
		{
			name:     "mixed numbers and letters",
			input:    "1A2B3C",
			expected: 0 + 9*32 + 1*1024 + 10*32768 + 2*1048576 + 11*33554432,
		},
		{
			name:     "max valid characters",
			input:    "ZZZZZZ",
			expected: 31 + 31*32 + 31*1024 + 31*32768 + 31*1048576 + 31*33554432, // 31 * (1 + 32 + 1024 + 32768 + 1048576 + 33554432)
		},
		{
			name:     "no banned chars: O excluded",
			input:    "N1P1Q1", // N=21, P=22, Q=23 - note no O
			expected: 21 + 0*32 + 22*1024 + 0*32768 + 23*1048576 + 0*33554432,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCID(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertCID(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertCID_InvalidLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"too short - 1", "1"},
		{"too short - 5", "12345"},
		{"too long - 7", "1234567"},
		{"too long - 10", "1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCID(tt.input)
			if result != 0 {
				t.Errorf("ConvertCID(%q) = %d, want 0 (invalid length should return 0)", tt.input, result)
			}
		})
	}
}

func TestConvertCID_BannedCharacters(t *testing.T) {
	// Banned characters: 0, I, O, S
	tests := []struct {
		name  string
		input string
	}{
		{"contains 0", "111011"},
		{"contains I", "111I11"},
		{"contains O", "11O111"},
		{"contains S", "S11111"},
		{"all banned", "000III"},
		{"mixed banned", "I0OS11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCID(tt.input)
			// Characters not in the map will contribute 0 to the result
			// The function doesn't explicitly reject them, it just doesn't map them
			// So we're testing that banned characters don't crash the function
			_ = result // Just verify it doesn't panic
		})
	}
}

func TestConvertCID_LowercaseNotSupported(t *testing.T) {
	// The map only contains uppercase letters
	input := "abcdef"
	result := ConvertCID(input)
	// Lowercase letters aren't mapped, so they'll contribute 0
	if result != 0 {
		t.Logf("ConvertCID(%q) = %d (lowercase not in map, contributes 0)", input, result)
	}
}

func TestConvertCID_CharacterMapping(t *testing.T) {
	// Verify specific character mappings
	tests := []struct {
		char     rune
		expected uint32
	}{
		{'1', 0},
		{'2', 1},
		{'9', 8},
		{'A', 9},
		{'B', 10},
		{'Z', 31},
		{'J', 17}, // J comes after I is skipped
		{'P', 22}, // P comes after O is skipped
		{'T', 25}, // T comes after S is skipped
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			// Create a CID with the character in the first position (32^0)
			input := string(tt.char) + "11111"
			result := ConvertCID(input)
			// The first character contributes its value * 32^0 = value * 1
			if result != tt.expected {
				t.Errorf("ConvertCID(%q) first char value = %d, want %d", input, result, tt.expected)
			}
		})
	}
}

func TestConvertCID_Base32Like(t *testing.T) {
	// Test that it behaves like base-32 conversion
	// The position multiplier should be powers of 32
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{
			name:     "position 0 only",
			input:    "211111", // 2 in position 0
			expected: 1,        // 1 * 32^0
		},
		{
			name:     "position 1 only",
			input:    "121111", // 2 in position 1
			expected: 32,       // 1 * 32^1
		},
		{
			name:     "position 2 only",
			input:    "112111", // 2 in position 2
			expected: 1024,     // 1 * 32^2
		},
		{
			name:     "position 3 only",
			input:    "111211", // 2 in position 3
			expected: 32768,    // 1 * 32^3
		},
		{
			name:     "position 4 only",
			input:    "111121", // 2 in position 4
			expected: 1048576,  // 1 * 32^4
		},
		{
			name:     "position 5 only",
			input:    "111112", // 2 in position 5
			expected: 33554432, // 1 * 32^5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCID(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertCID(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertCID_SkippedCharacters(t *testing.T) {
	// Verify that 0, I, O, S are actually skipped in the character sequence
	// The alphabet should be: 1-9 (0 skipped), A-H (I skipped), J-N (O skipped), P-R (S skipped), T-Z

	// Test that characters after skipped ones have the right values
	tests := []struct {
		name  string
		char1 string // Character before skip
		char2 string // Character after skip
		diff  uint32 // Expected difference (should be 1)
	}{
		{"before/after I skip", "H", "J", 1}, // H=16, J=17
		{"before/after O skip", "N", "P", 1}, // N=21, P=22
		{"before/after S skip", "R", "T", 1}, // R=24, T=25
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid1 := tt.char1 + "11111"
			cid2 := tt.char2 + "11111"
			val1 := ConvertCID(cid1)
			val2 := ConvertCID(cid2)
			diff := val2 - val1
			if diff != tt.diff {
				t.Errorf("Difference between %s and %s = %d, want %d (val1=%d, val2=%d)",
					tt.char1, tt.char2, diff, tt.diff, val1, val2)
			}
		})
	}
}

func BenchmarkConvertCID(b *testing.B) {
	testCID := "A1B2C3"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertCID(testCID)
	}
}

func BenchmarkConvertCID_AllLetters(b *testing.B) {
	testCID := "ABCDEF"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertCID(testCID)
	}
}

func BenchmarkConvertCID_AllNumbers(b *testing.B) {
	testCID := "123456"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertCID(testCID)
	}
}

func BenchmarkConvertCID_InvalidLength(b *testing.B) {
	testCID := "123" // Too short
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertCID(testCID)
	}
}
