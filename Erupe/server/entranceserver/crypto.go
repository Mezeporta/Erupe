package entranceserver

import (
	"encoding/binary"
)

var (
	_bin8Key     = []byte{0x01, 0x23, 0x34, 0x45, 0x56, 0xAB, 0xCD, 0xEF}
	_sum32Table0 = []byte{0x35, 0x7A, 0xAA, 0x97, 0x53, 0x66, 0x12}
	_sum32Table1 = []byte{0x7A, 0xAA, 0x97, 0x53, 0x66, 0x12, 0xDE, 0xDE, 0x35}
)

// CalcSum32 calculates the custom MHF "sum32" checksum of the given data.
func CalcSum32(data []byte) uint32 {
	tableIdx0 := int(len(data) & 0xFF)
	tableIdx1 := int(data[len(data)>>1] & 0xFF)

	out := make([]byte, 4)
	for i := 0; i < len(data); i++ {
		tableIdx0++
		tableIdx1++

		tmp := byte((_sum32Table1[tableIdx1%9] ^ _sum32Table0[tableIdx0%7]) ^ data[i])
		out[i&3] = (out[i&3] + tmp) & 0xFF
	}

	return binary.BigEndian.Uint32(out)
}

// EncryptBin8 encrypts the given data using MHF's "binary8" encryption.
func EncryptBin8(data []byte, key byte) []byte {
	curKey := uint32(((54323 * uint(key)) + 1) & 0xFFFFFFFF)

	var output []byte
	for i := 0; i < len(data); i++ {
		tmp := (_bin8Key[i&7] ^ byte((curKey>>13)&0xFF))
		output = append(output, data[i]^tmp)
		curKey = uint32(((54323 * uint(curKey)) + 1) & 0xFFFFFFFF)
	}

	return output
}

// DecryptBin8 decrypts the given MHF "binary8" data.
func DecryptBin8(data []byte, key byte) []byte {
	curKey := uint32(((54323 * uint(key)) + 1) & 0xFFFFFFFF)

	var output []byte
	for i := 0; i < len(data); i++ {
		tmp := (data[i] ^ byte((curKey>>13)&0xFF))
		output = append(output, tmp^_bin8Key[i&7])
		curKey = uint32(((54323 * uint(curKey)) + 1) & 0xFFFFFFFF)
	}

	return output
}
