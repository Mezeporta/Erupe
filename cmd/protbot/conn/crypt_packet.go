// Package conn provides MHF encrypted connection primitives.
//
// This is adapted from Erupe's network/crypt_packet.go to avoid importing
// erupe-ce/config (whose init() calls os.Exit without a config file).
package conn

import (
	"bytes"
	"encoding/binary"
)

const CryptPacketHeaderLength = 14

// CryptPacketHeader represents the parsed information of an encrypted packet header.
type CryptPacketHeader struct {
	Pf0                     byte
	KeyRotDelta             byte
	PacketNum               uint16
	DataSize                uint16
	PrevPacketCombinedCheck uint16
	Check0                  uint16
	Check1                  uint16
	Check2                  uint16
}

// NewCryptPacketHeader parses raw bytes into a CryptPacketHeader.
func NewCryptPacketHeader(data []byte) (*CryptPacketHeader, error) {
	var c CryptPacketHeader
	r := bytes.NewReader(data)

	if err := binary.Read(r, binary.BigEndian, &c.Pf0); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.KeyRotDelta); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.PacketNum); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.DataSize); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.PrevPacketCombinedCheck); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.Check0); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.Check1); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.Check2); err != nil {
		return nil, err
	}

	return &c, nil
}

// Encode encodes the CryptPacketHeader into raw bytes.
func (c *CryptPacketHeader) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	data := []interface{}{
		c.Pf0,
		c.KeyRotDelta,
		c.PacketNum,
		c.DataSize,
		c.PrevPacketCombinedCheck,
		c.Check0,
		c.Check1,
		c.Check2,
	}
	for _, v := range data {
		if err := binary.Write(buf, binary.BigEndian, v); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
