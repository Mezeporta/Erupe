package decryption

import "encoding/binary"

// PackSimple compresses data using JPK type-3 (LZ77) compression and wraps it
// in a JKR header. It is the inverse of UnpackSimple.
func PackSimple(data []byte) []byte {
	compressed := lzEncode(data)

	out := make([]byte, 16+len(compressed))
	binary.LittleEndian.PutUint32(out[0:4], 0x1A524B4A)  // JKR magic
	binary.LittleEndian.PutUint16(out[4:6], 0x0108)      // version
	binary.LittleEndian.PutUint16(out[6:8], 0x0003)      // type 3 = LZ only
	binary.LittleEndian.PutUint32(out[8:12], 0x00000010) // data offset = 16 (after header)
	binary.LittleEndian.PutUint32(out[12:16], uint32(len(data)))
	copy(out[16:], compressed)
	return out
}

// lzEncoder holds mutable state for the LZ77 compression loop.
// Ported from ReFrontier JPKEncodeLz.cs.
//
// The format groups 8 items behind a flag byte (MSB = item 0):
//
//	bit=0 → literal byte follows
//	bit=1 → back-reference follows (with sub-cases below)
//
// Back-reference sub-cases:
//
//	10xx + 1 byte  → length 3–6,  offset ≤ 255
//	11 + 2 bytes   → length 3–9,  offset ≤ 8191  (length encoded in hi byte bits 7–5)
//	11 + 2 bytes + 0 + 4 bits → length 10–25, offset ≤ 8191
//	11 + 2 bytes + 1 + 1 byte → length 26–280, offset ≤ 8191
type lzEncoder struct {
	flag         byte
	shiftIndex   int
	toWrite      [1024]byte // data bytes for the current flag group
	indexToWrite int
	out          []byte
}

func (e *lzEncoder) setFlag(value bool) {
	if e.shiftIndex <= 0 {
		e.flushFlag(false)
		e.shiftIndex = 7
	} else {
		e.shiftIndex--
	}
	if value {
		e.flag |= 1 << uint(e.shiftIndex)
	}
}

// setFlagsReverse writes `count` bits of value MSB-first.
func (e *lzEncoder) setFlagsReverse(value byte, count int) {
	for i := count - 1; i >= 0; i-- {
		e.setFlag(((value >> uint(i)) & 1) == 1)
	}
}

func (e *lzEncoder) writeByte(b byte) {
	e.toWrite[e.indexToWrite] = b
	e.indexToWrite++
}

func (e *lzEncoder) flushFlag(final bool) {
	if !final || e.indexToWrite > 0 {
		e.out = append(e.out, e.flag)
	}
	e.flag = 0
	e.out = append(e.out, e.toWrite[:e.indexToWrite]...)
	e.indexToWrite = 0
}

// lzEncode compresses data with the JPK LZ77 algorithm, producing the raw
// compressed bytes (without the JKR header).
func lzEncode(data []byte) []byte {
	const (
		compressionLevel = 280   // max match length
		maxIndexDist     = 0x300 // max look-back distance (768)
	)

	enc := &lzEncoder{shiftIndex: 8}

	for pos := 0; pos < len(data); {
		repLen, repOff := lzLongestRepetition(data, pos, compressionLevel, maxIndexDist)

		if repLen == 0 {
			// Literal byte
			enc.setFlag(false)
			enc.writeByte(data[pos])
			pos++
		} else {
			enc.setFlag(true)
			if repLen <= 6 && repOff <= 0xff {
				// Short: flag=10, 2-bit length, 1-byte offset
				enc.setFlag(false)
				enc.setFlagsReverse(byte(repLen-3), 2)
				enc.writeByte(byte(repOff))
			} else {
				// Long: flag=11, 2-byte offset/length header
				enc.setFlag(true)
				u16 := uint16(repOff)
				if repLen <= 9 {
					// Length fits in hi byte bits 7-5
					u16 |= uint16(repLen-2) << 13
				}
				enc.writeByte(byte(u16 >> 8))
				enc.writeByte(byte(u16 & 0xff))
				if repLen > 9 {
					if repLen <= 25 {
						// Extended: flag=0, 4-bit length
						enc.setFlag(false)
						enc.setFlagsReverse(byte(repLen-10), 4)
					} else {
						// Extended: flag=1, 1-byte length
						enc.setFlag(true)
						enc.writeByte(byte(repLen - 0x1a))
					}
				}
			}
			pos += repLen
		}
	}

	enc.flushFlag(true)
	return enc.out
}

// lzLongestRepetition finds the longest match for data[pos:] in the look-back
// window. Returns (matchLen, encodedOffset) where encodedOffset is
// (pos - matchStart - 1). Returns (0, 0) when no usable match exists.
func lzLongestRepetition(data []byte, pos, compressionLevel, maxIndexDist int) (int, uint) {
	const minLength = 3

	// Clamp threshold to available bytes
	threshold := compressionLevel
	if remaining := len(data) - pos; remaining < threshold {
		threshold = remaining
	}

	if pos == 0 || threshold < minLength {
		return 0, 0
	}

	windowStart := pos - maxIndexDist
	if windowStart < 0 {
		windowStart = 0
	}

	maxLen := 0
	var bestOffset uint

	for left := windowStart; left < pos; left++ {
		curLen := 0
		for curLen < threshold && data[left+curLen] == data[pos+curLen] {
			curLen++
		}
		if curLen >= minLength && curLen > maxLen {
			maxLen = curLen
			bestOffset = uint(pos - left - 1)
			if maxLen >= threshold {
				break
			}
		}
	}

	return maxLen, bestOffset
}
