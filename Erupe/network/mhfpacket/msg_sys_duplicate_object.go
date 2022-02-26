package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysDuplicateObject represents the MSG_SYS_DUPLICATE_OBJECT
type MsgSysDuplicateObject struct {
	ObjID       uint32
	X, Y, Z     float32
	Unk0        uint32
	OwnerCharID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDuplicateObject) Opcode() network.PacketID {
	return network.MSG_SYS_DUPLICATE_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysDuplicateObject) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.ObjID = bf.ReadUint32()
	m.X = bf.ReadFloat32()
	m.Y = bf.ReadFloat32()
	m.Z = bf.ReadFloat32()
	m.Unk0 = bf.ReadUint32()
	m.OwnerCharID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysDuplicateObject) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.ObjID)
	bf.WriteFloat32(m.X)
	bf.WriteFloat32(m.Y)
	bf.WriteFloat32(m.Z)
	bf.WriteUint32(m.Unk0)
	bf.WriteUint32(m.OwnerCharID)
	return nil
}
