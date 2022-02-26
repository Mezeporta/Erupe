package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfExchangeFpoint2Item represents the MSG_MHF_EXCHANGE_FPOINT_2_ITEM
type MsgMhfExchangeFpoint2Item struct{
	AckHandle uint32
	ItemHash uint32
	ItemType uint16
	ItemId uint16
	Quantity byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfExchangeFpoint2Item) Opcode() network.PacketID {
	return network.MSG_MHF_EXCHANGE_FPOINT_2_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfExchangeFpoint2Item) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.ItemHash = bf.ReadUint32()
	m.ItemType = bf.ReadUint16()
	m.ItemId = bf.ReadUint16()
	m.Quantity = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfExchangeFpoint2Item) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
