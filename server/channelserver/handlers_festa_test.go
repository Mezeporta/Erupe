package channelserver

import (
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"time"
)

func TestHandleMsgMhfEnumerateRanking_Default(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			TournamentOverride: 0, // Default state
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
	server.erupeConfig = &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			TournamentOverride: 1,
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
	server.erupeConfig = &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			TournamentOverride: 2,
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
	server.erupeConfig = &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			TournamentOverride: 3,
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

func TestHandleMsgMhfEnumerateRanking_DefaultBranch(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			TournamentOverride: 0,
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRanking{
		AckHandle: 99999,
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

func TestHandleMsgMhfEnumerateRanking_NegativeState(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &cfg.Config{
		DebugOptions: cfg.DebugOptions{
			TournamentOverride: -1,
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRanking{
		AckHandle: 99999,
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

func TestGenerateFestaTimestamps_Debug(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name  string
		start uint32
	}{
		{"Debug_Start1", 1},
		{"Debug_Start2", 2},
		{"Debug_Start3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamps := generateFestaTimestamps(session, tt.start, true)
			if len(timestamps) != 5 {
				t.Errorf("Expected 5 timestamps, got %d", len(timestamps))
			}
			for i, ts := range timestamps {
				if ts == 0 {
					t.Errorf("Timestamp %d should not be zero", i)
				}
			}
		})
	}
}

func TestGenerateFestaTimestamps_NonDebug_FutureStart(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Use a far-future start time so it does not trigger cleanup
	futureStart := uint32(TimeAdjusted().Unix() + 5000000)
	timestamps := generateFestaTimestamps(session, futureStart, false)

	if len(timestamps) != 5 {
		t.Errorf("Expected 5 timestamps, got %d", len(timestamps))
	}
	if timestamps[0] != futureStart {
		t.Errorf("First timestamp = %d, want %d", timestamps[0], futureStart)
	}
	// Verify intervals
	if timestamps[1] != timestamps[0]+604800 {
		t.Errorf("Second timestamp should be start+604800, got %d", timestamps[1])
	}
	if timestamps[2] != timestamps[1]+604800 {
		t.Errorf("Third timestamp should be second+604800, got %d", timestamps[2])
	}
	if timestamps[3] != timestamps[2]+9000 {
		t.Errorf("Fourth timestamp should be third+9000, got %d", timestamps[3])
	}
	if timestamps[4] != timestamps[3]+1240200 {
		t.Errorf("Fifth timestamp should be fourth+1240200, got %d", timestamps[4])
	}
}

func TestFestaTrialStruct(t *testing.T) {
	trial := FestaTrial{
		ID:        100,
		Objective: 2,
		GoalID:    500,
		TimesReq:  10,
		Locale:    1,
		Reward:    50,
	}
	if trial.ID != 100 {
		t.Errorf("ID = %d, want 100", trial.ID)
	}
	if trial.Objective != 2 {
		t.Errorf("Objective = %d, want 2", trial.Objective)
	}
	if trial.GoalID != 500 {
		t.Errorf("GoalID = %d, want 500", trial.GoalID)
	}
	if trial.TimesReq != 10 {
		t.Errorf("TimesReq = %d, want 10", trial.TimesReq)
	}
}

func TestPrizeStruct(t *testing.T) {
	prize := Prize{
		ID:       1,
		Tier:     2,
		SoulsReq: 100,
		ItemID:   0x1234,
		NumItem:  5,
		Claimed:  1,
	}
	if prize.ID != 1 {
		t.Errorf("ID = %d, want 1", prize.ID)
	}
	if prize.Tier != 2 {
		t.Errorf("Tier = %d, want 2", prize.Tier)
	}
	if prize.SoulsReq != 100 {
		t.Errorf("SoulsReq = %d, want 100", prize.SoulsReq)
	}
	if prize.Claimed != 1 {
		t.Errorf("Claimed = %d, want 1", prize.Claimed)
	}
}

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

func TestHandleMsgMhfInfoFesta_OverrideZero(t *testing.T) {
	srv := createMockServer()
	srv.festaRepo = &mockFestaRepo{}
	srv.erupeConfig.DebugOptions.FestaOverride = 0
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoFesta{AckHandle: 1}
	handleMsgMhfInfoFesta(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfInfoFesta_WithActiveEvent(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.DebugOptions.FestaOverride = 1
	srv.erupeConfig.RealClientMode = cfg.ZZ
	srv.erupeConfig.GameplayOptions.MaximumFP = 50000
	srv.festaRepo = &mockFestaRepo{
		events: []FestaEvent{{ID: 1, StartTime: uint32(time.Now().Add(-24 * time.Hour).Unix())}},
		trials: []FestaTrial{
			{ID: 1, Objective: 1, GoalID: 100, TimesReq: 5, Locale: 0, Reward: 10, Monopoly: "blue"},
		},
	}
	ensureFestaService(srv)
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoFesta{AckHandle: 1}
	handleMsgMhfInfoFesta(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfInfoFesta_FutureTimestamp(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.DebugOptions.FestaOverride = -1
	srv.festaRepo = &mockFestaRepo{
		events: []FestaEvent{{ID: 1, StartTime: uint32(time.Now().Add(72 * time.Hour).Unix())}},
	}
	ensureFestaService(srv)
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoFesta{AckHandle: 1}
	handleMsgMhfInfoFesta(s, pkt)
	<-s.sendPackets
}
