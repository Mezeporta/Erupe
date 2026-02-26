package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// --- Guild mission handler tests ---

func TestHandleMsgMhfAddGuildMissionCount(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfAddGuildMissionCount(session, &mhfpacket.MsgMhfAddGuildMissionCount{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfSetGuildMissionTarget(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfSetGuildMissionTarget(session, &mhfpacket.MsgMhfSetGuildMissionTarget{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfCancelGuildMissionTarget(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfCancelGuildMissionTarget(session, &mhfpacket.MsgMhfCancelGuildMissionTarget{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfGetGuildMissionRecord(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetGuildMissionRecord(session, &mhfpacket.MsgMhfGetGuildMissionRecord{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfGetGuildMissionList(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetGuildMissionList(session, &mhfpacket.MsgMhfGetGuildMissionList{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}
