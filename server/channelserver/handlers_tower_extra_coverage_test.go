package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"testing"
)

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
