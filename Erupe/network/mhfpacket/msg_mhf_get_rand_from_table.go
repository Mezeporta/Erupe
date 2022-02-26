package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetRandFromTable represents the MSG_MHF_GET_RAND_FROM_TABLE
type MsgMhfGetRandFromTable struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRandFromTable) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RAND_FROM_TABLE
}

// Parse parses the packet from binary
func (m *MsgMhfGetRandFromTable) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRandFromTable) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
