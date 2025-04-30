package mhfpacket

import (
	"errors"
	_config "erupe-ce/config"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// TerminalLogEntry represents an entry in the MSG_SYS_TERMINAL_LOG packet.
type TerminalLogEntry struct {
	Index uint32
	Type1 uint8
	Type2 uint8
	Unk0  int16
	Unk1  int32
	Unk2  int32
	Unk3  int32
	Unk4  []int32
}

// MsgSysTerminalLog represents the MSG_SYS_TERMINAL_LOG
type MsgSysTerminalLog struct {
	AckHandle uint32
	LogID     uint32 // 0 on the first packet, and the server sends back a value to use for subsequent requests.
	Entries   []TerminalLogEntry
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTerminalLog) Opcode() network.PacketID {
	return network.MSG_SYS_TERMINAL_LOG
}

// Parse parses the packet from binary
func (m *MsgSysTerminalLog) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.LogID = bf.ReadUint32()
	entryCount := int(bf.ReadUint16())
	bf.ReadUint16() // Zeroed

	for i := 0; i < entryCount; i++ {
		var e TerminalLogEntry
		e.Index = bf.ReadUint32()
		e.Type1 = bf.ReadUint8()
		e.Type2 = bf.ReadUint8()
		e.Unk0 = bf.ReadInt16()
		e.Unk1 = bf.ReadInt32()
		e.Unk2 = bf.ReadInt32()
		e.Unk3 = bf.ReadInt32()
		if _config.ErupeConfig.RealClientMode >= _config.G1 {
			for j := 0; j < 4; j++ {
				e.Unk4 = append(e.Unk4, bf.ReadInt32())
			}
		}
		m.Entries = append(m.Entries, e)
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysTerminalLog) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
