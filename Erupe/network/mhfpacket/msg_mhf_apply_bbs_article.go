package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfApplyBbsArticle represents the MSG_MHF_APPLY_BBS_ARTICLE
type MsgMhfApplyBbsArticle struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyBbsArticle) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_BBS_ARTICLE
}

// Parse parses the packet from binary
func (m *MsgMhfApplyBbsArticle) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyBbsArticle) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
