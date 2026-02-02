package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetUdInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdInfo{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdInfo(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetKijuInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetKijuInfo{
		AckHandle: 12345,
	}

	handleMsgMhfGetKijuInfo(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSetKiju(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetKiju{
		AckHandle: 12345,
	}

	handleMsgMhfSetKiju(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAddUdPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAddUdPoint{
		AckHandle: 12345,
	}

	handleMsgMhfAddUdPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdMyPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdMyPoint{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdMyPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdTotalPointInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdTotalPointInfo{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdTotalPointInfo(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdSelectedColorInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdSelectedColorInfo{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdSelectedColorInfo(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdMonsterPoint(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdMonsterPoint{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdMonsterPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdDailyPresentList(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdDailyPresentList{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdDailyPresentList(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdNormaPresentList(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdNormaPresentList{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdNormaPresentList(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireUdItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireUdItem{
		AckHandle: 12345,
	}

	handleMsgMhfAcquireUdItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdRanking{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetUdMyRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdMyRanking{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdMyRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestGenerateDivaTimestamps_Debug(t *testing.T) {
	// Test debug mode timestamps
	tests := []struct {
		name  string
		start uint32
	}{
		{"Debug_Start1", 1},
		{"Debug_Start2", 2},
		{"Debug_Start3", 3},
	}

	server := createMockServer()
	session := createMockSession(1, server)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamps := generateDivaTimestamps(session, tt.start, true)
			if len(timestamps) != 6 {
				t.Errorf("Expected 6 timestamps, got %d", len(timestamps))
			}
			// Verify timestamps are non-zero
			for i, ts := range timestamps {
				if ts == 0 {
					t.Errorf("Timestamp %d should not be zero", i)
				}
			}
		})
	}
}
