package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfLoadGuildCooking represents the MSG_MHF_LOAD_GUILD_COOKING
type MsgMhfLoadGuildCooking struct {
	AckHandle uint32
	MaxMeals  uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadGuildCooking) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_GUILD_COOKING
}

// Parse parses the packet from binary
func (m *MsgMhfLoadGuildCooking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	_ = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadGuildCooking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
