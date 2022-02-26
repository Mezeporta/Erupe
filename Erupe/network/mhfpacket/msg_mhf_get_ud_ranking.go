package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdRanking represents the MSG_MHF_GET_UD_RANKING
type MsgMhfGetUdRanking struct{
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdRanking) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
		m.AckHandle = bf.ReadUint32()
		return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
