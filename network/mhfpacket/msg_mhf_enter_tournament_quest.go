package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnterTournamentQuest represents the MSG_MHF_ENTER_TOURNAMENT_QUEST (opcode 0x00D2).
//
// Wire format derived from mhfo-hd.dll binary analysis (FUN_114f4280 = putEnterTournamentQuest).
// The client sends this packet when entering the actual tournament quest instance after
// completing the ENTRY_TOURNAMENT (0xD1) flow. Fields are all big-endian.
//
// Byte layout (after opcode):
//
//	[0..3]   uint32  AckHandle
//	[4..7]   uint32  TournamentID   — tournament being entered
//	[8..11]  uint32  EntryHandle    — slot handle assigned by server during ENTRY_TOURNAMENT response
//	[12..15] uint32  Unk2           — third field from server INFO response; semantics unclear
//	[16..19] uint32  QuestSlot      — derived from quest table (DAT_1e41d3b4); effectively uint16 in uint32
//	[20..23] uint32  StageHandle    — quest node offset (DAT_1e41d3b8); computed as quest_node + 0x10
//	[24..27] uint32  Unk5           — formatted string identifier (result of FUN_11586310)
//	[28]     uint8   StringLen      — length of optional trailing string (0 = absent in normal flow)
//	[29+]    bytes   String         — pascal-style string data (StringLen bytes, absent when 0)
type MsgMhfEnterTournamentQuest struct {
	AckHandle    uint32
	TournamentID uint32
	EntryHandle  uint32
	Unk2         uint32
	QuestSlot    uint32
	StageHandle  uint32
	Unk5         uint32
	String       []byte // pascal-style: 1-byte length prefix, then data; nil when absent
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnterTournamentQuest) Opcode() network.PacketID {
	return network.MSG_MHF_ENTER_TOURNAMENT_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfEnterTournamentQuest) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.TournamentID = bf.ReadUint32()
	m.EntryHandle = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.QuestSlot = bf.ReadUint32()
	m.StageHandle = bf.ReadUint32()
	m.Unk5 = bf.ReadUint32()
	strLen := bf.ReadUint8()
	if strLen > 0 {
		m.String = bf.ReadBytes(uint(strLen))
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnterTournamentQuest) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
