package binpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network"
)

// ChatType represents the chat message type (Thanks to @Alice on discord for identifying these!)
type ChatType uint8

// Chat types
const (
	ChatTypeWorld    ChatType = 0
	ChatTypeStage    ChatType = 1
	ChatTypeGuild    ChatType = 2
	ChatTypeAlliance ChatType = 3
	ChatTypeParty    ChatType = 4
	ChatTypeWhisper  ChatType = 5
)

// MsgBinChat is a binpacket for chat messages.
type MsgBinChat struct {
	Unk0       uint8
	Type       ChatType
	Flags      uint16
	Message    string
	SenderName string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgBinChat) Opcode() network.PacketID {
	return network.MSG_SYS_CAST_BINARY
}

// Parse parses the packet from binary
func (m *MsgBinChat) Parse(bf *byteframe.ByteFrame) error {
	m.Unk0 = bf.ReadUint8()
	m.Type = ChatType(bf.ReadUint8())
	m.Flags = bf.ReadUint16()
	_ = bf.ReadUint16() // lenSenderName
	_ = bf.ReadUint16() // lenMessage
	m.Message = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	m.SenderName = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgBinChat) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(m.Unk0)
	bf.WriteUint8(uint8(m.Type))
	bf.WriteUint16(m.Flags)
	cMessage := stringsupport.UTF8ToSJIS(m.Message)
	cSenderName := stringsupport.UTF8ToSJIS(m.SenderName)
	bf.WriteUint16(uint16(len(cSenderName) + 1))
	bf.WriteUint16(uint16(len(cMessage) + 1))
	bf.WriteNullTerminatedBytes(cMessage)
	bf.WriteNullTerminatedBytes(cSenderName)
	return nil
}
