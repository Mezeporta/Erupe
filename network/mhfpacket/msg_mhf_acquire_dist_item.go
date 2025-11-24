package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAcquireDistItem represents the MSG_MHF_ACQUIRE_DIST_ITEM
type MsgMhfAcquireDistItem struct {
	AckHandle        uint32
	DistributionType uint8
	DistributionID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireDistItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_DIST_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireDistItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.DistributionType = bf.ReadUint8()
	m.DistributionID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireDistItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
