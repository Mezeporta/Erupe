package scenario

import (
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
	"erupe-ce/common/byteframe"
)

// GachaPoints holds the parsed MSG_MHF_GET_GACHA_POINT response.
type GachaPoints struct {
	Premium  uint32
	Trial    uint32
	Frontier uint32
}

// GachaReward is one item returned by a normal gacha roll.
type GachaReward struct {
	ItemType uint8
	ItemID   uint16
	Quantity uint16
	Rarity   uint8
}

// GachaStoredItem is one item currently sitting in the character's
// gacha_items column (the client's "temp storage" display).
type GachaStoredItem struct {
	ItemType uint8
	ItemID   uint16
	Quantity uint16
}

// GetGachaPoint sends MSG_MHF_GET_GACHA_POINT and parses the response.
func GetGachaPoint(ch *protocol.ChannelConn) (*GachaPoints, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildGetGachaPointPacket(ack)
	fmt.Printf("[gacha] Sending GET_GACHA_POINT (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("get gacha point send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("get gacha point ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("get gacha point failed: error code %d", resp.ErrorCode)
	}
	if len(resp.Data) < 12 {
		return nil, fmt.Errorf("gacha point response too short: %d bytes", len(resp.Data))
	}
	bf := byteframe.NewByteFrameFromBytes(resp.Data)
	return &GachaPoints{
		Premium:  bf.ReadUint32(),
		Trial:    bf.ReadUint32(),
		Frontier: bf.ReadUint32(),
	}, nil
}

// PlayNormalGacha sends MSG_MHF_PLAY_NORMAL_GACHA and parses the reward list.
// Response layout: uint8 count, then count * (uint8 type, uint16 id, uint16 qty, uint8 rarity).
// The server returns a single 0x00 byte ACK payload on validation failure
// (e.g. empty reward pool), which this function reports as an error.
func PlayNormalGacha(ch *protocol.ChannelConn, gachaID uint32, rollType, gachaType uint8) ([]GachaReward, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildPlayNormalGachaPacket(ack, gachaID, rollType, gachaType)
	fmt.Printf("[gacha] Sending PLAY_NORMAL_GACHA (gachaID=%d, rollType=%d)...\n", gachaID, rollType)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("play normal gacha send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("play normal gacha ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("play normal gacha failed: error code %d", resp.ErrorCode)
	}
	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("play normal gacha response too short: %d bytes", len(resp.Data))
	}
	bf := byteframe.NewByteFrameFromBytes(resp.Data)
	count := bf.ReadUint8()
	// A single-byte (count=0) reply from an otherwise-successful ACK means
	// the server's validation failed. Surface it so callers can tell the
	// difference from a genuinely empty roll.
	if count == 0 && len(resp.Data) == 1 {
		return nil, fmt.Errorf("play normal gacha: server returned empty reward pool (check gacha config)")
	}
	rewards := make([]GachaReward, 0, count)
	for i := uint8(0); i < count; i++ {
		rewards = append(rewards, GachaReward{
			ItemType: bf.ReadUint8(),
			ItemID:   bf.ReadUint16(),
			Quantity: bf.ReadUint16(),
			Rarity:   bf.ReadUint8(),
		})
	}
	return rewards, nil
}

// ReceiveGachaItem sends MSG_MHF_RECEIVE_GACHA_ITEM and parses the pending
// items in the character's gacha_items column. When freeze=true the server
// does not clear the column, allowing non-destructive inspection.
// Response layout: uint8 count + count * (uint8 type, uint16 id, uint16 qty).
func ReceiveGachaItem(ch *protocol.ChannelConn, max uint8, freeze bool) ([]GachaStoredItem, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildReceiveGachaItemPacket(ack, max, freeze)
	fmt.Printf("[gacha] Sending RECEIVE_GACHA_ITEM (max=%d, freeze=%v)...\n", max, freeze)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("receive gacha item send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("receive gacha item ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("receive gacha item failed: error code %d", resp.ErrorCode)
	}
	if len(resp.Data) < 1 {
		return nil, fmt.Errorf("receive gacha item response too short: %d bytes", len(resp.Data))
	}
	bf := byteframe.NewByteFrameFromBytes(resp.Data)
	count := bf.ReadUint8()
	items := make([]GachaStoredItem, 0, count)
	for i := uint8(0); i < count; i++ {
		items = append(items, GachaStoredItem{
			ItemType: bf.ReadUint8(),
			ItemID:   bf.ReadUint16(),
			Quantity: bf.ReadUint16(),
		})
	}
	return items, nil
}
