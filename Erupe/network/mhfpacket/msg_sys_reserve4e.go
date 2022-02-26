package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve4E represents the MSG_SYS_reserve4E
type MsgSysReserve4E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve4E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve4E
}

// Parse parses the packet from binary
func (m *MsgSysReserve4E) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve4E) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
