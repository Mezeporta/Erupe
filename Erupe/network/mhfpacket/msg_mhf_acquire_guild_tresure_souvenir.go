package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireGuildTresureSouvenir represents the MSG_MHF_ACQUIRE_GUILD_TRESURE_SOUVENIR
type MsgMhfAcquireGuildTresureSouvenir struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireGuildTresureSouvenir) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_GUILD_TRESURE_SOUVENIR
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireGuildTresureSouvenir) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireGuildTresureSouvenir) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
