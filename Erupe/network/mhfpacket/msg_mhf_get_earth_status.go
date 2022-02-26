package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetEarthStatus represents the MSG_MHF_GET_EARTH_STATUS
type MsgMhfGetEarthStatus struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEarthStatus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_EARTH_STATUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetEarthStatus) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEarthStatus) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
