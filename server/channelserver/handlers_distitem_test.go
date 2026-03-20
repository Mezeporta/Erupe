package channelserver

import (
	"encoding/binary"
	"errors"
	"testing"
	"time"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

// --- mockDistRepo ---

type mockDistRepo struct {
	distributions []Distribution
	listErr       error
	items         map[uint32][]DistributionItem
	itemsErr      error
	description   string
	descErr       error
	recordedDist  uint32
	recordedChar  uint32
	recordErr     error
}

func (m *mockDistRepo) List(_ uint32, _ uint8) ([]Distribution, error) {
	return m.distributions, m.listErr
}

func (m *mockDistRepo) GetItems(distID uint32) ([]DistributionItem, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	if m.items != nil {
		return m.items[distID], nil
	}
	return nil, nil
}

func (m *mockDistRepo) RecordAccepted(distID, charID uint32) error {
	m.recordedDist = distID
	m.recordedChar = charID
	return m.recordErr
}

func (m *mockDistRepo) GetDescription(_ uint32) (string, error) {
	return m.description, m.descErr
}

func TestHandleMsgMhfEnumerateDistItem_Empty(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateDistItem{AckHandle: 100, DistType: 0}
	handleMsgMhfEnumerateDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0", errCode)
		}
		count := binary.BigEndian.Uint16(ackData[:2])
		if count != 0 {
			t.Errorf("dist count = %d, want 0", count)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfEnumerateDistItem_WithDistributions(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{
		distributions: []Distribution{
			{
				ID:              1,
				Deadline:        time.Unix(1000000, 0),
				Rights:          0,
				TimesAcceptable: 1,
				TimesAccepted:   0,
				MinHR:           1,
				MaxHR:           999,
				EventName:       "Test",
			},
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateDistItem{AckHandle: 100, DistType: 0}
	handleMsgMhfEnumerateDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		count := binary.BigEndian.Uint16(ackData[:2])
		if count != 1 {
			t.Errorf("dist count = %d, want 1", count)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfApplyDistItem_Empty(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfApplyDistItem{
		AckHandle:      100,
		DistributionID: 42,
	}
	handleMsgMhfApplyDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		// 4 (distID) + 2 (count=0) = 6
		distID := binary.BigEndian.Uint32(ackData[:4])
		if distID != 42 {
			t.Errorf("distID = %d, want 42", distID)
		}
		itemCount := binary.BigEndian.Uint16(ackData[4:6])
		if itemCount != 0 {
			t.Errorf("item count = %d, want 0", itemCount)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfApplyDistItem_WithItems(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{
		items: map[uint32][]DistributionItem{
			10: {
				{ItemType: 1, ID: 100, ItemID: 200, Quantity: 5},
				{ItemType: 2, ID: 101, ItemID: 300, Quantity: 3},
			},
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfApplyDistItem{
		AckHandle:      100,
		DistributionID: 10,
	}
	handleMsgMhfApplyDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		itemCount := binary.BigEndian.Uint16(ackData[4:6])
		if itemCount != 2 {
			t.Errorf("item count = %d, want 2", itemCount)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_ZeroID(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:      100,
		DistributionID: 0,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Should respond")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_RecordAccepted(t *testing.T) {
	server := createMockServer()
	distRepo := &mockDistRepo{
		items: map[uint32][]DistributionItem{
			5: {},
		},
	}
	server.distRepo = distRepo
	session := createMockSession(1, server)
	session.charID = 42

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:      100,
		DistributionID: 5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	if distRepo.recordedDist != 5 {
		t.Errorf("recorded dist ID = %d, want 5", distRepo.recordedDist)
	}
	if distRepo.recordedChar != 42 {
		t.Errorf("recorded char ID = %d, want 42", distRepo.recordedChar)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Should respond")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_RecordError(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{
		recordErr: errors.New("db error"),
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:      100,
		DistributionID: 5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	// Should still send success ack
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Should respond")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetDistDescription_Success(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{description: "Test event description"}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetDistDescription{
		AckHandle:      100,
		DistributionID: 1,
	}
	handleMsgMhfGetDistDescription(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0", errCode)
		}
		if len(ackData) == 0 {
			t.Fatal("AckData should not be empty")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetDistDescription_Error(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{descErr: errors.New("not found")}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetDistDescription{
		AckHandle:      100,
		DistributionID: 999,
	}
	handleMsgMhfGetDistDescription(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0 (still buf succeed)", errCode)
		}
		if len(ackData) != 4 {
			t.Errorf("AckData len = %d, want 4 (fallback)", len(ackData))
		}
	default:
		t.Fatal("No response queued")
	}
}
