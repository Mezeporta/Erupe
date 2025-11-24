package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetRestrictionEvent represents the MSG_MHF_GET_RESTRICTION_EVENT
type MsgMhfGetRestrictionEvent struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRestrictionEvent) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RESTRICTION_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfGetRestrictionEvent) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRestrictionEvent) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
