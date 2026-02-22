package stringsupport

import (
	"bytes"
	"testing"
)

func TestUTF8ToSJIS(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ascii", "Hello World"},
		{"numbers", "12345"},
		{"symbols", "!@#$%"},
		{"japanese_hiragana", "あいうえお"},
		{"japanese_katakana", "アイウエオ"},
		{"japanese_kanji", "日本語"},
		{"mixed", "Hello世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UTF8ToSJIS(tt.input)
			if len(result) == 0 && len(tt.input) > 0 {
				t.Error("UTF8ToSJIS returned empty result for non-empty input")
			}
		})
	}
}

func TestSJISToUTF8(t *testing.T) {
	// Test ASCII characters (which are the same in SJIS and UTF-8)
	asciiBytes := []byte("Hello World")
	result, err := SJISToUTF8(asciiBytes)
	if err != nil {
		t.Fatalf("SJISToUTF8() unexpected error: %v", err)
	}
	if result != "Hello World" {
		t.Errorf("SJISToUTF8() = %q, want %q", result, "Hello World")
	}
}

func TestUTF8ToSJIS_RoundTrip(t *testing.T) {
	// Test round-trip conversion for ASCII
	original := "Hello World 123"
	sjis := UTF8ToSJIS(original)
	back, _ := SJISToUTF8(sjis)

	if back != original {
		t.Errorf("Round-trip failed: got %q, want %q", back, original)
	}
}

func TestToNGWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minLen   int
		checkFn  func(t *testing.T, result []uint16)
	}{
		{
			name:   "ascii characters",
			input:  "ABC",
			minLen: 3,
			checkFn: func(t *testing.T, result []uint16) {
				if result[0] != uint16('A') {
					t.Errorf("result[0] = %d, want %d", result[0], 'A')
				}
			},
		},
		{
			name:   "numbers",
			input:  "123",
			minLen: 3,
			checkFn: func(t *testing.T, result []uint16) {
				if result[0] != uint16('1') {
					t.Errorf("result[0] = %d, want %d", result[0], '1')
				}
			},
		},
		{
			name:   "japanese characters",
			input:  "あ",
			minLen: 1,
			checkFn: func(t *testing.T, result []uint16) {
				if len(result) == 0 {
					t.Error("result should not be empty")
				}
			},
		},
		{
			name:   "empty string",
			input:  "",
			minLen: 0,
			checkFn: func(t *testing.T, result []uint16) {
				if len(result) != 0 {
					t.Errorf("result length = %d, want 0", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToNGWord(tt.input)
			if len(result) < tt.minLen {
				t.Errorf("ToNGWord() length = %d, want at least %d", len(result), tt.minLen)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

func TestPaddedString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		size      uint
		transform bool
		wantLen   uint
	}{
		{"short string", "Hello", 10, false, 10},
		{"exact size", "Test", 5, false, 5},
		{"longer than size", "This is a long string", 10, false, 10},
		{"empty string", "", 5, false, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PaddedString(tt.input, tt.size, tt.transform)
			if uint(len(result)) != tt.wantLen {
				t.Errorf("PaddedString() length = %d, want %d", len(result), tt.wantLen)
			}
			// Verify last byte is null
			if result[len(result)-1] != 0 {
				t.Error("PaddedString() should end with null byte")
			}
		})
	}
}

func TestPaddedString_NullTermination(t *testing.T) {
	result := PaddedString("Test", 10, false)
	if result[9] != 0 {
		t.Error("Last byte should be null")
	}
	// First 4 bytes should be "Test"
	if !bytes.Equal(result[0:4], []byte("Test")) {
		t.Errorf("First 4 bytes = %v, want %v", result[0:4], []byte("Test"))
	}
}

func TestCSVAdd(t *testing.T) {
	tests := []struct {
		name     string
		csv      string
		value    int
		expected string
	}{
		{"add to empty", "", 1, "1"},
		{"add to existing", "1,2,3", 4, "1,2,3,4"},
		{"add duplicate", "1,2,3", 2, "1,2,3"},
		{"add to single", "5", 10, "5,10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVAdd(tt.csv, tt.value)
			if result != tt.expected {
				t.Errorf("CSVAdd(%q, %d) = %q, want %q", tt.csv, tt.value, result, tt.expected)
			}
		})
	}
}

func TestCSVRemove(t *testing.T) {
	tests := []struct {
		name  string
		csv   string
		value int
		check func(t *testing.T, result string)
	}{
		{
			name:  "remove from middle",
			csv:   "1,2,3,4,5",
			value: 3,
			check: func(t *testing.T, result string) {
				if CSVContains(result, 3) {
					t.Error("Result should not contain 3")
				}
				if CSVLength(result) != 4 {
					t.Errorf("Result length = %d, want 4", CSVLength(result))
				}
			},
		},
		{
			name:  "remove from start",
			csv:   "1,2,3",
			value: 1,
			check: func(t *testing.T, result string) {
				if CSVContains(result, 1) {
					t.Error("Result should not contain 1")
				}
			},
		},
		{
			name:  "remove non-existent",
			csv:   "1,2,3",
			value: 99,
			check: func(t *testing.T, result string) {
				if CSVLength(result) != 3 {
					t.Errorf("Length should remain 3, got %d", CSVLength(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVRemove(tt.csv, tt.value)
			tt.check(t, result)
		})
	}
}

func TestCSVContains(t *testing.T) {
	tests := []struct {
		name     string
		csv      string
		value    int
		expected bool
	}{
		{"contains in middle", "1,2,3,4,5", 3, true},
		{"contains at start", "1,2,3", 1, true},
		{"contains at end", "1,2,3", 3, true},
		{"does not contain", "1,2,3", 5, false},
		{"empty csv", "", 1, false},
		{"single value match", "42", 42, true},
		{"single value no match", "42", 43, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVContains(tt.csv, tt.value)
			if result != tt.expected {
				t.Errorf("CSVContains(%q, %d) = %v, want %v", tt.csv, tt.value, result, tt.expected)
			}
		})
	}
}

func TestCSVLength(t *testing.T) {
	tests := []struct {
		name     string
		csv      string
		expected int
	}{
		{"empty", "", 0},
		{"single", "1", 1},
		{"multiple", "1,2,3,4,5", 5},
		{"two", "10,20", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVLength(tt.csv)
			if result != tt.expected {
				t.Errorf("CSVLength(%q) = %d, want %d", tt.csv, result, tt.expected)
			}
		})
	}
}

func TestCSVElems(t *testing.T) {
	tests := []struct {
		name     string
		csv      string
		expected []int
	}{
		{"empty", "", []int{}},
		{"single", "42", []int{42}},
		{"multiple", "1,2,3,4,5", []int{1, 2, 3, 4, 5}},
		{"negative numbers", "-1,0,1", []int{-1, 0, 1}},
		{"large numbers", "100,200,300", []int{100, 200, 300}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVElems(tt.csv)
			if len(result) != len(tt.expected) {
				t.Errorf("CSVElems(%q) length = %d, want %d", tt.csv, len(result), len(tt.expected))
			}
			for i, v := range tt.expected {
				if i >= len(result) || result[i] != v {
					t.Errorf("CSVElems(%q)[%d] = %d, want %d", tt.csv, i, result[i], v)
				}
			}
		})
	}
}

func TestCSVGetIndex(t *testing.T) {
	csv := "10,20,30,40,50"

	tests := []struct {
		name     string
		index    int
		expected int
	}{
		{"first", 0, 10},
		{"middle", 2, 30},
		{"last", 4, 50},
		{"out of bounds", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVGetIndex(csv, tt.index)
			if result != tt.expected {
				t.Errorf("CSVGetIndex(%q, %d) = %d, want %d", csv, tt.index, result, tt.expected)
			}
		})
	}
}

func TestCSVSetIndex(t *testing.T) {
	tests := []struct {
		name     string
		csv      string
		index    int
		value    int
		check    func(t *testing.T, result string)
	}{
		{
			name:  "set first",
			csv:   "10,20,30",
			index: 0,
			value: 99,
			check: func(t *testing.T, result string) {
				if CSVGetIndex(result, 0) != 99 {
					t.Errorf("Index 0 = %d, want 99", CSVGetIndex(result, 0))
				}
			},
		},
		{
			name:  "set middle",
			csv:   "10,20,30",
			index: 1,
			value: 88,
			check: func(t *testing.T, result string) {
				if CSVGetIndex(result, 1) != 88 {
					t.Errorf("Index 1 = %d, want 88", CSVGetIndex(result, 1))
				}
			},
		},
		{
			name:  "set last",
			csv:   "10,20,30",
			index: 2,
			value: 77,
			check: func(t *testing.T, result string) {
				if CSVGetIndex(result, 2) != 77 {
					t.Errorf("Index 2 = %d, want 77", CSVGetIndex(result, 2))
				}
			},
		},
		{
			name:  "set out of bounds",
			csv:   "10,20,30",
			index: 10,
			value: 99,
			check: func(t *testing.T, result string) {
				// Should not modify the CSV
				if CSVLength(result) != 3 {
					t.Errorf("CSV length changed when setting out of bounds")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CSVSetIndex(tt.csv, tt.index, tt.value)
			tt.check(t, result)
		})
	}
}

func TestCSV_CompleteWorkflow(t *testing.T) {
	// Test a complete workflow
	csv := ""

	// Add elements
	csv = CSVAdd(csv, 10)
	csv = CSVAdd(csv, 20)
	csv = CSVAdd(csv, 30)

	if CSVLength(csv) != 3 {
		t.Errorf("Length = %d, want 3", CSVLength(csv))
	}

	// Check contains
	if !CSVContains(csv, 20) {
		t.Error("Should contain 20")
	}

	// Get element
	if CSVGetIndex(csv, 1) != 20 {
		t.Errorf("Index 1 = %d, want 20", CSVGetIndex(csv, 1))
	}

	// Set element
	csv = CSVSetIndex(csv, 1, 99)
	if CSVGetIndex(csv, 1) != 99 {
		t.Errorf("Index 1 = %d, want 99 after set", CSVGetIndex(csv, 1))
	}

	// Remove element
	csv = CSVRemove(csv, 99)
	if CSVContains(csv, 99) {
		t.Error("Should not contain 99 after removal")
	}

	if CSVLength(csv) != 2 {
		t.Errorf("Length = %d, want 2 after removal", CSVLength(csv))
	}
}

func BenchmarkCSVAdd(b *testing.B) {
	csv := "1,2,3,4,5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CSVAdd(csv, 6)
	}
}

func BenchmarkCSVContains(b *testing.B) {
	csv := "1,2,3,4,5,6,7,8,9,10"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CSVContains(csv, 5)
	}
}

func BenchmarkCSVRemove(b *testing.B) {
	csv := "1,2,3,4,5,6,7,8,9,10"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CSVRemove(csv, 5)
	}
}

func BenchmarkCSVElems(b *testing.B) {
	csv := "1,2,3,4,5,6,7,8,9,10"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CSVElems(csv)
	}
}

func TestSJISToUTF8Lossy(t *testing.T) {
	// Valid SJIS (ASCII subset) decodes correctly.
	got := SJISToUTF8Lossy([]byte("Hello"))
	if got != "Hello" {
		t.Errorf("SJISToUTF8Lossy(valid) = %q, want %q", got, "Hello")
	}

	// Truncated multi-byte SJIS sequence (lead byte 0x82 without trail byte)
	// does not panic and returns some result (lossy).
	got = SJISToUTF8Lossy([]byte{0x82})
	_ = got // must not panic

	// Nil input returns empty string.
	got = SJISToUTF8Lossy(nil)
	if got != "" {
		t.Errorf("SJISToUTF8Lossy(nil) = %q, want %q", got, "")
	}
}

func TestUTF8ToSJIS_UnsupportedCharacters(t *testing.T) {
	// Regression test for PR #116: Characters outside the Shift-JIS range
	// (e.g. Lenny face, cuneiform) previously caused a panic in UTF8ToSJIS,
	// crashing the server when relayed from Discord.
	tests := []struct {
		name  string
		input string
	}{
		{"lenny_face", "( ͡° ͜ʖ ͡°)"},
		{"cuneiform", "𒀜"},
		{"emoji", "Hello 🎮 World"},
		{"mixed_unsupported", "Test ͡° message 𒀜 here"},
		{"zalgo_text", "H̷e̸l̵l̶o̷"},
		{"only_unsupported", "🎮🎲🎯"},
		{"cyrillic", "Привет"},
		{"arabic", "مرحبا"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Must not panic - the old code would panic here
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("UTF8ToSJIS panicked on input %q: %v", tt.input, r)
				}
			}()
			result := UTF8ToSJIS(tt.input)
			if result == nil {
				t.Error("UTF8ToSJIS returned nil")
			}
		})
	}
}

func TestUTF8ToSJIS_PreservesValidContent(t *testing.T) {
	// Verify that valid Shift-JIS content is preserved when mixed with
	// unsupported characters.
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ascii_with_emoji", "Hello 🎮 World", "Hello  World"},
		{"japanese_with_emoji", "テスト🎮データ", "テストデータ"},
		{"only_valid", "Hello World", "Hello World"},
		{"only_invalid", "🎮🎲🎯", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sjis := UTF8ToSJIS(tt.input)
			roundTripped, _ := SJISToUTF8(sjis)
			if roundTripped != tt.expected {
				t.Errorf("UTF8ToSJIS(%q) round-tripped to %q, want %q", tt.input, roundTripped, tt.expected)
			}
		})
	}
}

func TestToNGWord_UnsupportedCharacters(t *testing.T) {
	// ToNGWord also calls UTF8ToSJIS internally, so it must not panic either.
	inputs := []string{"( ͡° ͜ʖ ͡°)", "🎮", "Hello 🎮 World"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ToNGWord panicked on input %q: %v", input, r)
				}
			}()
			_ = ToNGWord(input)
		})
	}
}

func BenchmarkUTF8ToSJIS(b *testing.B) {
	text := "Hello World テスト"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UTF8ToSJIS(text)
	}
}

func BenchmarkSJISToUTF8(b *testing.B) {
	text := []byte("Hello World")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SJISToUTF8(text)
	}
}

func BenchmarkPaddedString(b *testing.B) {
	text := "Test String"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PaddedString(text, 50, false)
	}
}

func BenchmarkToNGWord(b *testing.B) {
	text := "TestString"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToNGWord(text)
	}
}
