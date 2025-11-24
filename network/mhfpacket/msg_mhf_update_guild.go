package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateGuild represents the MSG_MHF_UPDATE_GUILD
type MsgMhfUpdateGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
