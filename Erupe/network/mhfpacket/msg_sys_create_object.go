package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCreateObject represents the MSG_SYS_CREATE_OBJECT
type MsgSysCreateObject struct {
	AckHandle uint32
	X, Y, Z   float32
	Unk0      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateObject) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysCreateObject) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.X = bf.ReadFloat32()
	m.Y = bf.ReadFloat32()
	m.Z = bf.ReadFloat32()
	m.Unk0 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateObject) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
