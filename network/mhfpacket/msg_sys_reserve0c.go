package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve0C represents the MSG_SYS_reserve0C
type MsgSysReserve0C struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve0C) Opcode() network.PacketID {
	return network.MSG_SYS_reserve0C
}

// Parse parses the packet from binary
func (m *MsgSysReserve0C) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve0C) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
