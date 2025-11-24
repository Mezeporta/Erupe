package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateGuildMessageBoard represents the MSG_MHF_ENUMERATE_GUILD_MESSAGE_BOARD
type MsgMhfEnumerateGuildMessageBoard struct {
	AckHandle uint32
	Unk0      uint32
	MaxPosts  uint32 // always 100, even on news (00000064)
	// returning more than 4 news posts WILL softlock
	BoardType uint32 // 0 => message, 1 => news
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuildMessageBoard) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD_MESSAGE_BOARD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuildMessageBoard) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.MaxPosts = bf.ReadUint32()
	m.BoardType = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuildMessageBoard) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
