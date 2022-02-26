package mhfpacket

import (
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/common/bfutil"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysCreateStage represents the MSG_SYS_CREATE_STAGE
type MsgSysCreateStage struct {
	AckHandle   uint32
	Unk0        uint8 // Likely only has 1 and 2 as values.
	PlayerCount uint8
	StageID     string // NULL terminated string.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateStage) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysCreateStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.PlayerCount = bf.ReadUint8()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	panic("Not implemented")
}
