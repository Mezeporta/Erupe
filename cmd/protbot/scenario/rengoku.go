package scenario

import (
	"encoding/binary"
	"fmt"
	"time"

	"erupe-ce/cmd/protbot/protocol"
)

// GetRengokuBinary sends MSG_MHF_GET_RENGOKU_BINARY and returns the raw
// ECD-encrypted Hunting Road binary exactly as served to the real client
// (regression check for #206: spawn slots must never carry a zero candidate
// count, or the real client crashes on Hunting Road entry).
func GetRengokuBinary(ch *protocol.ChannelConn) ([]byte, error) {
	ack := ch.NextAckHandle()
	pkt := protocol.BuildGetRengokuBinaryPacket(ack)
	fmt.Printf("[rengoku] Sending GET_RENGOKU_BINARY (ackHandle=%d)...\n", ack)
	if err := ch.SendPacket(pkt); err != nil {
		return nil, fmt.Errorf("get rengoku binary send: %w", err)
	}
	resp, err := ch.WaitForAck(ack, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("get rengoku binary ack: %w", err)
	}
	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("get rengoku binary failed: error code %d", resp.ErrorCode)
	}
	return resp.Data, nil
}

// RengokuSlotCounts holds the per-slot candidate counts read back from one
// road mode (multiDef or soloDef) of a decrypted rengoku binary.
type RengokuSlotCounts struct {
	Label      string
	SlotCounts []uint32
}

// VerifyRengokuBinary reads the per-slot spawn candidate counts out of a
// decrypted rengoku binary for both road modes, mirroring the layout
// documented in server/channelserver/rengoku_binary.go (RoadMode struct:
// floorStatsCount, spawnCountCount, spawnTablePtrCount, floorStatsPtr,
// spawnTablePtrsPtr, spawnCountPtrsPtr — 6 x u32 starting at 0x14/0x2C).
// It is a standalone reader (parseRengokuBinary itself is unexported in the
// channelserver package) used to confirm, from the client's point of view,
// that no slot ever carries a zero count (issue #206: a zero count crashes
// the real client on Hunting Road entry).
func VerifyRengokuBinary(data []byte) ([]RengokuSlotCounts, error) {
	const rengokuMinSize = 0x44
	if len(data) < rengokuMinSize {
		return nil, fmt.Errorf("rengoku binary too small: %d bytes (need %d)", len(data), rengokuMinSize)
	}
	if data[0] != 'r' || data[1] != 'e' || data[2] != 'f' || data[3] != 0x1A {
		return nil, fmt.Errorf("bad magic: %02x %02x %02x %02x", data[0], data[1], data[2], data[3])
	}

	le := binary.LittleEndian
	var results []RengokuSlotCounts
	for _, rm := range []struct {
		label  string
		offset int
	}{{"multiDef", 0x14}, {"soloDef", 0x2C}} {
		slotCount := le.Uint32(data[rm.offset+8:])
		countPtrsPtr := le.Uint32(data[rm.offset+20:])

		counts := make([]uint32, slotCount)
		for i := uint32(0); i < slotCount; i++ {
			counts[i] = le.Uint32(data[countPtrsPtr+i*4:])
		}
		results = append(results, RengokuSlotCounts{Label: rm.label, SlotCounts: counts})
	}
	return results, nil
}
