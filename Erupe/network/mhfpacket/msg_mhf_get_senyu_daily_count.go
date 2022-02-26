package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetSenyuDailyCount represents the MSG_MHF_GET_SENYU_DAILY_COUNT
type MsgMhfGetSenyuDailyCount struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetSenyuDailyCount) Opcode() network.PacketID {
	return network.MSG_MHF_GET_SENYU_DAILY_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfGetSenyuDailyCount) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetSenyuDailyCount) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
