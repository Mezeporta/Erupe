package channelserver

import (
	"erupe-ce/common/mhfcourse"
	"erupe-ce/network/mhfpacket"
	"errors"
	"testing"
	"time"
)

// --- Cafe Duration Bonus Info tests ---

func TestHandleMsgMhfGetCafeDurationBonusInfo_Error(t *testing.T) {
	srv := createMockServer()
	srv.cafeRepo = &mockCafeRepo{bonusesErr: errors.New("db error")}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetCafeDurationBonusInfo{AckHandle: 1}
	handleMsgMhfGetCafeDurationBonusInfo(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetCafeDurationBonusInfo_WithBonuses(t *testing.T) {
	srv := createMockServer()
	srv.cafeRepo = &mockCafeRepo{
		bonuses: []CafeBonus{
			{ID: 1, TimeReq: 100, ItemType: 5, ItemID: 10, Quantity: 2, Claimed: false},
			{ID: 2, TimeReq: 200, ItemType: 6, ItemID: 20, Quantity: 1, Claimed: true},
		},
	}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetCafeDurationBonusInfo{AckHandle: 1}
	handleMsgMhfGetCafeDurationBonusInfo(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

// --- Receive Cafe Duration Bonus tests ---

func TestHandleMsgMhfReceiveCafeDurationBonus_Error(t *testing.T) {
	srv := createMockServer()
	srv.cafeRepo = &mockCafeRepo{claimableErr: errors.New("db error")}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfReceiveCafeDurationBonus{AckHandle: 1}
	handleMsgMhfReceiveCafeDurationBonus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfReceiveCafeDurationBonus_WithClaimable(t *testing.T) {
	srv := createMockServer()
	srv.cafeRepo = &mockCafeRepo{
		claimable: []CafeBonus{
			{ID: 1, ItemType: 5, ItemID: 10, Quantity: 2},
		},
	}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)
	// Course 30 is required for claimable items
	s.courses = []mhfcourse.Course{{ID: 30}}

	pkt := &mhfpacket.MsgMhfReceiveCafeDurationBonus{AckHandle: 1}
	handleMsgMhfReceiveCafeDurationBonus(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

// --- Post Cafe Duration Bonus Received tests ---

func TestHandleMsgMhfPostCafeDurationBonusReceived_Empty(t *testing.T) {
	srv := createMockServer()
	srv.cafeRepo = &mockCafeRepo{}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfPostCafeDurationBonusReceived{AckHandle: 1, CafeBonusID: []uint32{}}
	handleMsgMhfPostCafeDurationBonusReceived(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfPostCafeDurationBonusReceived_WithBonusIDs(t *testing.T) {
	srv := createMockServer()
	srv.cafeRepo = &mockCafeRepo{
		bonusItemType: 17, // netcafe point type
		bonusItemQty:  100,
	}
	charRepo := newMockCharacterRepo()
	charRepo.ints["netcafe_points"] = 50
	srv.charRepo = charRepo
	s := createMockSession(100, srv)
	srv.erupeConfig.GameplayOptions.MaximumNP = 99999

	pkt := &mhfpacket.MsgMhfPostCafeDurationBonusReceived{AckHandle: 1, CafeBonusID: []uint32{1, 2}}
	handleMsgMhfPostCafeDurationBonusReceived(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

// --- Daily Cafe Point tests ---

func TestHandleMsgMhfCheckDailyCafepoint_Eligible(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	// Set daily_time far in the past so midday.After(dailyTime) is true
	charRepo.times["daily_time"] = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	charRepo.ints["netcafe_points"] = 10
	srv.charRepo = charRepo
	srv.erupeConfig.GameplayOptions.MaximumNP = 99999
	srv.erupeConfig.GameplayOptions.BonusQuestAllowance = 10
	srv.erupeConfig.GameplayOptions.DailyQuestAllowance = 5
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfCheckDailyCafepoint{AckHandle: 1}
	handleMsgMhfCheckDailyCafepoint(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfCheckDailyCafepoint_NotEligible(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	// Set daily_time far in the future so midday.After(dailyTime) is false
	charRepo.times["daily_time"] = time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfCheckDailyCafepoint{AckHandle: 1}
	handleMsgMhfCheckDailyCafepoint(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

// --- Cafe Duration tests ---

func TestHandleMsgMhfGetCafeDuration(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	// cafe_reset in the future so we don't trigger reset logic
	charRepo.times["cafe_reset"] = time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	charRepo.ints["cafe_time"] = 3600
	srv.charRepo = charRepo
	srv.cafeRepo = &mockCafeRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetCafeDuration{AckHandle: 1}
	handleMsgMhfGetCafeDuration(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}
