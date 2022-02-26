package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCastedBinary represents the MSG_SYS_CASTED_BINARY
type MsgSysCastedBinary struct {
	CharID         uint32
	BroadcastType  uint8
	MessageType    uint8
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCastedBinary) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysCastedBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.CharID = bf.ReadUint32()
	m.BroadcastType = bf.ReadUint8()
	m.MessageType = bf.ReadUint8()
	dataSize := bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(dataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCastedBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.CharID)
	bf.WriteUint8(m.BroadcastType)
	bf.WriteUint8(m.MessageType)
	bf.WriteUint16(uint16(len(m.RawDataPayload)))
	bf.WriteBytes(m.RawDataPayload)
	return nil
}
