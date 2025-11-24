package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve1AE represents the MSG_SYS_reserve1AE
type MsgSysReserve1AE struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve1AE) Opcode() network.PacketID {
	return network.MSG_SYS_reserve1AE
}

// Parse parses the packet from binary
func (m *MsgSysReserve1AE) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve1AE) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
