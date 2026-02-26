package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetTenrouirai_Type1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{
		AckHandle: 12345,
		Unk0:      1,
	}

	handleMsgMhfGetTenrouirai(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetTenrouirai_Default(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{
		AckHandle: 12345,
		Unk0:      0,
		DataType:  0,
	}

	handleMsgMhfGetTenrouirai(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostTowerInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostTowerInfo{
		AckHandle: 12345,
	}

	handleMsgMhfPostTowerInfo(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostTenrouirai(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostTenrouirai{
		AckHandle: 12345,
	}

	handleMsgMhfPostTenrouirai(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetBreakSeibatuLevelReward(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBreakSeibatuLevelReward{
		AckHandle: 12345,
	}

	handleMsgMhfGetBreakSeibatuLevelReward(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetWeeklySeibatuRankingReward(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetWeeklySeibatuRankingReward{
		AckHandle: 12345,
	}

	handleMsgMhfGetWeeklySeibatuRankingReward(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPresentBox(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPresentBox{
		AckHandle: 12345,
	}

	handleMsgMhfPresentBox(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetTenrouirai_Type2_Rewards(t *testing.T) {
	srv := createMockServer()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, DataType: 2}
	handleMsgMhfGetTenrouirai(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetTenrouirai_Type4_Progress(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	ensureTowerService(srv)
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, DataType: 4, GuildID: 1}
	handleMsgMhfGetTenrouirai(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetTenrouirai_Type5_Scores(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, DataType: 5, GuildID: 1, MissionIndex: 0}
	handleMsgMhfGetTenrouirai(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetTenrouirai_Type6_RP(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, DataType: 6, GuildID: 1}
	handleMsgMhfGetTenrouirai(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfPostTowerInfo_SkillUpdate(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostTowerInfo{AckHandle: 1, InfoType: 2, Skill: 3, Cost: -10}
	handleMsgMhfPostTowerInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfPostTowerInfo_ProgressUpdate(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostTowerInfo{AckHandle: 1, InfoType: 1, TR: 5, TRP: 100, Cost: -20, Block1: 1}
	handleMsgMhfPostTowerInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfPostTowerInfo_ProgressType7(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostTowerInfo{AckHandle: 1, InfoType: 7, TR: 10, TRP: 200}
	handleMsgMhfPostTowerInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfPostTowerInfo_QuestToolsDebug(t *testing.T) {
	srv := createMockServer()
	srv.towerRepo = &mockTowerRepo{}
	srv.erupeConfig.DebugOptions.QuestTools = true
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostTowerInfo{AckHandle: 1, InfoType: 2, Skill: 1}
	handleMsgMhfPostTowerInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfPostTenrouirai_Op1(t *testing.T) {
	srv := createMockServer()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostTenrouirai{AckHandle: 1, Op: 1}
	handleMsgMhfPostTenrouirai(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfPostTenrouirai_QuestToolsDebug(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.DebugOptions.QuestTools = true
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostTenrouirai{AckHandle: 1, Op: 1, Floors: 10, Slays: 5}
	handleMsgMhfPostTenrouirai(s, pkt)
	<-s.sendPackets
}

// Tests consolidated from handlers_coverage3_test.go

func TestNonTrivialHandlers_TowerGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetTenrouirai_Type1_C3", func(s *Session) {
			handleMsgMhfGetTenrouirai(s, &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, Unk0: 1})
		}},
		{"handleMsgMhfGetTenrouirai_Unknown_C3", func(s *Session) {
			handleMsgMhfGetTenrouirai(s, &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, Unk0: 0, DataType: 0})
		}},
		{"handleMsgMhfGetWeeklySeibatuRankingReward_C3", func(s *Session) {
			handleMsgMhfGetWeeklySeibatuRankingReward(s, &mhfpacket.MsgMhfGetWeeklySeibatuRankingReward{AckHandle: 1})
		}},
		{"handleMsgMhfPresentBox_C3", func(s *Session) {
			handleMsgMhfPresentBox(s, &mhfpacket.MsgMhfPresentBox{AckHandle: 1})
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
