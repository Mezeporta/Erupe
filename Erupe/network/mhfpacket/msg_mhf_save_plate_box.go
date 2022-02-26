package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSavePlateBox represents the MSG_MHF_SAVE_PLATE_BOX
type MsgMhfSavePlateBox struct {
	AckHandle      uint32
	DataSize       uint32
	IsDataDiff     bool
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavePlateBox) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_PLATE_BOX
}

// Parse parses the packet from binary
func (m *MsgMhfSavePlateBox) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.DataSize = bf.ReadUint32()
	m.IsDataDiff = bf.ReadBool()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavePlateBox) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
