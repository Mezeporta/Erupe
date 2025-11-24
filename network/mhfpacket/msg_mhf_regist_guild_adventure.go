package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfRegistGuildAdventure represents the MSG_MHF_REGIST_GUILD_ADVENTURE
type MsgMhfRegistGuildAdventure struct {
	AckHandle   uint32
	Destination uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildAdventure) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Destination = bf.ReadUint32()
	_ = bf.ReadUint32() // CharID
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildAdventure) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
