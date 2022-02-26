package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadMercenaryM represents the MSG_MHF_READ_MERCENARY_M
type MsgMhfReadMercenaryM struct{
	AckHandle   uint32
	CharID   uint32
	Unk0        uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMercenaryM) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MERCENARY_M
}

// Parse parses the packet from binary
func (m *MsgMhfReadMercenaryM) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMercenaryM) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
