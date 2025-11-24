package mhfpacket

import (
	"errors"
	"erupe-ce/common/bfutil"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysReserveStage represents the MSG_SYS_RESERVE_STAGE
type MsgSysReserveStage struct {
	AckHandle uint32
	Ready     uint8  // Bitfield but hex (0x11 or 0x01)
	StageID   string // NULL terminated string.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserveStage) Opcode() network.PacketID {
	return network.MSG_SYS_RESERVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysReserveStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Ready = bf.ReadUint8()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserveStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
