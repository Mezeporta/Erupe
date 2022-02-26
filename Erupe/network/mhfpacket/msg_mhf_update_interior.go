package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateInterior represents the MSG_MHF_UPDATE_INTERIOR
type MsgMhfUpdateInterior struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateInterior) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_INTERIOR
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateInterior) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateInterior) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
