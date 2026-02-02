package channelserver

import (
	"testing"
)

func TestFindSubSliceIndices(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		sub      []byte
		expected []int
	}{
		{
			name:     "empty data",
			data:     []byte{},
			sub:      []byte{0x01},
			expected: nil,
		},
		{
			name:     "empty sub",
			data:     []byte{0x01, 0x02, 0x03},
			sub:      []byte{},
			expected: []int{0, 1, 2},
		},
		{
			name:     "single match at start",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			sub:      []byte{0x01, 0x02},
			expected: []int{0},
		},
		{
			name:     "single match at end",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			sub:      []byte{0x03, 0x04},
			expected: []int{2},
		},
		{
			name:     "single match in middle",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			sub:      []byte{0x02, 0x03, 0x04},
			expected: []int{1},
		},
		{
			name:     "multiple matches",
			data:     []byte{0x01, 0x02, 0x01, 0x02, 0x01, 0x02},
			sub:      []byte{0x01, 0x02},
			expected: []int{0, 2, 4},
		},
		{
			name:     "no match",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			sub:      []byte{0x05, 0x06},
			expected: nil,
		},
		{
			name:     "sub larger than data",
			data:     []byte{0x01, 0x02},
			sub:      []byte{0x01, 0x02, 0x03},
			expected: nil,
		},
		{
			name:     "overlapping matches",
			data:     []byte{0x01, 0x01, 0x01, 0x01},
			sub:      []byte{0x01, 0x01},
			expected: []int{0, 1, 2},
		},
		{
			name:     "single byte match",
			data:     []byte{0xAA, 0xBB, 0xAA, 0xCC, 0xAA},
			sub:      []byte{0xAA},
			expected: []int{0, 2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSubSliceIndices(tt.data, tt.sub)
			if !intSlicesEqual(result, tt.expected) {
				t.Errorf("findSubSliceIndices(%v, %v) = %v, want %v", tt.data, tt.sub, result, tt.expected)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{
			name:     "both empty",
			a:        []byte{},
			b:        []byte{},
			expected: true,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "equal slices",
			a:        []byte{0x01, 0x02, 0x03},
			b:        []byte{0x01, 0x02, 0x03},
			expected: true,
		},
		{
			name:     "different length",
			a:        []byte{0x01, 0x02},
			b:        []byte{0x01, 0x02, 0x03},
			expected: false,
		},
		{
			name:     "same length different content",
			a:        []byte{0x01, 0x02, 0x03},
			b:        []byte{0x01, 0x02, 0x04},
			expected: false,
		},
		{
			name:     "single byte equal",
			a:        []byte{0xFF},
			b:        []byte{0xFF},
			expected: true,
		},
		{
			name:     "single byte different",
			a:        []byte{0xFF},
			b:        []byte{0xFE},
			expected: false,
		},
		{
			name:     "one empty one not",
			a:        []byte{},
			b:        []byte{0x01},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equal(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("equal(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestEqual_Symmetry(t *testing.T) {
	// equal(a, b) should always equal equal(b, a)
	testCases := [][]byte{
		{0x01, 0x02, 0x03},
		{0x01, 0x02},
		{},
		{0xFF},
	}

	for i, a := range testCases {
		for j, b := range testCases {
			resultAB := equal(a, b)
			resultBA := equal(b, a)
			if resultAB != resultBA {
				t.Errorf("Symmetry failed: equal(case[%d], case[%d])=%v but equal(case[%d], case[%d])=%v",
					i, j, resultAB, j, i, resultBA)
			}
		}
	}
}

// Helper function to compare int slices
func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// BackportQuest tests are skipped because they require runtime configuration
// (ErupeConfig.RealClientMode) which is not available in unit tests.
// Integration tests should cover this function.
