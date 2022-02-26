package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetFpointExchangeList represents the MSG_MHF_GET_FPOINT_EXCHANGE_LIST
type MsgMhfGetFpointExchangeList struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetFpointExchangeList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_FPOINT_EXCHANGE_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetFpointExchangeList) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetFpointExchangeList) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
