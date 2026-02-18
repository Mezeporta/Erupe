package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfMercenaryHuntdata represents the MSG_MHF_MERCENARY_HUNTDATA
type MsgMhfMercenaryHuntdata struct {
	AckHandle uint32
	RequestType uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfMercenaryHuntdata) Opcode() network.PacketID {
	return network.MSG_MHF_MERCENARY_HUNTDATA
}

// Parse parses the packet from binary
func (m *MsgMhfMercenaryHuntdata) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.RequestType = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfMercenaryHuntdata) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
