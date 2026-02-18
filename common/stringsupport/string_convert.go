package stringsupport

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// UTF8ToSJIS encodes a UTF-8 string to Shift-JIS bytes, silently dropping any
// runes that cannot be represented in Shift-JIS.
func UTF8ToSJIS(x string) []byte {
	e := japanese.ShiftJIS.NewEncoder()
	xt, _, err := transform.String(e, x)
	if err != nil {
		// Filter out runes that can't be encoded to Shift-JIS instead of
		// crashing the server (see PR #116).
		var filtered []rune
		for _, r := range x {
			if _, _, err := transform.String(japanese.ShiftJIS.NewEncoder(), string(r)); err == nil {
				filtered = append(filtered, r)
			}
		}
		xt, _, _ = transform.String(japanese.ShiftJIS.NewEncoder(), string(filtered))
	}
	return []byte(xt)
}

// SJISToUTF8 decodes Shift-JIS bytes to a UTF-8 string.
func SJISToUTF8(b []byte) string {
	d := japanese.ShiftJIS.NewDecoder()
	result, err := io.ReadAll(transform.NewReader(bytes.NewReader(b), d))
	if err != nil {
		panic(err)
	}
	return string(result)
}

// ToNGWord converts a UTF-8 string into a slice of uint16 values in the
// Shift-JIS byte-swapped format used by the MHF NG-word (chat filter) system.
func ToNGWord(x string) []uint16 {
	var w []uint16
	for _, r := range x {
		if r > 0xFF {
			t := UTF8ToSJIS(string(r))
			if len(t) > 1 {
				w = append(w, uint16(t[1])<<8|uint16(t[0]))
			} else if len(t) == 1 {
				w = append(w, uint16(t[0]))
			}
			// Skip runes that produced no SJIS output (unsupported characters)
		} else {
			w = append(w, uint16(r))
		}
	}
	return w
}

// PaddedString returns a fixed-width null-terminated byte slice of the given
// size. If t is true the string is first encoded to Shift-JIS.
func PaddedString(x string, size uint, t bool) []byte {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			return make([]byte, size)
		}
		x = xt
	}
	out := make([]byte, size)
	copy(out, x)
	out[len(out)-1] = 0
	return out
}

// CSVAdd appends v to the comma-separated integer list if not already present.
func CSVAdd(csv string, v int) string {
	if len(csv) == 0 {
		return strconv.Itoa(v)
	}
	if CSVContains(csv, v) {
		return csv
	} else {
		return csv + "," + strconv.Itoa(v)
	}
}

// CSVRemove removes v from the comma-separated integer list.
func CSVRemove(csv string, v int) string {
	s := strings.Split(csv, ",")
	for i, e := range s {
		if e == strconv.Itoa(v) {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
		}
	}
	return strings.Join(s, ",")
}

// CSVContains reports whether v is present in the comma-separated integer list.
func CSVContains(csv string, v int) bool {
	s := strings.Split(csv, ",")
	for i := 0; i < len(s); i++ {
		j, _ := strconv.ParseInt(s[i], 10, 32)
		if int(j) == v {
			return true
		}
	}
	return false
}

// CSVLength returns the number of elements in the comma-separated list.
func CSVLength(csv string) int {
	if csv == "" {
		return 0
	}
	s := strings.Split(csv, ",")
	return len(s)
}

// CSVElems parses the comma-separated integer list into an int slice.
func CSVElems(csv string) []int {
	var r []int
	if csv == "" {
		return r
	}
	s := strings.Split(csv, ",")
	for i := 0; i < len(s); i++ {
		j, _ := strconv.ParseInt(s[i], 10, 32)
		r = append(r, int(j))
	}
	return r
}

// CSVGetIndex returns the integer at position i in the comma-separated list,
// or 0 if i is out of range.
func CSVGetIndex(csv string, i int) int {
	s := CSVElems(csv)
	if i < len(s) {
		return s[i]
	}
	return 0
}

// CSVSetIndex replaces the integer at position i in the comma-separated list
// with v. If i is out of range the list is returned unchanged.
func CSVSetIndex(csv string, i int, v int) string {
	s := CSVElems(csv)
	if i < len(s) {
		s[i] = v
	}
	var r []string
	for j := 0; j < len(s); j++ {
		r = append(r, fmt.Sprintf(`%d`, s[j]))
	}
	return strings.Join(r, ",")
}
