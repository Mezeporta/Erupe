package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve5F represents the MSG_SYS_reserve5F
type MsgSysReserve5F struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve5F) Opcode() network.PacketID {
	return network.MSG_SYS_reserve5F
}

// Parse parses the packet from binary
func (m *MsgSysReserve5F) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve5F) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
