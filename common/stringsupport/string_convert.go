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

func UTF8ToSJIS(x string) []byte {
	e := japanese.ShiftJIS.NewEncoder()
	xt, _, err := transform.String(e, x)
	if err != nil {
		panic(err)
	}
	return []byte(xt)
}

func SJISToUTF8(b []byte) string {
	d := japanese.ShiftJIS.NewDecoder()
	result, err := io.ReadAll(transform.NewReader(bytes.NewReader(b), d))
	if err != nil {
		panic(err)
	}
	return string(result)
}

func ToNGWord(x string) []uint16 {
	var w []uint16
	for _, r := range []rune(x) {
		if r > 0xFF {
			t := UTF8ToSJIS(string(r))
			if len(t) > 1 {
				w = append(w, uint16(t[1])<<8|uint16(t[0]))
			} else {
				w = append(w, uint16(t[0]))
			}
		} else {
			w = append(w, uint16(r))
		}
	}
	return w
}

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

func CSVLength(csv string) int {
	if csv == "" {
		return 0
	}
	s := strings.Split(csv, ",")
	return len(s)
}

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

func CSVGetIndex(csv string, i int) int {
	s := CSVElems(csv)
	if i < len(s) {
		return s[i]
	}
	return 0
}

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
