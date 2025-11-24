package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfListMail represents the MSG_MHF_LIST_MAIL
type MsgMhfListMail struct {
	AckHandle uint32
	Unk0      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfListMail) Opcode() network.PacketID {
	return network.MSG_MHF_LIST_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfListMail) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfListMail) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
