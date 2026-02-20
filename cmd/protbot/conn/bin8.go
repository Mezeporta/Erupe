package conn

import "encoding/binary"

var (
	bin8Key     = []byte{0x01, 0x23, 0x34, 0x45, 0x56, 0xAB, 0xCD, 0xEF}
	sum32Table0 = []byte{0x35, 0x7A, 0xAA, 0x97, 0x53, 0x66, 0x12}
	sum32Table1 = []byte{0x7A, 0xAA, 0x97, 0x53, 0x66, 0x12, 0xDE, 0xDE, 0x35}
)

// CalcSum32 calculates the custom MHF "sum32" checksum.
func CalcSum32(data []byte) uint32 {
	tableIdx0 := (len(data) + 1) & 0xFF
	tableIdx1 := int((data[len(data)>>1] + 1) & 0xFF)
	out := make([]byte, 4)
	for i := 0; i < len(data); i++ {
		key := data[i] ^ sum32Table0[(tableIdx0+i)%7] ^ sum32Table1[(tableIdx1+i)%9]
		out[i&3] = (out[i&3] + key) & 0xFF
	}
	return binary.BigEndian.Uint32(out)
}

func rotate(k *uint32) {
	*k = uint32(((54323 * uint(*k)) + 1) & 0xFFFFFFFF)
}

// DecryptBin8 decrypts MHF "binary8" data.
func DecryptBin8(data []byte, key byte) []byte {
	k := uint32(key)
	output := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		rotate(&k)
		tmp := data[i] ^ byte((k>>13)&0xFF)
		output[i] = tmp ^ bin8Key[i&7]
	}
	return output
}
