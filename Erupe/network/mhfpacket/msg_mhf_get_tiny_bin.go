package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetTinyBin represents the MSG_MHF_GET_TINY_BIN
type MsgMhfGetTinyBin struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	Unk0      uint16
	Unk1      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetTinyBin) Opcode() network.PacketID {
	return network.MSG_MHF_GET_TINY_BIN
}

// Parse parses the packet from binary
func (m *MsgMhfGetTinyBin) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetTinyBin) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
