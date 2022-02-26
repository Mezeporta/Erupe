package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoFesta represents the MSG_MHF_INFO_FESTA
type MsgMhfInfoFesta struct {
	AckHandle uint32
	Unk0      uint16 // Hardcoded 0 in the binary
	Unk1      uint16 // Hardcoded 0 in the binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoFesta) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfInfoFesta) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoFesta) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
