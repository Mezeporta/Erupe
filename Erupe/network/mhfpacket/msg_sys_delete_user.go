package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysDeleteUser represents the MSG_SYS_DELETE_USER
type MsgSysDeleteUser struct {
	CharID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDeleteUser) Opcode() network.PacketID {
	return network.MSG_SYS_DELETE_USER
}

// Parse parses the packet from binary
func (m *MsgSysDeleteUser) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysDeleteUser) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.CharID)

	return nil
}
