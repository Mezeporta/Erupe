package channelserver

import (
	"strings"
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

// --- EmptyTowerCSV tests ---

func TestEmptyTowerCSV(t *testing.T) {
	result := EmptyTowerCSV(3)
	if result != "0,0,0" {
		t.Errorf("EmptyTowerCSV(3) = %q, want %q", result, "0,0,0")
	}

	result = EmptyTowerCSV(1)
	if result != "0" {
		t.Errorf("EmptyTowerCSV(1) = %q, want %q", result, "0")
	}

	result = EmptyTowerCSV(5)
	parts := strings.Split(result, ",")
	if len(parts) != 5 {
		t.Errorf("EmptyTowerCSV(5) has %d parts, want 5", len(parts))
	}
}

// --- handleMsgMhfGetTowerInfo tests ---

func TestGetTowerInfo_InfoType1_TRP(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{
		towerData: TowerData{TR: 10, TRP: 100},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{AckHandle: 100, InfoType: 1}
	handleMsgMhfGetTowerInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetTowerInfo_InfoType2_Skills(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{
		towerData: TowerData{TSP: 50},
		skills:    "1,2,3",
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{AckHandle: 100, InfoType: 2}
	handleMsgMhfGetTowerInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetTowerInfo_InfoType3_Level(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{
		towerData: TowerData{Block1: 5, Block2: 3},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{AckHandle: 100, InfoType: 3}
	handleMsgMhfGetTowerInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetTowerInfo_InfoType4_History(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{AckHandle: 100, InfoType: 4}
	handleMsgMhfGetTowerInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetTowerInfo_InfoType5_Level(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{
		towerData: TowerData{Block1: 10, Block2: 7},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{AckHandle: 100, InfoType: 5}
	handleMsgMhfGetTowerInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetTowerInfo_DBError(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{towerDataErr: errNotFound}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{AckHandle: 100, InfoType: 1}
	handleMsgMhfGetTowerInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfGetGemInfo tests ---

func TestGetGemInfo_QueryType1_Gems(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{gems: "1,2,3,4,5"}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetGemInfo{AckHandle: 100, QueryType: 1}
	handleMsgMhfGetGemInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetGemInfo_QueryType2_History(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{gems: "0,0,0"}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetGemInfo{AckHandle: 100, QueryType: 2}
	handleMsgMhfGetGemInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGetGemInfo_NoGems(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{gems: ""}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetGemInfo{AckHandle: 100, QueryType: 1}
	handleMsgMhfGetGemInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfPostGemInfo tests ---

func TestPostGemInfo_AddGem(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{gems: "0,0,0,0,0"}
	ensureTowerService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostGemInfo{AckHandle: 100, Op: 1, Gem: 0x0101, Quantity: 5}
	handleMsgMhfPostGemInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestPostGemInfo_Transfer(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{gems: "0,0,0,0,0"}
	ensureTowerService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostGemInfo{AckHandle: 100, Op: 2, Gem: 0x0101, Quantity: 1}
	handleMsgMhfPostGemInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestPostGemInfo_DebugMode(t *testing.T) {
	server := createMockServer()
	server.towerRepo = &mockTowerRepo{gems: "0,0,0,0,0"}
	server.erupeConfig.DebugOptions.QuestTools = true
	ensureTowerService(server)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostGemInfo{AckHandle: 100, Op: 1, Gem: 0x0101, Quantity: 3}
	handleMsgMhfPostGemInfo(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfGetNotice / handleMsgMhfPostNotice tests ---

func TestGetNotice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetNotice{AckHandle: 100}
	handleMsgMhfGetNotice(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestPostNotice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostNotice{AckHandle: 100}
	handleMsgMhfPostNotice(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
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
