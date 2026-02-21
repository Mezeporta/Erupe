package channelserver

import (
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
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
