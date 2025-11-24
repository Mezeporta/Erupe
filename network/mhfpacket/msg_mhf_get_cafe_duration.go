package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetCafeDuration represents the MSG_MHF_GET_CAFE_DURATION
type MsgMhfGetCafeDuration struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetCafeDuration) Opcode() network.PacketID {
	return network.MSG_MHF_GET_CAFE_DURATION
}

// Parse parses the packet from binary
func (m *MsgMhfGetCafeDuration) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetCafeDuration) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
