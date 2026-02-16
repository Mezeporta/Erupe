package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetAdditionalBeatReward(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetAdditionalBeatReward{
		AckHandle: 12345,
	}

	handleMsgMhfGetAdditionalBeatReward(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdRankingRewardList(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdRankingRewardList{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdRankingRewardList(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetRewardSong(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRewardSong{
		AckHandle: 12345,
	}

	handleMsgMhfGetRewardSong(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUseRewardSong(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfUseRewardSong panicked: %v", r)
		}
	}()

	handleMsgMhfUseRewardSong(session, nil)
}

func TestHandleMsgMhfAddRewardSongCount(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfAddRewardSongCount panicked: %v", r)
		}
	}()

	handleMsgMhfAddRewardSongCount(session, nil)
}

func TestHandleMsgMhfAcquireMonthlyReward(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireMonthlyReward{
		AckHandle: 12345,
	}

	handleMsgMhfAcquireMonthlyReward(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAcceptReadReward(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfAcceptReadReward panicked: %v", r)
		}
	}()

	handleMsgMhfAcceptReadReward(session, nil)
}
