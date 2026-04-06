package decryption

/*
	This code is HEAVILY based from
	https://github.com/Chakratos/ReFrontier/blob/master/ReFrontier/Unpack.cs
*/

import (
	"erupe-ce/common/byteframe"
	"io"
)

// jpkState holds the mutable bit-reader state for a single JPK decompression.
// This is local to each call, making concurrent UnpackSimple calls safe.
type jpkState struct {
	shiftIndex int
	flag       byte
}

// UnpackSimple decompresses a JPK type-3 compressed byte slice. If the data
// does not start with the JKR magic header it is returned unchanged.
func UnpackSimple(data []byte) []byte {
	bf := byteframe.NewByteFrameFromBytes(data)
	bf.SetLE()
	header := bf.ReadUint32()

	if header == 0x1A524B4A {
		_, _ = bf.Seek(0x2, io.SeekCurrent)
		jpkType := bf.ReadUint16()

		switch jpkType {
		case 3:
			startOffset := bf.ReadInt32()
			outSize := bf.ReadInt32()
			outBuffer := make([]byte, outSize)
			_, _ = bf.Seek(int64(startOffset), io.SeekStart)
			s := &jpkState{}
			s.processDecode(bf, outBuffer)

			return outBuffer
		}
	}

	return data
}

// ProcessDecode runs the JPK LZ-style decompression loop, reading compressed
// tokens from data and writing decompressed bytes into outBuffer.
func ProcessDecode(data *byteframe.ByteFrame, outBuffer []byte) {
	s := &jpkState{}
	s.processDecode(data, outBuffer)
}

func (s *jpkState) processDecode(data *byteframe.ByteFrame, outBuffer []byte) {
	outIndex := 0

	for int(data.Index()) < len(data.Data()) && outIndex < len(outBuffer) {
		if s.bitShift(data) == 0 {
			outBuffer[outIndex] = ReadByte(data)
			outIndex++
			continue
		} else {
			if s.bitShift(data) == 0 {
				length := (s.bitShift(data) << 1) | s.bitShift(data)
				off := ReadByte(data)
				JPKCopy(outBuffer, int(off), int(length)+3, &outIndex)
				continue
			} else {
				hi := ReadByte(data)
				lo := ReadByte(data)
				length := int(hi&0xE0) >> 5
				off := ((int(hi) & 0x1F) << 8) | int(lo)
				if length != 0 {
					JPKCopy(outBuffer, off, length+2, &outIndex)
					continue
				} else {
					if s.bitShift(data) == 0 {
						length := (s.bitShift(data) << 3) | (s.bitShift(data) << 2) | (s.bitShift(data) << 1) | s.bitShift(data)
						JPKCopy(outBuffer, off, int(length)+2+8, &outIndex)
						continue
					} else {
						temp := ReadByte(data)
						if temp == 0xFF {
							for i := 0; i < off+0x1B; i++ {
								outBuffer[outIndex] = ReadByte(data)
								outIndex++
								continue
							}
						} else {
							JPKCopy(outBuffer, off, int(temp)+0x1a, &outIndex)
						}
					}
				}
			}
		}
	}
}

// bitShift reads one bit from the compressed stream's flag byte, refilling
// the flag from the next byte in data when all 8 bits have been consumed.
func (s *jpkState) bitShift(data *byteframe.ByteFrame) byte {
	s.shiftIndex--

	if s.shiftIndex < 0 {
		s.shiftIndex = 7
		s.flag = ReadByte(data)
	}

	return (s.flag >> s.shiftIndex) & 1
}

// JPKCopy copies length bytes from a previous position in outBuffer (determined
// by offset back from the current index) to implement LZ back-references.
func JPKCopy(outBuffer []byte, offset int, length int, index *int) {
	for i := 0; i < length; i++ {
		outBuffer[*index] = outBuffer[*index-offset-1]
		*index++
	}
}

// ReadByte reads a single byte from the ByteFrame.
func ReadByte(bf *byteframe.ByteFrame) byte {
	value := bf.ReadUint8()
	return value
}
