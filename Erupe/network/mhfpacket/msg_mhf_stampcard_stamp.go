package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStampcardStamp represents the MSG_MHF_STAMPCARD_STAMP
type MsgMhfStampcardStamp struct {
	// Field-size accurate.
	AckHandle uint32
	Unk0      uint16
	Unk1      uint16
	Unk2      uint16
	Unk3      uint16 // Hardcoded 0 in binary
	Unk4      uint32
	Unk5      uint32
	Unk6      uint32
	Unk7      uint32
	Unk8      uint32
	Unk9      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStampcardStamp) Opcode() network.PacketID {
	return network.MSG_MHF_STAMPCARD_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfStampcardStamp) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	m.Unk3 = bf.ReadUint16()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint32()
	m.Unk6 = bf.ReadUint32()
	m.Unk7 = bf.ReadUint32()
	m.Unk8 = bf.ReadUint32()
	m.Unk9 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStampcardStamp) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
