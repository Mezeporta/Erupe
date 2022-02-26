package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysUnlockGlobalSema represents the MSG_SYS_UNLOCK_GLOBAL_SEMA
type MsgSysUnlockGlobalSema struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUnlockGlobalSema) Opcode() network.PacketID {
	return network.MSG_SYS_UNLOCK_GLOBAL_SEMA
}

// Parse parses the packet from binary
func (m *MsgSysUnlockGlobalSema) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysUnlockGlobalSema) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	return nil
}
