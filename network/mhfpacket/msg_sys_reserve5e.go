package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve5E represents the MSG_SYS_reserve5E
type MsgSysReserve5E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve5E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve5E
}

// Parse parses the packet from binary
func (m *MsgSysReserve5E) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve5E) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
