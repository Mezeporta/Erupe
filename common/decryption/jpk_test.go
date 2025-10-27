package decryption

import (
	"bytes"
	"erupe-ce/common/byteframe"
	"io"
	"testing"
)

func TestUnpackSimple_UncompressedData(t *testing.T) {
	// Test data that doesn't have JPK header - should be returned as-is
	input := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	result := UnpackSimple(input)

	if !bytes.Equal(result, input) {
		t.Errorf("UnpackSimple() with uncompressed data should return input as-is, got %v, want %v", result, input)
	}
}

func TestUnpackSimple_InvalidHeader(t *testing.T) {
	// Test data with wrong header
	input := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x02, 0x03, 0x04}
	result := UnpackSimple(input)

	if !bytes.Equal(result, input) {
		t.Errorf("UnpackSimple() with invalid header should return input as-is, got %v, want %v", result, input)
	}
}

func TestUnpackSimple_JPKHeaderWrongType(t *testing.T) {
	// Test JPK header but wrong type (not type 3)
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0x1A524B4A) // JPK header
	bf.WriteUint16(0x00)       // Reserved
	bf.WriteUint16(1)          // Type 1 instead of 3
	bf.WriteInt32(12)          // Start offset
	bf.WriteInt32(10)          // Out size

	result := UnpackSimple(bf.Data())
	// Should return the input as-is since it's not type 3
	if !bytes.Equal(result, bf.Data()) {
		t.Error("UnpackSimple() with non-type-3 JPK should return input as-is")
	}
}

func TestUnpackSimple_ValidJPKType3_EmptyData(t *testing.T) {
	// Create a valid JPK type 3 header with minimal compressed data
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0x1A524B4A) // JPK header "JKR\x1A"
	bf.WriteUint16(0x00)       // Reserved
	bf.WriteUint16(3)          // Type 3
	bf.WriteInt32(12)          // Start offset (points to byte 12, after header)
	bf.WriteInt32(0)           // Out size (empty output)

	result := UnpackSimple(bf.Data())
	// Should return empty buffer
	if len(result) != 0 {
		t.Errorf("UnpackSimple() with zero output size should return empty slice, got length %d", len(result))
	}
}

func TestUnpackSimple_JPKHeader(t *testing.T) {
	// Test that the function correctly identifies JPK header (0x1A524B4A = "JKR\x1A" in little endian)
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0x1A524B4A) // Correct JPK magic

	data := bf.Data()
	if len(data) < 4 {
		t.Fatal("Not enough data written")
	}

	// Verify the header bytes are correct
	bf.Seek(0, io.SeekStart)
	header := bf.ReadUint32()
	if header != 0x1A524B4A {
		t.Errorf("Header = 0x%X, want 0x1A524B4A", header)
	}
}

func TestJPKBitShift_Initialization(t *testing.T) {
	// Test that the function doesn't crash with bad initial global state
	mShiftIndex = 10
	mFlag = 0xFF

	// Create data without JPK header (will return as-is)
	// Need at least 4 bytes since UnpackSimple reads a uint32 header
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD) // Not a JPK header

	data := bf.Data()
	result := UnpackSimple(data)

	// Without JPK header, should return data as-is
	if !bytes.Equal(result, data) {
		t.Error("UnpackSimple with non-JPK data should return input as-is")
	}
}

func TestReadByte(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0x42)
	bf.WriteUint8(0xAB)

	bf.Seek(0, io.SeekStart)
	b1 := ReadByte(bf)
	b2 := ReadByte(bf)

	if b1 != 0x42 {
		t.Errorf("ReadByte() = 0x%X, want 0x42", b1)
	}
	if b2 != 0xAB {
		t.Errorf("ReadByte() = 0x%X, want 0xAB", b2)
	}
}

func TestJPKCopy(t *testing.T) {
	outBuffer := make([]byte, 20)
	// Set up some initial data
	outBuffer[0] = 'A'
	outBuffer[1] = 'B'
	outBuffer[2] = 'C'

	index := 3
	// Copy 3 bytes from offset 2 (looking back 2+1=3 positions)
	JPKCopy(outBuffer, 2, 3, &index)

	// Should have copied 'A', 'B', 'C' to positions 3, 4, 5
	if outBuffer[3] != 'A' || outBuffer[4] != 'B' || outBuffer[5] != 'C' {
		t.Errorf("JPKCopy failed: got %v at positions 3-5, want ['A', 'B', 'C']", outBuffer[3:6])
	}
	if index != 6 {
		t.Errorf("index = %d, want 6", index)
	}
}

func TestJPKCopy_OverlappingCopy(t *testing.T) {
	// Test copying with overlapping regions (common in LZ-style compression)
	outBuffer := make([]byte, 20)
	outBuffer[0] = 'X'

	index := 1
	// Copy from 1 position back, 5 times - should repeat the pattern
	JPKCopy(outBuffer, 0, 5, &index)

	// Should produce: X X X X X (repeating X)
	for i := 1; i < 6; i++ {
		if outBuffer[i] != 'X' {
			t.Errorf("outBuffer[%d] = %c, want 'X'", i, outBuffer[i])
		}
	}
	if index != 6 {
		t.Errorf("index = %d, want 6", index)
	}
}

func TestProcessDecode_EmptyOutput(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0x00)

	outBuffer := make([]byte, 0)
	// Should not panic with empty output buffer
	ProcessDecode(bf, outBuffer)
}

func TestUnpackSimple_EdgeCases(t *testing.T) {
	// Test with data that has at least 4 bytes (header size required)
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "four bytes non-JPK",
			input: []byte{0x00, 0x01, 0x02, 0x03},
		},
		{
			name:  "partial header padded",
			input: []byte{0x4A, 0x4B, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnpackSimple(tt.input)
			// Should return input as-is without crashing
			if !bytes.Equal(result, tt.input) {
				t.Errorf("UnpackSimple() = %v, want %v", result, tt.input)
			}
		})
	}
}

func BenchmarkUnpackSimple_Uncompressed(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnpackSimple(data)
	}
}

func BenchmarkUnpackSimple_JPKHeader(b *testing.B) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0x1A524B4A) // JPK header
	bf.WriteUint16(0x00)
	bf.WriteUint16(3)
	bf.WriteInt32(12)
	bf.WriteInt32(0)
	data := bf.Data()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnpackSimple(data)
	}
}

func BenchmarkReadByte(b *testing.B) {
	bf := byteframe.NewByteFrame()
	for i := 0; i < 1000; i++ {
		bf.WriteUint8(byte(i % 256))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Seek(0, io.SeekStart)
		_ = ReadByte(bf)
	}
}
