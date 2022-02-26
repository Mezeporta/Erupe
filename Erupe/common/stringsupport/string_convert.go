package stringsupport

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// StringConverter is a small helper for encoding/decoding strings.
type StringConverter struct {
	Encoding encoding.Encoding
}

// Decode decodes the given bytes as the set encoding.
func (sc *StringConverter) Decode(data []byte) (string, error) {
	decoded, err := ioutil.ReadAll(transform.NewReader(bytes.NewBuffer(data), sc.Encoding.NewDecoder()))

	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// MustDecode decodes the given bytes as the set encoding. Panics on decode failure.
func (sc *StringConverter) MustDecode(data []byte) string {
	decoded, err := sc.Decode(data)
	if err != nil {
		panic(err)
	}

	return decoded
}

// Encode encodes the given string as the set encoding.
func (sc *StringConverter) Encode(data string) ([]byte, error) {
	encoded, err := ioutil.ReadAll(transform.NewReader(bytes.NewBuffer([]byte(data)), sc.Encoding.NewEncoder()))

	if err != nil {
		return nil, err
	}

	return encoded, nil
}

// MustEncode encodes the given string as the set encoding. Panics on encode failure.
func (sc *StringConverter) MustEncode(data string) []byte {
	encoded, err := sc.Encode(data)
	if err != nil {
		panic(err)
	}

	return encoded
}

/*
func MustConvertShiftJISToUTF8(text string) string {
	result, err := ConvertShiftJISToUTF8(text)
	if err != nil {
		panic(err)
	}
	return result
}
func MustConvertUTF8ToShiftJIS(text string) string {
	result, err := ConvertUTF8ToShiftJIS(text)
	if err != nil {
		panic(err)
	}
	return result
}
func ConvertShiftJISToUTF8(text string) (string, error) {
	r := bytes.NewBuffer([]byte(text))
	decoded, err := ioutil.ReadAll(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
*/

// ConvertUTF8ToShiftJIS converts a UTF8 string to a Shift-JIS []byte.
func ConvertUTF8ToShiftJIS(text string) ([]byte, error) {
	r := bytes.NewBuffer([]byte(text))
	encoded, err := ioutil.ReadAll(transform.NewReader(r, japanese.ShiftJIS.NewEncoder()))
	if err != nil {
		return nil, err
	}

	return encoded, nil
}
