package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve18B represents the MSG_SYS_reserve18B
type MsgSysReserve18B struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve18B) Opcode() network.PacketID {
	return network.MSG_SYS_reserve18B
}

// Parse parses the packet from binary
func (m *MsgSysReserve18B) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve18B) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
