package network

import (
	"bytes"
	"testing"
)

func TestNewCryptPacketHeader_ValidData(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected *CryptPacketHeader
	}{
		{
			name: "basic header",
			data: []byte{
				0x03,       // Pf0
				0x03,       // KeyRotDelta
				0x00, 0x01, // PacketNum (1)
				0x00, 0x0A, // DataSize (10)
				0x00, 0x00, // PrevPacketCombinedCheck (0)
				0x12, 0x34, // Check0 (0x1234)
				0x56, 0x78, // Check1 (0x5678)
				0x9A, 0xBC, // Check2 (0x9ABC)
			},
			expected: &CryptPacketHeader{
				Pf0:                     0x03,
				KeyRotDelta:             0x03,
				PacketNum:               1,
				DataSize:                10,
				PrevPacketCombinedCheck: 0,
				Check0:                  0x1234,
				Check1:                  0x5678,
				Check2:                  0x9ABC,
			},
		},
		{
			name: "all zero values",
			data: []byte{
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
			},
			expected: &CryptPacketHeader{
				Pf0:                     0x00,
				KeyRotDelta:             0x00,
				PacketNum:               0,
				DataSize:                0,
				PrevPacketCombinedCheck: 0,
				Check0:                  0,
				Check1:                  0,
				Check2:                  0,
			},
		},
		{
			name: "max values",
			data: []byte{
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
			},
			expected: &CryptPacketHeader{
				Pf0:                     0xFF,
				KeyRotDelta:             0xFF,
				PacketNum:               0xFFFF,
				DataSize:                0xFFFF,
				PrevPacketCombinedCheck: 0xFFFF,
				Check0:                  0xFFFF,
				Check1:                  0xFFFF,
				Check2:                  0xFFFF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewCryptPacketHeader(tt.data)
			if err != nil {
				t.Fatalf("NewCryptPacketHeader() error = %v, want nil", err)
			}

			if result.Pf0 != tt.expected.Pf0 {
				t.Errorf("Pf0 = 0x%X, want 0x%X", result.Pf0, tt.expected.Pf0)
			}
			if result.KeyRotDelta != tt.expected.KeyRotDelta {
				t.Errorf("KeyRotDelta = 0x%X, want 0x%X", result.KeyRotDelta, tt.expected.KeyRotDelta)
			}
			if result.PacketNum != tt.expected.PacketNum {
				t.Errorf("PacketNum = 0x%X, want 0x%X", result.PacketNum, tt.expected.PacketNum)
			}
			if result.DataSize != tt.expected.DataSize {
				t.Errorf("DataSize = 0x%X, want 0x%X", result.DataSize, tt.expected.DataSize)
			}
			if result.PrevPacketCombinedCheck != tt.expected.PrevPacketCombinedCheck {
				t.Errorf("PrevPacketCombinedCheck = 0x%X, want 0x%X", result.PrevPacketCombinedCheck, tt.expected.PrevPacketCombinedCheck)
			}
			if result.Check0 != tt.expected.Check0 {
				t.Errorf("Check0 = 0x%X, want 0x%X", result.Check0, tt.expected.Check0)
			}
			if result.Check1 != tt.expected.Check1 {
				t.Errorf("Check1 = 0x%X, want 0x%X", result.Check1, tt.expected.Check1)
			}
			if result.Check2 != tt.expected.Check2 {
				t.Errorf("Check2 = 0x%X, want 0x%X", result.Check2, tt.expected.Check2)
			}
		})
	}
}

func TestNewCryptPacketHeader_InvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "too short - 1 byte",
			data: []byte{0x03},
		},
		{
			name: "too short - 13 bytes",
			data: []byte{0x03, 0x03, 0x00, 0x01, 0x00, 0x0A, 0x00, 0x00, 0x12, 0x34, 0x56, 0x78, 0x9A},
		},
		{
			name: "too short - 7 bytes",
			data: []byte{0x03, 0x03, 0x00, 0x01, 0x00, 0x0A, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCryptPacketHeader(tt.data)
			if err == nil {
				t.Fatal("NewCryptPacketHeader() error = nil, want error")
			}
		})
	}
}

func TestNewCryptPacketHeader_ExtraDataIgnored(t *testing.T) {
	// Test that extra data beyond 14 bytes is ignored
	data := []byte{
		0x03, 0x03,
		0x00, 0x01,
		0x00, 0x0A,
		0x00, 0x00,
		0x12, 0x34,
		0x56, 0x78,
		0x9A, 0xBC,
		0xFF, 0xFF, 0xFF, // Extra bytes
	}

	result, err := NewCryptPacketHeader(data)
	if err != nil {
		t.Fatalf("NewCryptPacketHeader() error = %v, want nil", err)
	}

	expected := &CryptPacketHeader{
		Pf0:                     0x03,
		KeyRotDelta:             0x03,
		PacketNum:               1,
		DataSize:                10,
		PrevPacketCombinedCheck: 0,
		Check0:                  0x1234,
		Check1:                  0x5678,
		Check2:                  0x9ABC,
	}

	if result.Pf0 != expected.Pf0 || result.KeyRotDelta != expected.KeyRotDelta ||
		result.PacketNum != expected.PacketNum || result.DataSize != expected.DataSize {
		t.Errorf("Extra data affected parsing")
	}
}

func TestCryptPacketHeader_Encode(t *testing.T) {
	tests := []struct {
		name     string
		header   *CryptPacketHeader
		expected []byte
	}{
		{
			name: "basic header",
			header: &CryptPacketHeader{
				Pf0:                     0x03,
				KeyRotDelta:             0x03,
				PacketNum:               1,
				DataSize:                10,
				PrevPacketCombinedCheck: 0,
				Check0:                  0x1234,
				Check1:                  0x5678,
				Check2:                  0x9ABC,
			},
			expected: []byte{
				0x03, 0x03,
				0x00, 0x01,
				0x00, 0x0A,
				0x00, 0x00,
				0x12, 0x34,
				0x56, 0x78,
				0x9A, 0xBC,
			},
		},
		{
			name: "all zeros",
			header: &CryptPacketHeader{
				Pf0:                     0x00,
				KeyRotDelta:             0x00,
				PacketNum:               0,
				DataSize:                0,
				PrevPacketCombinedCheck: 0,
				Check0:                  0,
				Check1:                  0,
				Check2:                  0,
			},
			expected: []byte{
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00, 0x00,
			},
		},
		{
			name: "max values",
			header: &CryptPacketHeader{
				Pf0:                     0xFF,
				KeyRotDelta:             0xFF,
				PacketNum:               0xFFFF,
				DataSize:                0xFFFF,
				PrevPacketCombinedCheck: 0xFFFF,
				Check0:                  0xFFFF,
				Check1:                  0xFFFF,
				Check2:                  0xFFFF,
			},
			expected: []byte{
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
				0xFF, 0xFF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.header.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v, want nil", err)
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}

			// Check that the length is always 14
			if len(result) != CryptPacketHeaderLength {
				t.Errorf("Encode() length = %d, want %d", len(result), CryptPacketHeaderLength)
			}
		})
	}
}

func TestCryptPacketHeader_RoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		header *CryptPacketHeader
	}{
		{
			name: "basic header",
			header: &CryptPacketHeader{
				Pf0:                     0x03,
				KeyRotDelta:             0x03,
				PacketNum:               100,
				DataSize:                1024,
				PrevPacketCombinedCheck: 0x1234,
				Check0:                  0xABCD,
				Check1:                  0xEF01,
				Check2:                  0x2345,
			},
		},
		{
			name: "zero values",
			header: &CryptPacketHeader{
				Pf0:                     0x00,
				KeyRotDelta:             0x00,
				PacketNum:               0,
				DataSize:                0,
				PrevPacketCombinedCheck: 0,
				Check0:                  0,
				Check1:                  0,
				Check2:                  0,
			},
		},
		{
			name: "max values",
			header: &CryptPacketHeader{
				Pf0:                     0xFF,
				KeyRotDelta:             0xFF,
				PacketNum:               0xFFFF,
				DataSize:                0xFFFF,
				PrevPacketCombinedCheck: 0xFFFF,
				Check0:                  0xFFFF,
				Check1:                  0xFFFF,
				Check2:                  0xFFFF,
			},
		},
		{
			name: "realistic values",
			header: &CryptPacketHeader{
				Pf0:                     0x07,
				KeyRotDelta:             0x03,
				PacketNum:               523,
				DataSize:                2048,
				PrevPacketCombinedCheck: 0x2A56,
				Check0:                  0x06EA,
				Check1:                  0x0215,
				Check2:                  0x8FB3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tt.header.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v, want nil", err)
			}

			// Decode
			decoded, err := NewCryptPacketHeader(encoded)
			if err != nil {
				t.Fatalf("NewCryptPacketHeader() error = %v, want nil", err)
			}

			// Compare
			if decoded.Pf0 != tt.header.Pf0 {
				t.Errorf("Pf0 = 0x%X, want 0x%X", decoded.Pf0, tt.header.Pf0)
			}
			if decoded.KeyRotDelta != tt.header.KeyRotDelta {
				t.Errorf("KeyRotDelta = 0x%X, want 0x%X", decoded.KeyRotDelta, tt.header.KeyRotDelta)
			}
			if decoded.PacketNum != tt.header.PacketNum {
				t.Errorf("PacketNum = 0x%X, want 0x%X", decoded.PacketNum, tt.header.PacketNum)
			}
			if decoded.DataSize != tt.header.DataSize {
				t.Errorf("DataSize = 0x%X, want 0x%X", decoded.DataSize, tt.header.DataSize)
			}
			if decoded.PrevPacketCombinedCheck != tt.header.PrevPacketCombinedCheck {
				t.Errorf("PrevPacketCombinedCheck = 0x%X, want 0x%X", decoded.PrevPacketCombinedCheck, tt.header.PrevPacketCombinedCheck)
			}
			if decoded.Check0 != tt.header.Check0 {
				t.Errorf("Check0 = 0x%X, want 0x%X", decoded.Check0, tt.header.Check0)
			}
			if decoded.Check1 != tt.header.Check1 {
				t.Errorf("Check1 = 0x%X, want 0x%X", decoded.Check1, tt.header.Check1)
			}
			if decoded.Check2 != tt.header.Check2 {
				t.Errorf("Check2 = 0x%X, want 0x%X", decoded.Check2, tt.header.Check2)
			}
		})
	}
}

func TestCryptPacketHeaderLength_Constant(t *testing.T) {
	if CryptPacketHeaderLength != 14 {
		t.Errorf("CryptPacketHeaderLength = %d, want 14", CryptPacketHeaderLength)
	}
}
