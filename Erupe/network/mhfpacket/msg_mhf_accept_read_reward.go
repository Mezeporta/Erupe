package mhfpacket

import (
	"errors"

	"github.com/Solenataris/Erupe/network/clientctx"

	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcceptReadReward represents the MSG_MHF_ACCEPT_READ_REWARD
type MsgMhfAcceptReadReward struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcceptReadReward) Opcode() network.PacketID {
	return network.MSG_MHF_ACCEPT_READ_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfAcceptReadReward) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcceptReadReward) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
