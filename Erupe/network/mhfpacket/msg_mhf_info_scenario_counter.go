package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoScenarioCounter represents the MSG_MHF_INFO_SCENARIO_COUNTER
type MsgMhfInfoScenarioCounter struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoScenarioCounter) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_SCENARIO_COUNTER
}

// Parse parses the packet from binary
func (m *MsgMhfInfoScenarioCounter) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoScenarioCounter) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
