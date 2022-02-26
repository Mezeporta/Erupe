package binpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgBinTargeted is a format used for some broadcast types
// to target specific players, instead of groups (world, stage, etc).
// It forwards a normal binpacket in it's RawDataPayload
type MsgBinTargeted struct {
	TargetCount    uint16
	TargetCharIDs  []uint32
	RawDataPayload []byte // The regular binary payload to be forwarded to the targets.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgBinTargeted) Opcode() network.PacketID {
	return network.MSG_SYS_CAST_BINARY
}

// Parse parses the packet from binary
func (m *MsgBinTargeted) Parse(bf *byteframe.ByteFrame) error {
	m.TargetCount = bf.ReadUint16()

	m.TargetCharIDs = make([]uint32, m.TargetCount)
	for i := uint16(0); i < m.TargetCount; i++ {
		m.TargetCharIDs[i] = bf.ReadUint32()
	}

	m.RawDataPayload = bf.DataFromCurrent()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgBinTargeted) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint16(m.TargetCount)

	for i := 0; i < int(m.TargetCount); i++ {
		bf.WriteUint32(m.TargetCharIDs[i])
	}

	bf.WriteBytes(m.RawDataPayload)
	return nil
}
