package deltacomp

import (
	"bytes"
	"io"
	"fmt"
)

func checkReadUint8(r *bytes.Reader) (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func checkReadUint16(r *bytes.Reader) (uint16, error) {
	data := make([]byte, 2)
	n, err := r.Read(data)
	if err != nil {
		return 0, err
	} else if n != len(data) {
		return 0, io.EOF
	}

	return uint16(data[0])<<8 | uint16(data[1]), nil
}

func readCount(r *bytes.Reader) (int, error) {
	var count int

	count8, err := checkReadUint8(r)
	if err != nil {
		return 0, err
	}
	count = int(count8)

	if count == 0 {
		count16, err := checkReadUint16(r)
		if err != nil {
			return 0, err
		}
		count = int(count16)
	}

	return int(count), nil
}

// ApplyDataDiff applies a delta data diff patch onto given base data.
func ApplyDataDiff(diff []byte, baseData []byte) []byte {
	// Make a copy of the base data to return,
	// (probably just make this modify the given slice in the future).
	baseCopy := make([]byte, len(baseData))
	copy(baseCopy, baseData)

	patch := bytes.NewReader(diff)

	// The very first matchCount is +1 more than it should be, so we start at -1.
	dataOffset := -1
	for {
		// Read the amount of matching bytes.
		matchCount, err := readCount(patch)
		if err != nil {
			// No more data
			break
		}

		dataOffset += matchCount

		// Read the amount of differing bytes.
		differentCount, err := readCount(patch)
		if err != nil {
			// No more data
			break
		}
		differentCount--

		// Grow slice if it's required
		if(len(baseCopy) < dataOffset){
			fmt.Printf("Slice smaller than data offset, growing slice...")
 			baseCopy = append(baseCopy, make([]byte, (dataOffset + differentCount) - len(baseData))...)
		} else {
			length := len(baseCopy[dataOffset:])
			if length < differentCount {
				length -= differentCount
				baseCopy = append(baseCopy, make([]byte, length)...)
			}
		}


		// Apply the patch bytes.
		for i := 0; i < differentCount; i++ {
			b, err := checkReadUint8(patch)
			if err != nil {
				panic("Invalid or misunderstood patch format!")
			}


			baseCopy[dataOffset+i] = b
		}

		dataOffset += differentCount - 1

	}

	return baseCopy
}
