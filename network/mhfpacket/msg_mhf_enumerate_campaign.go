package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateCampaign represents the MSG_MHF_ENUMERATE_CAMPAIGN
type MsgMhfEnumerateCampaign struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateCampaign) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_CAMPAIGN
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateCampaign) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	bf.ReadUint16() // Zeroed in Z2
	bf.ReadUint16() // Zeroed in Z2
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateCampaign) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	return nil
}
