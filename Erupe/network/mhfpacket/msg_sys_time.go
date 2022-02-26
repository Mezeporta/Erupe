package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysTime represents the MSG_SYS_TIME
type MsgSysTime struct {
	GetRemoteTime bool   // Ask the other end to send it's time as well.
	Timestamp     uint32 // Unix timestamp, e.g. 1577105879
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTime) Opcode() network.PacketID {
	return network.MSG_SYS_TIME
}

// Parse parses the packet from binary
func (m *MsgSysTime) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.GetRemoteTime = bf.ReadBool()
	m.Timestamp = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysTime) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteBool(m.GetRemoteTime)
	bf.WriteUint32(m.Timestamp)
	return nil
}
