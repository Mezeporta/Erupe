package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadOtomoAirou represents the MSG_MHF_LOAD_OTOMO_AIROU
type MsgMhfLoadOtomoAirou struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadOtomoAirou) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_OTOMO_AIROU
}

// Parse parses the packet from binary
func (m *MsgMhfLoadOtomoAirou) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadOtomoAirou) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
