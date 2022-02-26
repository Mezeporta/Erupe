package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve06 represents the MSG_SYS_reserve06
type MsgSysReserve06 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve06) Opcode() network.PacketID {
	return network.MSG_SYS_reserve06
}

// Parse parses the packet from binary
func (m *MsgSysReserve06) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve06) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
