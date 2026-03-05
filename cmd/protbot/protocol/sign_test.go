package protocol

import (
	"testing"

	"erupe-ce/common/byteframe"
)

// buildSignResponse constructs a binary sign server response for testing.
// Format mirrors Erupe server/signserver/dsgn_resp.go:makeSignResponse.
func buildSignResponse(resultCode uint8, tokenID uint32, tokenString [16]byte, timestamp uint32, patchURLs []string, entranceAddr string, chars []testCharEntry) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(resultCode)
	bf.WriteUint8(uint8(len(patchURLs))) // patchCount
	bf.WriteUint8(1)                     // entranceCount (always 1 in tests)
	bf.WriteUint8(uint8(len(chars)))     // charCount
	bf.WriteUint32(tokenID)
	bf.WriteBytes(tokenString[:])
	bf.WriteUint32(timestamp)

	// Patch server URLs (pascal strings with uint8 length prefix)
	for _, url := range patchURLs {
		bf.WriteUint8(uint8(len(url)))
		bf.WriteBytes([]byte(url))
	}

	// Entrance server address (pascal string with null terminator included in length)
	bf.WriteUint8(uint8(len(entranceAddr) + 1))
	bf.WriteBytes([]byte(entranceAddr))
	bf.WriteUint8(0) // null terminator

	// Character entries
	for _, c := range chars {
		bf.WriteUint32(c.charID)
		bf.WriteUint16(c.hr)
		bf.WriteUint16(c.weaponType)
		bf.WriteUint32(c.lastLogin)
		bf.WriteUint8(c.isFemale)
		bf.WriteUint8(c.isNewChar)
		bf.WriteUint8(c.oldGR)
		bf.WriteUint8(c.useU16GR)
		// Name: 16 bytes padded
		name := make([]byte, 16)
		copy(name, []byte(c.name))
		bf.WriteBytes(name)
		// Desc: 32 bytes padded
		desc := make([]byte, 32)
		copy(desc, []byte(c.desc))
		bf.WriteBytes(desc)
		bf.WriteUint16(c.gr)
		bf.WriteUint8(c.unk1)
		bf.WriteUint8(c.unk2)
	}

	return bf.Data()
}

type testCharEntry struct {
	charID     uint32
	hr         uint16
	weaponType uint16
	lastLogin  uint32
	isFemale   uint8
	isNewChar  uint8
	oldGR      uint8
	useU16GR   uint8
	name       string
	desc       string
	gr         uint16
	unk1       uint8
	unk2       uint8
}

func TestParseSignResponse_Success(t *testing.T) {
	tokenID := uint32(12345)
	var tokenString [16]byte
	copy(tokenString[:], []byte("ABCDEFGHIJKLMNOP"))
	timestamp := uint32(1700000000)
	patchURLs := []string{"http://patch1.example.com", "http://patch2.example.com"}
	entranceAddr := "192.168.1.1:53310"
	chars := []testCharEntry{
		{
			charID:     100,
			hr:         999,
			weaponType: 3,
			lastLogin:  1699999999,
			isFemale:   1,
			isNewChar:  0,
			oldGR:      50,
			useU16GR:   1,
			name:       "Hunter",
			desc:       "A brave hunter",
			gr:         200,
			unk1:       0,
			unk2:       0,
		},
	}

	data := buildSignResponse(1, tokenID, tokenString, timestamp, patchURLs, entranceAddr, chars)

	result, err := parseSignResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TokenID != tokenID {
		t.Errorf("TokenID: got %d, want %d", result.TokenID, tokenID)
	}
	if result.TokenString != string(tokenString[:]) {
		t.Errorf("TokenString: got %q, want %q", result.TokenString, string(tokenString[:]))
	}
	if result.Timestamp != timestamp {
		t.Errorf("Timestamp: got %d, want %d", result.Timestamp, timestamp)
	}
	if result.EntranceAddr != entranceAddr {
		t.Errorf("EntranceAddr: got %q, want %q", result.EntranceAddr, entranceAddr)
	}
	if len(result.CharIDs) != 1 {
		t.Fatalf("CharIDs length: got %d, want 1", len(result.CharIDs))
	}
	if result.CharIDs[0] != 100 {
		t.Errorf("CharIDs[0]: got %d, want 100", result.CharIDs[0])
	}
}

func TestParseSignResponse_MultipleCharacters(t *testing.T) {
	var tokenString [16]byte
	copy(tokenString[:], []byte("0123456789ABCDEF"))
	chars := []testCharEntry{
		{charID: 10, name: "Char1"},
		{charID: 20, name: "Char2"},
		{charID: 30, name: "Char3"},
	}

	data := buildSignResponse(1, 42, tokenString, 0, nil, "127.0.0.1:53310", chars)

	result, err := parseSignResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.CharIDs) != 3 {
		t.Fatalf("CharIDs length: got %d, want 3", len(result.CharIDs))
	}
	expectedIDs := []uint32{10, 20, 30}
	for i, want := range expectedIDs {
		if result.CharIDs[i] != want {
			t.Errorf("CharIDs[%d]: got %d, want %d", i, result.CharIDs[i], want)
		}
	}
}

func TestParseSignResponse_NoCharacters(t *testing.T) {
	var tokenString [16]byte
	data := buildSignResponse(1, 1, tokenString, 0, nil, "127.0.0.1:53310", nil)

	result, err := parseSignResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.CharIDs) != 0 {
		t.Errorf("CharIDs length: got %d, want 0", len(result.CharIDs))
	}
}

func TestParseSignResponse_FailCode(t *testing.T) {
	// resultCode=0 means failure; the rest of the data is irrelevant
	// but we still need the 3 count bytes for the parser to read before checking
	data := []byte{0} // resultCode = 0

	_, err := parseSignResponse(data)
	if err == nil {
		t.Fatal("expected error for failure result code, got nil")
	}
}

func TestParseSignResponse_FailCode5(t *testing.T) {
	data := []byte{5} // resultCode = 5 (some other failure code)

	_, err := parseSignResponse(data)
	if err == nil {
		t.Fatal("expected error for result code 5, got nil")
	}
}

func TestParseSignResponse_Empty(t *testing.T) {
	_, err := parseSignResponse(nil)
	if err == nil {
		t.Fatal("expected error for nil data, got nil")
	}

	_, err = parseSignResponse([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
}
