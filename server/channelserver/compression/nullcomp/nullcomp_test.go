package nullcomp

import (
	"bytes"
	"testing"
)

func TestDecompress_WithValidHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "empty data after header",
			input:    []byte("cmp\x2020110113\x20\x20\x20\x00"),
			expected: []byte{},
		},
		{
			name:     "single regular byte",
			input:    []byte("cmp\x2020110113\x20\x20\x20\x00\x42"),
			expected: []byte{0x42},
		},
		{
			name:     "multiple regular bytes",
			input:    []byte("cmp\x2020110113\x20\x20\x20\x00\x48\x65\x6c\x6c\x6f"),
			expected: []byte("Hello"),
		},
		{
			name:     "single null byte compression",
			input:    []byte("cmp\x2020110113\x20\x20\x20\x00\x00\x05"),
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "multiple null bytes with max count",
			input:    []byte("cmp\x2020110113\x20\x20\x20\x00\x00\xFF"),
			expected: make([]byte, 255),
		},
		{
			name: "mixed regular and null bytes",
			input: append(
				[]byte("cmp\x2020110113\x20\x20\x20\x00\x48\x65\x6c\x6c\x6f"),
				[]byte{0x00, 0x03, 0x57, 0x6f, 0x72, 0x6c, 0x64}...,
			),
			expected: []byte("Hello\x00\x00\x00World"),
		},
		{
			name: "multiple null compressions",
			input: append(
				[]byte("cmp\x2020110113\x20\x20\x20\x00"),
				[]byte{0x41, 0x00, 0x02, 0x42, 0x00, 0x03, 0x43}...,
			),
			expected: []byte{0x41, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00, 0x43},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decompress(tt.input)
			if err != nil {
				t.Fatalf("Decompress() error = %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Decompress() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecompress_WithoutHeader(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectError   bool
		expectOriginal bool // Expect original data returned
	}{
		{
			name: "plain data without header (16+ bytes)",
			// Data must be at least 16 bytes to read header
			input:          []byte("Hello, World!!!!"), // Exactly 16 bytes
			expectError:    false,
			expectOriginal: true,
		},
		{
			name: "binary data without header (16+ bytes)",
			input: []byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
			},
			expectError:    false,
			expectOriginal: true,
		},
		{
			name: "data shorter than 16 bytes",
			// When data is shorter than 16 bytes, Read returns what it can with err=nil
			// Then n != len(header) returns nil, nil (not an error)
			input:          []byte("Short"),
			expectError:    false,
			expectOriginal: false, // Returns empty slice
		},
		{
			name:        "empty data",
			input:       []byte{},
			expectError: true, // EOF on first read
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decompress(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("Decompress() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Decompress() error = %v", err)
			}
			if tt.expectOriginal && !bytes.Equal(result, tt.input) {
				t.Errorf("Decompress() = %v, want %v (original data)", result, tt.input)
			}
		})
	}
}

func TestDecompress_InvalidData(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expectErr bool
	}{
		{
			name: "incomplete header",
			// Less than 16 bytes: Read returns what it can (no error),
			// but n != len(header) returns nil, nil
			input:     []byte("cmp\x20201"),
			expectErr: false,
		},
		{
			name:      "header with missing null count",
			input:     []byte("cmp\x2020110113\x20\x20\x20\x00\x00"),
			expectErr: false, // Valid header, EOF during decompression is handled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decompress(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Decompress() expected error but got none, result = %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Decompress() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestCompress_BasicData(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty data",
			input: []byte{},
		},
		{
			name:  "regular bytes without nulls",
			input: []byte("Hello, World!"),
		},
		{
			name:  "single null byte",
			input: []byte{0x00},
		},
		{
			name:  "multiple consecutive nulls",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "mixed data with nulls",
			input: []byte("Hello\x00\x00\x00World"),
		},
		{
			name:  "data starting with nulls",
			input: []byte{0x00, 0x00, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:  "data ending with nulls",
			input: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x00, 0x00},
		},
		{
			name:  "alternating nulls and bytes",
			input: []byte{0x41, 0x00, 0x42, 0x00, 0x43},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := Compress(tt.input)
			if err != nil {
				t.Fatalf("Compress() error = %v", err)
			}

			// Verify it has the correct header
			expectedHeader := []byte("cmp\x2020110113\x20\x20\x20\x00")
			if !bytes.HasPrefix(compressed, expectedHeader) {
				t.Errorf("Compress() result doesn't have correct header")
			}

			// Verify round-trip
			decompressed, err := Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress() error = %v", err)
			}
			if !bytes.Equal(decompressed, tt.input) {
				t.Errorf("Round-trip failed: got %v, want %v", decompressed, tt.input)
			}
		})
	}
}

func TestCompress_LargeNullSequences(t *testing.T) {
	tests := []struct {
		name      string
		nullCount int
	}{
		{
			name:      "exactly 255 nulls",
			nullCount: 255,
		},
		{
			name:      "256 nulls (overflow case)",
			nullCount: 256,
		},
		{
			name:      "500 nulls",
			nullCount: 500,
		},
		{
			name:      "1000 nulls",
			nullCount: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := make([]byte, tt.nullCount)
			compressed, err := Compress(input)
			if err != nil {
				t.Fatalf("Compress() error = %v", err)
			}

			// Verify round-trip
			decompressed, err := Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress() error = %v", err)
			}
			if !bytes.Equal(decompressed, input) {
				t.Errorf("Round-trip failed: got len=%d, want len=%d", len(decompressed), len(input))
			}
		})
	}
}

func TestCompressDecompress_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "binary data with mixed nulls",
			data: []byte{0x01, 0x02, 0x00, 0x00, 0x03, 0x04, 0x00, 0x05},
		},
		{
			name: "large binary data",
			data: append(append([]byte{0xFF, 0xFE, 0xFD}, make([]byte, 300)...), []byte{0x01, 0x02, 0x03}...),
		},
		{
			name: "text with embedded nulls",
			data: []byte("Test\x00\x00Data\x00\x00\x00End"),
		},
		{
			name: "all non-null bytes",
			data: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
		},
		{
			name: "only null bytes",
			data: make([]byte, 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress
			compressed, err := Compress(tt.data)
			if err != nil {
				t.Fatalf("Compress() error = %v", err)
			}

			// Decompress
			decompressed, err := Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress() error = %v", err)
			}

			// Verify
			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("Round-trip failed:\ngot  = %v\nwant = %v", decompressed, tt.data)
			}
		})
	}
}

func TestCompress_CompressionEfficiency(t *testing.T) {
	// Test that data with many nulls is actually compressed
	input := make([]byte, 1000)
	compressed, err := Compress(input)
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	// The compressed size should be much smaller than the original
	// With 1000 nulls, we expect roughly 16 (header) + 4*3 (for 255*3 + 235) bytes
	if len(compressed) >= len(input) {
		t.Errorf("Compression failed: compressed size (%d) >= input size (%d)", len(compressed), len(input))
	}
}

func TestDecompress_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "only header",
			input: []byte("cmp\x2020110113\x20\x20\x20\x00"),
		},
		{
			name:  "null with count 1",
			input: []byte("cmp\x2020110113\x20\x20\x20\x00\x00\x01"),
		},
		{
			name:  "multiple sections of compressed nulls",
			input: append([]byte("cmp\x2020110113\x20\x20\x20\x00"), []byte{0x00, 0x10, 0x41, 0x00, 0x20, 0x42}...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decompress(tt.input)
			if err != nil {
				t.Fatalf("Decompress() unexpected error = %v", err)
			}
			// Just ensure it doesn't crash and returns something
			_ = result
		})
	}
}

func BenchmarkCompress(b *testing.B) {
	data := make([]byte, 10000)
	// Fill with some pattern (half nulls, half data)
	for i := 0; i < len(data); i++ {
		if i%2 == 0 {
			data[i] = 0x00
		} else {
			data[i] = byte(i % 256)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Compress(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompress(b *testing.B) {
	data := make([]byte, 10000)
	for i := 0; i < len(data); i++ {
		if i%2 == 0 {
			data[i] = 0x00
		} else {
			data[i] = byte(i % 256)
		}
	}

	compressed, err := Compress(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}
