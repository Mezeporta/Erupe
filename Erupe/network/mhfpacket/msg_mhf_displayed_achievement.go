package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfDisplayedAchievement represents the MSG_MHF_DISPLAYED_ACHIEVEMENT
type MsgMhfDisplayedAchievement struct {
	Unk0 uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfDisplayedAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_DISPLAYED_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfDisplayedAchievement) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.Unk0 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfDisplayedAchievement) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint8(m.Unk0)
	return nil
}
