package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve5C represents the MSG_SYS_reserve5C
type MsgSysReserve5C struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve5C) Opcode() network.PacketID {
	return network.MSG_SYS_reserve5C
}

// Parse parses the packet from binary
func (m *MsgSysReserve5C) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve5C) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
