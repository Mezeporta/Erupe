package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadBeatLevelMyRanking represents the MSG_MHF_READ_BEAT_LEVEL_MY_RANKING
type MsgMhfReadBeatLevelMyRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevelMyRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL_MY_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevelMyRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevelMyRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
