package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfLoadFavoriteQuest(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadFavoriteQuest{AckHandle: 1}
	handleMsgMhfLoadFavoriteQuest(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("expected response")
	}
}

func TestHandleMsgMhfSaveFavoriteQuest(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfSaveFavoriteQuest{
		AckHandle: 1,
		Data:      []byte{0x01, 0x00, 0x01, 0x00, 0x01},
	}
	handleMsgMhfSaveFavoriteQuest(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("expected response")
	}
}
