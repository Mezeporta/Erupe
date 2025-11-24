package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateBeatLevel represents the MSG_MHF_UPDATE_BEAT_LEVEL
type MsgMhfUpdateBeatLevel struct {
	AckHandle   uint32
	Unk1        uint32
	Unk2        uint32
	MonsterData []byte
	Unk3        uint8
	Unk4        uint32
	Unk5        uint16
	Unk6        uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateBeatLevel) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_BEAT_LEVEL
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateBeatLevel) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.MonsterData = bf.ReadBytes(uint(120))
	m.Unk3 = bf.ReadUint8()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint16()
	m.Unk6 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateBeatLevel) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
