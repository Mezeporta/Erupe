package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type OperateGuildMemberAction uint8

const (
	_ = iota
	OPERATE_GUILD_MEMBER_ACTION_ACCEPT
	OPERATE_GUILD_MEMBER_ACTION_REJECT
	OPERATE_GUILD_MEMBER_ACTION_KICK
)

// MsgMhfOperateGuildMember represents the MSG_MHF_OPERATE_GUILD_MEMBER
type MsgMhfOperateGuildMember struct {
	AckHandle uint32
	GuildID   uint32
	CharID    uint32
	Action    uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfOperateGuildMember) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	m.Action = bf.ReadUint8()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateGuildMember) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
