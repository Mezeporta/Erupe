package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadHouse represents the MSG_MHF_LOAD_HOUSE
type MsgMhfLoadHouse struct {
	AckHandle      uint32
	Unk0           uint32
	Unk1           uint8
	Unk2           uint8
	Unk3           uint16 // Hardcoded 0 in binary
	DataSize       uint8
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadHouse) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadHouse) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint8()
	m.Unk3 = bf.ReadUint16()
	m.DataSize = bf.ReadUint8()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadHouse) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
