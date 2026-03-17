package deltacomp

import (
	"bytes"
	"fmt"
	"io"

	"go.uber.org/zap"
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
	result, err := ApplyDataDiffWithLimit(diff, baseData, 0)
	if err != nil {
		zap.L().Error("ApplyDataDiff failed", zap.Error(err))
		// Return original data on error to avoid corruption
		out := make([]byte, len(baseData))
		copy(out, baseData)
		return out
	}
	return result
}

// ApplyDataDiffWithLimit applies a delta data diff patch onto given base data.
// If maxOutput > 0, the result is capped at that size; exceeding it returns an error.
// If maxOutput == 0, no limit is enforced (backwards-compatible behavior).
func ApplyDataDiffWithLimit(diff []byte, baseData []byte, maxOutput int) ([]byte, error) {
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

		if dataOffset < 0 {
			return nil, fmt.Errorf("negative data offset %d", dataOffset)
		}
		if differentCount < 0 {
			return nil, fmt.Errorf("negative different count %d at offset %d", differentCount, dataOffset)
		}

		endOffset := dataOffset + differentCount
		if maxOutput > 0 && endOffset > maxOutput {
			return nil, fmt.Errorf("patch writes to offset %d, exceeds limit %d", endOffset, maxOutput)
		}

		// Grow slice if required
		if endOffset > len(baseCopy) {
			baseCopy = append(baseCopy, make([]byte, endOffset-len(baseCopy))...)
		}

		// Apply the patch bytes.
		for i := 0; i < differentCount; i++ {
			b, err := checkReadUint8(patch)
			if err != nil {
				return nil, fmt.Errorf("truncated patch at offset %d+%d: %w", dataOffset, i, err)
			}

			baseCopy[dataOffset+i] = b
		}

		dataOffset += differentCount - 1
	}

	return baseCopy, nil
}
