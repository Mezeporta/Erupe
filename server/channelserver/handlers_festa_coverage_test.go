package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"testing"
)

func TestHandleMsgMhfSaveMezfesData(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfSaveMezfesData{AckHandle: 1, RawDataPayload: []byte{0x01, 0x02}}
	handleMsgMhfSaveMezfesData(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfLoadMezfesData(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfLoadMezfesData{AckHandle: 1}
	handleMsgMhfLoadMezfesData(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfVoteFesta(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfVoteFesta{AckHandle: 1, TrialID: 42}
	handleMsgMhfVoteFesta(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfEntryFesta_NoGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{getErr: errNotFound}
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEntryFesta{AckHandle: 1}
	handleMsgMhfEntryFesta(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfEntryFesta_WithGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{guild: &Guild{ID: 1}}
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEntryFesta{AckHandle: 1}
	handleMsgMhfEntryFesta(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfChargeFesta(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	srv.guildRepo = &mockGuildRepo{guild: &Guild{ID: 1}}
	ensureFestaService(srv)
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfChargeFesta{AckHandle: 1, GuildID: 1, Souls: []uint16{10, 20}}
	handleMsgMhfChargeFesta(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireFesta(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfAcquireFesta{AckHandle: 1}
	handleMsgMhfAcquireFesta(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireFestaPersonalPrize(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfAcquireFestaPersonalPrize{AckHandle: 1, PrizeID: 5}
	handleMsgMhfAcquireFestaPersonalPrize(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireFestaIntermediatePrize(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfAcquireFestaIntermediatePrize{AckHandle: 1, PrizeID: 3}
	handleMsgMhfAcquireFestaIntermediatePrize(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateFestaPersonalPrize(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{
		prizes: []Prize{
			{ID: 1, Tier: 1, SoulsReq: 100, ItemID: 5, NumItem: 1, Claimed: 0},
		},
	}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateFestaPersonalPrize{AckHandle: 1}
	handleMsgMhfEnumerateFestaPersonalPrize(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateFestaIntermediatePrize(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateFestaIntermediatePrize{AckHandle: 1}
	handleMsgMhfEnumerateFestaIntermediatePrize(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfStateFestaU_NoGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{getErr: errNotFound}
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfStateFestaU{AckHandle: 1}
	handleMsgMhfStateFestaU(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfStateFestaU_WithGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{guild: &Guild{ID: 1}}
	srv.festaRepo = &mockFestaRepo{charSouls: 50}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfStateFestaU{AckHandle: 1}
	handleMsgMhfStateFestaU(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfStateFestaG_NoGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{getErr: errNotFound}
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfStateFestaG{AckHandle: 1}
	handleMsgMhfStateFestaG(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfStateFestaG_WithGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{guild: &Guild{ID: 1, Souls: 500}}
	srv.festaRepo = &mockFestaRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfStateFestaG{AckHandle: 1}
	handleMsgMhfStateFestaG(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateFestaMember_NoGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{getErr: errNotFound}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateFestaMember{AckHandle: 1}
	handleMsgMhfEnumerateFestaMember(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateFestaMember_WithMembers(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{
		guild: &Guild{ID: 1},
		members: []*GuildMember{
			{CharID: 1, Souls: 100},
			{CharID: 2, Souls: 50},
			{CharID: 3, Souls: 0},
		},
	}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateFestaMember{AckHandle: 1}
	handleMsgMhfEnumerateFestaMember(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}
