package channelserver

import (
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"testing"
	"time"
)

func TestHandleMsgMhfInfoFesta_OverrideZero(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	srv.erupeConfig.DebugOptions.FestaOverride = 0
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoFesta{AckHandle: 1}
	handleMsgMhfInfoFesta(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfInfoFesta_WithActiveEvent(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.DebugOptions.FestaOverride = 1
	srv.erupeConfig.RealClientMode = cfg.ZZ
	srv.erupeConfig.GameplayOptions.MaximumFP = 50000
	srv.festaRepo = &mockFestaRepo{
		events: []FestaEvent{{ID: 1, StartTime: uint32(time.Now().Add(-24 * time.Hour).Unix())}},
		trials: []FestaTrial{
			{ID: 1, Objective: 1, GoalID: 100, TimesReq: 5, Locale: 0, Reward: 10, Monopoly: "blue"},
		},
	}
	ensureFestaService(srv)
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoFesta{AckHandle: 1}
	handleMsgMhfInfoFesta(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfInfoFesta_FutureTimestamp(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.DebugOptions.FestaOverride = -1
	srv.festaRepo = &mockFestaRepo{
		events: []FestaEvent{{ID: 1, StartTime: uint32(time.Now().Add(72 * time.Hour).Unix())}},
	}
	ensureFestaService(srv)
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoFesta{AckHandle: 1}
	handleMsgMhfInfoFesta(s, pkt)
	<-s.sendPackets
}
