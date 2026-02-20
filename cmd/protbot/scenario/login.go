// Package scenario provides high-level MHF protocol flows.
package scenario

import (
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
)

// LoginResult holds the outcome of a full login flow.
type LoginResult struct {
	Sign    *protocol.SignResult
	Servers []protocol.ServerEntry
	Channel *protocol.ChannelConn
}

// Login performs the full sign → entrance → channel login flow.
func Login(signAddr, username, password string) (*LoginResult, error) {
	// Step 1: Sign server authentication.
	fmt.Printf("[sign] Connecting to %s...\n", signAddr)
	sign, err := protocol.DoSign(signAddr, username, password)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	fmt.Printf("[sign] OK — tokenID=%d, %d character(s), entrance=%s\n",
		sign.TokenID, len(sign.CharIDs), sign.EntranceAddr)

	if len(sign.CharIDs) == 0 {
		return nil, fmt.Errorf("no characters on account")
	}

	// Step 2: Entrance server — get server/channel list.
	fmt.Printf("[entrance] Connecting to %s...\n", sign.EntranceAddr)
	servers, err := protocol.DoEntrance(sign.EntranceAddr)
	if err != nil {
		return nil, fmt.Errorf("entrance: %w", err)
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no channels available")
	}
	for i, s := range servers {
		fmt.Printf("[entrance]   [%d] %s — %s:%d\n", i, s.Name, s.IP, s.Port)
	}

	// Step 3: Connect to the first channel server.
	first := servers[0]
	channelAddr := fmt.Sprintf("%s:%d", first.IP, first.Port)
	fmt.Printf("[channel] Connecting to %s...\n", channelAddr)
	ch, err := protocol.ConnectChannel(channelAddr)
	if err != nil {
		return nil, fmt.Errorf("channel connect: %w", err)
	}

	// Step 4: Send MSG_SYS_LOGIN.
	charID := sign.CharIDs[0]
	ack := ch.NextAckHandle()
	loginPkt := protocol.BuildLoginPacket(ack, charID, sign.TokenID, sign.TokenString)
	fmt.Printf("[channel] Sending MSG_SYS_LOGIN (charID=%d, ackHandle=%d)...\n", charID, ack)
	if err := ch.SendPacket(loginPkt); err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("channel send login: %w", err)
	}

	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("channel login ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		_ = ch.Close()
		return nil, fmt.Errorf("channel login failed: error code %d", resp.ErrorCode)
	}
	fmt.Printf("[channel] Login ACK received (error=%d, %d bytes data)\n",
		resp.ErrorCode, len(resp.Data))

	return &LoginResult{
		Sign:    sign,
		Servers: servers,
		Channel: ch,
	}, nil
}
