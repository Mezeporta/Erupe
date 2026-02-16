package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfEnumerateCampaign(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateCampaign{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateCampaign(session, pkt)

	// Verify response packet was queued (fail response expected)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfStateCampaign(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfStateCampaign{
		AckHandle: 12345,
	}

	handleMsgMhfStateCampaign(session, pkt)

	// Verify response packet was queued (fail response expected)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfApplyCampaign(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfApplyCampaign{
		AckHandle: 12345,
	}

	handleMsgMhfApplyCampaign(session, pkt)

	// Verify response packet was queued (fail response expected)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}
