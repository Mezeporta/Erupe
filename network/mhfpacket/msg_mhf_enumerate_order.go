package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateOrder represents the MSG_MHF_ENUMERATE_ORDER
type MsgMhfEnumerateOrder struct {
	AckHandle uint32
	EventID   uint32
	ClanID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateOrder) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_ORDER
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateOrder) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.EventID = bf.ReadUint32()
	m.ClanID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateOrder) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
