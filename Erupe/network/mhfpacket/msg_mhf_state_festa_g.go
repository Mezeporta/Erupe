package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStateFestaG represents the MSG_MHF_STATE_FESTA_G
type MsgMhfStateFestaG struct {
	AckHandle uint32
	Unk0      uint32 // Shared ID of something.
	Unk1      uint32
	Unk2      uint16 // Hardcoded 0 in the binary.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStateFestaG) Opcode() network.PacketID {
	return network.MSG_MHF_STATE_FESTA_G
}

// Parse parses the packet from binary
func (m *MsgMhfStateFestaG) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint16()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStateFestaG) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
