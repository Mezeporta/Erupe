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

// Tests consolidated from handlers_coverage3_test.go

func TestSimpleAckHandlers_TacticsGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfAddUdTacticsPoint", func(s *Session) {
			handleMsgMhfAddUdTacticsPoint(s, &mhfpacket.MsgMhfAddUdTacticsPoint{AckHandle: 1})
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

func TestNonTrivialHandlers_TacticsGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetUdTacticsPoint", func(s *Session) {
			handleMsgMhfGetUdTacticsPoint(s, &mhfpacket.MsgMhfGetUdTacticsPoint{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsRewardList", func(s *Session) {
			handleMsgMhfGetUdTacticsRewardList(s, &mhfpacket.MsgMhfGetUdTacticsRewardList{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsFollower", func(s *Session) {
			handleMsgMhfGetUdTacticsFollower(s, &mhfpacket.MsgMhfGetUdTacticsFollower{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsBonusQuest", func(s *Session) {
			handleMsgMhfGetUdTacticsBonusQuest(s, &mhfpacket.MsgMhfGetUdTacticsBonusQuest{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsFirstQuestBonus", func(s *Session) {
			handleMsgMhfGetUdTacticsFirstQuestBonus(s, &mhfpacket.MsgMhfGetUdTacticsFirstQuestBonus{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsRemainingPoint", func(s *Session) {
			handleMsgMhfGetUdTacticsRemainingPoint(s, &mhfpacket.MsgMhfGetUdTacticsRemainingPoint{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsRanking", func(s *Session) {
			handleMsgMhfGetUdTacticsRanking(s, &mhfpacket.MsgMhfGetUdTacticsRanking{AckHandle: 1})
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

func TestEmptyHandlers_MiscFiles_Tactics(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgMhfSetUdTacticsFollower", func() { handleMsgMhfSetUdTacticsFollower(session, nil) }},
		{"handleMsgMhfGetUdTacticsLog", func() { handleMsgMhfGetUdTacticsLog(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}
