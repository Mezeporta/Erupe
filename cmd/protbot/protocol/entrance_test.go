package protocol

import (
	"testing"

	"erupe-ce/cmd/protbot/conn"
	"erupe-ce/common/byteframe"
)

// encryptBin8 encrypts plaintext using the Bin8 algorithm.
// Since Bin8 is a symmetric XOR cipher, DecryptBin8(plaintext, key) produces ciphertext.
func encryptBin8(plaintext []byte, key byte) []byte {
	return conn.DecryptBin8(plaintext, key)
}

// buildEntranceResponse constructs a valid Bin8-encrypted entrance server response.
// Format: [key byte] [encrypted: "SV2" + uint16 entryCount + uint16 dataLen + uint32 checksum + serverData]
func buildEntranceResponse(key byte, respType string, entries []testServerEntry) []byte {
	// Build server data blob first (to compute checksum and length).
	serverData := buildServerData(entries)

	// Build the plaintext (before encryption).
	bf := byteframe.NewByteFrame()
	bf.WriteBytes([]byte(respType)) // "SV2" or "SVR"
	bf.WriteUint16(uint16(len(entries)))
	bf.WriteUint16(uint16(len(serverData)))
	if len(serverData) > 0 {
		bf.WriteUint32(conn.CalcSum32(serverData))
		bf.WriteBytes(serverData)
	}

	plaintext := bf.Data()
	encrypted := encryptBin8(plaintext, key)

	// Final response: key byte + encrypted data.
	result := make([]byte, 1+len(encrypted))
	result[0] = key
	copy(result[1:], encrypted)
	return result
}

type testServerEntry struct {
	ip           [4]byte // big-endian IP bytes (reversed for MHF format)
	name         string
	channelCount uint16
	channelPorts []uint16
}

// buildServerData constructs the binary server entry data blob.
// Format mirrors Erupe server/entranceserver/make_resp.go:encodeServerInfo.
func buildServerData(entries []testServerEntry) []byte {
	if len(entries) == 0 {
		return nil
	}

	bf := byteframe.NewByteFrame()
	for _, e := range entries {
		// IP bytes (stored reversed in the protocol — client reads and reverses)
		bf.WriteBytes(e.ip[:])

		bf.WriteUint16(0x0010) // serverIdx | 16
		bf.WriteUint16(0)      // zero
		bf.WriteUint16(e.channelCount)
		bf.WriteUint8(0) // Type
		bf.WriteUint8(0) // Season

		// G1+ recommended
		bf.WriteUint8(0)

		// G51+ (ZZ): skip byte + 65-byte name
		bf.WriteUint8(0)
		nameBytes := make([]byte, 65)
		copy(nameBytes, []byte(e.name))
		bf.WriteBytes(nameBytes)

		// GG+: AllowedClientFlags
		bf.WriteUint32(0)

		// Channel entries (28 bytes each)
		for j := uint16(0); j < e.channelCount; j++ {
			port := uint16(54001)
			if j < uint16(len(e.channelPorts)) {
				port = e.channelPorts[j]
			}
			bf.WriteUint16(port)            // port
			bf.WriteUint16(0x0010)          // channelIdx | 16
			bf.WriteUint16(100)             // maxPlayers
			bf.WriteUint16(5)               // currentPlayers
			bf.WriteBytes(make([]byte, 18)) // remaining fields (9 x uint16)
			bf.WriteUint16(12345)           // sentinel
		}
	}
	return bf.Data()
}

func TestParseEntranceResponse_ValidSV2(t *testing.T) {
	entries := []testServerEntry{
		{
			ip:           [4]byte{1, 0, 0, 127}, // 127.0.0.1 reversed
			name:         "TestServer",
			channelCount: 2,
			channelPorts: []uint16{54001, 54002},
		},
	}

	data := buildEntranceResponse(0x42, "SV2", entries)

	result, err := parseEntranceResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("entry count: got %d, want 2", len(result))
	}

	if result[0].Port != 54001 {
		t.Errorf("entry[0].Port: got %d, want 54001", result[0].Port)
	}
	if result[1].Port != 54002 {
		t.Errorf("entry[1].Port: got %d, want 54002", result[1].Port)
	}
	if result[0].Name != "TestServer ch1" {
		t.Errorf("entry[0].Name: got %q, want %q", result[0].Name, "TestServer ch1")
	}
	if result[1].Name != "TestServer ch2" {
		t.Errorf("entry[1].Name: got %q, want %q", result[1].Name, "TestServer ch2")
	}
}

func TestParseEntranceResponse_ValidSVR(t *testing.T) {
	entries := []testServerEntry{
		{
			ip:           [4]byte{1, 0, 0, 127},
			name:         "World1",
			channelCount: 1,
			channelPorts: []uint16{54001},
		},
	}

	data := buildEntranceResponse(0x10, "SVR", entries)

	result, err := parseEntranceResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("entry count: got %d, want 1", len(result))
	}
	if result[0].Port != 54001 {
		t.Errorf("entry[0].Port: got %d, want 54001", result[0].Port)
	}
}

func TestParseEntranceResponse_InvalidType(t *testing.T) {
	// Build a response with an invalid type string "BAD" instead of "SV2"/"SVR".
	bf := byteframe.NewByteFrame()
	bf.WriteBytes([]byte("BAD"))
	bf.WriteUint16(0) // entryCount
	bf.WriteUint16(0) // dataLen

	plaintext := bf.Data()
	key := byte(0x55)
	encrypted := encryptBin8(plaintext, key)

	data := make([]byte, 1+len(encrypted))
	data[0] = key
	copy(data[1:], encrypted)

	_, err := parseEntranceResponse(data)
	if err == nil {
		t.Fatal("expected error for invalid response type, got nil")
	}
}

func TestParseEntranceResponse_EmptyData(t *testing.T) {
	_, err := parseEntranceResponse(nil)
	if err == nil {
		t.Fatal("expected error for nil data, got nil")
	}

	_, err = parseEntranceResponse([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}

	_, err = parseEntranceResponse([]byte{0x42}) // only key, no encrypted data
	if err == nil {
		t.Fatal("expected error for single-byte data, got nil")
	}
}

func TestParseEntranceResponse_ZeroEntries(t *testing.T) {
	// Valid response with zero entries and zero data length.
	data := buildEntranceResponse(0x30, "SV2", nil)

	result, err := parseEntranceResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for zero entries, got %v", result)
	}
}

func TestParseServerEntries_MultipleServers(t *testing.T) {
	entries := []testServerEntry{
		{
			ip:           [4]byte{100, 1, 168, 192}, // 192.168.1.100 reversed
			name:         "Server1",
			channelCount: 1,
			channelPorts: []uint16{54001},
		},
		{
			ip:           [4]byte{200, 1, 168, 192}, // 192.168.1.200 reversed
			name:         "Server2",
			channelCount: 1,
			channelPorts: []uint16{54010},
		},
	}

	serverData := buildServerData(entries)

	result, err := parseServerEntries(serverData, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("entry count: got %d, want 2", len(result))
	}

	if result[0].Port != 54001 {
		t.Errorf("entry[0].Port: got %d, want 54001", result[0].Port)
	}
	if result[0].Name != "Server1 ch1" {
		t.Errorf("entry[0].Name: got %q, want %q", result[0].Name, "Server1 ch1")
	}
	if result[1].Port != 54010 {
		t.Errorf("entry[1].Port: got %d, want 54010", result[1].Port)
	}
	if result[1].Name != "Server2 ch1" {
		t.Errorf("entry[1].Name: got %q, want %q", result[1].Name, "Server2 ch1")
	}
}
