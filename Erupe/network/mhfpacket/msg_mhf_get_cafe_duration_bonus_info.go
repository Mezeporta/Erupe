package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetCafeDurationBonusInfo represents the MSG_MHF_GET_CAFE_DURATION_BONUS_INFO
type MsgMhfGetCafeDurationBonusInfo struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetCafeDurationBonusInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_CAFE_DURATION_BONUS_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetCafeDurationBonusInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetCafeDurationBonusInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
