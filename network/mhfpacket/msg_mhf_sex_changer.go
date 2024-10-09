package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfSexChanger represents the MSG_MHF_SEX_CHANGER
type MsgMhfSexChanger struct {
	AckHandle uint32
	Gender    uint8
	Unk0      uint8
	Unk1      uint8
	Unk2      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSexChanger) Opcode() network.PacketID {
	return network.MSG_MHF_SEX_CHANGER
}

// Parse parses the packet from binary
func (m *MsgMhfSexChanger) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Gender = bf.ReadUint8()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSexChanger) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
