package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUnreserveSrg represents the MSG_MHF_UNRESERVE_SRG
type MsgMhfUnreserveSrg struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUnreserveSrg) Opcode() network.PacketID {
	return network.MSG_MHF_UNRESERVE_SRG
}

// Parse parses the packet from binary
func (m *MsgMhfUnreserveSrg) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUnreserveSrg) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
