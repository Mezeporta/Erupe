package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddUdTacticsPoint represents the MSG_MHF_ADD_UD_TACTICS_POINT
//
// Sent during Diva Defense interception phase to report tactics points.
// RE'd from ZZ DLL putAdd_ud_tactics_point (FUN_114fe9c0): QuestID is read
// from a character data field, TacticsPoints is the accumulated tactics value.
type MsgMhfAddUdTacticsPoint struct {
	AckHandle     uint32
	QuestID       uint16 // Quest/character identifier from savedata
	TacticsPoints uint32 // Accumulated tactics interception points
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdTacticsPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_TACTICS_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdTacticsPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.QuestID = bf.ReadUint16()
	m.TacticsPoints = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdTacticsPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(m.QuestID)
	bf.WriteUint32(m.TacticsPoints)
	return nil
}
