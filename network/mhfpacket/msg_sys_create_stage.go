package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysCreateStage represents the MSG_SYS_CREATE_STAGE
type MsgSysCreateStage struct {
	AckHandle   uint32
	CreateType  uint8 // 1 = new stage (lobby, my house, quest), 2 = existing stage (guild room, move)
	PlayerCount uint8
	StageID     string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateStage) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysCreateStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CreateType = bf.ReadUint8()
	m.PlayerCount = bf.ReadUint8()
	bf.ReadUint8() // Length StageID
	m.StageID = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
