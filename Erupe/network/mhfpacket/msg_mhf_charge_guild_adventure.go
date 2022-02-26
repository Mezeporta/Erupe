package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfChargeGuildAdventure represents the MSG_MHF_CHARGE_GUILD_ADVENTURE
type MsgMhfChargeGuildAdventure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfChargeGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_CHARGE_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfChargeGuildAdventure) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfChargeGuildAdventure) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
