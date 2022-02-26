package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfKickExportForce represents the MSG_MHF_KICK_EXPORT_FORCE
type MsgMhfKickExportForce struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfKickExportForce) Opcode() network.PacketID {
	return network.MSG_MHF_KICK_EXPORT_FORCE
}

// Parse parses the packet from binary
func (m *MsgMhfKickExportForce) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfKickExportForce) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
