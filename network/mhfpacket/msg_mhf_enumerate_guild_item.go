package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateGuildItem represents the MSG_MHF_ENUMERATE_GUILD_ITEM
type MsgMhfEnumerateGuildItem struct {
	AckHandle uint32
	GuildId   uint32
	Unk0      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuildItem) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuildItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildId = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuildItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
