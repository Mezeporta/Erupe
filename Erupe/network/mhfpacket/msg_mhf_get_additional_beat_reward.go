package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetAdditionalBeatReward represents the MSG_MHF_GET_ADDITIONAL_BEAT_REWARD
type MsgMhfGetAdditionalBeatReward struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      uint32
	Unk3      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetAdditionalBeatReward) Opcode() network.PacketID {
	return network.MSG_MHF_GET_ADDITIONAL_BEAT_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfGetAdditionalBeatReward) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetAdditionalBeatReward) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
