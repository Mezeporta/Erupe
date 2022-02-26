package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfRegistGuildTresure represents the MSG_MHF_REGIST_GUILD_TRESURE
type MsgMhfRegistGuildTresure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildTresure) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_TRESURE
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildTresure) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildTresure) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
