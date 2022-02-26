package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLockGlobalSema represents the MSG_SYS_LOCK_GLOBAL_SEMA
type MsgSysLockGlobalSema struct {
	AckHandle             uint32
	UserIDLength					uint16
	ServerChannelIDLength	uint16
	UserIDString          string
	ServerChannelIDString string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLockGlobalSema) Opcode() network.PacketID {
	return network.MSG_SYS_LOCK_GLOBAL_SEMA
}

// Parse parses the packet from binary
func (m *MsgSysLockGlobalSema) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.UserIDLength = bf.ReadUint16()
	m.ServerChannelIDLength = bf.ReadUint16()
	m.UserIDString = string(bf.ReadBytes(uint(m.UserIDLength)))
	m.ServerChannelIDString = string(bf.ReadBytes(uint(m.ServerChannelIDLength)))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLockGlobalSema) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
