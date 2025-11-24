package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfRegistGuildCooking represents the MSG_MHF_REGIST_GUILD_COOKING
type MsgMhfRegistGuildCooking struct {
	AckHandle   uint32
	OverwriteID uint32
	MealID      uint16
	Success     uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildCooking) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_COOKING
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildCooking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.OverwriteID = bf.ReadUint32()
	m.MealID = bf.ReadUint16()
	m.Success = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildCooking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
