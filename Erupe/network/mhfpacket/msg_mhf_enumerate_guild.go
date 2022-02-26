package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type EnumerateGuildType uint8

const (
	_ = iota
	ENUMERATE_GUILD_TYPE_NAME
	//Numbers correspond to order in guild search menu
	ENUMERATE_GUILD_TYPE_6
	ENUMERATE_GUILD_TYPE_LEADER_ID
	ENUMERATE_GUILD_TYPE_3
	ENUMERATE_GUILD_TYPE_2
	ENUMERATE_GUILD_TYPE_7
	ENUMERATE_GUILD_TYPE_8
	ENUMERATE_GUILD_TYPE_NEW
)

// MsgMhfEnumerateGuild represents the MSG_MHF_ENUMERATE_GUILD
type MsgMhfEnumerateGuild struct {
	AckHandle      uint32
	Type           uint8
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Type = bf.ReadUint8()
	m.RawDataPayload = bf.DataFromCurrent()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
