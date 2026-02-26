package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"testing"
	"time"
)

func TestHandleMsgMhfGetWeeklySchedule(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{}
	srv.erupeConfig.GameplayOptions.MinFeatureWeapons = 1
	srv.erupeConfig.GameplayOptions.MaxFeatureWeapons = 3
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetWeeklySchedule{AckHandle: 1}
	handleMsgMhfGetWeeklySchedule(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetKeepLoginBoostStatus_Disabled(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{}
	srv.erupeConfig.GameplayOptions.DisableLoginBoost = true
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetKeepLoginBoostStatus{AckHandle: 1}
	handleMsgMhfGetKeepLoginBoostStatus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetKeepLoginBoostStatus_EmptyBoosts(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetKeepLoginBoostStatus{AckHandle: 1}
	handleMsgMhfGetKeepLoginBoostStatus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetKeepLoginBoostStatus_WithBoosts(t *testing.T) {
	srv := createMockServer()
	srv.eventRepo = &mockEventRepo{
		loginBoosts: []loginBoost{
			{WeekReq: 1, Expiration: time.Now().Add(-7 * 24 * time.Hour)},
			{WeekReq: 2, Expiration: time.Now().Add(-14 * 24 * time.Hour)},
			{WeekReq: 3, Expiration: time.Now().Add(-21 * 24 * time.Hour)},
			{WeekReq: 4, Expiration: time.Now().Add(-28 * 24 * time.Hour)},
			{WeekReq: 5, Expiration: time.Now().Add(-35 * 24 * time.Hour)},
		},
	}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetKeepLoginBoostStatus{AckHandle: 1}
	handleMsgMhfGetKeepLoginBoostStatus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfUseKeepLoginBoost(t *testing.T) {
	tests := []struct {
		name          string
		boostWeekUsed uint8
	}{
		{"week1", 1},
		{"week2", 2},
		{"week3", 3},
		{"week4", 4},
		{"week5", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := createMockServer()
			srv.eventRepo = &mockEventRepo{}
			s := createMockSession(100, srv)

			pkt := &mhfpacket.MsgMhfUseKeepLoginBoost{AckHandle: 1, BoostWeekUsed: tt.boostWeekUsed}
			handleMsgMhfUseKeepLoginBoost(s, pkt)

			select {
			case p := <-s.sendPackets:
				if len(p.data) == 0 {
					t.Fatal("Expected non-empty response")
				}
			default:
				t.Fatal("No response packet queued")
			}
		})
	}
}

func TestHandleMsgMhfLoadScenarioData(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfLoadScenarioData{AckHandle: 1}
	handleMsgMhfLoadScenarioData(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfSaveScenarioData(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfSaveScenarioData{AckHandle: 1, RawDataPayload: []byte{0x01, 0x02, 0x03}}
	handleMsgMhfSaveScenarioData(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfListMember(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["blocked"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfListMember{AckHandle: 1}
	handleMsgMhfListMember(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfOprMember_AddBlacklist(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["blocked"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOprMember{AckHandle: 1, Blacklist: true, Operation: false, CharIDs: []uint32{42}}
	handleMsgMhfOprMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfOprMember_AddFriend(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["friends"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOprMember{AckHandle: 1, Blacklist: false, Operation: false, CharIDs: []uint32{42}}
	handleMsgMhfOprMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfOprMember_RemoveBlacklist(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.strings["blocked"] = "42"
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOprMember{AckHandle: 1, Blacklist: true, Operation: true, CharIDs: []uint32{42}}
	handleMsgMhfOprMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}
