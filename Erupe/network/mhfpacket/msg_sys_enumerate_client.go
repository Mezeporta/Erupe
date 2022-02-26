package mhfpacket

import (
	"errors"

	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/common/bfutil"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

// MsgSysEnumerateClient represents the MSG_SYS_ENUMERATE_CLIENT
type MsgSysEnumerateClient struct {
	AckHandle uint32
	Unk0      uint8 // Hardcoded 1 in the client
	Unk1      uint8
	StageID   string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnumerateClient) Opcode() network.PacketID {
	return network.MSG_SYS_ENUMERATE_CLIENT
}

// Parse parses the packet from binary
func (m *MsgSysEnumerateClient) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnumerateClient) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
