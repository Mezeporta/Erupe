package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve204 represents the MSG_SYS_reserve204
type MsgSysReserve204 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve204) Opcode() network.PacketID {
	return network.MSG_SYS_reserve204
}

// Parse parses the packet from binary
func (m *MsgSysReserve204) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve204) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
