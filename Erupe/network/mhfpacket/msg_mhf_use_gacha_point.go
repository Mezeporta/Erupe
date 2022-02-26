package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUseGachaPoint represents the MSG_MHF_USE_GACHA_POINT
type MsgMhfUseGachaPoint struct{
	AckHandle      uint32
	Unk0           uint16 // padding?
	TrialCoins     uint32
	PremiumCoins   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUseGachaPoint) Opcode() network.PacketID {
	return network.MSG_MHF_USE_GACHA_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfUseGachaPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.TrialCoins = bf.ReadUint32()
	m.PremiumCoins = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUseGachaPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
