package decryption

/*
	This code is HEAVILY based from
	https://github.com/Chakratos/ReFrontier/blob/master/ReFrontier/Unpack.cs
*/

import (
	"erupe-ce/common/byteframe"
	"io"
)

var mShiftIndex = 0
var mFlag = byte(0)

// UnpackSimple decompresses a JPK type-3 compressed byte slice. If the data
// does not start with the JKR magic header it is returned unchanged.
func UnpackSimple(data []byte) []byte {
	mShiftIndex = 0
	mFlag = byte(0)

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
			ProcessDecode(bf, outBuffer)

			return outBuffer
		}
	}

	return data
}

// ProcessDecode runs the JPK LZ-style decompression loop, reading compressed
// tokens from data and writing decompressed bytes into outBuffer.
func ProcessDecode(data *byteframe.ByteFrame, outBuffer []byte) {
	outIndex := 0

	for int(data.Index()) < len(data.Data()) && outIndex < len(outBuffer)-1 {
		if JPKBitShift(data) == 0 {
			outBuffer[outIndex] = ReadByte(data)
			outIndex++
			continue
		} else {
			if JPKBitShift(data) == 0 {
				length := (JPKBitShift(data) << 1) | JPKBitShift(data)
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
					if JPKBitShift(data) == 0 {
						length := (JPKBitShift(data) << 3) | (JPKBitShift(data) << 2) | (JPKBitShift(data) << 1) | JPKBitShift(data)
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

// JPKBitShift reads one bit from the compressed stream's flag byte, refilling
// the flag from the next byte in data when all 8 bits have been consumed.
func JPKBitShift(data *byteframe.ByteFrame) byte {
	mShiftIndex--

	if mShiftIndex < 0 {
		mShiftIndex = 7
		mFlag = ReadByte(data)
	}

	return (byte)((mFlag >> mShiftIndex) & 1)
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
