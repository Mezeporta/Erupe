package mhfpacket

import (
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/common/bfutil"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysReserveStage represents the MSG_SYS_RESERVE_STAGE
type MsgSysReserveStage struct {
	AckHandle uint32
	Unk0      uint8  // Made with: `16 * x | 1;`, unknown `x` values.
	StageID   string // NULL terminated string.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserveStage) Opcode() network.PacketID {
	return network.MSG_SYS_RESERVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysReserveStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserveStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	panic("Not implemented")
}
