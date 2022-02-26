package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysUpdateObjectBinary represents the MSG_SYS_UPDATE_OBJECT_BINARY
type MsgSysUpdateObjectBinary struct {
	Unk0 uint32 // Object handle ID
	Unk1 uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUpdateObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_UPDATE_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysUpdateObjectBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysUpdateObjectBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.Unk0)
	bf.WriteUint32(m.Unk1)
	return nil
}
