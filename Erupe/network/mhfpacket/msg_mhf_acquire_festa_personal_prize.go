package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireFestaPersonalPrize represents the MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE
type MsgMhfAcquireFestaPersonalPrize struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireFestaPersonalPrize) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireFestaPersonalPrize) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireFestaPersonalPrize) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
