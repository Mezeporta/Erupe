package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfChargeFesta represents the MSG_MHF_CHARGE_FESTA
type MsgMhfChargeFesta struct {
	AckHandle uint32
	FestaID   uint32
	GuildID   uint32
	Souls     int
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfChargeFesta) Opcode() network.PacketID {
	return network.MSG_MHF_CHARGE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfChargeFesta) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.FestaID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.Souls = 0
	for i := bf.ReadUint16(); i > 0; i-- {
		m.Souls += int(bf.ReadUint16())
	}
	_ = bf.ReadUint8() // Unk
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfChargeFesta) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
