package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPlayStepupGacha represents the MSG_MHF_PLAY_STEPUP_GACHA
type MsgMhfPlayStepupGacha struct{
	AckHandle uint32
	GachaHash uint32
	RollType uint8
	CurrencyMode uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPlayStepupGacha) Opcode() network.PacketID {
	return network.MSG_MHF_PLAY_STEPUP_GACHA
}

// Parse parses the packet from binary
func (m *MsgMhfPlayStepupGacha) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GachaHash = bf.ReadUint32()
	m.RollType = bf.ReadUint8()
	m.CurrencyMode = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPlayStepupGacha) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
