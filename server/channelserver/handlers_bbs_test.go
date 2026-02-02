package channelserver

import (
	"testing"

	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetBbsUserStatus(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBbsUserStatus{
		AckHandle: 12345,
	}

	handleMsgMhfGetBbsUserStatus(session, pkt)

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

func TestHandleMsgMhfGetBbsSnsStatus(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBbsSnsStatus{
		AckHandle: 12345,
	}

	handleMsgMhfGetBbsSnsStatus(session, pkt)

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

func TestHandleMsgMhfApplyBbsArticle(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &config.Config{
		ScreenshotAPIURL: "http://example.com/api",
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfApplyBbsArticle{
		AckHandle: 12345,
	}

	handleMsgMhfApplyBbsArticle(session, pkt)

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
