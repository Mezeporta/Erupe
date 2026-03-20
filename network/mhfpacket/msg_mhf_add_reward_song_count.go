package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddRewardSongCount represents the MSG_MHF_ADD_REWARD_SONG_COUNT packet.
// Request layout:
//
//	u32 ack_handle
//	u32 prayer_id
//	u16 array_size_bytes  (= count × 2)
//	u8  count
//	u16[count] entries
type MsgMhfAddRewardSongCount struct {
	AckHandle      uint32
	PrayerID       uint32
	ArraySizeBytes uint16
	Count          uint8
	Entries        []uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddRewardSongCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_REWARD_SONG_COUNT
}

// Parse parses the packet from binary.
func (m *MsgMhfAddRewardSongCount) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.PrayerID = bf.ReadUint32()
	m.ArraySizeBytes = bf.ReadUint16()
	m.Count = bf.ReadUint8()
	m.Entries = make([]uint16, m.Count)
	for i := range m.Entries {
		m.Entries[i] = bf.ReadUint16()
	}
	return bf.Err()
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddRewardSongCount) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.PrayerID)
	bf.WriteUint16(uint16(len(m.Entries) * 2))
	bf.WriteUint8(uint8(len(m.Entries)))
	for _, e := range m.Entries {
		bf.WriteUint16(e)
	}
	return nil
}
