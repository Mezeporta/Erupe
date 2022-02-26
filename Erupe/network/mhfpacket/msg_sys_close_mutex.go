package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCloseMutex represents the MSG_SYS_CLOSE_MUTEX
type MsgSysCloseMutex struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCloseMutex) Opcode() network.PacketID {
	return network.MSG_SYS_CLOSE_MUTEX
}

// Parse parses the packet from binary
func (m *MsgSysCloseMutex) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCloseMutex) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
