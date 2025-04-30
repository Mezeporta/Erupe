package mhfpacket

import (
	"errors"
	_config "erupe-ce/config"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAcquireCafeItem represents the MSG_MHF_ACQUIRE_CAFE_ITEM
type MsgMhfAcquireCafeItem struct {
	AckHandle uint32
	// Valid sizes, not sure if [un]signed.
	ItemType  uint16
	ItemID    uint16
	Quant     uint16
	PointCost uint32
	Unk0      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireCafeItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_CAFE_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireCafeItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.ItemType = bf.ReadUint16()
	m.ItemID = bf.ReadUint16()
	m.Quant = bf.ReadUint16()
	if _config.ErupeConfig.RealClientMode >= _config.G6 {
		m.PointCost = bf.ReadUint32()
	} else {
		m.PointCost = uint32(bf.ReadUint16())
	}
	m.Unk0 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireCafeItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
