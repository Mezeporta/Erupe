package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfApplyDistItem represents the MSG_MHF_APPLY_DIST_ITEM
type MsgMhfApplyDistItem struct {
	AckHandle        uint32
	DistributionType uint8
	DistributionID   uint32
	Unk2             uint32
	Unk3             uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyDistItem) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_DIST_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfApplyDistItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.DistributionType = bf.ReadUint8()
	m.DistributionID = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyDistItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint8(m.DistributionType)
	bf.WriteUint32(m.DistributionID)
	bf.WriteUint32(m.Unk2)
	bf.WriteUint32(m.Unk3)
	return nil
}
