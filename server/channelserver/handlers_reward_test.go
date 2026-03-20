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

	pkt := &mhfpacket.MsgMhfUseRewardSong{AckHandle: 12345}
	handleMsgMhfUseRewardSong(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
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

// Tests consolidated from handlers_coverage3_test.go

func TestSimpleAckHandlers_RewardGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetRewardSong", func(s *Session) {
			handleMsgMhfGetRewardSong(s, &mhfpacket.MsgMhfGetRewardSong{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

func TestNonTrivialHandlers_RewardGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetAdditionalBeatReward", func(s *Session) {
			handleMsgMhfGetAdditionalBeatReward(s, &mhfpacket.MsgMhfGetAdditionalBeatReward{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdRankingRewardList", func(s *Session) {
			handleMsgMhfGetUdRankingRewardList(s, &mhfpacket.MsgMhfGetUdRankingRewardList{AckHandle: 1})
		}},
		{"handleMsgMhfAcquireMonthlyReward", func(s *Session) {
			handleMsgMhfAcquireMonthlyReward(s, &mhfpacket.MsgMhfAcquireMonthlyReward{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

func TestEmptyHandlers_MiscFiles_Reward(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Handlers that accept nil and take no action (no AckHandle).
	nilSafeTests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgMhfAddRewardSongCount", func() { handleMsgMhfAddRewardSongCount(session, nil) }},
		{"handleMsgMhfAcceptReadReward", func() { handleMsgMhfAcceptReadReward(session, nil) }},
	}

	for _, tt := range nilSafeTests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}

	// handleMsgMhfUseRewardSong is a real handler (requires a typed packet).
	t.Run("handleMsgMhfUseRewardSong", func(t *testing.T) {
		pkt := &mhfpacket.MsgMhfUseRewardSong{AckHandle: 1}
		handleMsgMhfUseRewardSong(session, pkt)
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("handleMsgMhfUseRewardSong: response should have data")
			}
		default:
			t.Error("handleMsgMhfUseRewardSong: no response queued")
		}
	})
}
