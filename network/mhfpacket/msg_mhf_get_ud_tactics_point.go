package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetUdTacticsPoint represents the MSG_MHF_GET_UD_TACTICS_POINT
type MsgMhfGetUdTacticsPoint struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
