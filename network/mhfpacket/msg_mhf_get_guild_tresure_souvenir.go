package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetGuildTresureSouvenir represents the MSG_MHF_GET_GUILD_TRESURE_SOUVENIR
type MsgMhfGetGuildTresureSouvenir struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildTresureSouvenir) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_TRESURE_SOUVENIR
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildTresureSouvenir) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildTresureSouvenir) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
