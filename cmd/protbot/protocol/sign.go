package protocol

import (
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"

	"erupe-ce/cmd/protbot/conn"
)

// SignResult holds the parsed response from a successful DSGN sign-in.
type SignResult struct {
	TokenID     uint32
	TokenString string // 16 raw bytes as string
	Timestamp   uint32
	EntranceAddr string
	CharIDs     []uint32
}

// DoSign connects to the sign server and performs a DSGN login.
// Reference: Erupe server/signserver/session.go (handleDSGN) and dsgn_resp.go (makeSignResponse).
func DoSign(addr, username, password string) (*SignResult, error) {
	c, err := conn.DialWithInit(addr)
	if err != nil {
		return nil, fmt.Errorf("sign connect: %w", err)
	}
	defer c.Close()

	// Build DSGN request: "DSGN:041" + \x00 + SJIS(user) + \x00 + SJIS(pass) + \x00 + \x00
	// The server reads: null-terminated request type, null-terminated user, null-terminated pass, null-terminated unk.
	// The request type has a 3-char version suffix (e.g. "041" for ZZ client mode 41) that the server strips.
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:041")) // reqType with version suffix (server strips last 3 chars to get "DSGN:")
	bf.WriteNullTerminatedBytes(stringsupport.UTF8ToSJIS(username))
	bf.WriteNullTerminatedBytes(stringsupport.UTF8ToSJIS(password))
	bf.WriteUint8(0) // Unk null-terminated empty string

	if err := c.SendPacket(bf.Data()); err != nil {
		return nil, fmt.Errorf("sign send: %w", err)
	}

	resp, err := c.ReadPacket()
	if err != nil {
		return nil, fmt.Errorf("sign recv: %w", err)
	}

	return parseSignResponse(resp)
}

// parseSignResponse parses the binary response from the sign server.
// Reference: Erupe server/signserver/dsgn_resp.go:makeSignResponse
func parseSignResponse(data []byte) (*SignResult, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("empty sign response")
	}

	rbf := byteframe.NewByteFrameFromBytes(data)

	resultCode := rbf.ReadUint8()
	if resultCode != 1 { // SIGN_SUCCESS = 1
		return nil, fmt.Errorf("sign failed with code %d", resultCode)
	}

	patchCount := rbf.ReadUint8()    // patch server count (usually 2)
	_ = rbf.ReadUint8()              // entrance server count (usually 1)
	charCount := rbf.ReadUint8()     // character count

	result := &SignResult{}
	result.TokenID = rbf.ReadUint32()
	result.TokenString = string(rbf.ReadBytes(16)) // 16 raw bytes
	result.Timestamp = rbf.ReadUint32()

	// Skip patch server URLs (pascal strings with uint8 length prefix)
	for i := uint8(0); i < patchCount; i++ {
		strLen := rbf.ReadUint8()
		_ = rbf.ReadBytes(uint(strLen))
	}

	// Read entrance server address (pascal string with uint8 length prefix)
	entranceLen := rbf.ReadUint8()
	result.EntranceAddr = string(rbf.ReadBytes(uint(entranceLen - 1)))
	_ = rbf.ReadUint8() // null terminator

	// Read character entries
	for i := uint8(0); i < charCount; i++ {
		charID := rbf.ReadUint32()
		result.CharIDs = append(result.CharIDs, charID)

		_ = rbf.ReadUint16() // HR
		_ = rbf.ReadUint16() // WeaponType
		_ = rbf.ReadUint32() // LastLogin
		_ = rbf.ReadUint8()  // IsFemale
		_ = rbf.ReadUint8()  // IsNewCharacter
		_ = rbf.ReadUint8()  // Old GR
		_ = rbf.ReadUint8()  // Use uint16 GR flag
		_ = rbf.ReadBytes(16) // Character name (padded)
		_ = rbf.ReadBytes(32) // Unk desc string (padded)
		// ZZ mode: additional fields
		_ = rbf.ReadUint16() // GR
		_ = rbf.ReadUint8()  // Unk
		_ = rbf.ReadUint8()  // Unk
	}

	return result, nil
}
