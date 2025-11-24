package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetDailyMissionMaster represents the MSG_MHF_GET_DAILY_MISSION_MASTER
type MsgMhfGetDailyMissionMaster struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetDailyMissionMaster) Opcode() network.PacketID {
	return network.MSG_MHF_GET_DAILY_MISSION_MASTER
}

// Parse parses the packet from binary
func (m *MsgMhfGetDailyMissionMaster) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetDailyMissionMaster) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
