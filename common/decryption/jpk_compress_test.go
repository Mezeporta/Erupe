package decryption

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestPackSimpleRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"single byte", []byte{0x42}},
		{"ascii text", []byte("hello world")},
		{"repeated pattern", bytes.Repeat([]byte{0xAB, 0xCD}, 100)},
		{"all zeros", make([]byte, 256)},
		{"all 0xFF", bytes.Repeat([]byte{0xFF}, 128)},
		{"sequential bytes", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}()},
		{"long repeating run", bytes.Repeat([]byte("ABCDEFGH"), 50)},
		{"mixed", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD, 0x80, 0x81}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compressed := PackSimple(tc.data)
			got := UnpackSimple(compressed)
			if !bytes.Equal(got, tc.data) {
				t.Errorf("round-trip mismatch\n  want len=%d\n   got len=%d", len(tc.data), len(got))
			}
		})
	}
}

func TestPackSimpleHeader(t *testing.T) {
	data := []byte("test data")
	compressed := PackSimple(data)

	if len(compressed) < 16 {
		t.Fatalf("output too short: %d bytes", len(compressed))
	}

	magic := binary.LittleEndian.Uint32(compressed[0:4])
	if magic != 0x1A524B4A {
		t.Errorf("wrong magic: got 0x%08X, want 0x1A524B4A", magic)
	}

	jpkType := binary.LittleEndian.Uint16(compressed[6:8])
	if jpkType != 3 {
		t.Errorf("wrong type: got %d, want 3", jpkType)
	}

	decompSize := binary.LittleEndian.Uint32(compressed[12:16])
	if decompSize != uint32(len(data)) {
		t.Errorf("wrong decompressed size: got %d, want %d", decompSize, len(data))
	}
}

func TestPackSimpleLargeRepeating(t *testing.T) {
	// 4 KB of repeating pattern — should compress well
	data := bytes.Repeat([]byte{0xAA, 0xBB, 0xCC, 0xDD}, 1024)
	compressed := PackSimple(data)

	if len(compressed) >= len(data) {
		t.Logf("note: compressed (%d) not smaller than original (%d)", len(compressed), len(data))
	}

	got := UnpackSimple(compressed)
	if !bytes.Equal(got, data) {
		t.Errorf("round-trip failed for large repeating data")
	}
}
