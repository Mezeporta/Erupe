package byteframe

/*
	This is HEAVILY based on the code from
		https://github.com/sinni800/sgemu/blob/master/Core/Packet.go
*/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// ByteFrame is a struct for reading and writing raw byte data.
type ByteFrame struct {
	index     uint
	usedSize  uint
	buf       []byte
	byteOrder binary.ByteOrder
}

// NewByteFrame creates a new ByteFrame with valid default values.
// byteOrder defaults to big endian.
func NewByteFrame() *ByteFrame {
	b := &ByteFrame{
		index:     0,
		usedSize:  0,
		buf:       make([]byte, 4),
		byteOrder: binary.BigEndian,
	}
	return b
}

// NewByteFrameFromBytes creates a new ByteFrame with valid default values.
// makes a copy of the given buf and initalizes with it.
// byteOrder defaults to big endian.
func NewByteFrameFromBytes(buf []byte) *ByteFrame {
	b := &ByteFrame{
		index:     0,
		usedSize:  uint(len(buf)),
		buf:       make([]byte, len(buf)),
		byteOrder: binary.BigEndian,
	}
	copy(b.buf, buf)
	return b
}

// grow either doubles the backing buffer size, or grows it by the size specified, whichever is larger.
func (b *ByteFrame) grow(size uint) {
	bytesToAdd := uint(0)
	if size > uint(len(b.buf)) {
		bytesToAdd = size
	} else {
		bytesToAdd = uint(len(b.buf))
	}

	newBuf := make([]byte, uint(len(b.buf))+bytesToAdd)
	copy(newBuf, b.buf)
	b.buf = newBuf
}

// wcheck checks if we have enough space to write.
func (b *ByteFrame) wcheck(size uint) {
	if b.index+size > uint(len(b.buf)) {
		b.grow(size)
	}
}

// wprologue is a helpler function to update state after a write.
func (b *ByteFrame) wprologue(size uint) {

	tmp := int(b.index+size) - int(b.usedSize)
	if tmp > 0 {
		b.usedSize += uint(tmp)
	}

	b.index += size
}

// rcheck checks if we have enough data to read.
func (b *ByteFrame) rcheck(size uint) bool {
	if b.index+size > uint(len(b.buf)) || b.index+size > b.usedSize+1 {
		return false
	}
	return true
}

func (b *ByteFrame) rprologue(size uint) {
	b.index += size
}

func (b *ByteFrame) rerr() {
	panic("Error while reading!")
}

// Seek (implements the io.Seeker interface)
func (b *ByteFrame) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if offset > int64(b.usedSize) {
			return int64(b.index), errors.New("cannot seek beyond the max index")
		}
		b.index = uint(offset)
	case io.SeekCurrent:
		newPos := int64(b.index) + offset
		if newPos > int64(b.usedSize) {
			return int64(b.index), errors.New("cannot seek beyond the max index")
		} else if newPos < 0 {
			return int64(b.index), errors.New("cannot seek before the buffer start")
		}
		b.index = uint(newPos)
	case io.SeekEnd:
		newPos := int64(b.usedSize) + offset
		if newPos > int64(b.usedSize) {
			return int64(b.index), errors.New("cannot seek beyond the max index")
		} else if newPos < 0 {
			return int64(b.index), errors.New("cannot seek before the buffer start")
		}
		b.index = uint(newPos)

	}

	return int64(b.index), nil
}

// Data returns the data from the buffer start up to the max index.
func (b *ByteFrame) Data() []byte {
	return b.buf[:b.usedSize]
}

// DataFromCurrent returns the data from the current index up to the max index.
func (b *ByteFrame) DataFromCurrent() []byte {
	return b.buf[b.index:b.usedSize]
}

func (b *ByteFrame) Index() uint {
	return b.index
}

// SetLE sets the byte order to litte endian.
func (b *ByteFrame) SetLE() {
	b.byteOrder = binary.LittleEndian
}

// SetBE sets the byte order to big endian.
func (b *ByteFrame) SetBE() {
	b.byteOrder = binary.BigEndian
}

// WriteUint8 writes a uint8 at the current index.
func (b *ByteFrame) WriteUint8(x uint8) {
	b.wcheck(1)
	b.buf[b.index] = x
	b.wprologue(1)
}

// WriteBool writes a bool at the current index
// (1 byte. true -> 1, false -> 0)
func (b *ByteFrame) WriteBool(x bool) {
	if x {
		b.WriteUint8(1)
	} else {
		b.WriteUint8(0)
	}
}

// WriteUint16 writes a uint16 at the current index.
func (b *ByteFrame) WriteUint16(x uint16) {
	b.wcheck(2)
	b.byteOrder.PutUint16(b.buf[b.index:], x)
	b.wprologue(2)
}

// WriteUint32 writes a uint32 at the current index.
func (b *ByteFrame) WriteUint32(x uint32) {
	b.wcheck(4)
	b.byteOrder.PutUint32(b.buf[b.index:], x)
	b.wprologue(4)
}

// WriteUint64 writes a uint64 at the current index.
func (b *ByteFrame) WriteUint64(x uint64) {
	b.wcheck(8)
	b.byteOrder.PutUint64(b.buf[b.index:], x)
	b.wprologue(8)
}

// WriteInt8 writes a int8 at the current index.
func (b *ByteFrame) WriteInt8(x int8) {
	b.wcheck(1)
	b.buf[b.index] = byte(x)
	b.wprologue(1)
}

// WriteInt16 writes a int16 at the current index.
func (b *ByteFrame) WriteInt16(x int16) {
	b.wcheck(2)
	b.byteOrder.PutUint16(b.buf[b.index:], uint16(x))
	b.wprologue(2)
}

// WriteInt32 writes a int32 at the current index.
func (b *ByteFrame) WriteInt32(x int32) {
	b.wcheck(4)
	b.byteOrder.PutUint32(b.buf[b.index:], uint32(x))
	b.wprologue(4)
}

// WriteInt64 writes a int64 at the current index.
func (b *ByteFrame) WriteInt64(x int64) {
	b.wcheck(8)
	b.byteOrder.PutUint64(b.buf[b.index:], uint64(x))
	b.wprologue(8)
}

// WriteFloat32 writes a float32 at the current index.
func (b *ByteFrame) WriteFloat32(x float32) {
	b.wcheck(4)
	tmp := math.Float32bits(x)
	b.byteOrder.PutUint32(b.buf[b.index:], tmp)
	b.wprologue(4)
}

// WriteFloat64 writes a float64 at the current index
func (b *ByteFrame) WriteFloat64(x float64) {
	b.wcheck(8)
	tmp := math.Float64bits(x)
	b.byteOrder.PutUint64(b.buf[b.index:], tmp)
	b.wprologue(8)
}

// WriteBytes writes a slice of bytes at the current index.
func (b *ByteFrame) WriteBytes(x []byte) {
	b.wcheck(uint(len(x)))
	copy(b.buf[b.index:], x)
	b.wprologue(uint(len(x)))
}

// WriteNullTerminatedBytes write a slice bytes with an additional NULL terminator.
func (b *ByteFrame) WriteNullTerminatedBytes(x []byte) {
	b.WriteBytes(x)
	b.WriteUint8(0)
}

// ReadUint8 reads a uint8 at the current index.
func (b *ByteFrame) ReadUint8() (x uint8) {
	if !b.rcheck(1) {
		b.rerr()
	}
	x = uint8(b.buf[b.index])
	b.rprologue(1)
	return
}

// ReadBool reads a bool at the current index
// (1 byte. b > 0 -> true, b == 0 -> false)
func (b *ByteFrame) ReadBool() (x bool) {
	tmp := b.ReadUint8()
	x = tmp > 0
	return
}

// ReadUint16 reads a uint16 at the current index.
func (b *ByteFrame) ReadUint16() (x uint16) {
	if !b.rcheck(2) {
		b.rerr()
	}
	x = b.byteOrder.Uint16(b.buf[b.index:])
	b.rprologue(2)
	return
}

// ReadUint32 reads a uint32 at the current index.
func (b *ByteFrame) ReadUint32() (x uint32) {
	if !b.rcheck(4) {
		b.rerr()
	}
	x = b.byteOrder.Uint32(b.buf[b.index:])
	b.rprologue(4)
	return
}

// ReadUint64 reads a uint64 at the current index.
func (b *ByteFrame) ReadUint64() (x uint64) {
	if !b.rcheck(8) {
		b.rerr()
	}
	x = b.byteOrder.Uint64(b.buf[b.index:])
	b.rprologue(8)
	return
}

// ReadInt8 reads a int8 at the current index.
func (b *ByteFrame) ReadInt8() (x int8) {
	if !b.rcheck(1) {
		b.rerr()
	}
	x = int8(b.buf[b.index])
	b.rprologue(1)
	return
}

// ReadInt16 reads a int16 at the current index.
func (b *ByteFrame) ReadInt16() (x int16) {
	if !b.rcheck(2) {
		b.rerr()
	}
	x = int16(b.byteOrder.Uint16(b.buf[b.index:]))
	b.rprologue(2)
	return
}

// ReadInt32 reads a int32 at the current index.
func (b *ByteFrame) ReadInt32() (x int32) {
	if !b.rcheck(4) {
		b.rerr()
	}
	x = int32(b.byteOrder.Uint32(b.buf[b.index:]))
	b.rprologue(4)
	return
}

// ReadInt64 reads a int64 at the current index.
func (b *ByteFrame) ReadInt64() (x int64) {
	if !b.rcheck(8) {
		b.rerr()
	}
	x = int64(b.byteOrder.Uint64(b.buf[b.index:]))
	b.rprologue(8)
	return
}

// ReadFloat32 reads a float32 at the current index.
func (b *ByteFrame) ReadFloat32() (x float32) {
	if !b.rcheck(4) {
		b.rerr()
	}
	x = math.Float32frombits(b.byteOrder.Uint32(b.buf[b.index:]))
	b.rprologue(4)
	return
}

// ReadFloat64 reads a float64 at the current index.
func (b *ByteFrame) ReadFloat64() (x float64) {
	if !b.rcheck(8) {
		b.rerr()
	}
	x = math.Float64frombits(b.byteOrder.Uint64(b.buf[b.index:]))
	b.rprologue(8)
	return
}

// ReadBytes reads `size` many bytes at the current index.
func (b *ByteFrame) ReadBytes(size uint) (x []byte) {
	if !b.rcheck(size) {
		b.rerr()
	}
	x = b.buf[b.index : b.index+size]
	b.rprologue(size)
	return
}

// ReadNullTerminatedBytes reads bytes up to a NULL terminator.
func (b *ByteFrame) ReadNullTerminatedBytes() []byte {
	tmpData := b.DataFromCurrent()
	tmp := bytes.SplitN(tmpData, []byte{0x00}, 2)[0]

	if len(tmp) == len(tmpData) {
		return []byte{}
	}

	b.rprologue(uint(len(tmp)) + 1)
	return tmp
}
