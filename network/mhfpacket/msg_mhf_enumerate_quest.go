package mhfpacket

import (
	"errors"
	"erupe-ce/config"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateQuest represents the MSG_MHF_ENUMERATE_QUEST
type MsgMhfEnumerateQuest struct {
	AckHandle uint32
	Unk0      uint8 // Hardcoded 0 in the binary
	World     uint8
	Counter   uint16
	Offset    uint16 // Increments to request following batches of quests
	Unk4      uint8  // Hardcoded 0 in the binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateQuest) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateQuest) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.World = bf.ReadUint8()
	m.Counter = bf.ReadUint16()
	if config.GetConfig().ClientID <= config.Z1 {
		m.Offset = uint16(bf.ReadUint8())
	} else {
		m.Offset = bf.ReadUint16()
	}
	m.Unk4 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateQuest) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
