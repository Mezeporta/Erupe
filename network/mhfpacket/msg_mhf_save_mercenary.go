package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfSaveMercenary represents the MSG_MHF_SAVE_MERCENARY
type MsgMhfSaveMercenary struct {
	AckHandle   uint32
	GCP         uint32
	PactMercID  uint32
	RastaMercID uint32
	charID      uint32
	lastState   bool
	MercData    []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveMercenary) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_MERCENARY
}

// Parse parses the packet from binary
func (m *MsgMhfSaveMercenary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	bf.ReadUint32() // lenData
	m.GCP = bf.ReadUint32()
	m.PactMercID = bf.ReadUint32()
	dataSize := bf.ReadUint32()
	_ = bf.ReadUint32() // Merc index?
	raw := bf.ReadBytes(uint(dataSize))
	temp := byteframe.NewByteFrameFromBytes(raw)
	m.RastaMercID = temp.ReadUint32()
	m.charID = temp.ReadUint32()
	m.lastState = temp.ReadBool()

	m.MercData = raw[9:]
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveMercenary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
