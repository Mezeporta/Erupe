package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfInfoTournament represents the MSG_MHF_INFO_TOURNAMENT
type MsgMhfInfoTournament struct {
	AckHandle uint32
	QueryType    uint8
	TournamentID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoTournament) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfInfoTournament) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.QueryType = bf.ReadUint8()
	m.TournamentID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoTournament) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
