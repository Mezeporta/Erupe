package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetTowerInfo_TowerRankPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{
		AckHandle: 12345,
		InfoType:  mhfpacket.TowerInfoTypeTowerRankPoint,
	}

	handleMsgMhfGetTowerInfo(session, pkt)

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

func TestHandleMsgMhfGetTowerInfo_GetOwnTowerSkill(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{
		AckHandle: 12345,
		InfoType:  mhfpacket.TowerInfoTypeGetOwnTowerSkill,
	}

	handleMsgMhfGetTowerInfo(session, pkt)

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

func TestHandleMsgMhfGetTowerInfo_TowerTouhaHistory(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{
		AckHandle: 12345,
		InfoType:  mhfpacket.TowerInfoTypeTowerTouhaHistory,
	}

	handleMsgMhfGetTowerInfo(session, pkt)

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

func TestHandleMsgMhfGetTowerInfo_Unk5(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTowerInfo{
		AckHandle: 12345,
		InfoType:  mhfpacket.TowerInfoTypeUnk5,
	}

	handleMsgMhfGetTowerInfo(session, pkt)

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

func TestHandleMsgMhfGetTenrouirai_Type4(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetTenrouirai{
		AckHandle: 12345,
		Unk0:      0,
		Unk2:      4,
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
		Unk2:      0,
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

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetBreakSeibatuLevelReward panicked: %v", r)
		}
	}()

	handleMsgMhfGetBreakSeibatuLevelReward(session, nil)
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

func TestHandleMsgMhfGetGemInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetGemInfo{
		AckHandle: 12345,
	}

	handleMsgMhfGetGemInfo(session, pkt)

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

func TestHandleMsgMhfPostGemInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfPostGemInfo panicked: %v", r)
		}
	}()

	handleMsgMhfPostGemInfo(session, nil)
}
