package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEnumerateQuest is sent by the client to request a paginated list of available quests.
//
// This packet is used when:
//   - Accessing the quest counter/board in town
//   - Scrolling through quest lists
//   - Switching between quest categories/worlds
//
// The server responds with quest metadata and binary file paths. The client then
// loads quest details from binary files on disk or via MSG_SYS_GET_FILE.
//
// Pagination:
// Quest lists can be very large (hundreds of quests). The client requests quests
// in batches using the Offset field:
//   - Offset 0: First batch (quests 0-N)
//   - Offset N: Next batch (quests N-2N)
//   - Continues until server returns no more quests
//
// World Types:
//   - 0: Newbie World (beginner quests)
//   - 1: Normal World (standard quests)
//   - 2+: Other world categories (events, special quests)
type MsgMhfEnumerateQuest struct {
	AckHandle uint32 // Response handle for matching server response to request
	Unk0      uint8  // Hardcoded 0 in the binary (purpose unknown)
	World     uint8  // World ID/category to enumerate quests for
	Counter   uint16 // Client counter for tracking sequential requests
	Offset    uint16 // Pagination offset - increments by batch size for next page
	Unk4      uint8  // Hardcoded 0 in the binary (purpose unknown)
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateQuest) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateQuest) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.World = bf.ReadUint8()
	m.Counter = bf.ReadUint16()
	m.Offset = bf.ReadUint16()
	m.Unk4 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateQuest) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
