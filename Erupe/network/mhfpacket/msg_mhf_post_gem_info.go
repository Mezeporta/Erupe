package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostGemInfo represents the MSG_MHF_POST_GEM_INFO
type MsgMhfPostGemInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostGemInfo) Opcode() network.PacketID {
	return network.MSG_MHF_POST_GEM_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfPostGemInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostGemInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
