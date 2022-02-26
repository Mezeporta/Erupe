package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysExtendThreshold represents the MSG_SYS_EXTEND_THRESHOLD
type MsgSysExtendThreshold struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysExtendThreshold) Opcode() network.PacketID {
	return network.MSG_SYS_EXTEND_THRESHOLD
}

// Parse parses the packet from binary
func (m *MsgSysExtendThreshold) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	// No data aside from opcode.
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysExtendThreshold) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	// No data aside from opcode.
	return nil
}
