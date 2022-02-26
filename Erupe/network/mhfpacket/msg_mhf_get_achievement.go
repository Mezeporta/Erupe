package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetAchievement represents the MSG_MHF_GET_ACHIEVEMENT
type MsgMhfGetAchievement struct{
	AckHandle      uint32
	Unk0           uint16 // id?
	Unk1           uint32 // char?
	Unk2           uint32 // pad?
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfGetAchievement) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetAchievement) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
