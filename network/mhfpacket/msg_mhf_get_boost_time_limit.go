package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetBoostTimeLimit represents the MSG_MHF_GET_BOOST_TIME_LIMIT
type MsgMhfGetBoostTimeLimit struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBoostTimeLimit) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BOOST_TIME_LIMIT
}

// Parse parses the packet from binary
func (m *MsgMhfGetBoostTimeLimit) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBoostTimeLimit) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
