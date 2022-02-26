package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetRengokuBinary represents the MSG_MHF_GET_RENGOKU_BINARY
type MsgMhfGetRengokuBinary struct {
	AckHandle uint32
	Unk0      uint8 // Hardcoded 0 in binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRengokuBinary) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RENGOKU_BINARY
}

// Parse parses the packet from binary
func (m *MsgMhfGetRengokuBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRengokuBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
