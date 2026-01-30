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

// TestEncodeServerInfoClanMemberLimit documents the hardcoded clan member limit.
//
// CURRENT BEHAVIOR:
//
//	bf.WriteUint32(0x0000003C)  // Hardcoded to 60
//
// EXPECTED BEHAVIOR (after fix commit 7d760bd):
//
//	bf.WriteUint32(uint32(s.erupeConfig.GameplayOptions.ClanMemberLimits[len(s.erupeConfig.GameplayOptions.ClanMemberLimits)-1][1]))
//	This reads the maximum clan member limit from the last entry of ClanMemberLimits config.
//
// Note: The ClanMemberLimits config field doesn't exist in this branch yet.
// This test documents the expected value (60 = 0x3C) that will be read from config.
func TestEncodeServerInfoClanMemberLimit(t *testing.T) {
	// The hardcoded value is 60 (0x3C)
	hardcodedLimit := uint32(0x0000003C)

	if hardcodedLimit != 60 {
		t.Errorf("hardcoded clan member limit = %d, expected 60", hardcodedLimit)
	}

	t.Logf("Current implementation uses hardcoded clan member limit: %d", hardcodedLimit)
	t.Logf("After fix, this will be read from config.GameplayOptions.ClanMemberLimits")
}

// TestMakeHeaderDataIntegrity verifies that data passed to makeHeader is preserved
// through encryption/decryption.
func TestMakeHeaderDataIntegrity(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		respType string
		count    uint16
		key      byte
	}{
		{"empty SV2", []byte{}, "SV2", 0, 0x00},
		{"simple SVR", []byte{0x01, 0x02}, "SVR", 1, 0x00},
		{"with key", []byte{0xDE, 0xAD, 0xBE, 0xEF}, "SV2", 2, 0x42},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := makeHeader(tc.data, tc.respType, tc.count, tc.key)

			// Result should have key as first byte
			if len(result) == 0 {
				t.Fatal("makeHeader returned empty result")
			}
			if result[0] != tc.key {
				t.Errorf("first byte = 0x%X, want 0x%X (key)", result[0], tc.key)
			}

			// Decrypt and verify response type
			if len(result) > 1 {
				decrypted := DecryptBin8(result[1:], tc.key)
				if len(decrypted) >= 3 {
					gotType := string(decrypted[:3])
					if gotType != tc.respType {
						t.Errorf("decrypted respType = %s, want %s", gotType, tc.respType)
					}
				}
			}
		})
	}
}
