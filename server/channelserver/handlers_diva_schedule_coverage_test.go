package channelserver

import (
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"testing"
	"time"
)

func TestHandleMsgMhfGetUdSchedule_DivaOverrideZero_ZZ(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{}
	srv.erupeConfig.DebugOptions.DivaOverride = 0
	srv.erupeConfig.RealClientMode = cfg.ZZ
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetUdSchedule_DivaOverrideZero_OlderClient(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{}
	srv.erupeConfig.DebugOptions.DivaOverride = 0
	srv.erupeConfig.RealClientMode = cfg.G10
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetUdSchedule_DivaOverride1(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{}
	srv.erupeConfig.DebugOptions.DivaOverride = 1
	srv.erupeConfig.RealClientMode = cfg.ZZ
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetUdSchedule_DivaOverride2(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{}
	srv.erupeConfig.DebugOptions.DivaOverride = 2
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetUdSchedule_DivaOverride3(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{}
	srv.erupeConfig.DebugOptions.DivaOverride = 3
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetUdSchedule_WithExistingEvent(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{
		events: []DivaEvent{{ID: 1, StartTime: uint32(time.Now().Unix())}},
	}
	srv.erupeConfig.DebugOptions.DivaOverride = -1
	srv.erupeConfig.RealClientMode = cfg.ZZ
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetUdSchedule_NoEvents(t *testing.T) {
	srv := createMockServer()
	srv.divaRepo = &mockDivaRepo{}
	srv.erupeConfig.DebugOptions.DivaOverride = -1
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetUdSchedule{AckHandle: 1}
	handleMsgMhfGetUdSchedule(s, pkt)
	<-s.sendPackets
}
