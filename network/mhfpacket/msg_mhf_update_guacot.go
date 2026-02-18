package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// Goocoo represents a single Goocoo (guacot) companion entry in an update packet.
type Goocoo struct {
	Index uint32
	Data1 []int16
	Data2 []uint32
	Name  []byte
}

// MsgMhfUpdateGuacot represents the MSG_MHF_UPDATE_GUACOT
type MsgMhfUpdateGuacot struct {
	AckHandle  uint32
	EntryCount uint16
	Goocoos    []Goocoo
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuacot) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.EntryCount = bf.ReadUint16()
	bf.ReadUint16() // Zeroed
	for i := 0; i < int(m.EntryCount); i++ {
		var temp Goocoo
		temp.Index = bf.ReadUint32()
		for j := 0; j < 22; j++ {
			temp.Data1 = append(temp.Data1, bf.ReadInt16())
		}
		for j := 0; j < 2; j++ {
			temp.Data2 = append(temp.Data2, bf.ReadUint32())
		}
		temp.Name = bf.ReadBytes(uint(bf.ReadUint8()))
		m.Goocoos = append(m.Goocoos, temp)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuacot) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
