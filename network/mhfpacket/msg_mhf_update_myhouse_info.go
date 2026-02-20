package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateMyhouseInfo represents the MSG_MHF_UPDATE_MYHOUSE_INFO
type MsgMhfUpdateMyhouseInfo struct {
	AckHandle uint32
	Data      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateMyhouseInfo) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_MYHOUSE_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateMyhouseInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	if ctx.RealClientMode >= _config.G10 {
		m.Data = bf.ReadBytes(362)
	} else if ctx.RealClientMode >= _config.GG {
		m.Data = bf.ReadBytes(338)
	} else if ctx.RealClientMode >= _config.F5 {
		// G1 is a guess
		m.Data = bf.ReadBytes(314)
	} else {
		m.Data = bf.ReadBytes(290)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateMyhouseInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
