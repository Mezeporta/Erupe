package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdNormaPresentList represents the MSG_MHF_GET_UD_NORMA_PRESENT_LIST
type MsgMhfGetUdNormaPresentList struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdNormaPresentList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_NORMA_PRESENT_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdNormaPresentList) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdNormaPresentList) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
