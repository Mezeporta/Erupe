package mhfpacket

import (
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysSetStageBinary represents the MSG_SYS_SET_STAGE_BINARY
type MsgSysSetStageBinary struct {
	BinaryType0    uint8
	BinaryType1    uint8 // Index
	StageID        string
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetStageBinary) Opcode() network.PacketID {
	return network.MSG_SYS_SET_STAGE_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysSetStageBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.BinaryType0 = bf.ReadUint8()
	m.BinaryType1 = bf.ReadUint8()
	bf.ReadUint8()              // StageID length <= 0x20
	dataSize := bf.ReadUint16() // <= 0x400
	m.StageID = string(bf.ReadNullTerminatedBytes())
	m.RawDataPayload = bf.ReadBytes(uint(dataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetStageBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return fmt.Errorf("MsgSysSetStageBinary.Build: not implemented")
}
