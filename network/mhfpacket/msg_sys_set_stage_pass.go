package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysSetStagePass represents the MSG_SYS_SET_STAGE_PASS
type MsgSysSetStagePass struct {
	Unk0     uint8  // Hardcoded 0 in the binary
	Password string // NULL-terminated string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetStagePass) Opcode() network.PacketID {
	return network.MSG_SYS_SET_STAGE_PASS
}

// Parse parses the packet from binary
func (m *MsgSysSetStagePass) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.Unk0 = bf.ReadUint8()
	_ = bf.ReadUint8() // Password length
	m.Password = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetStagePass) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
