package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfSetGuildMissionTarget represents the MSG_MHF_SET_GUILD_MISSION_TARGET
type MsgMhfSetGuildMissionTarget struct {
	AckHandle uint32
	MissionID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetGuildMissionTarget) Opcode() network.PacketID {
	return network.MSG_MHF_SET_GUILD_MISSION_TARGET
}

// Parse parses the packet from binary
func (m *MsgMhfSetGuildMissionTarget) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.MissionID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetGuildMissionTarget) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
