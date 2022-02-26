package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysAuthTerminal represents the MSG_SYS_AUTH_TERMINAL
type MsgSysAuthTerminal struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAuthTerminal) Opcode() network.PacketID {
	return network.MSG_SYS_AUTH_TERMINAL
}

// Parse parses the packet from binary
func (m *MsgSysAuthTerminal) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysAuthTerminal) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
