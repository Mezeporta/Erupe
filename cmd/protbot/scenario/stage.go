package scenario

import (
	"encoding/binary"
	"fmt"
	"time"

	"erupe-ce/common/byteframe"

	"erupe-ce/cmd/protbot/protocol"
)

// StageInfo holds a parsed stage entry from MSG_SYS_ENUMERATE_STAGE response.
type StageInfo struct {
	ID         string
	Reserved   uint16
	Clients    uint16
	Displayed  uint16
	MaxPlayers uint16
	Flags      uint8
}

// EnterLobby enumerates available lobby stages and enters the first one.
func EnterLobby(ch *protocol.ChannelConn) error {
	// Step 1: Enumerate stages with "sl1Ns" prefix (main lobby stages).
	ack := ch.NextAckHandle()
	enumPkt := protocol.BuildEnumerateStagePacket(ack, "sl1Ns")
	fmt.Printf("[stage] Sending MSG_SYS_ENUMERATE_STAGE (prefix=\"sl1Ns\", ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(enumPkt); err != nil {
		return fmt.Errorf("enumerate stage send: %w", err)
	}

	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return fmt.Errorf("enumerate stage ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf("enumerate stage failed: error code %d", resp.ErrorCode)
	}

	stages := parseEnumerateStageResponse(resp.Data)
	fmt.Printf("[stage] Found %d stage(s)\n", len(stages))
	for i, s := range stages {
		fmt.Printf("[stage]   [%d] %s — %d/%d players, flags=0x%02X\n",
			i, s.ID, s.Clients, s.MaxPlayers, s.Flags)
	}

	// Step 2: Enter the default lobby stage.
	// Even if no stages were enumerated, use the default stage ID.
	stageID := "sl1Ns200p0a0u0"
	if len(stages) > 0 {
		stageID = stages[0].ID
	}

	ack = ch.NextAckHandle()
	enterPkt := protocol.BuildEnterStagePacket(ack, stageID)
	fmt.Printf("[stage] Sending MSG_SYS_ENTER_STAGE (stageID=%q, ackHandle=%d)...\n", stageID, ack)
	if err := ch.SendPacket(enterPkt); err != nil {
		return fmt.Errorf("enter stage send: %w", err)
	}

	resp, err = ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return fmt.Errorf("enter stage ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf("enter stage failed: error code %d", resp.ErrorCode)
	}
	fmt.Printf("[stage] Enter stage ACK received (error=%d)\n", resp.ErrorCode)

	return nil
}

// parseEnumerateStageResponse parses the ACK data from MSG_SYS_ENUMERATE_STAGE.
// Reference: Erupe server/channelserver/handlers_stage.go (handleMsgSysEnumerateStage)
func parseEnumerateStageResponse(data []byte) []StageInfo {
	if len(data) < 2 {
		return nil
	}

	bf := byteframe.NewByteFrameFromBytes(data)
	count := bf.ReadUint16()

	var stages []StageInfo
	for i := uint16(0); i < count; i++ {
		s := StageInfo{}
		s.Reserved = bf.ReadUint16()
		s.Clients = bf.ReadUint16()
		s.Displayed = bf.ReadUint16()
		s.MaxPlayers = bf.ReadUint16()
		s.Flags = bf.ReadUint8()

		// Stage ID is a pascal string with uint8 length prefix.
		strLen := bf.ReadUint8()
		if strLen > 0 {
			idBytes := bf.ReadBytes(uint(strLen))
			// Remove null terminator if present.
			if len(idBytes) > 0 && idBytes[len(idBytes)-1] == 0 {
				idBytes = idBytes[:len(idBytes)-1]
			}
			s.ID = string(idBytes)
		}

		stages = append(stages, s)
	}

	// After stages: uint32 timestamp, uint32 max clan members (we ignore these).
	_ = binary.BigEndian // suppress unused import if needed

	return stages
}
