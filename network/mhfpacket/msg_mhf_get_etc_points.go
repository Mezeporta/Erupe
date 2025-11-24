package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetEtcPoints represents the MSG_MHF_GET_ETC_POINTS
type MsgMhfGetEtcPoints struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEtcPoints) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ETC_POINTS
}

// Parse parses the packet from binary
func (m *MsgMhfGetEtcPoints) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEtcPoints) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return nil
}
