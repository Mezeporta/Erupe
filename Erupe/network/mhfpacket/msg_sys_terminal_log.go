package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// TerminalLogEntry represents an entry in the MSG_SYS_TERMINAL_LOG packet.
type TerminalLogEntry struct {
	// Unknown fields
	U0, U1, U2, U3, U4, U5, U6, U7, U8 uint32
}

// MsgSysTerminalLog represents the MSG_SYS_TERMINAL_LOG
type MsgSysTerminalLog struct {
	AckHandle  uint32
	LogID      uint32 // 0 on the first packet, and the server sends back a value to use for subsequent requests.
	EntryCount uint16
	Unk0       uint16 // Hardcoded 0 in the binary
	Entries    []*TerminalLogEntry
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTerminalLog) Opcode() network.PacketID {
	return network.MSG_SYS_TERMINAL_LOG
}

// Parse parses the packet from binary
func (m *MsgSysTerminalLog) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.LogID = bf.ReadUint32()
	m.EntryCount = bf.ReadUint16()
	m.Unk0 = bf.ReadUint16()

	for i := 0; i < int(m.EntryCount); i++ {
		e := &TerminalLogEntry{}
		e.U0 = bf.ReadUint32()
		e.U1 = bf.ReadUint32()
		e.U2 = bf.ReadUint32()
		e.U3 = bf.ReadUint32()
		e.U4 = bf.ReadUint32()
		e.U5 = bf.ReadUint32()
		e.U6 = bf.ReadUint32()
		e.U7 = bf.ReadUint32()
		e.U8 = bf.ReadUint32()
		m.Entries = append(m.Entries, e)
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysTerminalLog) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
