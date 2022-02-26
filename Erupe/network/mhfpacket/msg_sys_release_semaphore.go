package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReleaseSemaphore represents the MSG_SYS_RELEASE_SEMAPHORE
type MsgSysReleaseSemaphore struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReleaseSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_RELEASE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysReleaseSemaphore) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReleaseSemaphore) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
