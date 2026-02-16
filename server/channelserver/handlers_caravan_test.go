package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetRyoudama(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRyoudama{
		AckHandle: 12345,
	}

	handleMsgMhfGetRyoudama(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostRyoudama(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfPostRyoudama panicked: %v", r)
		}
	}()

	handleMsgMhfPostRyoudama(session, nil)
}

func TestHandleMsgMhfGetTinyBin(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTinyBin{
		AckHandle: 12345,
	}

	handleMsgMhfGetTinyBin(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Response might be empty bytes
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostTinyBin(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostTinyBin{
		AckHandle: 12345,
	}

	handleMsgMhfPostTinyBin(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfCaravanMyScore(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCaravanMyScore{
		AckHandle: 12345,
	}

	handleMsgMhfCaravanMyScore(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfCaravanRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCaravanRanking{
		AckHandle: 12345,
	}

	handleMsgMhfCaravanRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfCaravanMyRank(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCaravanMyRank{
		AckHandle: 12345,
	}

	handleMsgMhfCaravanMyRank(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}
