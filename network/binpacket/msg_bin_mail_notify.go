package binpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network"
)

// MsgBinMailNotify is a binpacket broadcast to notify a player of new mail.
type MsgBinMailNotify struct {
	SenderName string
}

// Parse parses the packet from binary.
func (m MsgBinMailNotify) Parse(bf *byteframe.ByteFrame) error {
	panic("implement me")
}

// Build builds a binary packet from the current data.
func (m MsgBinMailNotify) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(0x01) // Unk
	bf.WriteBytes(stringsupport.PaddedString(m.SenderName, 21, true))
	return nil
}

// Opcode returns the ID associated with this packet type.
func (m MsgBinMailNotify) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}
