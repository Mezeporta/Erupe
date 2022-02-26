package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetPaperData represents the MSG_MHF_GET_PAPER_DATA
type MsgMhfGetPaperData struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetPaperData) Opcode() network.PacketID {
	return network.MSG_MHF_GET_PAPER_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfGetPaperData) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetPaperData) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
