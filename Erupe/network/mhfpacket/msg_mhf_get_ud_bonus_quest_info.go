package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdBonusQuestInfo represents the MSG_MHF_GET_UD_BONUS_QUEST_INFO
type MsgMhfGetUdBonusQuestInfo struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdBonusQuestInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_BONUS_QUEST_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdBonusQuestInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdBonusQuestInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
