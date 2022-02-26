package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStampcardPrize represents the MSG_MHF_STAMPCARD_PRIZE
type MsgMhfStampcardPrize struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStampcardPrize) Opcode() network.PacketID {
	return network.MSG_MHF_STAMPCARD_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfStampcardPrize) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStampcardPrize) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
