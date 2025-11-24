package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfReserve10F represents the MSG_MHF_reserve10F
type MsgMhfReserve10F struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReserve10F) Opcode() network.PacketID {
	return network.MSG_MHF_reserve10F
}

// Parse parses the packet from binary
func (m *MsgMhfReserve10F) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReserve10F) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
