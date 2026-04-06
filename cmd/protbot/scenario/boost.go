package scenario

import (
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
	"erupe-ce/common/byteframe"
)

// BoostTimeStatus holds the parsed response of MSG_MHF_GET_BOOST_TIME_LIMIT.
// When boost time is disabled server-side, or has not been started yet,
// BoostLimitUnix is 0. Prior to the #187 fix, unset boost_time columns
// wrapped to a large uint32 the client interpreted as permanently active.
type BoostTimeStatus struct {
	BoostLimitUnix uint32
}

// BoostRight holds the parsed response of MSG_MHF_GET_BOOST_RIGHT.
// 0 = disabled, 1 = active, 2 = available.
type BoostRight struct {
	State uint32
}

// LoginBoostEntry holds a single entry of the 5-entry MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS response.
type LoginBoostEntry struct {
	WeekReq    uint8
	Active     bool
	WeekCount  uint8
	Expiration uint32
}

// LoginBoostStatus holds the full parsed keep login boost status response.
type LoginBoostStatus struct {
	Entries []LoginBoostEntry
}

// GetBoostTimeLimit sends MSG_MHF_GET_BOOST_TIME_LIMIT and parses the response.
func GetBoostTimeLimit(ch *protocol.ChannelConn) (*BoostTimeStatus, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildGetBoostTimeLimitPacket(ack)
	fmt.Printf("[boost] Sending GET_BOOST_TIME_LIMIT (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("get boost time limit send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("get boost time limit ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("get boost time limit failed: error code %d", resp.ErrorCode)
	}
	if len(resp.Data) < 4 {
		return nil, fmt.Errorf("get boost time limit response too short: %d bytes", len(resp.Data))
	}
	bf := byteframe.NewByteFrameFromBytes(resp.Data)
	return &BoostTimeStatus{BoostLimitUnix: bf.ReadUint32()}, nil
}

// GetBoostRight sends MSG_MHF_GET_BOOST_RIGHT and parses the response.
func GetBoostRight(ch *protocol.ChannelConn) (*BoostRight, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildGetBoostRightPacket(ack)
	fmt.Printf("[boost] Sending GET_BOOST_RIGHT (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("get boost right send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("get boost right ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("get boost right failed: error code %d", resp.ErrorCode)
	}
	if len(resp.Data) < 4 {
		return nil, fmt.Errorf("get boost right response too short: %d bytes", len(resp.Data))
	}
	bf := byteframe.NewByteFrameFromBytes(resp.Data)
	return &BoostRight{State: bf.ReadUint32()}, nil
}

// GetKeepLoginBoostStatus sends MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS and parses the response.
// The server returns either 35 bytes (5 entries × 7 bytes) or 35 zero bytes
// when DisableLoginBoost is set.
func GetKeepLoginBoostStatus(ch *protocol.ChannelConn) (*LoginBoostStatus, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildGetKeepLoginBoostStatusPacket(ack)
	fmt.Printf("[boost] Sending GET_KEEP_LOGIN_BOOST_STATUS (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("get login boost status send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("get login boost status ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("get login boost status failed: error code %d", resp.ErrorCode)
	}
	if len(resp.Data) < 35 {
		return nil, fmt.Errorf("login boost status response too short: %d bytes", len(resp.Data))
	}
	bf := byteframe.NewByteFrameFromBytes(resp.Data)
	status := &LoginBoostStatus{}
	for i := 0; i < 5; i++ {
		status.Entries = append(status.Entries, LoginBoostEntry{
			WeekReq:    bf.ReadUint8(),
			Active:     bf.ReadBool(),
			WeekCount:  bf.ReadUint8(),
			Expiration: bf.ReadUint32(),
		})
	}
	return status, nil
}
