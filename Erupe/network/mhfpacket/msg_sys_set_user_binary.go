package mhfpacket

import (
	"errors"

	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysSetUserBinary represents the MSG_SYS_SET_USER_BINARY
type MsgSysSetUserBinary struct {
	BinaryType     uint8
	DataSize       uint16
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetUserBinary) Opcode() network.PacketID {
	return network.MSG_SYS_SET_USER_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysSetUserBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.BinaryType = bf.ReadUint8()
	m.DataSize = bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetUserBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
