package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve1A9 represents the MSG_SYS_reserve1A9
type MsgSysReserve1A9 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve1A9) Opcode() network.PacketID {
	return network.MSG_SYS_reserve1A9
}

// Parse parses the packet from binary
func (m *MsgSysReserve1A9) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve1A9) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
