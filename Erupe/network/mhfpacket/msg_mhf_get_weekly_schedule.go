package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetWeeklySchedule represents the MSG_MHF_GET_WEEKLY_SCHEDULE
type MsgMhfGetWeeklySchedule struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetWeeklySchedule) Opcode() network.PacketID {
	return network.MSG_MHF_GET_WEEKLY_SCHEDULE
}

// Parse parses the packet from binary
func (m *MsgMhfGetWeeklySchedule) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetWeeklySchedule) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
