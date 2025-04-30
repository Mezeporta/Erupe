package mhfpacket

import (
	"errors"
	_config "erupe-ce/config"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfStampcardStamp represents the MSG_MHF_STAMPCARD_STAMP
type MsgMhfStampcardStamp struct {
	AckHandle uint32
	HR        uint16
	GR        uint16
	Stamps    uint16
	Reward1   uint16
	Reward2   uint16
	Item1     uint16
	Item2     uint16
	Quantity1 uint16
	Quantity2 uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStampcardStamp) Opcode() network.PacketID {
	return network.MSG_MHF_STAMPCARD_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfStampcardStamp) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.HR = bf.ReadUint16()
	if _config.ErupeConfig.RealClientMode >= _config.G1 {
		m.GR = bf.ReadUint16()
	}
	m.Stamps = bf.ReadUint16()
	bf.ReadUint16() // Zeroed
	if _config.ErupeConfig.RealClientMode >= _config.Z2 {
		m.Reward1 = uint16(bf.ReadUint32())
		m.Reward2 = uint16(bf.ReadUint32())
		m.Item1 = uint16(bf.ReadUint32())
		m.Item2 = uint16(bf.ReadUint32())
		m.Quantity1 = uint16(bf.ReadUint32())
		m.Quantity2 = uint16(bf.ReadUint32())
	} else {
		m.Reward1 = 10
		m.Reward2 = 10
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStampcardStamp) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
