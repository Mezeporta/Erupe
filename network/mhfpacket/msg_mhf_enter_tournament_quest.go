package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnterTournamentQuest represents the MSG_MHF_ENTER_TOURNAMENT_QUEST
type MsgMhfEnterTournamentQuest struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnterTournamentQuest) Opcode() network.PacketID {
	return network.MSG_MHF_ENTER_TOURNAMENT_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfEnterTournamentQuest) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnterTournamentQuest) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
