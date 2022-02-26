package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadBeatLevel represents the MSG_MHF_READ_BEAT_LEVEL
type MsgMhfReadBeatLevel struct {
	AckHandle    uint32
	Unk0         uint32
	ValidIDCount uint32 // Valid entries in the array
	IDs          [16]uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevel) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevel) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	// I assume this used to be dynamic, but as of the last JP client version, all of this data is hard-coded literals.
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()         // Always 1
	m.ValidIDCount = bf.ReadUint32() // Always 4

	// Always 0x74, 0x6B, 0x02, 0x24 followed by 12 zero values.
	for i := 0; i < len(m.IDs); i++ {
		m.IDs[i] = bf.ReadUint32()
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevel) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
