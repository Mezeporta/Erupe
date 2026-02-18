package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// OperateGuildAction identifies the guild management action to perform.
type OperateGuildAction uint8

const (
	OperateGuildDisband = iota + 1
	OperateGuildApply
	OperateGuildLeave
	OperateGuildResign
	OperateGuildSetApplicationDeny
	OperateGuildSetApplicationAllow
	OperateGuildSetAvoidLeadershipTrue
	OperateGuildSetAvoidLeadershipFalse
	OperateGuildUpdateComment
	OperateGuildDonateRank
	OperateGuildUpdateMotto
	OperateGuildRenamePugi1
	OperateGuildRenamePugi2
	OperateGuildRenamePugi3
	OperateGuildChangePugi1
	OperateGuildChangePugi2
	OperateGuildChangePugi3
	OperateGuildUnlockOutfit
	OperateGuildDonateRoom
	OperateGuildGraduateRookie
	OperateGuildDonateEvent
	OperateGuildEventExchange
	OperateGuildUnknown // I don't think this op exists
	OperateGuildGraduateReturn
	OperateGuildChangeDivaPugi1
	OperateGuildChangeDivaPugi2
	OperateGuildChangeDivaPugi3
)

// MsgMhfOperateGuild represents the MSG_MHF_OPERATE_GUILD
type MsgMhfOperateGuild struct {
	AckHandle uint32
	GuildID   uint32
	Action    OperateGuildAction
	Data1     *byteframe.ByteFrame
	Data2     *byteframe.ByteFrame
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfOperateGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.Action = OperateGuildAction(bf.ReadUint8())
	dataLen := uint(bf.ReadUint8())
	m.Data1 = byteframe.NewByteFrameFromBytes(bf.ReadBytes(4))
	m.Data2 = byteframe.NewByteFrameFromBytes(bf.ReadBytes(dataLen))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
