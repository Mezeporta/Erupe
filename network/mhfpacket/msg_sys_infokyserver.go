package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysInfokyserver represents the MSG_SYS_INFOKYSERVER
type MsgSysInfokyserver struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysInfokyserver) Opcode() network.PacketID {
	return network.MSG_SYS_INFOKYSERVER
}

// Parse parses the packet from binary
func (m *MsgSysInfokyserver) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysInfokyserver) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
