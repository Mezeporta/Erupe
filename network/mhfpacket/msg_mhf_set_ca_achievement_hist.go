package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// CaAchievementHist is a single entry in the CA achievement history packet.
type CaAchievementHist struct {
	Unk0 uint32
	Unk1 uint8
}

// MsgMhfSetCaAchievementHist represents the MSG_MHF_SET_CA_ACHIEVEMENT_HIST
type MsgMhfSetCaAchievementHist struct {
	AckHandle uint32
	Unk0      uint16
	Unk1      uint8
	Unk2      []CaAchievementHist
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetCaAchievementHist) Opcode() network.PacketID {
	return network.MSG_MHF_SET_CA_ACHIEVEMENT_HIST
}

// Parse parses the packet from binary
func (m *MsgMhfSetCaAchievementHist) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint8()
	for i := 0; i < int(m.Unk1); i++ {
		var temp CaAchievementHist
		temp.Unk0 = bf.ReadUint32()
		temp.Unk1 = bf.ReadUint8()
		m.Unk2 = append(m.Unk2, temp)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetCaAchievementHist) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
