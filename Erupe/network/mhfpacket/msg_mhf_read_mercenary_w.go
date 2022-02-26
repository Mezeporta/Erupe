package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadMercenaryW represents the MSG_MHF_READ_MERCENARY_W
type MsgMhfReadMercenaryW struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint8
	Unk2      uint16 // Hardcoded 0 in the binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMercenaryW) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MERCENARY_W
}

// Parse parses the packet from binary
func (m *MsgMhfReadMercenaryW) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMercenaryW) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
