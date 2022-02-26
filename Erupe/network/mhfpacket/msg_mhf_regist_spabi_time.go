package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfRegistSpabiTime represents the MSG_MHF_REGIST_SPABI_TIME
type MsgMhfRegistSpabiTime struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistSpabiTime) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_SPABI_TIME
}

// Parse parses the packet from binary
func (m *MsgMhfRegistSpabiTime) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistSpabiTime) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
