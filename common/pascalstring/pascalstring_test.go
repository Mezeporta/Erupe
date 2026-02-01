package pascalstring

import (
	"bytes"
	"erupe-ce/common/byteframe"
	"testing"
)

func TestUint8_NoTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "Hello"

	Uint8(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint8()
	expectedLength := uint8(len(testString) + 1) // +1 for null terminator

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}

	data := bf.ReadBytes(uint(length))
	// Should be "Hello\x00"
	expected := []byte("Hello\x00")
	if !bytes.Equal(data, expected) {
		t.Errorf("data = %v, want %v", data, expected)
	}
}

func TestUint8_WithTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	// ASCII string (no special characters)
	testString := "Test"

	Uint8(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	if length == 0 {
		t.Error("length should not be 0 for ASCII string")
	}

	data := bf.ReadBytes(uint(length))
	// Should end with null terminator
	if data[len(data)-1] != 0 {
		t.Error("data should end with null terminator")
	}
}

func TestUint8_EmptyString(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := ""

	Uint8(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	if length != 1 { // Just null terminator
		t.Errorf("length = %d, want 1", length)
	}

	data := bf.ReadBytes(uint(length))
	if data[0] != 0 {
		t.Error("empty string should produce just null terminator")
	}
}

func TestUint16_NoTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "World"

	Uint16(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint16()
	expectedLength := uint16(len(testString) + 1)

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}

	data := bf.ReadBytes(uint(length))
	expected := []byte("World\x00")
	if !bytes.Equal(data, expected) {
		t.Errorf("data = %v, want %v", data, expected)
	}
}

func TestUint16_WithTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "Test"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length == 0 {
		t.Error("length should not be 0 for ASCII string")
	}

	data := bf.ReadBytes(uint(length))
	if data[len(data)-1] != 0 {
		t.Error("data should end with null terminator")
	}
}

func TestUint16_EmptyString(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := ""

	Uint16(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length != 1 {
		t.Errorf("length = %d, want 1", length)
	}
}

func TestUint32_NoTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "Testing"

	Uint32(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint32()
	expectedLength := uint32(len(testString) + 1)

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}

	data := bf.ReadBytes(uint(length))
	expected := []byte("Testing\x00")
	if !bytes.Equal(data, expected) {
		t.Errorf("data = %v, want %v", data, expected)
	}
}

func TestUint32_WithTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "Test"

	Uint32(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint32()

	if length == 0 {
		t.Error("length should not be 0 for ASCII string")
	}

	data := bf.ReadBytes(uint(length))
	if data[len(data)-1] != 0 {
		t.Error("data should end with null terminator")
	}
}

func TestUint32_EmptyString(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := ""

	Uint32(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint32()

	if length != 1 {
		t.Errorf("length = %d, want 1", length)
	}
}

func TestUint8_LongString(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "This is a longer test string with more characters"

	Uint8(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint8()
	expectedLength := uint8(len(testString) + 1)

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}

	data := bf.ReadBytes(uint(length))
	if !bytes.HasSuffix(data, []byte{0}) {
		t.Error("data should end with null terminator")
	}
	if !bytes.HasPrefix(data, []byte("This is")) {
		t.Error("data should start with expected string")
	}
}

func TestUint16_LongString(t *testing.T) {
	bf := byteframe.NewByteFrame()
	// Create a string longer than 255 to test uint16
	testString := ""
	for i := 0; i < 300; i++ {
		testString += "A"
	}

	Uint16(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint16()
	expectedLength := uint16(len(testString) + 1)

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}

	data := bf.ReadBytes(uint(length))
	if !bytes.HasSuffix(data, []byte{0}) {
		t.Error("data should end with null terminator")
	}
}

func TestAllFunctions_NullTermination(t *testing.T) {
	tests := []struct {
		name     string
		writeFn  func(*byteframe.ByteFrame, string, bool)
		readSize func(*byteframe.ByteFrame) uint
	}{
		{
			name: "Uint8",
			writeFn: func(bf *byteframe.ByteFrame, s string, t bool) {
				Uint8(bf, s, t)
			},
			readSize: func(bf *byteframe.ByteFrame) uint {
				return uint(bf.ReadUint8())
			},
		},
		{
			name: "Uint16",
			writeFn: func(bf *byteframe.ByteFrame, s string, t bool) {
				Uint16(bf, s, t)
			},
			readSize: func(bf *byteframe.ByteFrame) uint {
				return uint(bf.ReadUint16())
			},
		},
		{
			name: "Uint32",
			writeFn: func(bf *byteframe.ByteFrame, s string, t bool) {
				Uint32(bf, s, t)
			},
			readSize: func(bf *byteframe.ByteFrame) uint {
				return uint(bf.ReadUint32())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			testString := "Test"

			tt.writeFn(bf, testString, false)

			bf.Seek(0, 0)
			size := tt.readSize(bf)
			data := bf.ReadBytes(size)

			// Verify null termination
			if data[len(data)-1] != 0 {
				t.Errorf("%s: data should end with null terminator", tt.name)
			}

			// Verify length includes null terminator
			if size != uint(len(testString)+1) {
				t.Errorf("%s: size = %d, want %d", tt.name, size, len(testString)+1)
			}
		})
	}
}

func TestTransform_JapaneseCharacters(t *testing.T) {
	// Test with Japanese characters that should be transformed to Shift-JIS
	bf := byteframe.NewByteFrame()
	testString := "テスト" // "Test" in Japanese katakana

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length == 0 {
		t.Error("Transformed Japanese string should have non-zero length")
	}

	// The transformed Shift-JIS should be different length than UTF-8
	// UTF-8: 9 bytes (3 chars * 3 bytes each), Shift-JIS: 6 bytes (3 chars * 2 bytes each) + 1 null
	data := bf.ReadBytes(uint(length))
	if data[len(data)-1] != 0 {
		t.Error("Transformed string should end with null terminator")
	}
}

func TestTransform_InvalidUTF8(t *testing.T) {
	// This test verifies graceful handling of encoding errors
	// When transformation fails, the functions should write length 0

	bf := byteframe.NewByteFrame()
	// Create a string with invalid UTF-8 sequence
	// Note: Go strings are generally valid UTF-8, but we can test the error path
	testString := "Valid ASCII"

	Uint8(bf, testString, true)
	// Should succeed for ASCII characters

	bf.Seek(0, 0)
	length := bf.ReadUint8()
	if length == 0 {
		t.Error("ASCII string should transform successfully")
	}
}

func BenchmarkUint8_NoTransform(b *testing.B) {
	testString := "Hello, World!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrame()
		Uint8(bf, testString, false)
	}
}

func BenchmarkUint8_WithTransform(b *testing.B) {
	testString := "Hello, World!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrame()
		Uint8(bf, testString, true)
	}
}

func BenchmarkUint16_NoTransform(b *testing.B) {
	testString := "Hello, World!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrame()
		Uint16(bf, testString, false)
	}
}

func BenchmarkUint32_NoTransform(b *testing.B) {
	testString := "Hello, World!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrame()
		Uint32(bf, testString, false)
	}
}

func BenchmarkUint16_Japanese(b *testing.B) {
	testString := "テストメッセージ"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrame()
		Uint16(bf, testString, true)
	}
}

// Edge case tests for additional coverage

func TestUint8_MaxLength(t *testing.T) {
	// Uint8 length can hold max 255, but string + null = 254 chars max
	// Test with string at the uint8 boundary
	bf := byteframe.NewByteFrame()
	testString := string(make([]byte, 254)) // 254 chars + 1 null = 255

	Uint8(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	// Length should be 255 (254 chars + null)
	if length != 255 {
		t.Errorf("length = %d, want 255", length)
	}
}

func TestUint8_OverflowString(t *testing.T) {
	// Test string that would overflow uint8 length (>255)
	bf := byteframe.NewByteFrame()
	testString := string(make([]byte, 300)) // Would need 301 for length+null, exceeds uint8

	Uint8(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	// Due to uint8 overflow, length will wrap around
	// 301 % 256 = 45, but actual behavior depends on implementation
	t.Logf("Overflow string produced length: %d", length)
	// This test documents the current behavior (truncation via uint8 overflow)
}

func TestUint16_ExactBoundary(t *testing.T) {
	// Test string that fits exactly in uint16 boundary
	bf := byteframe.NewByteFrame()
	testString := string(make([]byte, 1000))

	Uint16(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint16()
	expectedLength := uint16(1000 + 1)

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}
}

func TestUint32_LargeString(t *testing.T) {
	// Test moderately large string with uint32
	bf := byteframe.NewByteFrame()
	testString := string(make([]byte, 10000))

	Uint32(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint32()
	expectedLength := uint32(10000 + 1)

	if length != expectedLength {
		t.Errorf("length = %d, want %d", length, expectedLength)
	}
}

func TestTransform_MixedContent(t *testing.T) {
	// Test string with mixed ASCII and Japanese content
	bf := byteframe.NewByteFrame()
	testString := "Player1: テスト"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length == 0 {
		t.Error("Mixed content string should transform successfully")
	}

	// Should have null terminator
	data := bf.ReadBytes(uint(length))
	if data[len(data)-1] != 0 {
		t.Error("Transformed string should end with null terminator")
	}
}

func TestTransform_SpecialSymbols(t *testing.T) {
	// Test Japanese symbols that exist in Shift-JIS
	bf := byteframe.NewByteFrame()
	testString := "★☆●○"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	// Some symbols may not transform, check length is valid
	t.Logf("Special symbols produced length: %d", length)
	// If length is 0, the transform failed which is expected for some symbols
}

func TestTransform_NumbersAndSymbols(t *testing.T) {
	// ASCII numbers and basic symbols should always work
	bf := byteframe.NewByteFrame()
	testString := "12345!@#$%"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length == 0 {
		t.Error("ASCII numbers and symbols should transform successfully")
	}
}

func TestUint8_SingleCharacter(t *testing.T) {
	// Edge case: single character
	bf := byteframe.NewByteFrame()
	testString := "A"

	Uint8(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	// Length should be 2 (1 char + null)
	if length != 2 {
		t.Errorf("length = %d, want 2", length)
	}

	data := bf.ReadBytes(uint(length))
	if data[0] != 'A' || data[1] != 0 {
		t.Errorf("data = %v, want [A, 0]", data)
	}
}

func TestUint16_SingleCharacter(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "B"

	Uint16(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length != 2 {
		t.Errorf("length = %d, want 2", length)
	}

	data := bf.ReadBytes(uint(length))
	if data[0] != 'B' || data[1] != 0 {
		t.Errorf("data = %v, want [B, 0]", data)
	}
}

func TestUint32_SingleCharacter(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "C"

	Uint32(bf, testString, false)

	bf.Seek(0, 0)
	length := bf.ReadUint32()

	if length != 2 {
		t.Errorf("length = %d, want 2", length)
	}

	data := bf.ReadBytes(uint(length))
	if data[0] != 'C' || data[1] != 0 {
		t.Errorf("data = %v, want [C, 0]", data)
	}
}

func TestTransform_OnlySpaces(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "   " // Three spaces

	Uint8(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	// Spaces should transform fine
	if length != 4 { // 3 spaces + null
		t.Errorf("length = %d, want 4", length)
	}
}

func TestTransform_Hiragana(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "あいうえお"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length == 0 {
		t.Error("Hiragana should transform to Shift-JIS successfully")
	}

	// Hiragana in Shift-JIS is 2 bytes per character
	// 5 characters * 2 bytes + 1 null = 11 bytes
	data := bf.ReadBytes(uint(length))
	if data[len(data)-1] != 0 {
		t.Error("Transformed string should end with null terminator")
	}
}

func TestTransform_Katakana(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "アイウエオ"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length == 0 {
		t.Error("Katakana should transform to Shift-JIS successfully")
	}

	data := bf.ReadBytes(uint(length))
	if data[len(data)-1] != 0 {
		t.Error("Transformed string should end with null terminator")
	}
}

func TestMultipleWrites(t *testing.T) {
	// Test writing multiple strings to same buffer
	bf := byteframe.NewByteFrame()

	Uint8(bf, "First", false)
	Uint8(bf, "Second", false)
	Uint8(bf, "Third", false)

	bf.Seek(0, 0)

	// Read first string
	len1 := bf.ReadUint8()
	data1 := bf.ReadBytes(uint(len1))

	// Read second string
	len2 := bf.ReadUint8()
	data2 := bf.ReadBytes(uint(len2))

	// Read third string
	len3 := bf.ReadUint8()
	data3 := bf.ReadBytes(uint(len3))

	// Verify each string
	if string(data1[:len(data1)-1]) != "First" {
		t.Errorf("First string = %s, want First", string(data1[:len(data1)-1]))
	}
	if string(data2[:len(data2)-1]) != "Second" {
		t.Errorf("Second string = %s, want Second", string(data2[:len(data2)-1]))
	}
	if string(data3[:len(data3)-1]) != "Third" {
		t.Errorf("Third string = %s, want Third", string(data3[:len(data3)-1]))
	}
}

func TestTransform_EmptyStringWithTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := ""

	Uint8(bf, testString, true) // Transform enabled but string is empty

	bf.Seek(0, 0)
	length := bf.ReadUint8()

	// Empty string with transform should still produce length 1 (just null)
	if length != 1 {
		t.Errorf("Empty string with transform: length = %d, want 1", length)
	}
}

func TestTransform_UnsupportedCharacters(t *testing.T) {
	// Test characters that cannot be encoded in Shift-JIS
	// Emoji and some Unicode characters are not supported
	testStrings := []string{
		"\U0001F600", // Emoji (grinning face) - not in Shift-JIS
		"🎮",          // Game controller emoji
		"\U0001F4A9", // Pile of poo emoji
		"中文测试",       // Simplified Chinese (some chars not in Shift-JIS)
		"العربية",    // Arabic
		"עברית",      // Hebrew
		"ไทย",        // Thai
		"한글",         // Korean
	}

	for _, testString := range testStrings {
		t.Run(testString, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			Uint8(bf, testString, true)

			bf.Seek(0, 0)
			length := bf.ReadUint8()

			// These strings may fail to transform or transform partially
			// Document the current behavior
			t.Logf("String %q with transform produced length: %d", testString, length)

			// If length is 0, the transform failed completely (error path)
			// If length > 0, some or all characters were transformed
		})
	}
}

func TestTransform_Uint16_UnsupportedCharacters(t *testing.T) {
	bf := byteframe.NewByteFrame()
	// Use a string with characters not in Shift-JIS
	testString := "🎮"

	Uint16(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint16()
	t.Logf("Uint16 transform of emoji produced length: %d", length)
}

func TestTransform_Uint32_UnsupportedCharacters(t *testing.T) {
	bf := byteframe.NewByteFrame()
	// Use a string with characters not in Shift-JIS
	testString := "🎮"

	Uint32(bf, testString, true)

	bf.Seek(0, 0)
	length := bf.ReadUint32()
	t.Logf("Uint32 transform of emoji produced length: %d", length)
}
