package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsFirstQuestBonus represents the MSG_MHF_GET_UD_TACTICS_FIRST_QUEST_BONUS
type MsgMhfGetUdTacticsFirstQuestBonus struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsFirstQuestBonus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_FIRST_QUEST_BONUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsFirstQuestBonus) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsFirstQuestBonus) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
