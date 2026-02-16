package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetBoostTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostTime{
		AckHandle: 12345,
	}

	handleMsgMhfGetBoostTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Response should be empty bytes for this handler
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostBoostTimeQuestReturn(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostBoostTimeQuestReturn{
		AckHandle: 12345,
	}

	handleMsgMhfPostBoostTimeQuestReturn(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostBoostTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostBoostTime{
		AckHandle: 12345,
	}

	handleMsgMhfPostBoostTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostBoostTimeLimit(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostBoostTimeLimit{
		AckHandle: 12345,
	}

	handleMsgMhfPostBoostTimeLimit(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestCafeBonusStruct(t *testing.T) {
	// Test CafeBonus struct can be created
	bonus := CafeBonus{
		ID:       1,
		TimeReq:  3600,
		ItemType: 1,
		ItemID:   100,
		Quantity: 5,
		Claimed:  false,
	}

	if bonus.ID != 1 {
		t.Errorf("ID = %d, want 1", bonus.ID)
	}
	if bonus.TimeReq != 3600 {
		t.Errorf("TimeReq = %d, want 3600", bonus.TimeReq)
	}
	if bonus.Claimed {
		t.Error("Claimed should be false")
	}
}
