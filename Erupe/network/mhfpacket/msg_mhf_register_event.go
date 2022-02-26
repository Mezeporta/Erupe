package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfRegisterEvent represents the MSG_MHF_REGISTER_EVENT
type MsgMhfRegisterEvent struct {
	AckHandle uint32
	Unk0      uint16
	Unk1      uint8
	Unk2      uint8
	Unk3      uint8
	Unk4      uint8
	Unk5      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegisterEvent) Opcode() network.PacketID {
	return network.MSG_MHF_REGISTER_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfRegisterEvent) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint8()
	m.Unk3 = bf.ReadUint8()
	m.Unk4 = bf.ReadUint8()
	m.Unk5 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegisterEvent) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return nil
}
