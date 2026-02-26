package channelserver

import (
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"testing"
	"time"
)

func TestHandleMsgMhfGetCafeDuration_ResetPath(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	// cafe_reset in the past to trigger reset logic
	charRepo.times["cafe_reset"] = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	charRepo.ints["cafe_time"] = 1800
	srv.charRepo = charRepo
	srv.cafeRepo = &mockCafeRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetCafeDuration{AckHandle: 1}
	handleMsgMhfGetCafeDuration(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetCafeDuration_NoResetTime(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	// No cafe_reset set → ReadTime returns error → sets new reset time
	charRepo.ints["cafe_time"] = 100
	srv.charRepo = charRepo
	srv.cafeRepo = &mockCafeRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetCafeDuration{AckHandle: 1}
	handleMsgMhfGetCafeDuration(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetCafeDuration_ZZClient(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.RealClientMode = cfg.ZZ
	charRepo := newMockCharacterRepo()
	charRepo.times["cafe_reset"] = time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	charRepo.ints["cafe_time"] = 3600
	srv.charRepo = charRepo
	srv.cafeRepo = &mockCafeRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetCafeDuration{AckHandle: 1}
	handleMsgMhfGetCafeDuration(s, pkt)
	<-s.sendPackets
}
