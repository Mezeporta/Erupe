package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateUseTrendWeaponLog represents the MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG
type MsgMhfUpdateUseTrendWeaponLog struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint16 // Weapon/item ID probably?
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateUseTrendWeaponLog) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateUseTrendWeaponLog) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateUseTrendWeaponLog) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
