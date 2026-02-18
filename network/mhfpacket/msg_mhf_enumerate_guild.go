package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// EnumerateGuildType specifies the search/sort criteria for guild enumeration.
type EnumerateGuildType uint8

const (
	ENUMERATE_GUILD_UNKNOWN = iota
	ENUMERATE_GUILD_TYPE_GUILD_NAME
	ENUMERATE_GUILD_TYPE_LEADER_NAME
	ENUMERATE_GUILD_TYPE_LEADER_ID
	ENUMERATE_GUILD_TYPE_ORDER_MEMBERS
	ENUMERATE_GUILD_TYPE_ORDER_REGISTRATION
	ENUMERATE_GUILD_TYPE_ORDER_RANK
	ENUMERATE_GUILD_TYPE_MOTTO
	ENUMERATE_GUILD_TYPE_RECRUITING
	ENUMERATE_ALLIANCE_TYPE_ALLIANCE_NAME
	ENUMERATE_ALLIANCE_TYPE_LEADER_NAME
	ENUMERATE_ALLIANCE_TYPE_LEADER_ID
	ENUMERATE_ALLIANCE_TYPE_ORDER_MEMBERS
	ENUMERATE_ALLIANCE_TYPE_ORDER_REGISTRATION
)

// MsgMhfEnumerateGuild represents the MSG_MHF_ENUMERATE_GUILD
type MsgMhfEnumerateGuild struct {
	AckHandle uint32
	Type      EnumerateGuildType
	Page      uint8
	Sorting   bool
	Data1     *byteframe.ByteFrame
	Data2     *byteframe.ByteFrame
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Type = EnumerateGuildType(bf.ReadUint8())
	m.Page = bf.ReadUint8()
	m.Sorting = bf.ReadBool()
	bf.ReadUint8() // Zeroed
	m.Data1 = byteframe.NewByteFrameFromBytes(bf.ReadBytes(4))
	bf.ReadUint16() // Zeroed
	dataLen := uint(bf.ReadUint8())
	bf.ReadUint8() // Zeroed
	m.Data2 = byteframe.NewByteFrameFromBytes(bf.ReadBytes(dataLen))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
