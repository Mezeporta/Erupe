package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfArrangeGuildMember represents the MSG_MHF_ARRANGE_GUILD_MEMBER
type MsgMhfArrangeGuildMember struct {
	AckHandle uint32
	GuildID   uint32
	CharIDs   []uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfArrangeGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_ARRANGE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfArrangeGuildMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	charCount := bf.ReadUint16()

	m.CharIDs = make([]uint32, charCount)

	for i := uint16(0); i < charCount; i++ {
		m.CharIDs[i] = bf.ReadUint32()
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfArrangeGuildMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.GuildID)
	bf.WriteUint16(uint16(len(m.CharIDs)))

	for _, charID := range m.CharIDs {
		bf.WriteUint32(charID)
	}

	return nil
}
