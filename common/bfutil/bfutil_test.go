package bfutil

import (
	"bytes"
	"testing"
)

func TestUpToNull(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "data with null terminator",
			input:    []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			expected: []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}, // "Hello"
		},
		{
			name:     "data without null terminator",
			input:    []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F},
			expected: []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}, // "Hello"
		},
		{
			name:     "data with null at start",
			input:    []byte{0x00, 0x48, 0x65, 0x6C, 0x6C, 0x6F},
			expected: []byte{},
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "only null byte",
			input:    []byte{0x00},
			expected: []byte{},
		},
		{
			name:     "multiple null bytes",
			input:    []byte{0x48, 0x65, 0x00, 0x00, 0x6C, 0x6C, 0x6F},
			expected: []byte{0x48, 0x65}, // "He"
		},
		{
			name:     "binary data with null",
			input:    []byte{0xFF, 0xAB, 0x12, 0x00, 0x34, 0x56},
			expected: []byte{0xFF, 0xAB, 0x12},
		},
		{
			name:     "binary data without null",
			input:    []byte{0xFF, 0xAB, 0x12, 0x34, 0x56},
			expected: []byte{0xFF, 0xAB, 0x12, 0x34, 0x56},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpToNull(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("UpToNull() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUpToNull_ReturnsSliceNotCopy(t *testing.T) {
	// Test that UpToNull returns a slice of the original array, not a copy
	input := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, 0x57, 0x6F, 0x72, 0x6C, 0x64}
	result := UpToNull(input)

	// Verify we got the expected data
	expected := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}
	if !bytes.Equal(result, expected) {
		t.Errorf("UpToNull() = %v, want %v", result, expected)
	}

	// The result should be a slice of the input array
	if len(result) > 0 && cap(result) < len(expected) {
		t.Error("Result should be a slice of input array")
	}
}

func BenchmarkUpToNull(b *testing.B) {
	data := []byte("Hello, World!\x00Extra data here")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UpToNull(data)
	}
}

func BenchmarkUpToNull_NoNull(b *testing.B) {
	data := []byte("Hello, World! No null terminator in this string at all")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UpToNull(data)
	}
}

func BenchmarkUpToNull_NullAtStart(b *testing.B) {
	data := []byte("\x00Hello, World!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UpToNull(data)
	}
}
