package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfRegistGuildAdventureDiva represents the MSG_MHF_REGIST_GUILD_ADVENTURE_DIVA
type MsgMhfRegistGuildAdventureDiva struct {
	AckHandle   uint32
	Destination uint32
	Charge      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildAdventureDiva) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_ADVENTURE_DIVA
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildAdventureDiva) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Destination = bf.ReadUint32()
	m.Charge = bf.ReadUint32()
	_ = bf.ReadUint32() // CharID
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildAdventureDiva) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
