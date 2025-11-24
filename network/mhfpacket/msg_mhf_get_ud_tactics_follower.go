package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetUdTacticsFollower represents the MSG_MHF_GET_UD_TACTICS_FOLLOWER
type MsgMhfGetUdTacticsFollower struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsFollower) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_FOLLOWER
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsFollower) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsFollower) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
