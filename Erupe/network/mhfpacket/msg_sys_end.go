package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysEnd represents the MSG_SYS_END
type MsgSysEnd struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnd) Opcode() network.PacketID {
	return network.MSG_SYS_END
}

// Parse parses the packet from binary
func (m *MsgSysEnd) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	// No data aside from opcode.
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnd) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	// No data aside from opcode.
	return nil
}
