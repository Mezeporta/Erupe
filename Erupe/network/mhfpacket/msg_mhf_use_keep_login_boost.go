package mhfpacket

import (
	"errors"

	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUseKeepLoginBoost represents the MSG_MHF_USE_KEEP_LOGIN_BOOST
type MsgMhfUseKeepLoginBoost struct {
	AckHandle     uint32
	BoostWeekUsed uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUseKeepLoginBoost) Opcode() network.PacketID {
	return network.MSG_MHF_USE_KEEP_LOGIN_BOOST
}

// Parse parses the packet from binary
func (m *MsgMhfUseKeepLoginBoost) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.BoostWeekUsed = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUseKeepLoginBoost) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
