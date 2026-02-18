package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// OperateMailOperation identifies the mail operation to perform.
type OperateMailOperation uint8

const (
	OperateMailDelete = iota + 1
	OperateMailLock
	OperateMailUnlock
	OpreateMailNull
	OperateMailAcquireItem
)

// MsgMhfOprtMail represents the MSG_MHF_OPRT_MAIL
type MsgMhfOprtMail struct {
	AckHandle uint32
	AccIndex  uint8
	Index     uint8
	Operation OperateMailOperation
	Data      []byte
	Amount    uint16
	ItemID    uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOprtMail) Opcode() network.PacketID {
	return network.MSG_MHF_OPRT_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfOprtMail) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.AccIndex = bf.ReadUint8()
	m.Index = bf.ReadUint8()
	m.Operation = OperateMailOperation(bf.ReadUint8())
	bf.ReadUint8() // Zeroed
	if m.Operation == OperateMailAcquireItem {
		m.Amount = bf.ReadUint16()
		m.ItemID = bf.ReadUint16()
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOprtMail) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
