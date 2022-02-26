package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostTenrouirai represents the MSG_MHF_POST_TENROUIRAI
type MsgMhfPostTenrouirai struct{
	AckHandle      uint32
	Unk0           uint16
	Unk1           uint32
	Unk2           uint32
	Unk3           uint32
	Unk4           uint32
	Unk5           uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostTenrouirai) Opcode() network.PacketID {
	return network.MSG_MHF_POST_TENROUIRAI
}

// Parse parses the packet from binary
func (m *MsgMhfPostTenrouirai) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostTenrouirai) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
