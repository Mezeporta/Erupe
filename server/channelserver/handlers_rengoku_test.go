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
