package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCancelGuildScout represents the MSG_MHF_CANCEL_GUILD_SCOUT
type MsgMhfCancelGuildScout struct {
	AckHandle    uint32
	InvitationID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCancelGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_CANCEL_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfCancelGuildScout) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.InvitationID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCancelGuildScout) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
