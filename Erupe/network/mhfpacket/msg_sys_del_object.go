package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysDelObject represents the MSG_SYS_DEL_OBJECT
type MsgSysDelObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDelObject) Opcode() network.PacketID {
	return network.MSG_SYS_DEL_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysDelObject) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysDelObject) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
