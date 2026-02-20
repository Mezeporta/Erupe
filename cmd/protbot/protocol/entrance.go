package protocol

import (
	"encoding/binary"
	"fmt"
	"net"

	"erupe-ce/common/byteframe"

	"erupe-ce/cmd/protbot/conn"
)

// ServerEntry represents a channel server from the entrance server response.
type ServerEntry struct {
	IP       string
	Port     uint16
	Name     string
}

// DoEntrance connects to the entrance server and retrieves the server list.
// Reference: Erupe server/entranceserver/entrance_server.go and make_resp.go.
func DoEntrance(addr string) ([]ServerEntry, error) {
	c, err := conn.DialWithInit(addr)
	if err != nil {
		return nil, fmt.Errorf("entrance connect: %w", err)
	}
	defer func() { _ = c.Close() }()

	// Send a minimal packet (the entrance server reads it, checks len > 5 for USR data).
	// An empty/short packet triggers only SV2 response.
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	if err := c.SendPacket(bf.Data()); err != nil {
		return nil, fmt.Errorf("entrance send: %w", err)
	}

	resp, err := c.ReadPacket()
	if err != nil {
		return nil, fmt.Errorf("entrance recv: %w", err)
	}

	return parseEntranceResponse(resp)
}

// parseEntranceResponse parses the Bin8-encrypted entrance server response.
// Reference: Erupe server/entranceserver/make_resp.go (makeHeader, makeSv2Resp)
func parseEntranceResponse(data []byte) ([]ServerEntry, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("entrance response too short")
	}

	// First byte is the Bin8 encryption key.
	key := data[0]
	decrypted := conn.DecryptBin8(data[1:], key)

	rbf := byteframe.NewByteFrameFromBytes(decrypted)

	// Read response type header: "SV2" or "SVR"
	respType := string(rbf.ReadBytes(3))
	if respType != "SV2" && respType != "SVR" {
		return nil, fmt.Errorf("unexpected entrance response type: %s", respType)
	}

	entryCount := rbf.ReadUint16()
	dataLen := rbf.ReadUint16()
	if dataLen == 0 {
		return nil, nil
	}
	expectedSum := rbf.ReadUint32()
	serverData := rbf.ReadBytes(uint(dataLen))

	actualSum := conn.CalcSum32(serverData)
	if expectedSum != actualSum {
		return nil, fmt.Errorf("entrance checksum mismatch: expected %08X, got %08X", expectedSum, actualSum)
	}

	return parseServerEntries(serverData, entryCount)
}

// parseServerEntries parses the server info binary blob.
// Reference: Erupe server/entranceserver/make_resp.go (encodeServerInfo)
func parseServerEntries(data []byte, entryCount uint16) ([]ServerEntry, error) {
	bf := byteframe.NewByteFrameFromBytes(data)
	var entries []ServerEntry

	for i := uint16(0); i < entryCount; i++ {
		ipBytes := bf.ReadBytes(4)
		ip := net.IP([]byte{
			byte(ipBytes[3]), byte(ipBytes[2]),
			byte(ipBytes[1]), byte(ipBytes[0]),
		})

		_ = bf.ReadUint16() // serverIdx | 16
		_ = bf.ReadUint16() // 0
		channelCount := bf.ReadUint16()
		_ = bf.ReadUint8()  // Type
		_ = bf.ReadUint8()  // Season/rotation

		// G1+ recommended flag
		_ = bf.ReadUint8()

		// G51+ (ZZ): skip 1 byte, then read 65-byte padded name
		_ = bf.ReadUint8()
		nameBytes := bf.ReadBytes(65)

		// GG+: AllowedClientFlags
		_ = bf.ReadUint32()

		// Parse name (null-separated: name + description)
		name := ""
		for j := 0; j < len(nameBytes); j++ {
			if nameBytes[j] == 0 {
				break
			}
			name += string(nameBytes[j])
		}

		// Read channel entries (14 x uint16 = 28 bytes each)
		for j := uint16(0); j < channelCount; j++ {
			port := bf.ReadUint16()
			_ = bf.ReadUint16() // channelIdx | 16
			_ = bf.ReadUint16() // maxPlayers
			_ = bf.ReadUint16() // currentPlayers
			_ = bf.ReadBytes(18) // remaining channel fields (9 x uint16: 6 zeros + unk319 + unk254 + unk255)
			_ = bf.ReadUint16() // 12345

			serverIP := ip.String()
			// Convert 127.0.0.1 representation
			if binary.LittleEndian.Uint32(ipBytes) == 0x0100007F {
				serverIP = "127.0.0.1"
			}

			entries = append(entries, ServerEntry{
				IP:   serverIP,
				Port: port,
				Name: fmt.Sprintf("%s ch%d", name, j+1),
			})
		}
	}

	return entries, nil
}
