package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateRanking represents the MSG_MHF_ENUMERATE_RANKING
type MsgMhfEnumerateRanking struct {
	AckHandle uint32
	Unk0      uint16 // Hardcoded 0 in the binary
	Unk1      uint16 // Hardcoded 0 in the binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateRanking) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
