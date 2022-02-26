package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve4F represents the MSG_SYS_reserve4F
type MsgSysReserve4F struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve4F) Opcode() network.PacketID {
	return network.MSG_SYS_reserve4F
}

// Parse parses the packet from binary
func (m *MsgSysReserve4F) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve4F) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
