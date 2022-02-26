package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostTowerInfo represents the MSG_MHF_POST_TOWER_INFO
type MsgMhfPostTowerInfo struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      uint32
	Unk3      uint32
	Unk4      uint32
	Unk5      uint32
	Unk6      uint32
	Unk7      uint32
	Unk8      uint32
	Unk9      uint32
	Unk10     uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostTowerInfo) Opcode() network.PacketID {
	return network.MSG_MHF_POST_TOWER_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfPostTowerInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint32()
	m.Unk6 = bf.ReadUint32()
	m.Unk7 = bf.ReadUint32()
	m.Unk8 = bf.ReadUint32()
	m.Unk9 = bf.ReadUint32()
	m.Unk10 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostTowerInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
