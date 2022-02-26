package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSetEnhancedMinidata represents the MSG_MHF_SET_ENHANCED_MINIDATA
type MsgMhfSetEnhancedMinidata struct {
	AckHandle      uint32
	Unk0           uint16 // Hardcoded 4 in the binary.
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetEnhancedMinidata) Opcode() network.PacketID {
	return network.MSG_MHF_SET_ENHANCED_MINIDATA
}

// Parse parses the packet from binary
func (m *MsgMhfSetEnhancedMinidata) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(0x400)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetEnhancedMinidata) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
