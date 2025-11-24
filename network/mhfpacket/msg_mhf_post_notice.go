package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPostNotice represents the MSG_MHF_POST_NOTICE
type MsgMhfPostNotice struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostNotice) Opcode() network.PacketID {
	return network.MSG_MHF_POST_NOTICE
}

// Parse parses the packet from binary
func (m *MsgMhfPostNotice) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostNotice) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
