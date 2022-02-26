package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetRejectGuildScout represents the MSG_MHF_SET_REJECT_GUILD_SCOUT
type MsgMhfSetRejectGuildScout struct {
	AckHandle uint32
	Reject    bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetRejectGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_SET_REJECT_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfSetRejectGuildScout) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Reject = bf.ReadBool()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetRejectGuildScout) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
