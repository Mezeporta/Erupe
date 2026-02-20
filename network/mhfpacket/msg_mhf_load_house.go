package mhfpacket

import (
	"errors"
	"erupe-ce/common/stringsupport"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfLoadHouse represents the MSG_MHF_LOAD_HOUSE
type MsgMhfLoadHouse struct {
	AckHandle   uint32
	CharID      uint32
	Destination uint8
	// False if already in hosts My Series, in case host updates PW
	CheckPass bool
	Password  string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadHouse) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadHouse) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	m.Destination = bf.ReadUint8()
	m.CheckPass = bf.ReadBool()
	bf.ReadUint16() // Zeroed
	bf.ReadUint8()  // Password length
	m.Password, _ = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadHouse) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
