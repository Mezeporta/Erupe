package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve20B represents the MSG_SYS_reserve20B
type MsgSysReserve20B struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve20B) Opcode() network.PacketID {
	return network.MSG_SYS_reserve20B
}

// Parse parses the packet from binary
func (m *MsgSysReserve20B) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve20B) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
