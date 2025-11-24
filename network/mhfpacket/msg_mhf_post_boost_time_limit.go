package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPostBoostTimeLimit represents the MSG_MHF_POST_BOOST_TIME_LIMIT
type MsgMhfPostBoostTimeLimit struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostBoostTimeLimit) Opcode() network.PacketID {
	return network.MSG_MHF_POST_BOOST_TIME_LIMIT
}

// Parse parses the packet from binary
func (m *MsgMhfPostBoostTimeLimit) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostBoostTimeLimit) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
