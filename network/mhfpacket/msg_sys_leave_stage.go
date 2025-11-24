package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysLeaveStage represents the MSG_SYS_LEAVE_STAGE
type MsgSysLeaveStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLeaveStage) Opcode() network.PacketID {
	return network.MSG_SYS_LEAVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysLeaveStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysLeaveStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
