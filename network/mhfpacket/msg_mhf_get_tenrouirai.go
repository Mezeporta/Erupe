package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetTenrouirai represents the MSG_MHF_GET_TENROUIRAI
type MsgMhfGetTenrouirai struct {
	AckHandle uint32
	Unk0         uint8
	DataType     uint8
	GuildID      uint32
	MissionIndex uint8
	Unk4         uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetTenrouirai) Opcode() network.PacketID {
	return network.MSG_MHF_GET_TENROUIRAI
}

// Parse parses the packet from binary
func (m *MsgMhfGetTenrouirai) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.DataType = bf.ReadUint8()
	m.GuildID = bf.ReadUint32()
	m.MissionIndex = bf.ReadUint8()
	m.Unk4 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetTenrouirai) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
