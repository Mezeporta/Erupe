package scenario

import (
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
)

// EnumerateQuests sends MSG_MHF_ENUMERATE_QUEST and returns the raw quest list data.
func EnumerateQuests(ch *protocol.ChannelConn, world uint8, counter uint16) ([]byte, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildEnumerateQuestPacket(ack, world, counter, 0)
	fmt.Printf("[quest] Sending MSG_MHF_ENUMERATE_QUEST (world=%d, counter=%d, ackHandle=%d)...\n",
		world, counter, ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("enumerate quest send: %w", err)
	}

	resp, err := ch.WaitForAck(ack, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("enumerate quest ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("enumerate quest failed: error code %d", resp.ErrorCode)
	}
	fmt.Printf("[quest] ENUMERATE_QUEST ACK (error=%d, %d bytes data)\n",
		resp.ErrorCode, len(resp.Data))

	return resp.Data, nil
}
