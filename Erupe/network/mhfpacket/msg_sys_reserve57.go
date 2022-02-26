package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve57 represents the MSG_SYS_reserve57
type MsgSysReserve57 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve57) Opcode() network.PacketID {
	return network.MSG_SYS_reserve57
}

// Parse parses the packet from binary
func (m *MsgSysReserve57) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve57) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
