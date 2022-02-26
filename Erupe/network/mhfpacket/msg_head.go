package mhfpacket

import (
	"errors"

	"github.com/Solenataris/Erupe/network/clientctx"

	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgHead represents the MSG_HEAD
type MsgHead struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgHead) Opcode() network.PacketID {
	return network.MSG_HEAD
}

// Parse parses the packet from binary
func (m *MsgHead) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgHead) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
