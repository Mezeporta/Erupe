package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetRestrictionEvent represents the MSG_MHF_SET_RESTRICTION_EVENT
type MsgMhfSetRestrictionEvent struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetRestrictionEvent) Opcode() network.PacketID {
	return network.MSG_MHF_SET_RESTRICTION_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfSetRestrictionEvent) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetRestrictionEvent) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
