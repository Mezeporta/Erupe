package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireMonthlyItem represents the MSG_MHF_ACQUIRE_MONTHLY_ITEM
type MsgMhfAcquireMonthlyItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireMonthlyItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_MONTHLY_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireMonthlyItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireMonthlyItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
