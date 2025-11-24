package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserve19B represents the MSG_SYS_reserve19B
type MsgSysReserve19B struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve19B) Opcode() network.PacketID {
	return network.MSG_SYS_reserve19B
}

// Parse parses the packet from binary
func (m *MsgSysReserve19B) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve19B) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
