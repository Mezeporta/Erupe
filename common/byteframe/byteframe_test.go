package byteframe

import (
	"io"
	"math"
	"testing"
)

func TestNewByteFrame(t *testing.T) {
	bf := NewByteFrame()

	if bf == nil {
		t.Fatal("NewByteFrame() returned nil")
	}
	if len(bf.Data()) != 0 {
		t.Errorf("NewByteFrame().Data() len = %d, want 0", len(bf.Data()))
	}
}

func TestNewByteFrameFromBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	bf := NewByteFrameFromBytes(data)

	if bf == nil {
		t.Fatal("NewByteFrameFromBytes() returned nil")
	}
	if len(bf.Data()) != len(data) {
		t.Errorf("NewByteFrameFromBytes().Data() len = %d, want %d", len(bf.Data()), len(data))
	}

	// Verify data is copied, not referenced
	data[0] = 99
	if bf.Data()[0] == 99 {
		t.Error("NewByteFrameFromBytes() did not copy data")
	}
}

func TestWriteReadUint8(t *testing.T) {
	tests := []uint8{0, 1, 127, 128, 255}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint8(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadUint8()
			if got != val {
				t.Errorf("Write/ReadUint8(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadUint16(t *testing.T) {
	tests := []uint16{0, 1, 255, 256, 32767, 65535}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint16(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadUint16()
			if got != val {
				t.Errorf("Write/ReadUint16(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadUint32(t *testing.T) {
	tests := []uint32{0, 1, 255, 65535, 2147483647, 4294967295}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint32(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadUint32()
			if got != val {
				t.Errorf("Write/ReadUint32(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadUint64(t *testing.T) {
	tests := []uint64{0, 1, 255, 65535, 4294967295, 18446744073709551615}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteUint64(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadUint64()
			if got != val {
				t.Errorf("Write/ReadUint64(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadInt8(t *testing.T) {
	tests := []int8{-128, -1, 0, 1, 127}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteInt8(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadInt8()
			if got != val {
				t.Errorf("Write/ReadInt8(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadInt16(t *testing.T) {
	tests := []int16{-32768, -1, 0, 1, 32767}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteInt16(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadInt16()
			if got != val {
				t.Errorf("Write/ReadInt16(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadInt32(t *testing.T) {
	tests := []int32{-2147483648, -1, 0, 1, 2147483647}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteInt32(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadInt32()
			if got != val {
				t.Errorf("Write/ReadInt32(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadInt64(t *testing.T) {
	tests := []int64{-9223372036854775808, -1, 0, 1, 9223372036854775807}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteInt64(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadInt64()
			if got != val {
				t.Errorf("Write/ReadInt64(%d) = %d", val, got)
			}
		})
	}
}

func TestWriteReadFloat32(t *testing.T) {
	tests := []float32{0, 1.5, -1.5, 3.14159, math.MaxFloat32, math.SmallestNonzeroFloat32}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteFloat32(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadFloat32()
			if got != val {
				t.Errorf("Write/ReadFloat32(%f) = %f", val, got)
			}
		})
	}
}

func TestWriteReadFloat64(t *testing.T) {
	tests := []float64{0, 1.5, -1.5, 3.14159265358979, math.MaxFloat64, math.SmallestNonzeroFloat64}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteFloat64(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadFloat64()
			if got != val {
				t.Errorf("Write/ReadFloat64(%f) = %f", val, got)
			}
		})
	}
}

func TestWriteReadBool(t *testing.T) {
	tests := []bool{true, false}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteBool(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadBool()
			if got != val {
				t.Errorf("Write/ReadBool(%v) = %v", val, got)
			}
		})
	}
}

func TestWriteReadBytes(t *testing.T) {
	tests := [][]byte{
		{},
		{1},
		{1, 2, 3, 4, 5},
		{0, 255, 128, 64, 32},
	}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteBytes(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadBytes(uint(len(val)))
			if len(got) != len(val) {
				t.Errorf("Write/ReadBytes len = %d, want %d", len(got), len(val))
				return
			}
			for i := range got {
				if got[i] != val[i] {
					t.Errorf("Write/ReadBytes[%d] = %d, want %d", i, got[i], val[i])
				}
			}
		})
	}
}

func TestWriteReadNullTerminatedBytes(t *testing.T) {
	tests := [][]byte{
		{},
		{65},
		{65, 66, 67},
	}

	for _, val := range tests {
		t.Run("", func(t *testing.T) {
			bf := NewByteFrame()
			bf.WriteNullTerminatedBytes(val)
			bf.Seek(0, io.SeekStart)
			got := bf.ReadNullTerminatedBytes()
			if len(got) != len(val) {
				t.Errorf("Write/ReadNullTerminatedBytes len = %d, want %d", len(got), len(val))
				return
			}
			for i := range got {
				if got[i] != val[i] {
					t.Errorf("Write/ReadNullTerminatedBytes[%d] = %d, want %d", i, got[i], val[i])
				}
			}
		})
	}
}

func TestSeek(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint32(0x12345678)
	bf.WriteUint32(0xDEADBEEF)

	// SeekStart
	pos, err := bf.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf("Seek(0, SeekStart) error = %v", err)
	}
	if pos != 0 {
		t.Errorf("Seek(0, SeekStart) pos = %d, want 0", pos)
	}

	val := bf.ReadUint32()
	if val != 0x12345678 {
		t.Errorf("After Seek(0, SeekStart) ReadUint32() = %x, want 0x12345678", val)
	}

	// SeekCurrent
	pos, err = bf.Seek(-4, io.SeekCurrent)
	if err != nil {
		t.Errorf("Seek(-4, SeekCurrent) error = %v", err)
	}
	if pos != 0 {
		t.Errorf("Seek(-4, SeekCurrent) pos = %d, want 0", pos)
	}

	// SeekEnd
	pos, err = bf.Seek(-4, io.SeekEnd)
	if err != nil {
		t.Errorf("Seek(-4, SeekEnd) error = %v", err)
	}
	if pos != 4 {
		t.Errorf("Seek(-4, SeekEnd) pos = %d, want 4", pos)
	}

	val = bf.ReadUint32()
	if val != 0xDEADBEEF {
		t.Errorf("After Seek(-4, SeekEnd) ReadUint32() = %x, want 0xDEADBEEF", val)
	}
}

func TestSeekErrors(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint32(0x12345678)

	// Seek beyond end
	_, err := bf.Seek(100, io.SeekStart)
	if err == nil {
		t.Error("Seek(100, SeekStart) should return error")
	}

	// Seek before start
	_, err = bf.Seek(-100, io.SeekCurrent)
	if err == nil {
		t.Error("Seek(-100, SeekCurrent) should return error")
	}

	// Seek before start from end
	_, err = bf.Seek(-100, io.SeekEnd)
	if err == nil {
		t.Error("Seek(-100, SeekEnd) should return error")
	}
}

func TestEndianness(t *testing.T) {
	// Test big endian (default)
	bf := NewByteFrame()
	bf.WriteUint16(0x1234)
	data := bf.Data()
	if data[0] != 0x12 || data[1] != 0x34 {
		t.Errorf("Big endian WriteUint16(0x1234) = %v, want [0x12, 0x34]", data)
	}

	// Test little endian
	bf = NewByteFrame()
	bf.SetLE()
	bf.WriteUint16(0x1234)
	data = bf.Data()
	if data[0] != 0x34 || data[1] != 0x12 {
		t.Errorf("Little endian WriteUint16(0x1234) = %v, want [0x34, 0x12]", data)
	}

	// Test switching back to big endian
	bf = NewByteFrame()
	bf.SetLE()
	bf.SetBE()
	bf.WriteUint16(0x1234)
	data = bf.Data()
	if data[0] != 0x12 || data[1] != 0x34 {
		t.Errorf("Switched back to big endian WriteUint16(0x1234) = %v, want [0x12, 0x34]", data)
	}
}

func TestDataFromCurrent(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint8(1)
	bf.WriteUint8(2)
	bf.WriteUint8(3)
	bf.WriteUint8(4)

	bf.Seek(2, io.SeekStart)
	remaining := bf.DataFromCurrent()

	if len(remaining) != 2 {
		t.Errorf("DataFromCurrent() len = %d, want 2", len(remaining))
	}
	if remaining[0] != 3 || remaining[1] != 4 {
		t.Errorf("DataFromCurrent() = %v, want [3, 4]", remaining)
	}
}

func TestBufferGrowth(t *testing.T) {
	bf := NewByteFrame()

	// Write more data than initial buffer size (4 bytes)
	for i := 0; i < 100; i++ {
		bf.WriteUint32(uint32(i))
	}

	if len(bf.Data()) != 400 {
		t.Errorf("After writing 100 uint32s, Data() len = %d, want 400", len(bf.Data()))
	}

	// Verify data integrity
	bf.Seek(0, io.SeekStart)
	for i := 0; i < 100; i++ {
		val := bf.ReadUint32()
		if val != uint32(i) {
			t.Errorf("After growth, ReadUint32()[%d] = %d, want %d", i, val, i)
		}
	}
}

func TestMultipleWrites(t *testing.T) {
	bf := NewByteFrame()

	bf.WriteUint8(0x01)
	bf.WriteUint16(0x0203)
	bf.WriteUint32(0x04050607)
	bf.WriteUint64(0x08090A0B0C0D0E0F)

	expected := []byte{
		0x01,
		0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	}

	data := bf.Data()
	if len(data) != len(expected) {
		t.Errorf("Multiple writes Data() len = %d, want %d", len(data), len(expected))
		return
	}

	for i := range expected {
		if data[i] != expected[i] {
			t.Errorf("Multiple writes Data()[%d] = %x, want %x", i, data[i], expected[i])
		}
	}
}

func TestReadPanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadUint32 on empty buffer should panic")
		}
	}()

	bf := NewByteFrame()
	bf.ReadUint32()
}

func TestReadBoolNonZero(t *testing.T) {
	// Test that any non-zero value is considered true
	bf := NewByteFrameFromBytes([]byte{0, 1, 2, 255})

	if bf.ReadBool() != false {
		t.Error("ReadBool(0) should be false")
	}
	if bf.ReadBool() != true {
		t.Error("ReadBool(1) should be true")
	}
	if bf.ReadBool() != true {
		t.Error("ReadBool(2) should be true")
	}
	if bf.ReadBool() != true {
		t.Error("ReadBool(255) should be true")
	}
}

func TestReadNullTerminatedBytesNoTerminator(t *testing.T) {
	// Test behavior when there's no null terminator
	bf := NewByteFrameFromBytes([]byte{65, 66, 67})
	result := bf.ReadNullTerminatedBytes()

	if len(result) != 0 {
		t.Errorf("ReadNullTerminatedBytes with no terminator should return empty, got %v", result)
	}
}

func TestReadUint8PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadUint8 past end should panic")
		}
	}()
	bf := NewByteFrameFromBytes([]byte{0x01})
	bf.ReadUint8() // consume the one byte
	bf.ReadUint8() // should panic - no more data
}

func TestReadUint16PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadUint16 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadUint16()
}

func TestReadUint64PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadUint64 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadUint64()
}

func TestReadInt8PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadInt8 past end should panic")
		}
	}()
	bf := NewByteFrameFromBytes([]byte{0x01})
	bf.ReadInt8() // consume the one byte
	bf.ReadInt8() // should panic - no more data
}

func TestReadInt16PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadInt16 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadInt16()
}

func TestReadInt32PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadInt32 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadInt32()
}

func TestReadInt64PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadInt64 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadInt64()
}

func TestReadFloat32PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadFloat32 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadFloat32()
}

func TestReadFloat64PanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadFloat64 on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadFloat64()
}

func TestReadBytesPanicsOnOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadBytes on empty buffer should panic")
		}
	}()
	bf := NewByteFrame()
	bf.ReadBytes(10)
}

func TestSeekInvalidWhence(t *testing.T) {
	bf := NewByteFrame()
	bf.WriteUint32(0x12345678)

	// Invalid whence value should not crash, just not change position
	pos, _ := bf.Seek(0, 99)
	if pos != 4 {
		t.Errorf("Seek with invalid whence pos = %d, want 4", pos)
	}
}

func TestLittleEndianReadWrite(t *testing.T) {
	bf := NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0x12345678)
	bf.WriteInt16(-1234)
	bf.WriteFloat32(3.14)

	bf.Seek(0, io.SeekStart)
	bf.SetLE()

	if val := bf.ReadUint32(); val != 0x12345678 {
		t.Errorf("LE ReadUint32 = 0x%X, want 0x12345678", val)
	}
	if val := bf.ReadInt16(); val != -1234 {
		t.Errorf("LE ReadInt16 = %d, want -1234", val)
	}
	if val := bf.ReadFloat32(); val < 3.13 || val > 3.15 {
		t.Errorf("LE ReadFloat32 = %f, want ~3.14", val)
	}
}

func TestGrowWithLargeWrite(t *testing.T) {
	bf := NewByteFrame()
	// Initial buffer is 4 bytes. Write 1000 bytes to trigger grow with size > buf
	largeData := make([]byte, 1000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	bf.WriteBytes(largeData)

	if len(bf.Data()) != 1000 {
		t.Errorf("Data() len after large write = %d, want 1000", len(bf.Data()))
	}

	bf.Seek(0, io.SeekStart)
	readBack := bf.ReadBytes(1000)
	for i := range readBack {
		if readBack[i] != byte(i%256) {
			t.Errorf("Data mismatch at position %d: got %d, want %d", i, readBack[i], byte(i%256))
			break
		}
	}
}
