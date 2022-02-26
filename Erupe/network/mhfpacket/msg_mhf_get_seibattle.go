package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetSeibattle represents the MSG_MHF_GET_SEIBATTLE
type MsgMhfGetSeibattle struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	Unk0      uint8
	Unk1      uint8
	Unk2      uint32 // Some shared ID with MSG_SYS_RECORD_LOG, world ID?
	Unk3      uint8
	Unk4      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetSeibattle) Opcode() network.PacketID {
	return network.MSG_MHF_GET_SEIBATTLE
}

// Parse parses the packet from binary
func (m *MsgMhfGetSeibattle) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint8()
	m.Unk4 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetSeibattle) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
