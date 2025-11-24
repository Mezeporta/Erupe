package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfExchangeKouryouPoint represents the MSG_MHF_EXCHANGE_KOURYOU_POINT
type MsgMhfExchangeKouryouPoint struct {
	AckHandle     uint32
	KouryouPoints uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfExchangeKouryouPoint) Opcode() network.PacketID {
	return network.MSG_MHF_EXCHANGE_KOURYOU_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfExchangeKouryouPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.KouryouPoints = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfExchangeKouryouPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
