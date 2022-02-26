package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysPositionObject represents the MSG_SYS_POSITION_OBJECT
type MsgSysPositionObject struct {
	ObjID   uint32
	X, Y, Z float32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysPositionObject) Opcode() network.PacketID {
	return network.MSG_SYS_POSITION_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysPositionObject) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.ObjID = bf.ReadUint32()
	m.X = bf.ReadFloat32()
	m.Y = bf.ReadFloat32()
	m.Z = bf.ReadFloat32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysPositionObject) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.ObjID)
	bf.WriteFloat32(m.X)
	bf.WriteFloat32(m.Y)
	bf.WriteFloat32(m.Z)
	return nil
}
