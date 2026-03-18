package scenario

import (
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
	"erupe-ce/common/byteframe"
)

// AchievementEntry holds parsed achievement data from the server response.
type AchievementEntry struct {
	ID       uint8
	Level    uint8
	Next     uint16
	Required uint32
	Notify   bool
	Trophy   uint8
	Progress uint32
}

// AchievementResult holds the full parsed GET_ACHIEVEMENT response.
type AchievementResult struct {
	Points  uint32
	Entries []AchievementEntry
}

// GetAchievements sends MSG_MHF_GET_ACHIEVEMENT and returns the parsed response.
func GetAchievements(ch *protocol.ChannelConn, charID uint32) (*AchievementResult, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildGetAchievementPacket(ack, charID)
	fmt.Printf("[achievement] Sending GET_ACHIEVEMENT (charID=%d, ackHandle=%d)...\n", charID, ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("get achievement send: %w", err)
	}

	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("get achievement ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("get achievement failed: error code %d", resp.ErrorCode)
	}
	fmt.Printf("[achievement] ACK received (%d bytes)\n", len(resp.Data))

	return parseAchievementResponse(resp.Data)
}

// IncrementAchievement sends MSG_MHF_ADD_ACHIEVEMENT (fire-and-forget, no ACK).
func IncrementAchievement(ch *protocol.ChannelConn, achievementID uint8) error {
	pkt := protocol.BuildAddAchievementPacket(achievementID)
	fmt.Printf("[achievement] Sending ADD_ACHIEVEMENT (id=%d)...\n", achievementID)
	return ch.SendPacket(pkt)
}

// DisplayedAchievement sends MSG_MHF_DISPLAYED_ACHIEVEMENT to tell the server
// the client has seen all rank-up notifications (fire-and-forget, no ACK).
func DisplayedAchievement(ch *protocol.ChannelConn) error {
	pkt := protocol.BuildDisplayedAchievementPacket()
	fmt.Printf("[achievement] Sending DISPLAYED_ACHIEVEMENT...\n")
	return ch.SendPacket(pkt)
}

func parseAchievementResponse(data []byte) (*AchievementResult, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("achievement response too short: %d bytes", len(data))
	}

	bf := byteframe.NewByteFrameFromBytes(data)
	result := &AchievementResult{}

	// Header: 4x uint32 points (all same value), 3 bytes unk, 1 byte count
	result.Points = bf.ReadUint32()
	bf.ReadUint32() // Points repeated
	bf.ReadUint32() // Points repeated
	bf.ReadUint32() // Points repeated
	bf.ReadBytes(3) // Unk (0x02, 0x00, 0x00)
	count := bf.ReadUint8()

	for i := uint8(0); i < count; i++ {
		entry := AchievementEntry{}
		entry.ID = bf.ReadUint8()
		entry.Level = bf.ReadUint8()
		entry.Next = bf.ReadUint16()
		entry.Required = bf.ReadUint32()
		entry.Notify = bf.ReadBool()
		entry.Trophy = bf.ReadUint8()
		bf.ReadUint16() // Unk
		entry.Progress = bf.ReadUint32()
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}
