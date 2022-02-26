package mhfpacket

import (
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysNotifyRegister represents the MSG_SYS_NOTIFY_REGISTER
type MsgSysNotifyRegister struct {
	RegisterID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysNotifyRegister) Opcode() network.PacketID {
	return network.MSG_SYS_NOTIFY_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysNotifyRegister) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.RegisterID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysNotifyRegister) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.RegisterID)
	return nil
}
