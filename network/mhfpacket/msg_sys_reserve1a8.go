package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve1A8 represents the MSG_SYS_reserve1A8
type MsgSysReserve1A8 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve1A8) Opcode() network.PacketID {
	return network.MSG_SYS_reserve1A8
}

// Parse parses the packet from binary
func (m *MsgSysReserve1A8) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve1A8) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
