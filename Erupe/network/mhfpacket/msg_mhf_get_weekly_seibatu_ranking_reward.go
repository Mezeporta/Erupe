package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetWeeklySeibatuRankingReward represents the MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD
type MsgMhfGetWeeklySeibatuRankingReward struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetWeeklySeibatuRankingReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetWeeklySeibatuRankingReward) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetWeeklySeibatuRankingReward) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
