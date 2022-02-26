package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfTransitMessage represents the MSG_MHF_TRANSIT_MESSAGE
type MsgMhfTransitMessage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfTransitMessage) Opcode() network.PacketID {
	return network.MSG_MHF_TRANSIT_MESSAGE
}

// Parse parses the packet from binary
func (m *MsgMhfTransitMessage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfTransitMessage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
