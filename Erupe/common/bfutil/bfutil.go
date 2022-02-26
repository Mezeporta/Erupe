package bfutil

import "bytes"

// UpToNull returns the given byte slice's data, up to (not including) the first null byte.
func UpToNull(data []byte) []byte {
	return bytes.SplitN(data, []byte{0x00}, 2)[0]
}
