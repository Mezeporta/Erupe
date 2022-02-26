package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysHideClient represents the MSG_SYS_HIDE_CLIENT
type MsgSysHideClient struct {
	Hide bool
	Unk0 uint16 // Hardcoded 0 in binary
	Unk1 uint8  // Hardcoded 0 in binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysHideClient) Opcode() network.PacketID {
	return network.MSG_SYS_HIDE_CLIENT
}

// Parse parses the packet from binary
func (m *MsgSysHideClient) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.Hide = bf.ReadBool()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysHideClient) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
