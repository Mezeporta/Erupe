package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve0D represents the MSG_SYS_reserve0D
type MsgSysReserve0D struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve0D) Opcode() network.PacketID {
	return network.MSG_SYS_reserve0D
}

// Parse parses the packet from binary
func (m *MsgSysReserve0D) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve0D) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
