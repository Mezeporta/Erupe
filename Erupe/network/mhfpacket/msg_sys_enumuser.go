package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysEnumuser represents the MSG_SYS_ENUMUSER
type MsgSysEnumuser struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnumuser) Opcode() network.PacketID {
	return network.MSG_SYS_ENUMUSER
}

// Parse parses the packet from binary
func (m *MsgSysEnumuser) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnumuser) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
