package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysSetObjectBinary represents the MSG_SYS_SET_OBJECT_BINARY
type MsgSysSetObjectBinary struct {
	ObjID          uint32
	DataSize       uint16
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_SET_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysSetObjectBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.ObjID = bf.ReadUint32()
	m.DataSize = bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetObjectBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
