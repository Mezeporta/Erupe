package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfSetRestrictionEvent represents the MSG_MHF_SET_RESTRICTION_EVENT
type MsgMhfSetRestrictionEvent struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      uint32
	Unk3      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetRestrictionEvent) Opcode() network.PacketID {
	return network.MSG_MHF_SET_RESTRICTION_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfSetRestrictionEvent) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetRestrictionEvent) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
