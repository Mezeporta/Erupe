package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfSetUdTacticsFollower represents the MSG_MHF_SET_UD_TACTICS_FOLLOWER
type MsgMhfSetUdTacticsFollower struct {
	AckHandle uint32
	Unk0      uint16
	Unk1      uint16
	Unk2      uint16
	Unk3      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetUdTacticsFollower) Opcode() network.PacketID {
	return network.MSG_MHF_SET_UD_TACTICS_FOLLOWER
}

// Parse parses the packet from binary
func (m *MsgMhfSetUdTacticsFollower) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	m.Unk3 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetUdTacticsFollower) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
