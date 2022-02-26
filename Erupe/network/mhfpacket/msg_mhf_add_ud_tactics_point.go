package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddUdTacticsPoint represents the MSG_MHF_ADD_UD_TACTICS_POINT
type MsgMhfAddUdTacticsPoint struct {
	AckHandle uint32
	Unk0      uint16
	Unk1      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdTacticsPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_TACTICS_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdTacticsPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdTacticsPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(m.Unk0)
	bf.WriteUint32(m.Unk1)
	return nil
}
