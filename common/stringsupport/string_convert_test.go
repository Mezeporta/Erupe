package stringsupport

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding/japanese"
)

func TestStringConverterDecode(t *testing.T) {
	sc := &StringConverter{Encoding: japanese.ShiftJIS}

	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr bool
	}{
		{"empty", []byte{}, "", false},
		{"ascii", []byte("Hello"), "Hello", false},
		{"japanese hello", []byte{0x82, 0xb1, 0x82, 0xf1, 0x82, 0xc9, 0x82, 0xbf, 0x82, 0xcd}, "こんにちは", false},
		{"mixed", []byte{0x41, 0x42, 0x43}, "ABC", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sc.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringConverterEncode(t *testing.T) {
	sc := &StringConverter{Encoding: japanese.ShiftJIS}

	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"empty", "", []byte{}, false},
		{"ascii", "Hello", []byte("Hello"), false},
		{"japanese hello", "こんにちは", []byte{0x82, 0xb1, 0x82, 0xf1, 0x82, 0xc9, 0x82, 0xbf, 0x82, 0xcd}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sc.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringConverterMustDecode(t *testing.T) {
	sc := &StringConverter{Encoding: japanese.ShiftJIS}

	// Valid input should not panic
	result := sc.MustDecode([]byte("Hello"))
	if result != "Hello" {
		t.Errorf("MustDecode() = %q, want %q", result, "Hello")
	}
}

func TestStringConverterMustEncode(t *testing.T) {
	sc := &StringConverter{Encoding: japanese.ShiftJIS}

	// Valid input should not panic
	result := sc.MustEncode("Hello")
	if !bytes.Equal(result, []byte("Hello")) {
		t.Errorf("MustEncode() = %v, want %v", result, []byte("Hello"))
	}
}

func TestUTF8ToSJIS(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{"empty", "", []byte{}},
		{"ascii", "ABC", []byte("ABC")},
		{"japanese", "こんにちは", []byte{0x82, 0xb1, 0x82, 0xf1, 0x82, 0xc9, 0x82, 0xbf, 0x82, 0xcd}},
		{"mixed", "Hello世界", []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x90, 0xa2, 0x8a, 0x45}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UTF8ToSJIS(tt.input)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("UTF8ToSJIS(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSJISToUTF8(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"empty", []byte{}, ""},
		{"ascii", []byte("ABC"), "ABC"},
		{"japanese", []byte{0x82, 0xb1, 0x82, 0xf1, 0x82, 0xc9, 0x82, 0xbf, 0x82, 0xcd}, "こんにちは"},
		{"mixed", []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x90, 0xa2, 0x8a, 0x45}, "Hello世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SJISToUTF8(tt.input)
			if got != tt.want {
				t.Errorf("SJISToUTF8(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUTF8ToSJISRoundTrip(t *testing.T) {
	tests := []string{
		"Hello",
		"ABC123",
		"こんにちは",
		"テスト",
		"モンスターハンター",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			encoded := UTF8ToSJIS(input)
			decoded := SJISToUTF8(encoded)
			if decoded != input {
				t.Errorf("Round trip failed: %q -> %v -> %q", input, encoded, decoded)
			}
		})
	}
}

func TestPaddedString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		size      uint
		transform bool
		wantLen   int
		wantEnd   byte
	}{
		{"empty ascii", "", 10, false, 10, 0},
		{"short ascii", "Hi", 10, false, 10, 0},
		{"exact ascii", "1234567890", 10, false, 10, 0},
		{"empty sjis", "", 10, true, 10, 0},
		{"short sjis", "Hi", 10, true, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PaddedString(tt.input, tt.size, tt.transform)
			if len(got) != tt.wantLen {
				t.Errorf("PaddedString() len = %d, want %d", len(got), tt.wantLen)
			}
			if got[len(got)-1] != tt.wantEnd {
				t.Errorf("PaddedString() last byte = %d, want %d", got[len(got)-1], tt.wantEnd)
			}
		})
	}
}

func TestPaddedStringContent(t *testing.T) {
	// Verify the content is correctly placed at the beginning
	result := PaddedString("ABC", 10, false)

	if result[0] != 'A' || result[1] != 'B' || result[2] != 'C' {
		t.Errorf("PaddedString() content mismatch: got %v", result[:3])
	}

	// Rest should be zeros (except last which is forced to 0)
	for i := 3; i < 10; i++ {
		if result[i] != 0 {
			t.Errorf("PaddedString() byte at %d = %d, want 0", i, result[i])
		}
	}
}

func TestCSVAdd(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		v    int
		want string
	}{
		{"empty add", "", 5, "5"},
		{"add to existing", "1,2,3", 4, "1,2,3,4"},
		{"add duplicate", "1,2,3", 2, "1,2,3"},
		{"add to single", "1", 2, "1,2"},
		{"add zero", "", 0, "0"},
		{"add negative", "1,2", -5, "1,2,-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSVAdd(tt.csv, tt.v)
			if got != tt.want {
				t.Errorf("CSVAdd(%q, %d) = %q, want %q", tt.csv, tt.v, got, tt.want)
			}
		})
	}
}

func TestCSVRemove(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		v    int
		want string
	}{
		{"remove from middle", "1,2,3", 2, "1,3"},
		{"remove first", "1,2,3", 1, "3,2"},
		{"remove last", "1,2,3", 3, "1,2"},
		{"remove only", "5", 5, ""},
		{"remove nonexistent", "1,2,3", 99, "1,2,3"},
		{"remove from empty", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSVRemove(tt.csv, tt.v)
			if got != tt.want {
				t.Errorf("CSVRemove(%q, %d) = %q, want %q", tt.csv, tt.v, got, tt.want)
			}
		})
	}
}

func TestCSVContains(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		v    int
		want bool
	}{
		{"contains first", "1,2,3", 1, true},
		{"contains middle", "1,2,3", 2, true},
		{"contains last", "1,2,3", 3, true},
		{"not contains", "1,2,3", 99, false},
		{"empty csv", "", 5, false},
		{"single contains", "5", 5, true},
		{"single not contains", "5", 3, false},
		{"contains zero", "0,1,2", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSVContains(tt.csv, tt.v)
			if got != tt.want {
				t.Errorf("CSVContains(%q, %d) = %v, want %v", tt.csv, tt.v, got, tt.want)
			}
		})
	}
}

func TestCSVLength(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		want int
	}{
		{"empty", "", 0},
		{"single", "5", 1},
		{"two", "1,2", 2},
		{"three", "1,2,3", 3},
		{"many", "1,2,3,4,5,6,7,8,9,10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSVLength(tt.csv)
			if got != tt.want {
				t.Errorf("CSVLength(%q) = %d, want %d", tt.csv, got, tt.want)
			}
		})
	}
}

func TestCSVElems(t *testing.T) {
	tests := []struct {
		name string
		csv  string
		want []int
	}{
		{"empty", "", nil},
		{"single", "5", []int{5}},
		{"multiple", "1,2,3", []int{1, 2, 3}},
		{"with zero", "0,1,2", []int{0, 1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSVElems(tt.csv)
			if len(got) != len(tt.want) {
				t.Errorf("CSVElems(%q) len = %d, want %d", tt.csv, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CSVElems(%q)[%d] = %d, want %d", tt.csv, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCSVAddRemoveRoundTrip(t *testing.T) {
	csv := ""
	csv = CSVAdd(csv, 1)
	csv = CSVAdd(csv, 2)
	csv = CSVAdd(csv, 3)

	if !CSVContains(csv, 1) || !CSVContains(csv, 2) || !CSVContains(csv, 3) {
		t.Error("CSVAdd did not add all elements")
	}

	csv = CSVRemove(csv, 2)
	if CSVContains(csv, 2) {
		t.Error("CSVRemove did not remove element")
	}
	if CSVLength(csv) != 2 {
		t.Errorf("CSVLength after remove = %d, want 2", CSVLength(csv))
	}
}
