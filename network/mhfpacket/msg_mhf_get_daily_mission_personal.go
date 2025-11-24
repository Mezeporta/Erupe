package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetDailyMissionPersonal represents the MSG_MHF_GET_DAILY_MISSION_PERSONAL
type MsgMhfGetDailyMissionPersonal struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetDailyMissionPersonal) Opcode() network.PacketID {
	return network.MSG_MHF_GET_DAILY_MISSION_PERSONAL
}

// Parse parses the packet from binary
func (m *MsgMhfGetDailyMissionPersonal) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetDailyMissionPersonal) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
