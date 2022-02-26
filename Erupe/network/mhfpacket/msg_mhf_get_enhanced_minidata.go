package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetEnhancedMinidata represents the MSG_MHF_GET_ENHANCED_MINIDATA
type MsgMhfGetEnhancedMinidata struct {
	AckHandle uint32
	CharID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEnhancedMinidata) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ENHANCED_MINIDATA
}

// Parse parses the packet from binary
func (m *MsgMhfGetEnhancedMinidata) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEnhancedMinidata) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
