package decryption

import (
	"bytes"
	"testing"
)

// TestEncodeDecodeECD_RoundTrip verifies that encoding then decoding returns
// the original plaintext for various payloads and key indices.
func TestEncodeDecodeECD_RoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		payload []byte
		key     int
	}{
		{"empty", []byte{}, DefaultECDKey},
		{"single_byte", []byte{0x42}, DefaultECDKey},
		{"all_zeros", make([]byte, 64), DefaultECDKey},
		{"all_ones", bytes.Repeat([]byte{0xFF}, 64), DefaultECDKey},
		{"sequential", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}(), DefaultECDKey},
		{"key0", []byte("hello world"), 0},
		{"key1", []byte("hello world"), 1},
		{"key5", []byte("hello world"), 5},
		{"large", bytes.Repeat([]byte("MHFrontier"), 1000), DefaultECDKey},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := EncodeECD(tc.payload, tc.key)
			if err != nil {
				t.Fatalf("EncodeECD: %v", err)
			}

			// Encoded output must start with ECD magic.
			if len(enc) < 4 {
				t.Fatalf("encoded output too short: %d bytes", len(enc))
			}

			dec, err := DecodeECD(enc)
			if err != nil {
				t.Fatalf("DecodeECD: %v", err)
			}

			if !bytes.Equal(dec, tc.payload) {
				t.Errorf("round-trip mismatch:\n  got  %x\n  want %x", dec, tc.payload)
			}
		})
	}
}

// TestDecodeECD_Errors verifies that invalid inputs are rejected with errors.
func TestDecodeECD_Errors(t *testing.T) {
	cases := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{
			name:    "too_small",
			data:    []byte{0x65, 0x63, 0x64},
			wantErr: "too small",
		},
		{
			name: "bad_magic",
			data: func() []byte {
				b := make([]byte, 16)
				b[0] = 0xDE
				return b
			}(),
			wantErr: "invalid magic",
		},
		{
			name: "invalid_key",
			data: func() []byte {
				b := make([]byte, 16)
				// ECD magic
				b[0], b[1], b[2], b[3] = 0x65, 0x63, 0x64, 0x1A
				// key index = 99 (out of range)
				b[4] = 99
				return b
			}(),
			wantErr: "invalid key",
		},
		{
			name: "payload_exceeds_buffer",
			data: func() []byte {
				b := make([]byte, 16)
				b[0], b[1], b[2], b[3] = 0x65, 0x63, 0x64, 0x1A
				// key 4
				b[4] = DefaultECDKey
				// declare payload size larger than the buffer
				b[8], b[9], b[10], b[11] = 0xFF, 0xFF, 0xFF, 0x00
				return b
			}(),
			wantErr: "exceeds buffer",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeECD(tc.data)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tc.wantErr)) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestEncodeECD_InvalidKey verifies that an out-of-range key is rejected.
func TestEncodeECD_InvalidKey(t *testing.T) {
	_, err := EncodeECD([]byte("test"), 99)
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
}

// TestDecodeECD_EmptyPayload verifies that a valid header with zero payload
// decodes to an empty slice without error.
func TestDecodeECD_EmptyPayload(t *testing.T) {
	enc, err := EncodeECD([]byte{}, DefaultECDKey)
	if err != nil {
		t.Fatalf("EncodeECD: %v", err)
	}
	dec, err := DecodeECD(enc)
	if err != nil {
		t.Fatalf("DecodeECD: %v", err)
	}
	if len(dec) != 0 {
		t.Errorf("expected empty payload, got %d bytes", len(dec))
	}
}
