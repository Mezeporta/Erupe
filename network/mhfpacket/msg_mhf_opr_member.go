package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfOprMember represents the MSG_MHF_OPR_MEMBER
type MsgMhfOprMember struct {
	AckHandle uint32
	Blacklist bool
	Operation bool
	Unk       uint16
	CharID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOprMember) Opcode() network.PacketID {
	return network.MSG_MHF_OPR_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfOprMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Blacklist = bf.ReadBool()
	m.Operation = bf.ReadBool()
	m.Unk = bf.ReadUint16()
	m.CharID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOprMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
