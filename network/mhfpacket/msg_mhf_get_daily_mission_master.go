package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetDailyMissionMaster requests the server-side daily mission master list.
// Full request payload beyond the AckHandle is not yet reverse-engineered.
type MsgMhfGetDailyMissionMaster struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetDailyMissionMaster) Opcode() network.PacketID {
	return network.MSG_MHF_GET_DAILY_MISSION_MASTER
}

// Parse parses the packet from binary.
// Only the AckHandle is parsed; additional fields are unknown.
func (m *MsgMhfGetDailyMissionMaster) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetDailyMissionMaster) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
