package mhfpacket

import (
	"errors"
	"erupe-ce/common/stringsupport"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfSendMail represents the MSG_MHF_SEND_MAIL
type MsgMhfSendMail struct {
	AckHandle     uint32
	RecipientID   uint32
	SubjectLength uint16
	BodyLength    uint16
	Quantity      uint16
	ItemID        uint16
	Subject       string
	Body          string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSendMail) Opcode() network.PacketID {
	return network.MSG_MHF_SEND_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfSendMail) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.RecipientID = bf.ReadUint32()
	m.SubjectLength = bf.ReadUint16()
	m.BodyLength = bf.ReadUint16()
	bf.ReadUint16() // Zeroed
	m.Quantity = bf.ReadUint16()
	m.ItemID = bf.ReadUint16()
	m.Subject, _ = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	m.Body, _ = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSendMail) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
