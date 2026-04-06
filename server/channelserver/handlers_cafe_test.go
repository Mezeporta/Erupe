package channelserver

import (
	"errors"
	"erupe-ce/common/mhfcourse"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"testing"
	"time"
)

func TestHandleMsgMhfGetBoostTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostTime{
		AckHandle: 12345,
	}

	handleMsgMhfGetBoostTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Response should be empty bytes for this handler
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostBoostTimeQuestReturn(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostBoostTimeQuestReturn{
		AckHandle: 12345,
	}

	handleMsgMhfPostBoostTimeQuestReturn(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostBoostTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostBoostTime{
		AckHandle: 12345,
	}

	handleMsgMhfPostBoostTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostBoostTimeLimit(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostBoostTimeLimit{
		AckHandle: 12345,
	}

	handleMsgMhfPostBoostTimeLimit(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestCafeBonusStruct(t *testing.T) {
	// Test CafeBonus struct can be created
	bonus := CafeBonus{
		ID:       1,
		TimeReq:  3600,
		ItemType: 1,
		ItemID:   100,
		Quantity: 5,
		Claimed:  false,
	}

	if bonus.ID != 1 {
		t.Errorf("ID = %d, want 1", bonus.ID)
	}
	if bonus.TimeReq != 3600 {
		t.Errorf("TimeReq = %d, want 3600", bonus.TimeReq)
	}
	if bonus.Claimed {
		t.Error("Claimed should be false")
	}
}

// --- Mock-based handler tests ---

func TestHandleMsgMhfUpdateCafepoint(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.ints["netcafe_points"] = 150
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateCafepoint{AckHandle: 100}

	handleMsgMhfUpdateCafepoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireCafeItem(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.ints["netcafe_points"] = 500
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireCafeItem{
		AckHandle: 100,
		PointCost: 200,
	}

	handleMsgMhfAcquireCafeItem(session, pkt)

	if charMock.ints["netcafe_points"] != 300 {
		t.Errorf("netcafe_points = %d, want 300 (500-200)", charMock.ints["netcafe_points"])
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfStartBoostTime_Disabled(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.GameplayOptions.DisableBoostTime = true
	charMock := newMockCharacterRepo()
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfStartBoostTime{AckHandle: 100}

	handleMsgMhfStartBoostTime(session, pkt)

	// When disabled, boost_time should NOT be saved
	if _, ok := charMock.times["boost_time"]; ok {
		t.Error("boost_time should not be saved when disabled")
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfStartBoostTime_Enabled(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.GameplayOptions.DisableBoostTime = false
	server.erupeConfig.GameplayOptions.BoostTimeDuration = 3600
	charMock := newMockCharacterRepo()
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfStartBoostTime{AckHandle: 100}

	handleMsgMhfStartBoostTime(session, pkt)

	savedTime, ok := charMock.times["boost_time"]
	if !ok {
		t.Fatal("boost_time should be saved")
	}
	if savedTime.Before(time.Now()) {
		t.Error("boost_time should be in the future")
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetBoostTimeLimit(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	future := time.Now().Add(1 * time.Hour)
	charMock.times["boost_time"] = future
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostTimeLimit{AckHandle: 100}

	handleMsgMhfGetBoostTimeLimit(session, pkt)

	// This handler sends two responses (doAckBufSucceed + doAckSimpleSucceed)
	count := 0
	for {
		select {
		case <-session.sendPackets:
			count++
		default:
			goto done
		}
	}
done:
	if count != 2 {
		t.Errorf("Expected 2 response packets, got %d", count)
	}
}

func TestHandleMsgMhfGetBoostTimeLimit_NoBoost(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.readErr = errNotFound
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostTimeLimit{AckHandle: 100}

	handleMsgMhfGetBoostTimeLimit(session, pkt)

	// Should still send responses even on error
	count := 0
	for {
		select {
		case <-session.sendPackets:
			count++
		default:
			goto done2
		}
	}
done2:
	if count < 1 {
		t.Error("Should queue at least one response packet")
	}
}

func TestHandleMsgMhfGetBoostRight_Active(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.times["boost_time"] = time.Now().Add(1 * time.Hour) // Future = active
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostRight{AckHandle: 100}

	handleMsgMhfGetBoostRight(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetBoostRight_Expired(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.times["boost_time"] = time.Now().Add(-1 * time.Hour) // Past = expired
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostRight{AckHandle: 100}

	handleMsgMhfGetBoostRight(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetBoostRight_NoRecord(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.readErr = errNotFound
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoostRight{AckHandle: 100}

	handleMsgMhfGetBoostRight(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

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
	// No cafe_reset set -> ReadTime returns error -> sets new reset time
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

// ackBufPayload extracts the payload bytes from a queued buffered-ack packet.
// Layout: opcode(u16) + ackHandle(u32) + isBuf(u8) + err(u8) + len(u16) + data.
func ackBufPayload(t *testing.T, data []byte) []byte {
	t.Helper()
	const headerLen = 10
	if len(data) < headerLen {
		t.Fatalf("ack packet too short: %d bytes", len(data))
	}
	return data[headerLen:]
}

// Regression for #187: GetBoostTimeLimit must return 0 when DisableBoostTime
// is set, overriding any stored boost_time.
func TestHandleMsgMhfGetBoostTimeLimit_DisableBoostTime(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.GameplayOptions.DisableBoostTime = true
	charMock := newMockCharacterRepo()
	charMock.times["boost_time"] = time.Now().Add(1 * time.Hour)
	server.charRepo = charMock
	session := createMockSession(1, server)

	handleMsgMhfGetBoostTimeLimit(session, &mhfpacket.MsgMhfGetBoostTimeLimit{AckHandle: 100})

	p := <-session.sendPackets
	payload := ackBufPayload(t, p.data)
	if len(payload) != 4 || payload[0] != 0 || payload[1] != 0 || payload[2] != 0 || payload[3] != 0 {
		t.Errorf("expected zero uint32 payload, got %x", payload)
	}
}

// Regression for #187: GetBoostRight must report "no right" when disabled.
func TestHandleMsgMhfGetBoostRight_DisableBoostTime(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.GameplayOptions.DisableBoostTime = true
	charMock := newMockCharacterRepo()
	charMock.times["boost_time"] = time.Now().Add(1 * time.Hour)
	server.charRepo = charMock
	session := createMockSession(1, server)

	handleMsgMhfGetBoostRight(session, &mhfpacket.MsgMhfGetBoostRight{AckHandle: 100})

	p := <-session.sendPackets
	payload := ackBufPayload(t, p.data)
	want := []byte{0x00, 0x00, 0x00, 0x00}
	if len(payload) != 4 || payload[0] != want[0] || payload[1] != want[1] || payload[2] != want[2] || payload[3] != want[3] {
		t.Errorf("expected %x, got %x", want, payload)
	}
}
