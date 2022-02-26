package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPlayBoxGacha represents the MSG_MHF_PLAY_BOX_GACHA
type MsgMhfPlayBoxGacha struct{
	AckHandle uint32
	GachaHash uint32
	RollType uint8
	CurrencyMode uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPlayBoxGacha) Opcode() network.PacketID {
	return network.MSG_MHF_PLAY_BOX_GACHA
}

// Parse parses the packet from binary
func (m *MsgMhfPlayBoxGacha) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.GachaHash = bf.ReadUint32()
	m.RollType = bf.ReadUint8()
	m.CurrencyMode = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPlayBoxGacha) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
