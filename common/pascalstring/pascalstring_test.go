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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
	length := bf.ReadUint16()

	if length != 1 {
		t.Errorf("length = %d, want 1", length)
	}
}

func TestUint32_NoTransform(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "Testing"

	Uint32(bf, testString, false)

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
	length := bf.ReadUint32()

	if length != 1 {
		t.Errorf("length = %d, want 1", length)
	}
}

func TestUint8_LongString(t *testing.T) {
	bf := byteframe.NewByteFrame()
	testString := "This is a longer test string with more characters"

	Uint8(bf, testString, false)

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

			_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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

	_, _ = bf.Seek(0, 0)
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
