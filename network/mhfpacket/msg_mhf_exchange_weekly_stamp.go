package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfExchangeWeeklyStamp represents the MSG_MHF_EXCHANGE_WEEKLY_STAMP
type MsgMhfExchangeWeeklyStamp struct {
	AckHandle uint32
	StampType string
	ExchangeType uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfExchangeWeeklyStamp) Opcode() network.PacketID {
	return network.MSG_MHF_EXCHANGE_WEEKLY_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfExchangeWeeklyStamp) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	stampType := bf.ReadUint8()
	switch stampType {
	case 1:
		m.StampType = "hl"
	case 2:
		m.StampType = "ex"
	}
	m.ExchangeType = bf.ReadUint8()
	bf.ReadUint16() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfExchangeWeeklyStamp) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
