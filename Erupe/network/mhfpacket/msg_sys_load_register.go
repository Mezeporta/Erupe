package mhfpacket

import (
	"fmt"

	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysLoadRegister represents the MSG_SYS_LOAD_REGISTER
type MsgSysLoadRegister struct {
	AckHandle  uint32
	RegisterID uint32
	Unk1       uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLoadRegister) Opcode() network.PacketID {
	return network.MSG_SYS_LOAD_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysLoadRegister) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.RegisterID = bf.ReadUint32()
	m.Unk1 = bf.ReadUint8()
	fixedZero0 := bf.ReadUint16()
	fixedZero1 := bf.ReadUint8()

	if fixedZero0 != 0 || fixedZero1 != 0 {
		return fmt.Errorf("expected fixed-0 values, got %d %d", fixedZero0, fixedZero1)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLoadRegister) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.RegisterID)
	bf.WriteUint8(m.Unk1)
	bf.WriteUint16(0)
	bf.WriteUint8(0)

	return nil
}
