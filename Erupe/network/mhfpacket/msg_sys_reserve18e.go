package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve18E represents the MSG_SYS_reserve18E
type MsgSysReserve18E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve18E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve18E
}

// Parse parses the packet from binary
func (m *MsgSysReserve18E) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve18E) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
