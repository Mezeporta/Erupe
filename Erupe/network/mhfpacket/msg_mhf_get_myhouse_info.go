package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetMyhouseInfo represents the MSG_MHF_GET_MYHOUSE_INFO
type MsgMhfGetMyhouseInfo struct {
	AckHandle uint32
	Unk0      uint32

	// No idea why it would send a buffer of data on a _GET_, but w/e.
	DataSize       uint8
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetMyhouseInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_MYHOUSE_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetMyhouseInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.DataSize = bf.ReadUint8()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetMyhouseInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
