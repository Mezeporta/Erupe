package byteframe

import (
	"io"
	"testing"
)

func TestReadUint32_UnderRead(t *testing.T) {
	bf := NewByteFrameFromBytes([]byte{0x01})
	got := bf.ReadUint32()
	if got != 0 {
		t.Errorf("ReadUint32 on 1-byte frame = %d, want 0", got)
	}
	if bf.Err() == nil {
		t.Error("expected ErrReadOverflow")
	}
}

func TestStickyError_ReadAfterFailed(t *testing.T) {
	bf := NewByteFrameFromBytes([]byte{0x01})
	_ = bf.ReadUint32() // triggers error
	// All subsequent reads should return zero
	if bf.ReadUint8() != 0 {
		t.Error("ReadUint8 after error should return 0")
	}
	if bf.ReadUint16() != 0 {
		t.Error("ReadUint16 after error should return 0")
	}
	if bf.ReadUint64() != 0 {
		t.Error("ReadUint64 after error should return 0")
	}
	if bf.ReadInt8() != 0 {
		t.Error("ReadInt8 after error should return 0")
	}
	if bf.ReadInt16() != 0 {
		t.Error("ReadInt16 after error should return 0")
	}
	if bf.ReadInt32() != 0 {
		t.Error("ReadInt32 after error should return 0")
	}
	if bf.ReadInt64() != 0 {
		t.Error("ReadInt64 after error should return 0")
	}
	if bf.ReadFloat32() != 0 {
		t.Error("ReadFloat32 after error should return 0")
	}
	if bf.ReadFloat64() != 0 {
		t.Error("ReadFloat64 after error should return 0")
	}
	if bf.ReadBytes(1) != nil {
		t.Error("ReadBytes after error should return nil")
	}
}

func TestReadOverflow_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		size int
		fn   func(bf *ByteFrame)
	}{
		{"Uint8", 0, func(bf *ByteFrame) { bf.ReadUint8() }},
		{"Uint16", 1, func(bf *ByteFrame) { bf.ReadUint16() }},
		{"Uint32", 3, func(bf *ByteFrame) { bf.ReadUint32() }},
		{"Uint64", 7, func(bf *ByteFrame) { bf.ReadUint64() }},
		{"Int8", 0, func(bf *ByteFrame) { bf.ReadInt8() }},
		{"Int16", 1, func(bf *ByteFrame) { bf.ReadInt16() }},
		{"Int32", 3, func(bf *ByteFrame) { bf.ReadInt32() }},
		{"Int64", 7, func(bf *ByteFrame) { bf.ReadInt64() }},
		{"Float32", 3, func(bf *ByteFrame) { bf.ReadFloat32() }},
		{"Float64", 7, func(bf *ByteFrame) { bf.ReadFloat64() }},
		{"Bytes", 2, func(bf *ByteFrame) { bf.ReadBytes(5) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			bf := NewByteFrameFromBytes(data)
			tt.fn(bf)
			if bf.Err() == nil {
				t.Errorf("expected overflow error for %s with %d bytes", tt.name, tt.size)
			}
		})
	}
}

func TestReadBytes_Exact(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	bf := NewByteFrameFromBytes(data)
	got := bf.ReadBytes(4)
	if len(got) != 4 {
		t.Errorf("ReadBytes(4) returned %d bytes", len(got))
	}
	if bf.Err() != nil {
		t.Errorf("unexpected error: %v", bf.Err())
	}
	// Reading 1 more byte should fail
	_ = bf.ReadUint8()
	if bf.Err() == nil {
		t.Error("expected overflow after reading all bytes")
	}
}

func TestWriteThenRead_RoundTrip(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint8(0xFF)
	bf.WriteUint16(0x1234)
	bf.WriteUint32(0xDEADBEEF)
	bf.WriteUint64(0x0102030405060708)
	bf.WriteInt8(-1)
	bf.WriteInt16(-256)
	bf.WriteInt32(-100000)
	bf.WriteInt64(-999999999)
	bf.WriteFloat32(3.14)
	bf.WriteFloat64(2.718281828)

	_, _ = bf.Seek(0, io.SeekStart)

	if v := bf.ReadUint8(); v != 0xFF {
		t.Errorf("uint8 = %d", v)
	}
	if v := bf.ReadUint16(); v != 0x1234 {
		t.Errorf("uint16 = %d", v)
	}
	if v := bf.ReadUint32(); v != 0xDEADBEEF {
		t.Errorf("uint32 = %x", v)
	}
	if v := bf.ReadUint64(); v != 0x0102030405060708 {
		t.Errorf("uint64 = %x", v)
	}
	if v := bf.ReadInt8(); v != -1 {
		t.Errorf("int8 = %d", v)
	}
	if v := bf.ReadInt16(); v != -256 {
		t.Errorf("int16 = %d", v)
	}
	if v := bf.ReadInt32(); v != -100000 {
		t.Errorf("int32 = %d", v)
	}
	if v := bf.ReadInt64(); v != -999999999 {
		t.Errorf("int64 = %d", v)
	}
	if v := bf.ReadFloat32(); v < 3.13 || v > 3.15 {
		t.Errorf("float32 = %f", v)
	}
	if v := bf.ReadFloat64(); v < 2.71 || v > 2.72 {
		t.Errorf("float64 = %f", v)
	}
	if bf.Err() != nil {
		t.Errorf("unexpected error: %v", bf.Err())
	}
}
