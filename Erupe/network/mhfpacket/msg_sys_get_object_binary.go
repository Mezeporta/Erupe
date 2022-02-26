package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysGetObjectBinary represents the MSG_SYS_GET_OBJECT_BINARY
type MsgSysGetObjectBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_GET_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysGetObjectBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetObjectBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
