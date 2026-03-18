package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddUdPoint represents the MSG_MHF_ADD_UD_POINT
//
// Sent by the client after completing a Diva Defense quest to report earned points.
// RE'd from ZZ DLL putAdd_ud_point (FUN_114fd490): the client sums 11 point
// category accumulators into QuestPoints, and computes BonusPoints from the
// kiju prayer song multiplier applied to the base categories.
type MsgMhfAddUdPoint struct {
	AckHandle   uint32
	QuestPoints uint32 // Total points earned from the quest (sum of all categories)
	BonusPoints uint32 // Extra points from kiju/prayer song multiplier
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.QuestPoints = bf.ReadUint32()
	m.BonusPoints = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.QuestPoints)
	bf.WriteUint32(m.BonusPoints)
	return nil
}
