package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// The server sends different responses based on these values.
const (
	TowerInfoTypeUnk0 = iota
	TowerInfoTypeTowerRankPoint
	TowerInfoTypeGetOwnTowerSkill
	TowerInfoTypeUnk3
	TowerInfoTypeTowerTouhaHistory
	TowerInfoTypeUnk5
)

// MsgMhfGetTowerInfo represents the MSG_MHF_GET_TOWER_INFO
type MsgMhfGetTowerInfo struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	InfoType  uint32 // Requested response type
	Unk0      uint32
	Unk1      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetTowerInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_TOWER_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetTowerInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.InfoType = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetTowerInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
