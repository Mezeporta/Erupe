package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/common/bfutil"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCheckSemaphore represents the MSG_SYS_CHECK_SEMAPHORE
type MsgSysCheckSemaphore struct{
	AckHandle uint32
	StageID   string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCheckSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_CHECK_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysCheckSemaphore) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCheckSemaphore) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}