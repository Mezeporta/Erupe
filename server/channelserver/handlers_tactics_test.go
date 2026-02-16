package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetUdTacticsPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsPoint{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAddUdTacticsPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAddUdTacticsPoint{
		AckHandle: 12345,
	}

	handleMsgMhfAddUdTacticsPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTacticsRewardList(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsRewardList{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsRewardList(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTacticsFollower(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsFollower{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsFollower(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTacticsBonusQuest(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsBonusQuest{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsBonusQuest(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTacticsFirstQuestBonus(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsFirstQuestBonus{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsFirstQuestBonus(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTacticsRemainingPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsRemainingPoint{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsRemainingPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTacticsRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTacticsRanking{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTacticsRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSetUdTacticsFollower(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfSetUdTacticsFollower panicked: %v", r)
		}
	}()

	handleMsgMhfSetUdTacticsFollower(session, nil)
}

func TestHandleMsgMhfGetUdTacticsLog(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetUdTacticsLog panicked: %v", r)
		}
	}()

	handleMsgMhfGetUdTacticsLog(session, nil)
}
