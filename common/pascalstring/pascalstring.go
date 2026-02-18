package pascalstring

import (
	"erupe-ce/common/byteframe"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Uint8 writes x as a null-terminated string with a uint8 length prefix. If t
// is true the string is first encoded to Shift-JIS.
func Uint8(bf *byteframe.ByteFrame, x string, t bool) {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			bf.WriteUint8(0)
			return
		}
		x = xt
	}
	bf.WriteUint8(uint8(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

// Uint16 writes x as a null-terminated string with a uint16 length prefix. If
// t is true the string is first encoded to Shift-JIS.
func Uint16(bf *byteframe.ByteFrame, x string, t bool) {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			bf.WriteUint16(0)
			return
		}
		x = xt
	}
	bf.WriteUint16(uint16(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

// Uint32 writes x as a null-terminated string with a uint32 length prefix. If
// t is true the string is first encoded to Shift-JIS.
func Uint32(bf *byteframe.ByteFrame, x string, t bool) {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			bf.WriteUint32(0)
			return
		}
		x = xt
	}
	bf.WriteUint32(uint32(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}
