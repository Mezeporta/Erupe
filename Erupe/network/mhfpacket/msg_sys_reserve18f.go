package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve18F represents the MSG_SYS_reserve18F
type MsgSysReserve18F struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve18F) Opcode() network.PacketID {
	return network.MSG_SYS_reserve18F
}

// Parse parses the packet from binary
func (m *MsgSysReserve18F) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve18F) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
