package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfStateFestaG represents the MSG_MHF_STATE_FESTA_G
type MsgMhfStateFestaG struct {
	AckHandle uint32
	FestaID   uint32
	GuildID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStateFestaG) Opcode() network.PacketID {
	return network.MSG_MHF_STATE_FESTA_G
}

// Parse parses the packet from binary
func (m *MsgMhfStateFestaG) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.FestaID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	_ = bf.ReadUint16() // Hardcoded 0 in the binary.
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStateFestaG) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
