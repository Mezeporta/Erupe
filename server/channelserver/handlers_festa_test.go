package channelserver

import (
	"testing"

	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfEnumerateRanking_Default(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			TournamentEvent: 0, // Default state
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRanking{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateRanking_State1(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			TournamentEvent: 1,
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRanking{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateRanking_State2(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			TournamentEvent: 2,
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRanking{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateRanking_State3(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			TournamentEvent: 3,
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRanking{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfVoteFesta(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfVoteFesta{
		AckHandle: 12345,
	}

	handleMsgMhfVoteFesta(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestEmptyFestaHandlers(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name    string
		handler func(s *Session, p mhfpacket.MHFPacket)
	}{
		{"handleMsgMhfEntryTournament", handleMsgMhfEntryTournament},
		{"handleMsgMhfAcquireTournament", handleMsgMhfAcquireTournament},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.handler(session, nil)
		})
	}
}
