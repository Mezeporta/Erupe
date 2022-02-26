package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLogout represents the MSG_SYS_LOGOUT
type MsgSysLogout struct {
	Unk0 uint8 // Hardcoded 1 in binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLogout) Opcode() network.PacketID {
	return network.MSG_SYS_LOGOUT
}

// Parse parses the packet from binary
func (m *MsgSysLogout) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.Unk0 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLogout) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.Unk0 = bf.ReadUint8()
	return nil
}
