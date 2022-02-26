package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveOtomoAirou represents the MSG_MHF_SAVE_OTOMO_AIROU
type MsgMhfSaveOtomoAirou struct {
	AckHandle      uint32
	DataSize       uint16
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveOtomoAirou) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_OTOMO_AIROU
}

// Parse parses the packet from binary
func (m *MsgMhfSaveOtomoAirou) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.DataSize = bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveOtomoAirou) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
