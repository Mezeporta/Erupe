package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAnswerGuildScout represents the MSG_MHF_ANSWER_GUILD_SCOUT
type MsgMhfAnswerGuildScout struct {
	AckHandle uint32
	LeaderID  uint32
	Answer    bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAnswerGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_ANSWER_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfAnswerGuildScout) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.LeaderID = bf.ReadUint32()
	m.Answer = bf.ReadBool()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAnswerGuildScout) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
