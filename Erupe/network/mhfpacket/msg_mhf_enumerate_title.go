package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateTitle represents the MSG_MHF_ENUMERATE_TITLE
type MsgMhfEnumerateTitle struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateTitle) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_TITLE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateTitle) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateTitle) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
