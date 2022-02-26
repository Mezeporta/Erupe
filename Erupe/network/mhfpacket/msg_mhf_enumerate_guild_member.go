package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateGuildMember represents the MSG_MHF_ENUMERATE_GUILD_MEMBER
type MsgMhfEnumerateGuildMember struct {
	AckHandle uint32
	Unk0      uint16 // Hardcoded 00 01 in the binary
	Unk1      uint32 // Alliance related
	GuildID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuildMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuildMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(m.Unk0)
	bf.WriteUint32(m.Unk1)
	bf.WriteUint32(m.GuildID)
	return nil
}
