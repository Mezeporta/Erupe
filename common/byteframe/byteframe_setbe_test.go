package byteframe

import (
	"encoding/binary"
	"io"
	"testing"
)

func TestByteFrame_SetBE(t *testing.T) {
	bf := NewByteFrame()
	// Default is already BigEndian, switch to LE first
	bf.SetLE()
	if bf.byteOrder != binary.LittleEndian {
		t.Error("SetLE() should set LittleEndian")
	}

	// Now test SetBE
	bf.SetBE()
	if bf.byteOrder != binary.BigEndian {
		t.Error("SetBE() should set BigEndian")
	}

	// Verify write/read works correctly in BE mode after switching
	bf.WriteUint16(0x1234)
	bf.Seek(0, io.SeekStart)
	got := bf.ReadUint16()
	if got != 0x1234 {
		t.Errorf("ReadUint16() = 0x%04X, want 0x1234", got)
	}

	// Verify raw bytes are in big endian order
	bf2 := NewByteFrame()
	bf2.SetLE()
	bf2.SetBE()
	bf2.WriteUint32(0xDEADBEEF)
	data := bf2.Data()
	if data[0] != 0xDE || data[1] != 0xAD || data[2] != 0xBE || data[3] != 0xEF {
		t.Errorf("SetBE bytes: got %X, want DEADBEEF", data)
	}
}

func TestByteFrame_LEReadWrite(t *testing.T) {
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteUint32(0x12345678)
	data := bf.Data()
	// In LE, LSB first
	if data[0] != 0x78 || data[1] != 0x56 || data[2] != 0x34 || data[3] != 0x12 {
		t.Errorf("LE WriteUint32 bytes: got %X, want 78563412", data)
	}

	bf.Seek(0, io.SeekStart)
	got := bf.ReadUint32()
	if got != 0x12345678 {
		t.Errorf("LE ReadUint32() = 0x%08X, want 0x12345678", got)
	}
}
