package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCheckMonthlyItem represents the MSG_MHF_CHECK_MONTHLY_ITEM
type MsgMhfCheckMonthlyItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckMonthlyItem) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_MONTHLY_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfCheckMonthlyItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCheckMonthlyItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
