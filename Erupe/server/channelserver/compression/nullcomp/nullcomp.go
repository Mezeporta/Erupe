package nullcomp

import (
	"bytes"
	"io"
)

// Decompress decompresses null-compressesed data.
func Decompress(compData []byte) ([]byte, error) {
	r := bytes.NewReader(compData)

	header := make([]byte, 16)
	n, err := r.Read(header)
	if err != nil {
		return nil, err
	} else if n != len(header) {
		return nil, err
	}

	// Just return the data if it doesn't contain the cmp header.
	if !bytes.Equal(header, []byte("cmp\x2020110113\x20\x20\x20\x00")) {
		return compData, nil
	}

	var output []byte
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if b == 0 {
			// If it's a null byte, then the next byte is how many nulls to add.
			nullCount, err := r.ReadByte()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			output = append(output, make([]byte, int(nullCount))...)
		} else {
			output = append(output, b)
		}
	}

	return output, nil
}

// Compress null compresses give given data.
func Compress(rawData []byte) ([]byte, error) {
	r := bytes.NewReader(rawData)
	var output []byte
	output = append(output, []byte("cmp\x2020110113\x20\x20\x20\x00")...)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if b == 0 {
			output = append(output, []byte{0x00}...)
			// read to get null count
			nullCount := 1
			for {
				i, err := r.ReadByte()
				if err == io.EOF {
					output = append(output, []byte{byte(nullCount)}...)
					break
				} else if i != 0 {
					r.UnreadByte()
					output = append(output, []byte{byte(nullCount)}...)
					break
				} else if err != nil {
					return nil, err
				}
				nullCount++

				// Flush the null-count if it gets to 255, start on the next null count.
				if nullCount == 255 {
					output = append(output, []byte{0xFF, 0x00}...)
					nullCount = 0
				}
			}
		} else {
			output = append(output, b)
		}
	}
	return output, nil
}
