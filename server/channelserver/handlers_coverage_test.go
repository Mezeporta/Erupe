package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// Tests for handlers that do NOT require database access, exercising additional
// code paths not covered by existing test files (handlers_core_test.go,
// handlers_rengoku_test.go, etc.).

// TestHandleMsgSysPing_DifferentAckHandles verifies ping works with various ack handles.
func TestHandleMsgSysPing_DifferentAckHandles(t *testing.T) {
	server := createMockServer()

	ackHandles := []uint32{0, 1, 99999, 0xFFFFFFFF}
	for _, ack := range ackHandles {
		session := createMockSession(1, server)
		pkt := &mhfpacket.MsgSysPing{AckHandle: ack}

		handleMsgSysPing(session, pkt)

		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Errorf("AckHandle=%d: Response packet should have data", ack)
			}
		default:
			t.Errorf("AckHandle=%d: No response packet queued", ack)
		}
	}
}

// TestHandleMsgSysTerminalLog_NoEntries verifies the handler works with nil entries.
func TestHandleMsgSysTerminalLog_NoEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 99999,
		LogID:     0,
		Entries:   nil,
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysTerminalLog_ManyEntries verifies the handler with many log entries.
func TestHandleMsgSysTerminalLog_ManyEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	entries := make([]*mhfpacket.TerminalLogEntry, 20)
	for i := range entries {
		entries[i] = &mhfpacket.TerminalLogEntry{
			Index: uint32(i),
			Type1: uint8(i % 256),
			Type2: uint8((i + 1) % 256),
			Data:  make([]int16, 15),
		}
	}

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 55555,
		LogID:     42,
		Entries:   entries,
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysTime_MultipleCalls verifies calling time handler repeatedly.
func TestHandleMsgSysTime_MultipleCalls(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
		Timestamp:     0,
	}

	for i := 0; i < 5; i++ {
		handleMsgSysTime(session, pkt)
	}

	// Should have 5 queued responses
	count := 0
	for {
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("Response packet should have data")
			}
			count++
		default:
			goto done
		}
	}
done:
	if count != 5 {
		t.Errorf("Expected 5 queued responses, got %d", count)
	}
}

// TestHandleMsgMhfGetRengokuRankingRank_DifferentAck verifies rengoku ranking
// works with different ack handles.
func TestHandleMsgMhfGetRengokuRankingRank_DifferentAck(t *testing.T) {
	server := createMockServer()

	ackHandles := []uint32{0, 1, 54321, 0xDEADBEEF}
	for _, ack := range ackHandles {
		session := createMockSession(1, server)
		pkt := &mhfpacket.MsgMhfGetRengokuRankingRank{AckHandle: ack}

		handleMsgMhfGetRengokuRankingRank(session, pkt)

		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Errorf("AckHandle=%d: Response packet should have data", ack)
			}
		default:
			t.Errorf("AckHandle=%d: No response packet queued", ack)
		}
	}
}
