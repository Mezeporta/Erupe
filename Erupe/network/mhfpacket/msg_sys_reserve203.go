package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// TODO(Andoryuuta): Make up a name for this packet, not reserved anymore. Called "Is_update_guild_msg_board"

// MsgSysReserve203 represents the MSG_SYS_reserve203
type MsgSysReserve203 struct {
	AckHandle uint32
	Unk0      uint16 // Hardcoded 0x0000 in the binary
	Unk1      uint16 // Hardcoded 0x0500 in the binary.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve203) Opcode() network.PacketID {
	return network.MSG_SYS_reserve203
}

// Parse parses the packet from binary
func (m *MsgSysReserve203) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve203) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
