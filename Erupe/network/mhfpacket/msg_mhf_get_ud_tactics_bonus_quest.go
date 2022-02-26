package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsBonusQuest represents the MSG_MHF_GET_UD_TACTICS_BONUS_QUEST
type MsgMhfGetUdTacticsBonusQuest struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsBonusQuest) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_BONUS_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsBonusQuest) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsBonusQuest) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
