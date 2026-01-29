package network

import (
	"bytes"
	"testing"
)

func TestCryptPacketHeaderLength(t *testing.T) {
	if CryptPacketHeaderLength != 14 {
		t.Errorf("CryptPacketHeaderLength = %d, want 14", CryptPacketHeaderLength)
	}
}

func TestNewCryptPacketHeader(t *testing.T) {
	// Create a valid 14-byte header
	data := []byte{
		0x01,       // Pf0
		0x02,       // KeyRotDelta
		0x00, 0x03, // PacketNum
		0x00, 0x04, // DataSize
		0x00, 0x05, // PrevPacketCombinedCheck
		0x00, 0x06, // Check0
		0x00, 0x07, // Check1
		0x00, 0x08, // Check2
	}

	header, err := NewCryptPacketHeader(data)
	if err != nil {
		t.Fatalf("NewCryptPacketHeader() error = %v", err)
	}

	if header.Pf0 != 0x01 {
		t.Errorf("Pf0 = %d, want 1", header.Pf0)
	}
	if header.KeyRotDelta != 0x02 {
		t.Errorf("KeyRotDelta = %d, want 2", header.KeyRotDelta)
	}
	if header.PacketNum != 0x03 {
		t.Errorf("PacketNum = %d, want 3", header.PacketNum)
	}
	if header.DataSize != 0x04 {
		t.Errorf("DataSize = %d, want 4", header.DataSize)
	}
	if header.PrevPacketCombinedCheck != 0x05 {
		t.Errorf("PrevPacketCombinedCheck = %d, want 5", header.PrevPacketCombinedCheck)
	}
	if header.Check0 != 0x06 {
		t.Errorf("Check0 = %d, want 6", header.Check0)
	}
	if header.Check1 != 0x07 {
		t.Errorf("Check1 = %d, want 7", header.Check1)
	}
	if header.Check2 != 0x08 {
		t.Errorf("Check2 = %d, want 8", header.Check2)
	}
}

func TestNewCryptPacketHeaderTooShort(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"1 byte", []byte{0x01}},
		{"5 bytes", []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
		{"13 bytes", make([]byte, 13)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCryptPacketHeader(tt.data)
			if err == nil {
				t.Errorf("NewCryptPacketHeader(%v) should return error for short data", tt.data)
			}
		})
	}
}

func TestCryptPacketHeaderEncode(t *testing.T) {
	header := &CryptPacketHeader{
		Pf0:                     0x01,
		KeyRotDelta:             0x02,
		PacketNum:               0x0003,
		DataSize:                0x0004,
		PrevPacketCombinedCheck: 0x0005,
		Check0:                  0x0006,
		Check1:                  0x0007,
		Check2:                  0x0008,
	}

	encoded, err := header.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if len(encoded) != CryptPacketHeaderLength {
		t.Errorf("Encode() len = %d, want %d", len(encoded), CryptPacketHeaderLength)
	}

	expected := []byte{
		0x01,       // Pf0
		0x02,       // KeyRotDelta
		0x00, 0x03, // PacketNum
		0x00, 0x04, // DataSize
		0x00, 0x05, // PrevPacketCombinedCheck
		0x00, 0x06, // Check0
		0x00, 0x07, // Check1
		0x00, 0x08, // Check2
	}

	if !bytes.Equal(encoded, expected) {
		t.Errorf("Encode() = %v, want %v", encoded, expected)
	}
}

func TestCryptPacketHeaderRoundTrip(t *testing.T) {
	tests := []CryptPacketHeader{
		{
			Pf0:                     0x00,
			KeyRotDelta:             0x00,
			PacketNum:               0x0000,
			DataSize:                0x0000,
			PrevPacketCombinedCheck: 0x0000,
			Check0:                  0x0000,
			Check1:                  0x0000,
			Check2:                  0x0000,
		},
		{
			Pf0:                     0xFF,
			KeyRotDelta:             0xFF,
			PacketNum:               0xFFFF,
			DataSize:                0xFFFF,
			PrevPacketCombinedCheck: 0xFFFF,
			Check0:                  0xFFFF,
			Check1:                  0xFFFF,
			Check2:                  0xFFFF,
		},
		{
			Pf0:                     0x12,
			KeyRotDelta:             0x34,
			PacketNum:               0x5678,
			DataSize:                0x9ABC,
			PrevPacketCombinedCheck: 0xDEF0,
			Check0:                  0x1234,
			Check1:                  0x5678,
			Check2:                  0x9ABC,
		},
	}

	for i, original := range tests {
		t.Run("", func(t *testing.T) {
			encoded, err := original.Encode()
			if err != nil {
				t.Fatalf("Test %d: Encode() error = %v", i, err)
			}

			decoded, err := NewCryptPacketHeader(encoded)
			if err != nil {
				t.Fatalf("Test %d: NewCryptPacketHeader() error = %v", i, err)
			}

			if decoded.Pf0 != original.Pf0 {
				t.Errorf("Test %d: Pf0 = %d, want %d", i, decoded.Pf0, original.Pf0)
			}
			if decoded.KeyRotDelta != original.KeyRotDelta {
				t.Errorf("Test %d: KeyRotDelta = %d, want %d", i, decoded.KeyRotDelta, original.KeyRotDelta)
			}
			if decoded.PacketNum != original.PacketNum {
				t.Errorf("Test %d: PacketNum = %d, want %d", i, decoded.PacketNum, original.PacketNum)
			}
			if decoded.DataSize != original.DataSize {
				t.Errorf("Test %d: DataSize = %d, want %d", i, decoded.DataSize, original.DataSize)
			}
			if decoded.PrevPacketCombinedCheck != original.PrevPacketCombinedCheck {
				t.Errorf("Test %d: PrevPacketCombinedCheck = %d, want %d", i, decoded.PrevPacketCombinedCheck, original.PrevPacketCombinedCheck)
			}
			if decoded.Check0 != original.Check0 {
				t.Errorf("Test %d: Check0 = %d, want %d", i, decoded.Check0, original.Check0)
			}
			if decoded.Check1 != original.Check1 {
				t.Errorf("Test %d: Check1 = %d, want %d", i, decoded.Check1, original.Check1)
			}
			if decoded.Check2 != original.Check2 {
				t.Errorf("Test %d: Check2 = %d, want %d", i, decoded.Check2, original.Check2)
			}
		})
	}
}

func TestCryptPacketHeaderBigEndian(t *testing.T) {
	// Verify big-endian encoding
	header := &CryptPacketHeader{
		PacketNum: 0x1234,
	}

	encoded, err := header.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// PacketNum is at bytes 2-3 (after Pf0 and KeyRotDelta)
	if encoded[2] != 0x12 || encoded[3] != 0x34 {
		t.Errorf("PacketNum encoding is not big-endian: %v", encoded[2:4])
	}
}

func TestNewCryptPacketHeaderExtraBytes(t *testing.T) {
	// Test with more than required bytes (should still work)
	data := make([]byte, 20)
	data[0] = 0x01 // Pf0

	header, err := NewCryptPacketHeader(data)
	if err != nil {
		t.Fatalf("NewCryptPacketHeader() with extra bytes error = %v", err)
	}

	if header.Pf0 != 0x01 {
		t.Errorf("Pf0 = %d, want 1", header.Pf0)
	}
}

func TestCryptPacketHeaderZeroValues(t *testing.T) {
	header := &CryptPacketHeader{}

	encoded, err := header.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	expected := make([]byte, CryptPacketHeaderLength)
	if !bytes.Equal(encoded, expected) {
		t.Errorf("Encode() zero header = %v, want all zeros", encoded)
	}
}
