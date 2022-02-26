package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysRecordLog represents the MSG_SYS_RECORD_LOG
type MsgSysRecordLog struct {
	AckHandle         uint32
	Unk0              uint32
	Unk1              uint16 // Hardcoded 0
	HardcodedDataSize uint16 // Hardcoded 0x4AC
	Unk3              uint32 // Some shared ID with MSG_MHF_GET_SEIBATTLE. World ID??
	DataBuf           []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRecordLog) Opcode() network.PacketID {
	return network.MSG_SYS_RECORD_LOG
}

// Parse parses the packet from binary
func (m *MsgSysRecordLog) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint16()
	m.HardcodedDataSize = bf.ReadUint16()
	m.Unk3 = bf.ReadUint32()
	m.DataBuf = bf.ReadBytes(uint(m.HardcodedDataSize))
	return nil

}

// Build builds a binary packet from the current data.
func (m *MsgSysRecordLog) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
