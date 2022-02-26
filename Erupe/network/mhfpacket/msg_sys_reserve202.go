package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve202 represents the MSG_SYS_reserve202
type MsgSysReserve202 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve202) Opcode() network.PacketID {
	return network.MSG_SYS_reserve202
}

// Parse parses the packet from binary
func (m *MsgSysReserve202) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve202) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
