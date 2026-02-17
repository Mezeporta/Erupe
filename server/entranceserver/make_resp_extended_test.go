package entranceserver

import (
	"testing"
)

// TestMakeHeader tests the makeHeader function with various inputs
func TestMakeHeader(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		respType   string
		entryCount uint16
		key        byte
	}{
		{"empty data", []byte{}, "SV2", 0, 0x00},
		{"small data", []byte{0x01, 0x02, 0x03}, "SV2", 1, 0x00},
		{"SVR type", []byte{0xAA, 0xBB}, "SVR", 2, 0x42},
		{"USR type", []byte{0x01}, "USR", 1, 0x00},
		{"larger data", make([]byte, 100), "SV2", 5, 0xFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeHeader(tt.data, tt.respType, tt.entryCount, tt.key)
			if len(result) == 0 {
				t.Error("makeHeader returned empty result")
			}
			// First byte should be the key
			if result[0] != tt.key {
				t.Errorf("first byte = %x, want %x", result[0], tt.key)
			}
		})
	}
}
