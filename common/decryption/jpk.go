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

// UnpackSimple decompresses a JPK type-3 (LZ) or type-4 (HFI: Huffman+LZ)
// compressed byte slice. If the data does not start with the JKR magic
// header, or uses an unsupported compression type, it is returned unchanged.
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
		case 4:
			startOffset := bf.ReadInt32()
			outSize := bf.ReadInt32()
			outBuffer := make([]byte, outSize)
			_, _ = bf.Seek(int64(startOffset), io.SeekStart)
			s := &hfiState{}
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

// hfiState holds the mutable state for a single JPK type-4 (HFI: Huffman +
// LZ77) decompression. It embeds the outer LZ77 flag-bit register (shared
// shape with jpkState) and adds a second bit register for the inner
// Huffman-coded byte stream, plus the tree table geometry needed to walk it.
// This is local to each call, making concurrent decompressions safe.
type hfiState struct {
	jpkState     // outer LZ77 flag-bit state, fed by readHFByte
	hfShiftIndex int
	hfFlag       byte
	hfDataOffset int64 // cursor into bf for the Huffman bit-stream
	tableOffset  int64 // byte offset of the huffman tree table
	tableLen     int   // number of int16 entries in the table
}

// readHFByte decodes one byte by walking the Huffman tree stored at the
// start of the compressed block. Starting from the tree root (tableLen),
// each bit read from the Huffman bit-stream selects the left (0) or right
// (1) child until a leaf value (<0x100) is reached.
func (s *hfiState) readHFByte(bf *byteframe.ByteFrame) byte {
	node := s.tableLen
	for node >= 0x100 {
		s.hfShiftIndex--
		if s.hfShiftIndex < 0 {
			s.hfShiftIndex = 7
			_, _ = bf.Seek(s.hfDataOffset, io.SeekStart)
			s.hfFlag = bf.ReadUint8()
			s.hfDataOffset++
		}
		bit := (s.hfFlag >> uint(s.hfShiftIndex)) & 1
		idx := (int64(node*2-0x200+int(bit)))*2 + s.tableOffset
		_, _ = bf.Seek(idx, io.SeekStart)
		node = int(bf.ReadInt16())
	}
	return byte(node)
}

// hfiBitShift reads one bit from the outer LZ77 flag stream, mirroring
// jpkState.bitShift but sourcing its bytes through readHFByte instead of
// directly from the stream: for HFI, the LZ77 layer operates on bytes that
// have already been Huffman-decoded.
func (s *hfiState) hfiBitShift(bf *byteframe.ByteFrame) byte {
	s.shiftIndex--
	if s.shiftIndex < 0 {
		s.shiftIndex = 7
		s.flag = s.readHFByte(bf)
	}
	return (s.flag >> uint(s.shiftIndex)) & 1
}

// processDecode runs the same LZ77 decompression loop as jpkState's, but
// reads its bytes through the Huffman decoder (readHFByte) instead of
// directly from bf. bf must be positioned at the 2-byte Huffman table length
// field that begins the compressed block.
func (s *hfiState) processDecode(bf *byteframe.ByteFrame, outBuffer []byte) {
	s.tableLen = int(bf.ReadInt16())
	s.tableOffset = int64(bf.Index())
	s.hfDataOffset = s.tableOffset + int64(s.tableLen)*4 - 0x3fc

	outIndex := 0
	for outIndex < len(outBuffer) {
		if s.hfiBitShift(bf) == 0 {
			outBuffer[outIndex] = s.readHFByte(bf)
			outIndex++
			continue
		}

		if s.hfiBitShift(bf) == 0 {
			length := (s.hfiBitShift(bf) << 1) | s.hfiBitShift(bf)
			off := int(s.readHFByte(bf))
			JPKCopy(outBuffer, off, int(length)+3, &outIndex)
			continue
		}

		hi := s.readHFByte(bf)
		lo := s.readHFByte(bf)
		length := int(hi&0xE0) >> 5
		off := ((int(hi) & 0x1F) << 8) | int(lo)
		if length != 0 {
			JPKCopy(outBuffer, off, length+2, &outIndex)
			continue
		}

		if s.hfiBitShift(bf) == 0 {
			var length byte
			for i := 3; i >= 0; i-- {
				length |= s.hfiBitShift(bf) << uint(i)
			}
			JPKCopy(outBuffer, off, int(length)+2+8, &outIndex)
			continue
		}

		temp := s.readHFByte(bf)
		if temp == 0xFF {
			for i := 0; i < off+0x1B && outIndex < len(outBuffer); i++ {
				outBuffer[outIndex] = s.readHFByte(bf)
				outIndex++
			}
			continue
		}
		JPKCopy(outBuffer, off, int(temp)+0x1a, &outIndex)
	}
}
