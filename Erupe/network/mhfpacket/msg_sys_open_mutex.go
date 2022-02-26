package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysOpenMutex represents the MSG_SYS_OPEN_MUTEX
type MsgSysOpenMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysOpenMutex) Opcode() network.PacketID {
	return network.MSG_SYS_OPEN_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysOpenMutex) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysOpenMutex) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
