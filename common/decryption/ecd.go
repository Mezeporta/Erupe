package decryption

/*
	ECD encryption/decryption ported from:
	  - ReFrontier (C#): https://github.com/Chakratos/ReFrontier (LibReFrontier/Crypto.cs)
	  - FrontierTextHandler (Python): src/crypto.py

	ECD is a stream cipher used to protect MHF game data files. All known
	MHF files use key index 4 (DefaultECDKey). The cipher uses a 32-bit LCG
	for key-stream generation with a Feistel-like nibble transformation and
	ciphertext-feedback (CFB) chaining.
*/

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

// ECDMagic is the ECD container magic ("ecd\x1a"), stored little-endian on disk.
// On-disk bytes: 65 63 64 1A; decoded as LE uint32: 0x1A646365.
const ECDMagic = uint32(0x1A646365)

// DefaultECDKey is the LCG key index used by all known MHF game files.
const DefaultECDKey = 4

const ecdHeaderSize = 16

// rndBufECD holds the 6 LCG key-parameter sets. Each entry is an 8-byte pair
// of (multiplier, increment) stored big-endian, indexed by the key field in
// the ECD header.
var rndBufECD = [...]byte{
	0x4A, 0x4B, 0x52, 0x2E, 0x00, 0x00, 0x00, 0x01, // key 0
	0x00, 0x01, 0x0D, 0xCD, 0x00, 0x00, 0x00, 0x01, // key 1
	0x00, 0x01, 0x0D, 0xCD, 0x00, 0x00, 0x00, 0x01, // key 2
	0x00, 0x01, 0x0D, 0xCD, 0x00, 0x00, 0x00, 0x01, // key 3
	0x00, 0x19, 0x66, 0x0D, 0x00, 0x00, 0x00, 0x03, // key 4 (default; all MHF files)
	0x7D, 0x2B, 0x89, 0xDD, 0x00, 0x00, 0x00, 0x01, // key 5
}

const numECDKeys = len(rndBufECD) / 8

// getRndECD advances the LCG by one step using the selected key's parameters
// and returns the new 32-bit state.
func getRndECD(key int, rnd uint32) uint32 {
	offset := key * 8
	multiplier := binary.BigEndian.Uint32(rndBufECD[offset:])
	increment := binary.BigEndian.Uint32(rndBufECD[offset+4:])
	return rnd*multiplier + increment
}

// DecodeECD decrypts an ECD-encrypted buffer and returns the plaintext payload.
// The 16-byte ECD header is consumed; only the decrypted payload is returned.
//
// The cipher uses the CRC32 stored in the header to seed the LCG key stream.
// No post-decryption CRC check is performed (matching reference implementations).
func DecodeECD(data []byte) ([]byte, error) {
	if len(data) < ecdHeaderSize {
		return nil, errors.New("ecd: buffer too small for header")
	}
	if binary.LittleEndian.Uint32(data[:4]) != ECDMagic {
		return nil, errors.New("ecd: invalid magic")
	}

	key := int(binary.LittleEndian.Uint16(data[4:6]))
	if key >= numECDKeys {
		return nil, fmt.Errorf("ecd: invalid key index %d", key)
	}

	payloadSize := int(binary.LittleEndian.Uint32(data[8:12]))
	if len(data) < ecdHeaderSize+payloadSize {
		return nil, fmt.Errorf("ecd: declared payload size %d exceeds buffer (%d bytes available)",
			payloadSize, len(data)-ecdHeaderSize)
	}

	// Seed LCG: rotate the stored CRC32 by 16 bits and set LSB to 1.
	storedCRC := binary.LittleEndian.Uint32(data[12:16])
	rnd := (storedCRC<<16 | storedCRC>>16) | 1

	// Initial LCG step establishes the cipher-feedback byte r8.
	rnd = getRndECD(key, rnd)
	r8 := byte(rnd)

	out := make([]byte, payloadSize)
	for i := 0; i < payloadSize; i++ {
		rnd = getRndECD(key, rnd)
		xorpad := rnd

		// Nibble-feedback decryption: XOR with previous decrypted byte, then
		// apply 8 rounds of Feistel-like nibble mixing using the key stream.
		r11 := uint32(data[ecdHeaderSize+i]) ^ uint32(r8)
		r12 := (r11 >> 4) & 0xFF

		for j := 0; j < 8; j++ {
			r10 := xorpad ^ r11
			r11 = r12
			r12 = (r12 ^ r10) & 0xFF
			xorpad >>= 4
		}

		r8 = byte((r12 & 0xF) | ((r11 & 0xF) << 4))
		out[i] = r8
	}

	return out, nil
}

// EncodeECD encrypts plaintext using the ECD cipher and returns the complete
// ECD container (16-byte header + encrypted payload). Use DefaultECDKey (4)
// for all MHF-compatible output.
func EncodeECD(data []byte, key int) ([]byte, error) {
	if key < 0 || key >= numECDKeys {
		return nil, fmt.Errorf("ecd: invalid key index %d", key)
	}

	payloadSize := len(data)
	checksum := crc32.ChecksumIEEE(data)

	out := make([]byte, ecdHeaderSize+payloadSize)
	binary.LittleEndian.PutUint32(out[0:], ECDMagic)
	binary.LittleEndian.PutUint16(out[4:], uint16(key))
	// out[6:8] = 0 (reserved padding)
	binary.LittleEndian.PutUint32(out[8:], uint32(payloadSize))
	binary.LittleEndian.PutUint32(out[12:], checksum)

	// Seed LCG identically to decryption so the streams stay in sync.
	rnd := (checksum<<16 | checksum>>16) | 1
	rnd = getRndECD(key, rnd)
	r8 := byte(rnd)

	for i := 0; i < payloadSize; i++ {
		rnd = getRndECD(key, rnd)
		xorpad := rnd

		// Inverse Feistel: compute the nibble-mixed values using a zeroed
		// initial state, then XOR the plaintext nibbles through.
		r11 := uint32(0)
		r12 := uint32(0)

		for j := 0; j < 8; j++ {
			r10 := xorpad ^ r11
			r11 = r12
			r12 = (r12 ^ r10) & 0xFF
			xorpad >>= 4
		}

		b := data[i]
		dig2 := uint32(b)
		dig1 := (dig2 >> 4) & 0xFF
		dig1 ^= r11
		dig2 ^= r12
		dig1 ^= dig2

		rr := byte((dig2 & 0xF) | ((dig1 & 0xF) << 4))
		rr ^= r8
		out[ecdHeaderSize+i] = rr
		r8 = b // Cipher-feedback: next iteration uses current plaintext byte.
	}

	return out, nil
}
