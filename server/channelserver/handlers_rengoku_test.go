package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetRengokuRankingRank(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRengokuRankingRank{
		AckHandle: 12345,
	}

	handleMsgMhfGetRengokuRankingRank(session, pkt)

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

func TestRengokuScoreStruct(t *testing.T) {
	score := RengokuScore{
		Name:  "TestPlayer",
		Score: 12345,
	}

	if score.Name != "TestPlayer" {
		t.Errorf("Name = %s, want TestPlayer", score.Name)
	}
	if score.Score != 12345 {
		t.Errorf("Score = %d, want 12345", score.Score)
	}
}

func TestRengokuScoreStruct_DefaultValues(t *testing.T) {
	score := RengokuScore{}

	if score.Name != "" {
		t.Errorf("Default Name should be empty, got %s", score.Name)
	}
	if score.Score != 0 {
		t.Errorf("Default Score should be 0, got %d", score.Score)
	}
}

func TestHandleMsgMhfGetRengokuRankingRank_ResponseData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRengokuRankingRank{
		AckHandle: 55555,
	}

	handleMsgMhfGetRengokuRankingRank(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestRengokuScoreStruct_Fields(t *testing.T) {
	score := RengokuScore{
		Name:  "Hunter",
		Score: 99999,
	}

	if score.Name != "Hunter" {
		t.Errorf("Name = %s, want Hunter", score.Name)
	}
	if score.Score != 99999 {
		t.Errorf("Score = %d, want 99999", score.Score)
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

// Tests consolidated from handlers_coverage3_test.go

func TestNonTrivialHandlers_RengokuGo(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgMhfGetRengokuRankingRank", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetRengokuRankingRank(session, &mhfpacket.MsgMhfGetRengokuRankingRank{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})
}
