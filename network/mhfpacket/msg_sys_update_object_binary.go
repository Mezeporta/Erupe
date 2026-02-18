package mhfpacket

import (
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/common/byteframe"
)

// MsgSysUpdateObjectBinary represents the MSG_SYS_UPDATE_OBJECT_BINARY
type MsgSysUpdateObjectBinary struct {
	ObjectHandleID uint32
	Unk1           uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUpdateObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_UPDATE_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysUpdateObjectBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.ObjectHandleID = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysUpdateObjectBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.ObjectHandleID)
	bf.WriteUint32(m.Unk1)
	return nil
}
