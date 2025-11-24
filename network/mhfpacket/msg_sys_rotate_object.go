package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysRotateObject represents the MSG_SYS_ROTATE_OBJECT
type MsgSysRotateObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRotateObject) Opcode() network.PacketID {
	return network.MSG_SYS_ROTATE_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysRotateObject) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysRotateObject) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
