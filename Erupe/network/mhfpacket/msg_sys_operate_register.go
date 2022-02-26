package mhfpacket

import (
	"fmt"

	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysOperateRegister represents the MSG_SYS_OPERATE_REGISTER
type MsgSysOperateRegister struct {
	AckHandle      uint32
	RegisterID     uint32
	fixedZero      uint16
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysOperateRegister) Opcode() network.PacketID {
	return network.MSG_SYS_OPERATE_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysOperateRegister) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.RegisterID = bf.ReadUint32()
	m.fixedZero = bf.ReadUint16()

	if m.fixedZero != 0 {
		return fmt.Errorf("expected fixed-0 values, got %d", m.fixedZero)
	}

	dataSize := bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(dataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysOperateRegister) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.RegisterID)
	bf.WriteUint16(0)
	bf.WriteUint16(uint16(len(m.RawDataPayload)))
	bf.WriteBytes(m.RawDataPayload)

	return nil
}
