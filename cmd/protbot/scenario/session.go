package scenario

import (
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
)

// SetupSession performs the post-login session setup: ISSUE_LOGKEY, RIGHTS_RELOAD, LOADDATA.
// Returns the loaddata response blob for inspection.
func SetupSession(ch *protocol.ChannelConn, charID uint32) ([]byte, error) {
	// Step 1: Issue logkey.
	ack := ch.NextAckHandle()
	fmt.Printf("[session] Sending MSG_SYS_ISSUE_LOGKEY (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(protocol.BuildIssueLogkeyPacket(ack)); err != nil {
		return nil, fmt.Errorf("issue logkey send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("issue logkey ack: %w", err)
	}
	fmt.Printf("[session] ISSUE_LOGKEY ACK (error=%d, %d bytes)\n", resp.ErrorCode, len(resp.Data))

	// Step 2: Rights reload.
	ack = ch.NextAckHandle()
	fmt.Printf("[session] Sending MSG_SYS_RIGHTS_RELOAD (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(protocol.BuildRightsReloadPacket(ack)); err != nil {
		return nil, fmt.Errorf("rights reload send: %w", err)
	}
	resp, err = ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("rights reload ack: %w", err)
	}
	fmt.Printf("[session] RIGHTS_RELOAD ACK (error=%d, %d bytes)\n", resp.ErrorCode, len(resp.Data))

	// Step 3: Load save data.
	ack = ch.NextAckHandle()
	fmt.Printf("[session] Sending MSG_MHF_LOADDATA (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(protocol.BuildLoaddataPacket(ack)); err != nil {
		return nil, fmt.Errorf("loaddata send: %w", err)
	}
	resp, err = ch.WaitForAck(ack, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("loaddata ack: %w", err)
	}
	fmt.Printf("[session] LOADDATA ACK (error=%d, %d bytes)\n", resp.ErrorCode, len(resp.Data))

	return resp.Data, nil
}
