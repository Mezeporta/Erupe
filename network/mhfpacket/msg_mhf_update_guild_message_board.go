package mhfpacket

import (
	"errors"
	"erupe-ce/common/stringsupport"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateGuildMessageBoard represents the MSG_MHF_UPDATE_GUILD_MESSAGE_BOARD
type MsgMhfUpdateGuildMessageBoard struct {
	AckHandle   uint32
	MessageOp   uint32
	PostType    uint32
	StampID     uint32
	TitleLength uint32
	BodyLength  uint32
	Title       string
	Body        string
	PostID      uint32
	LikeState   bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuildMessageBoard) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUILD_MESSAGE_BOARD
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuildMessageBoard) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.MessageOp = bf.ReadUint32()
	switch m.MessageOp {
	case 0:
		m.PostType = bf.ReadUint32()
		m.StampID = bf.ReadUint32()
		m.TitleLength = bf.ReadUint32()
		m.BodyLength = bf.ReadUint32()
		m.Title, _ = stringsupport.SJISToUTF8(bf.ReadBytes(uint(m.TitleLength)))
		m.Body, _ = stringsupport.SJISToUTF8(bf.ReadBytes(uint(m.BodyLength)))
	case 1:
		m.PostID = bf.ReadUint32()
	case 2:
		m.PostID = bf.ReadUint32()
		bf.ReadBytes(8)
		m.TitleLength = bf.ReadUint32()
		m.BodyLength = bf.ReadUint32()
		m.Title, _ = stringsupport.SJISToUTF8(bf.ReadBytes(uint(m.TitleLength)))
		m.Body, _ = stringsupport.SJISToUTF8(bf.ReadBytes(uint(m.BodyLength)))
	case 3:
		m.PostID = bf.ReadUint32()
		bf.ReadBytes(8)
		m.StampID = bf.ReadUint32()
	case 4:
		m.PostID = bf.ReadUint32()
		bf.ReadBytes(8)
		m.LikeState = bf.ReadBool()
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuildMessageBoard) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
