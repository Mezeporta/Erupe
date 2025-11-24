package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddRewardSongCount represents the MSG_MHF_ADD_REWARD_SONG_COUNT
type MsgMhfAddRewardSongCount struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddRewardSongCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_REWARD_SONG_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfAddRewardSongCount) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddRewardSongCount) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
