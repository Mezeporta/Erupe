package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysGetObjectOwner represents the MSG_SYS_GET_OBJECT_OWNER
type MsgSysGetObjectOwner struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetObjectOwner) Opcode() network.PacketID {
	return network.MSG_SYS_GET_OBJECT_OWNER
}

// Parse parses the packet from binary
func (m *MsgSysGetObjectOwner) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetObjectOwner) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
