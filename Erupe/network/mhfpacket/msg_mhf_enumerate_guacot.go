package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateGuacot represents the MSG_MHF_ENUMERATE_GUACOT
type MsgMhfEnumerateGuacot struct {
	AckHandle uint32
	Unk0      uint16 // Hardcoded 0 in binary
	Unk1      uint16 // Hardcoded 0 in binary
	Unk2      uint16 // Hardcoded 0 in binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuacot) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuacot) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(m.Unk0)
	bf.WriteUint16(m.Unk1)
	bf.WriteUint16(m.Unk2)
	return nil
}
