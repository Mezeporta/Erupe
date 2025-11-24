package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAcquireFesta represents the MSG_MHF_ACQUIRE_FESTA
type MsgMhfAcquireFesta struct {
	AckHandle uint32
	FestaID   uint32
	GuildID   uint32
	Unk       uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireFesta) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireFesta) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.FestaID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.Unk = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireFesta) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
