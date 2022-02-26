package mhfpacket

import (
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/common/bfutil"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysGetStageBinary represents the MSG_SYS_GET_STAGE_BINARY
type MsgSysGetStageBinary struct {
	AckHandle   uint32
	BinaryType0 uint8
	BinaryType1 uint8
	Unk0        uint32 // Hardcoded 0
	StageID     string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetStageBinary) Opcode() network.PacketID {
	return network.MSG_SYS_GET_STAGE_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysGetStageBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.BinaryType0 = bf.ReadUint8()
	m.BinaryType1 = bf.ReadUint8()
	m.Unk0 = bf.ReadUint32()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetStageBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	panic("Not implemented")
}
