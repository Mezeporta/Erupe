package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddKouryouPoint represents the MSG_MHF_ADD_KOURYOU_POINT
type MsgMhfAddKouryouPoint struct {
	AckHandle     uint32
	KouryouPoints uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddKouryouPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_KOURYOU_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddKouryouPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.KouryouPoints = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddKouryouPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.KouryouPoints)
	return nil
}
