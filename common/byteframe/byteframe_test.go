package byteframe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"
)

func TestNewByteFrame(t *testing.T) {
	bf := NewByteFrame()
	if bf == nil {
		t.Fatal("NewByteFrame() returned nil")
	}
	if bf.index != 0 {
		t.Errorf("index = %d, want 0", bf.index)
	}
	if bf.usedSize != 0 {
		t.Errorf("usedSize = %d, want 0", bf.usedSize)
	}
	if len(bf.buf) != 4 {
		t.Errorf("buf length = %d, want 4", len(bf.buf))
	}
	if bf.byteOrder != binary.BigEndian {
		t.Error("byteOrder should be BigEndian by default")
	}
}

func TestNewByteFrameFromBytes(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03, 0x04}
	bf := NewByteFrameFromBytes(input)
	if bf == nil {
		t.Fatal("NewByteFrameFromBytes() returned nil")
	}
	if bf.index != 0 {
		t.Errorf("index = %d, want 0", bf.index)
	}
	if bf.usedSize != uint(len(input)) {
		t.Errorf("usedSize = %d, want %d", bf.usedSize, len(input))
	}
	if !bytes.Equal(bf.buf, input) {
		t.Errorf("buf = %v, want %v", bf.buf, input)
	}
	// Verify it's a copy, not the same slice
	input[0] = 0xFF
	if bf.buf[0] == 0xFF {
		t.Error("NewByteFrameFromBytes should make a copy, not use the same slice")
	}
}

func TestByteFrame_WriteAndReadUint8(t *testing.T) {
	bf := NewByteFrame()
	values := []uint8{0, 1, 127, 128, 255}

	for _, v := range values {
		bf.WriteUint8(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadUint8()
		if got != expected {
			t.Errorf("ReadUint8()[%d] = %d, want %d", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadUint16(t *testing.T) {
	tests := []struct {
		name  string
		value uint16
	}{
		{"zero", 0},
		{"one", 1},
		{"max_int8", 127},
		{"max_uint8", 255},
		{"max_int16", 32767},
		{"max_uint16", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint16(tt.value)
			_, _ = bf.Seek(0, io.SeekStart)
			got := bf.ReadUint16()
			if got != tt.value {
				t.Errorf("ReadUint16() = %d, want %d", got, tt.value)
			}
		})
	}
}

func TestByteFrame_WriteAndReadUint32(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
	}{
		{"zero", 0},
		{"one", 1},
		{"max_uint16", 65535},
		{"max_uint32", 4294967295},
		{"arbitrary", 0x12345678},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint32(tt.value)
			_, _ = bf.Seek(0, io.SeekStart)
			got := bf.ReadUint32()
			if got != tt.value {
				t.Errorf("ReadUint32() = %d, want %d", got, tt.value)
			}
		})
	}
}

func TestByteFrame_WriteAndReadUint64(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
	}{
		{"zero", 0},
		{"one", 1},
		{"max_uint32", 4294967295},
		{"max_uint64", 18446744073709551615},
		{"arbitrary", 0x123456789ABCDEF0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint64(tt.value)
			_, _ = bf.Seek(0, io.SeekStart)
			got := bf.ReadUint64()
			if got != tt.value {
				t.Errorf("ReadUint64() = %d, want %d", got, tt.value)
			}
		})
	}
}

func TestByteFrame_WriteAndReadInt8(t *testing.T) {
	values := []int8{-128, -1, 0, 1, 127}
	bf := NewByteFrame()

	for _, v := range values {
		bf.WriteInt8(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadInt8()
		if got != expected {
			t.Errorf("ReadInt8()[%d] = %d, want %d", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadInt16(t *testing.T) {
	values := []int16{-32768, -1, 0, 1, 32767}
	bf := NewByteFrame()

	for _, v := range values {
		bf.WriteInt16(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadInt16()
		if got != expected {
			t.Errorf("ReadInt16()[%d] = %d, want %d", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadInt32(t *testing.T) {
	values := []int32{-2147483648, -1, 0, 1, 2147483647}
	bf := NewByteFrame()

	for _, v := range values {
		bf.WriteInt32(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadInt32()
		if got != expected {
			t.Errorf("ReadInt32()[%d] = %d, want %d", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadInt64(t *testing.T) {
	values := []int64{-9223372036854775808, -1, 0, 1, 9223372036854775807}
	bf := NewByteFrame()

	for _, v := range values {
		bf.WriteInt64(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadInt64()
		if got != expected {
			t.Errorf("ReadInt64()[%d] = %d, want %d", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadFloat32(t *testing.T) {
	values := []float32{0.0, -1.5, 1.5, 3.14159, math.MaxFloat32, -math.MaxFloat32}
	bf := NewByteFrame()

	for _, v := range values {
		bf.WriteFloat32(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadFloat32()
		if got != expected {
			t.Errorf("ReadFloat32()[%d] = %f, want %f", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadFloat64(t *testing.T) {
	values := []float64{0.0, -1.5, 1.5, 3.14159265358979, math.MaxFloat64, -math.MaxFloat64}
	bf := NewByteFrame()

	for _, v := range values {
		bf.WriteFloat64(v)
	}

	_, _ = bf.Seek(0, io.SeekStart)
	for i, expected := range values {
		got := bf.ReadFloat64()
		if got != expected {
			t.Errorf("ReadFloat64()[%d] = %f, want %f", i, got, expected)
		}
	}
}

func TestByteFrame_WriteAndReadBool(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteBool(true)
	bf.WriteBool(false)
	bf.WriteBool(true)

	_, _ = bf.Seek(0, io.SeekStart)
	if got := bf.ReadBool(); got != true {
		t.Errorf("ReadBool()[0] = %v, want true", got)
	}
	if got := bf.ReadBool(); got != false {
		t.Errorf("ReadBool()[1] = %v, want false", got)
	}
	if got := bf.ReadBool(); got != true {
		t.Errorf("ReadBool()[2] = %v, want true", got)
	}
}

func TestByteFrame_WriteAndReadBytes(t *testing.T) {
	bf := NewByteFrame()
	input := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	bf.WriteBytes(input)

	_, _ = bf.Seek(0, io.SeekStart)
	got := bf.ReadBytes(uint(len(input)))
	if !bytes.Equal(got, input) {
		t.Errorf("ReadBytes() = %v, want %v", got, input)
	}
}

func TestByteFrame_WriteAndReadNullTerminatedBytes(t *testing.T) {
	bf := NewByteFrame()
	input := []byte("Hello, World!")
	bf.WriteNullTerminatedBytes(input)

	_, _ = bf.Seek(0, io.SeekStart)
	got := bf.ReadNullTerminatedBytes()
	if !bytes.Equal(got, input) {
		t.Errorf("ReadNullTerminatedBytes() = %v, want %v", got, input)
	}
}

func TestByteFrame_ReadNullTerminatedBytes_NoNull(t *testing.T) {
	bf := NewByteFrame()
	input := []byte("Hello")
	bf.WriteBytes(input)

	_, _ = bf.Seek(0, io.SeekStart)
	got := bf.ReadNullTerminatedBytes()
	// When there's no null terminator, it should return empty slice
	if len(got) != 0 {
		t.Errorf("ReadNullTerminatedBytes() = %v, want empty slice", got)
	}
}

func TestByteFrame_Endianness(t *testing.T) {
	// Test BigEndian (default)
	bfBE := NewByteFrame()
	bfBE.WriteUint16(0x1234)
	dataBE := bfBE.Data()
	if dataBE[0] != 0x12 || dataBE[1] != 0x34 {
		t.Errorf("BigEndian: got %X %X, want 12 34", dataBE[0], dataBE[1])
	}

	// Test LittleEndian
	bfLE := NewByteFrame()
	bfLE.SetLE()
	bfLE.WriteUint16(0x1234)
	dataLE := bfLE.Data()
	if dataLE[0] != 0x34 || dataLE[1] != 0x12 {
		t.Errorf("LittleEndian: got %X %X, want 34 12", dataLE[0], dataLE[1])
	}
}

func TestByteFrame_Seek(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04, 0x05})

	tests := []struct {
		name      string
		offset    int64
		whence    int
		wantIndex uint
		wantErr   bool
	}{
		{"seek_start_0", 0, io.SeekStart, 0, false},
		{"seek_start_2", 2, io.SeekStart, 2, false},
		{"seek_start_5", 5, io.SeekStart, 5, false},
		{"seek_start_beyond", 6, io.SeekStart, 5, true},
		{"seek_current_forward", 2, io.SeekCurrent, 5, true}, // Will go beyond max
		{"seek_current_backward", -3, io.SeekCurrent, 2, false},
		{"seek_current_before_start", -10, io.SeekCurrent, 2, true},
		{"seek_end_0", 0, io.SeekEnd, 5, false},
		{"seek_end_negative", -2, io.SeekEnd, 3, false},
		{"seek_end_beyond", 1, io.SeekEnd, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to known position for each test
			_, _ = bf.Seek(5, io.SeekStart)

			pos, err := bf.Seek(tt.offset, tt.whence)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Seek() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Seek() unexpected error: %v", err)
				}
				if bf.index != tt.wantIndex {
					t.Errorf("index = %d, want %d", bf.index, tt.wantIndex)
				}
				if uint(pos) != tt.wantIndex {
					t.Errorf("returned position = %d, want %d", pos, tt.wantIndex)
				}
			}
		})
	}
}

func TestByteFrame_Data(t *testing.T) {
	bf := NewByteFrame()
	input := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	bf.WriteBytes(input)

	data := bf.Data()
	if !bytes.Equal(data, input) {
		t.Errorf("Data() = %v, want %v", data, input)
	}
}

func TestByteFrame_DataFromCurrent(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	_, _ = bf.Seek(2, io.SeekStart)

	data := bf.DataFromCurrent()
	expected := []byte{0x03, 0x04, 0x05}
	if !bytes.Equal(data, expected) {
		t.Errorf("DataFromCurrent() = %v, want %v", data, expected)
	}
}

func TestByteFrame_Index(t *testing.T) {
	bf := NewByteFrame()
	if bf.Index() != 0 {
		t.Errorf("Index() = %d, want 0", bf.Index())
	}

	bf.WriteUint8(0x01)
	if bf.Index() != 1 {
		t.Errorf("Index() = %d, want 1", bf.Index())
	}

	bf.WriteUint16(0x0102)
	if bf.Index() != 3 {
		t.Errorf("Index() = %d, want 3", bf.Index())
	}
}

func TestByteFrame_BufferGrowth(t *testing.T) {
	bf := NewByteFrame()
	initialCap := len(bf.buf)

	// Write enough data to force growth
	for i := 0; i < 100; i++ {
		bf.WriteUint32(uint32(i))
	}

	if len(bf.buf) <= initialCap {
		t.Errorf("Buffer should have grown, initial cap: %d, current: %d", initialCap, len(bf.buf))
	}

	// Verify all data is still accessible
	_, _ = bf.Seek(0, io.SeekStart)
	for i := 0; i < 100; i++ {
		got := bf.ReadUint32()
		if got != uint32(i) {
			t.Errorf("After growth, ReadUint32()[%d] = %d, want %d", i, got, i)
			break
		}
	}
}

func TestByteFrame_ReadOverflowSetsError(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint8(0x01)
	_, _ = bf.Seek(0, io.SeekStart)
	bf.ReadUint8()

	if bf.Err() != nil {
		t.Fatal("Err() should be nil before overflow")
	}

	// Should set sticky error - trying to read 2 bytes when only 1 was written
	got := bf.ReadUint16()
	if got != 0 {
		t.Errorf("ReadUint16() after overflow = %d, want 0", got)
	}
	if bf.Err() == nil {
		t.Error("Err() should be non-nil after read overflow")
	}
	if !errors.Is(bf.Err(), ErrReadOverflow) {
		t.Errorf("Err() = %v, want ErrReadOverflow", bf.Err())
	}

	// Subsequent reads should also return zero without changing the error
	got32 := bf.ReadUint32()
	if got32 != 0 {
		t.Errorf("ReadUint32() after overflow = %d, want 0", got32)
	}
}

func TestByteFrame_SequentialWrites(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint8(0x01)
	bf.WriteUint16(0x0203)
	bf.WriteUint32(0x04050607)
	bf.WriteUint64(0x08090A0B0C0D0E0F)

	expected := []byte{
		0x01,                               // uint8
		0x02, 0x03,                         // uint16
		0x04, 0x05, 0x06, 0x07,             // uint32
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, // uint64
	}

	data := bf.Data()
	if !bytes.Equal(data, expected) {
		t.Errorf("Sequential writes: got %X, want %X", data, expected)
	}
}

func BenchmarkByteFrame_WriteUint8(b *testing.B) {
	bf := NewByteFrame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.WriteUint8(0x42)
	}
}

func BenchmarkByteFrame_WriteUint32(b *testing.B) {
	bf := NewByteFrame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.WriteUint32(0x12345678)
	}
}

func BenchmarkByteFrame_ReadUint32(b *testing.B) {
	bf := NewByteFrame()
	for i := 0; i < 1000; i++ {
		bf.WriteUint32(0x12345678)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bf.Seek(0, io.SeekStart)
		bf.ReadUint32()
	}
}

func BenchmarkByteFrame_WriteBytes(b *testing.B) {
	bf := NewByteFrame()
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.WriteBytes(data)
	}
}
