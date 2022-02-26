package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEntryTournament represents the MSG_MHF_ENTRY_TOURNAMENT
type MsgMhfEntryTournament struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEntryTournament) Opcode() network.PacketID {
	return network.MSG_MHF_ENTRY_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfEntryTournament) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEntryTournament) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
