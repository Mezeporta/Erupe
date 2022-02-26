package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysBackStage represents the MSG_SYS_BACK_STAGE
type MsgSysBackStage struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysBackStage) Opcode() network.PacketID {
	return network.MSG_SYS_BACK_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysBackStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysBackStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
