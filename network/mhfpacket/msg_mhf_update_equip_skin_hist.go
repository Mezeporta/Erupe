package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateEquipSkinHist represents the MSG_MHF_UPDATE_EQUIP_SKIN_HIST
type MsgMhfUpdateEquipSkinHist struct {
	AckHandle uint32
	MogType   uint8
	ArmourID  uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateEquipSkinHist) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_EQUIP_SKIN_HIST
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateEquipSkinHist) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.MogType = bf.ReadUint8()
	m.ArmourID = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateEquipSkinHist) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
