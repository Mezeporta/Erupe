package mhfpacket

import (
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/common/byteframe"
)

// MsgSysLogout represents the MSG_SYS_LOGOUT
type MsgSysLogout struct {
	LogoutType uint8 // Hardcoded 1 in binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLogout) Opcode() network.PacketID {
	return network.MSG_SYS_LOGOUT
}

// Parse parses the packet from binary
func (m *MsgSysLogout) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.LogoutType = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLogout) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint8(m.LogoutType)
	return nil
}
