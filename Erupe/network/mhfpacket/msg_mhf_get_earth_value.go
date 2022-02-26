package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetEarthValue represents the MSG_MHF_GET_EARTH_VALUE
type MsgMhfGetEarthValue struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	ReqType      uint32
	Unk3      uint32
	Unk4      uint32
	Unk5      uint32
	Unk6      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEarthValue) Opcode() network.PacketID {
	return network.MSG_MHF_GET_EARTH_VALUE
}

// Parse parses the packet from binary
func (m *MsgMhfGetEarthValue) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.ReqType = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint32()
	m.Unk6 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEarthValue) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
