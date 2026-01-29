package entranceserver

import (
	"bytes"
	"testing"
)

func TestMakeHeader(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		respType   string
		entryCount uint16
		key        byte
	}{
		{"empty data", []byte{}, "SV2", 0, 0x00},
		{"single byte", []byte{0x01}, "SVR", 1, 0x00},
		{"multiple bytes", []byte{0x01, 0x02, 0x03, 0x04}, "SV2", 2, 0x00},
		{"with key", []byte{0xDE, 0xAD, 0xBE, 0xEF}, "USR", 5, 0x42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeHeader(tt.data, tt.respType, tt.entryCount, tt.key)

			// Result should not be empty
			if len(result) == 0 {
				t.Error("makeHeader() returned empty result")
			}

			// First byte should be the key
			if result[0] != tt.key {
				t.Errorf("makeHeader() first byte = %x, want %x", result[0], tt.key)
			}

			// Result should be longer than just the key
			if len(result) <= 1 {
				t.Error("makeHeader() result too short")
			}
		})
	}
}

func TestMakeHeaderEncryption(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	result1 := makeHeader(data, "SV2", 1, 0x00)
	result2 := makeHeader(data, "SV2", 1, 0x01)

	// Different keys should produce different encrypted output
	if bytes.Equal(result1, result2) {
		t.Error("makeHeader() with different keys should produce different output")
	}
}

func TestMakeHeaderRespTypes(t *testing.T) {
	data := []byte{0x01}

	// Test different response types produce valid output
	types := []string{"SV2", "SVR", "USR"}

	for _, respType := range types {
		t.Run(respType, func(t *testing.T) {
			result := makeHeader(data, respType, 1, 0x00)
			if len(result) == 0 {
				t.Errorf("makeHeader() with type %s returned empty result", respType)
			}
		})
	}
}

func TestMakeHeaderEmptyData(t *testing.T) {
	// Empty data should still produce a valid (shorter) header
	result := makeHeader([]byte{}, "SV2", 0, 0x00)

	if len(result) == 0 {
		t.Error("makeHeader() with empty data returned empty result")
	}
}

func TestMakeHeaderLargeData(t *testing.T) {
	// Test with larger data
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(i % 256)
	}

	result := makeHeader(data, "SV2", 100, 0x55)

	if len(result) == 0 {
		t.Error("makeHeader() with large data returned empty result")
	}

	// Result should be data + overhead
	if len(result) <= len(data) {
		t.Error("makeHeader() result should be larger than input data due to header")
	}
}

func TestMakeHeaderEntryCount(t *testing.T) {
	data := []byte{0x01, 0x02}

	// Different entry counts should work
	for _, count := range []uint16{0, 1, 10, 100, 65535} {
		result := makeHeader(data, "SV2", count, 0x00)
		if len(result) == 0 {
			t.Errorf("makeHeader() with entryCount=%d returned empty result", count)
		}
	}
}

func TestMakeHeaderDecryptable(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	key := byte(0x00)

	result := makeHeader(data, "SV2", 1, key)

	// Remove key byte and decrypt
	encrypted := result[1:]
	decrypted := DecryptBin8(encrypted, key)

	// Decrypted data should start with "SV2"
	if len(decrypted) >= 3 && string(decrypted[:3]) != "SV2" {
		t.Errorf("makeHeader() decrypted data should start with SV2, got %s", string(decrypted[:3]))
	}
}

func TestMakeHeaderConsistency(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	key := byte(0x10)

	// Same input should produce same output
	result1 := makeHeader(data, "SV2", 5, key)
	result2 := makeHeader(data, "SV2", 5, key)

	if !bytes.Equal(result1, result2) {
		t.Error("makeHeader() with same input should produce same output")
	}
}
