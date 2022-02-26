package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSavedata represents the MSG_MHF_SAVEDATA
type MsgMhfSavedata struct {
	AckHandle      uint32
	AllocMemSize   uint32
	SaveType       uint8 // Either 1 or 2, representing a true or false value for some reason.
	Unk1           uint32
	DataSize       uint32
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavedata) Opcode() network.PacketID {
	return network.MSG_MHF_SAVEDATA
}

// Parse parses the packet from binary
func (m *MsgMhfSavedata) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.AllocMemSize = bf.ReadUint32()
	m.SaveType = bf.ReadUint8()
	m.Unk1 = bf.ReadUint32()
	m.DataSize = bf.ReadUint32()
	if m.DataSize == 0 { // seems to be used when DataSize = 0 rather than on savetype?
		m.RawDataPayload = bf.ReadBytes(uint(m.AllocMemSize))
	} else {
		m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavedata) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
