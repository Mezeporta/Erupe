package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUseUdShopCoin represents the MSG_MHF_USE_UD_SHOP_COIN
type MsgMhfUseUdShopCoin struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUseUdShopCoin) Opcode() network.PacketID {
	return network.MSG_MHF_USE_UD_SHOP_COIN
}

// Parse parses the packet from binary
func (m *MsgMhfUseUdShopCoin) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUseUdShopCoin) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
