package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReceiveCafeDurationBonus represents the MSG_MHF_RECEIVE_CAFE_DURATION_BONUS
type MsgMhfReceiveCafeDurationBonus struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReceiveCafeDurationBonus) Opcode() network.PacketID {
	return network.MSG_MHF_RECEIVE_CAFE_DURATION_BONUS
}

// Parse parses the packet from binary
func (m *MsgMhfReceiveCafeDurationBonus) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReceiveCafeDurationBonus) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
