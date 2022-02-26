package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAddGuildWeeklyBonusExceptionalUser represents the MSG_MHF_ADD_GUILD_WEEKLY_BONUS_EXCEPTIONAL_USER
type MsgMhfAddGuildWeeklyBonusExceptionalUser struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddGuildWeeklyBonusExceptionalUser) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_GUILD_WEEKLY_BONUS_EXCEPTIONAL_USER
}

// Parse parses the packet from binary
func (m *MsgMhfAddGuildWeeklyBonusExceptionalUser) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddGuildWeeklyBonusExceptionalUser) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
