package mhfpacket

import (
	"errors"

	"erupe-ce/common/bfutil"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysEnterStage is sent by the client to enter an existing stage.
//
// This packet is used when:
//   - Moving from one town area to another (e.g., Mezeporta -> Pallone)
//   - Joining another player's room or quest
//   - Entering a persistent stage that already exists
//
// The stage must already exist on the server. For creating new stages (quests, rooms),
// use MSG_SYS_CREATE_STAGE followed by MSG_SYS_ENTER_STAGE.
//
// Stage ID Format:
// Stage IDs are encoded strings like "sl1Ns200p0a0u0" that identify specific
// game areas:
//   - sl1Ns200p0a0u0: Mezeporta (main town)
//   - sl1Ns211p0a0u0: Rasta bar
//   - Quest stages: Dynamic IDs created when quests start
//
// After entering, the session's stage pointer is updated and the player receives
// broadcasts from other players in that stage.
type MsgSysEnterStage struct {
	AckHandle uint32 // Response handle for acknowledgment
	UnkBool   uint8  // Boolean flag (purpose unknown, possibly force-enter)
	StageID   string // ID of the stage to enter (length-prefixed string)
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnterStage) Opcode() network.PacketID {
	return network.MSG_SYS_ENTER_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysEnterStage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.UnkBool = bf.ReadUint8()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnterStage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
