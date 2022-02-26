package mhfpacket

import (
	"errors"

	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireDistItem represents the MSG_MHF_ACQUIRE_DIST_ITEM
type MsgMhfAcquireDistItem struct {
	AckHandle uint32
	// Valid field size(s), not sure about the types.
	Unk0 uint8
	Unk1 uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireDistItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_DIST_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireDistItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireDistItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
