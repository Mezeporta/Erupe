package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetUdGuildMapInfo represents the MSG_MHF_GET_UD_GUILD_MAP_INFO
type MsgMhfGetUdGuildMapInfo struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdGuildMapInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_GUILD_MAP_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdGuildMapInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdGuildMapInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
