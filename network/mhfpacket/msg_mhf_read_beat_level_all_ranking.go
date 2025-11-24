package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfReadBeatLevelAllRanking represents the MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING
type MsgMhfReadBeatLevelAllRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevelAllRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevelAllRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevelAllRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
