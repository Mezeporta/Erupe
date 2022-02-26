package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireExchangeShop represents the MSG_MHF_ACQUIRE_EXCHANGE_SHOP
type MsgMhfAcquireExchangeShop struct {
	AckHandle      uint32
	DataSize       uint16
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireExchangeShop) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_EXCHANGE_SHOP
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireExchangeShop) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.DataSize = bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireExchangeShop) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(m.DataSize)
	bf.WriteBytes(m.RawDataPayload)
	return nil
}
