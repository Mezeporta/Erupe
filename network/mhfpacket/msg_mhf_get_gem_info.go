package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetGemInfo represents the MSG_MHF_GET_GEM_INFO
type MsgMhfGetGemInfo struct {
	AckHandle uint32
	QueryType uint32
	Unk1      uint32
	Unk2      int32
	Unk3      int32
	Unk4      int32
	Unk5      int32
	Unk6      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGemInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GEM_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetGemInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.QueryType = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadInt32()
	m.Unk3 = bf.ReadInt32()
	m.Unk4 = bf.ReadInt32()
	m.Unk5 = bf.ReadInt32()
	m.Unk6 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGemInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
