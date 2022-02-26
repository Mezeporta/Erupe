package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGemInfo represents the MSG_MHF_GET_GEM_INFO
type MsgMhfGetGemInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGemInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GEM_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetGemInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGemInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
