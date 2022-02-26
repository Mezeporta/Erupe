package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type GuildIconMsgPart struct {
	Index    uint16
	ID       uint16
	Unk0     uint8
	Size     uint8
	Rotation uint8
	Unk1     uint8
	Unk2     uint16
	PosX     uint16
	PosY     uint16
}

// MsgMhfUpdateGuildIcon represents the MSG_MHF_UPDATE_GUILD_ICON
type MsgMhfUpdateGuildIcon struct {
	AckHandle uint32
	GuildID   uint32
	PartCount uint16
	Unk1      uint16
	IconParts []GuildIconMsgPart
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuildIcon) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUILD_ICON
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuildIcon) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.PartCount = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()

	m.IconParts = make([]GuildIconMsgPart, m.PartCount)

	for i := 0; i < int(m.PartCount); i++ {
		m.IconParts[i] = GuildIconMsgPart{
			Index:    bf.ReadUint16(),
			ID:       bf.ReadUint16(),
			Unk0:     bf.ReadUint8(),
			Size:     bf.ReadUint8(),
			Rotation: bf.ReadUint8(),
			Unk1:     bf.ReadUint8(),
			Unk2:     bf.ReadUint16(),
			PosX:     bf.ReadUint16(),
			PosY:     bf.ReadUint16(),
		}
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuildIcon) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
