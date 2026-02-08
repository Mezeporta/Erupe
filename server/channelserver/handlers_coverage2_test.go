package channelserver

import (
	"testing"

	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

// Tests for guild handlers that do not require database access.

func TestHandleMsgMhfEntryRookieGuild(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEntryRookieGuild{
		AckHandle: 12345,
		Unk:       42,
	}

	handleMsgMhfEntryRookieGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGenerateUdGuildMap(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGenerateUdGuildMap{
		AckHandle: 12345,
	}

	handleMsgMhfGenerateUdGuildMap(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfCheckMonthlyItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCheckMonthlyItem{
		AckHandle: 12345,
		Type:      0,
	}

	handleMsgMhfCheckMonthlyItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireMonthlyItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireMonthlyItem{
		AckHandle: 12345,
	}

	handleMsgMhfAcquireMonthlyItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateInvGuild(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateInvGuild{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateInvGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfOperationInvGuild(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfOperationInvGuild{
		AckHandle: 12345,
		Operation: 1,
	}

	handleMsgMhfOperationInvGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Tests for mercenary handlers that do not require database access.

func TestHandleMsgMhfMercenaryHuntdata_Unk0Is1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle: 12345,
		Unk0:      1,
	}

	handleMsgMhfMercenaryHuntdata(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfMercenaryHuntdata_Unk0Is0(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle: 12345,
		Unk0:      0,
	}

	handleMsgMhfMercenaryHuntdata(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfMercenaryHuntdata_Unk0Is2(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle: 12345,
		Unk0:      2,
	}

	handleMsgMhfMercenaryHuntdata(session, pkt)

	// Unk0=2 takes the else branch (same as 0)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Tests for festa/ranking handlers.

func TestHandleMsgMhfEnumerateRanking_DefaultBranch(t *testing.T) {
	server := createMockServer()
	server.erupeConfig = &config.Config{
		DevMode: false,
		DevModeOptions: config.DevModeOptions{
			TournamentEvent: 0,
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
	server.erupeConfig = &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			TournamentEvent: -1,
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

// Tests for rengoku handlers.

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

// Tests for empty handlers that are not covered in other test files.

func TestEmptyHandlers_Coverage2(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name    string
		handler func(s *Session, p mhfpacket.MHFPacket)
	}{
		{"handleMsgSysCastedBinary", handleMsgSysCastedBinary},
		{"handleMsgMhfResetTitle", handleMsgMhfResetTitle},
		{"handleMsgMhfUpdateForceGuildRank", handleMsgMhfUpdateForceGuildRank},
		{"handleMsgMhfUpdateGuild", handleMsgMhfUpdateGuild},
		{"handleMsgMhfUpdateGuildcard", handleMsgMhfUpdateGuildcard},
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

// Tests for handlers.go - handlers that produce responses without DB access.

func TestHandleMsgSysTerminalLog_MultipleEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 12345,
		LogID:     200,
		Entries: []*mhfpacket.TerminalLogEntry{
			{Type1: 10, Type2: 20, Data: []int16{100, 200, 300}},
			{Type1: 11, Type2: 21, Data: []int16{400, 500}},
			{Type1: 12, Type2: 22, Data: []int16{600}},
		},
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

func TestHandleMsgSysTerminalLog_ZeroLogID(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 12345,
		LogID:     0,
		Entries:   []*mhfpacket.TerminalLogEntry{},
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

func TestHandleMsgSysPing_DifferentAckHandle(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysPing{
		AckHandle: 0xFFFFFFFF,
	}

	handleMsgSysPing(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysTime_GetRemoteTimeFalse(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
	}

	handleMsgSysTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysIssueLogkey_LogKeyGenerated(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysIssueLogkey{
		AckHandle: 77777,
	}

	handleMsgSysIssueLogkey(session, pkt)

	// Verify that the logKey was set on the session
	session.Lock()
	keyLen := len(session.logKey)
	session.Unlock()

	if keyLen != 16 {
		t.Errorf("logKey length = %d, want 16", keyLen)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysIssueLogkey_Uniqueness(t *testing.T) {
	server := createMockServer()

	// Generate two logkeys and verify they differ
	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)

	pkt1 := &mhfpacket.MsgSysIssueLogkey{AckHandle: 1}
	pkt2 := &mhfpacket.MsgSysIssueLogkey{AckHandle: 2}

	handleMsgSysIssueLogkey(session1, pkt1)
	handleMsgSysIssueLogkey(session2, pkt2)

	// Drain send packets
	<-session1.sendPackets
	<-session2.sendPackets

	session1.Lock()
	key1 := make([]byte, len(session1.logKey))
	copy(key1, session1.logKey)
	session1.Unlock()

	session2.Lock()
	key2 := make([]byte, len(session2.logKey))
	copy(key2, session2.logKey)
	session2.Unlock()

	if len(key1) != 16 || len(key2) != 16 {
		t.Fatalf("logKeys should be 16 bytes each, got %d and %d", len(key1), len(key2))
	}

	same := true
	for i := range key1 {
		if key1[i] != key2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("Two generated logkeys should differ (extremely unlikely to be the same)")
	}
}

// Tests for event handlers.

func TestHandleMsgMhfReleaseEvent_ErrorCode(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReleaseEvent{
		AckHandle: 88888,
	}

	handleMsgMhfReleaseEvent(session, pkt)

	// This handler manually sends a response with error code 0x41
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateEvent_Stub(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateEvent{
		AckHandle: 77777,
	}

	handleMsgMhfEnumerateEvent(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Tests for achievement handler.

func TestHandleMsgMhfSetCaAchievementHist_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetCaAchievementHist{
		AckHandle: 44444,
	}

	handleMsgMhfSetCaAchievementHist(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test concurrent handler invocations to catch potential data races.

func TestHandlersConcurrentInvocations(t *testing.T) {
	server := createMockServer()

	done := make(chan struct{})
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		go func(id uint32) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("goroutine %d panicked: %v", id, r)
				}
				done <- struct{}{}
			}()

			session := createMockSession(id, server)

			// Run several handlers concurrently
			handleMsgSysPing(session, &mhfpacket.MsgSysPing{AckHandle: id})
			<-session.sendPackets

			handleMsgSysTime(session, &mhfpacket.MsgSysTime{GetRemoteTime: true})
			<-session.sendPackets

			handleMsgSysIssueLogkey(session, &mhfpacket.MsgSysIssueLogkey{AckHandle: id})
			<-session.sendPackets

			handleMsgMhfMercenaryHuntdata(session, &mhfpacket.MsgMhfMercenaryHuntdata{AckHandle: id, Unk0: 1})
			<-session.sendPackets

			handleMsgMhfEnumerateMercenaryLog(session, &mhfpacket.MsgMhfEnumerateMercenaryLog{AckHandle: id})
			<-session.sendPackets
		}(uint32(i + 100))
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Test festa handler with various config states.

func TestHandleMsgMhfVoteFesta_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfVoteFesta{
		AckHandle: 33333,
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

// Test record log handler with stage setup.

func TestHandleMsgSysRecordLog_RemovesReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	stage := NewStage("test_stage_record")
	session.stage = stage
	stage.reservedClientSlots[session.charID] = true

	pkt := &mhfpacket.MsgSysRecordLog{
		AckHandle: 55555,
	}

	handleMsgSysRecordLog(session, pkt)

	if _, exists := stage.reservedClientSlots[session.charID]; exists {
		t.Error("charID should be removed from reserved slots after record log")
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysRecordLog_NoExistingReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	stage := NewStage("test_stage_no_reservation")
	session.stage = stage
	// No reservation exists for this charID

	pkt := &mhfpacket.MsgSysRecordLog{
		AckHandle: 55556,
	}

	// Should not panic even if charID is not in reservedClientSlots
	handleMsgSysRecordLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test unlock global sema handler.

func TestHandleMsgSysUnlockGlobalSema_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysUnlockGlobalSema{
		AckHandle: 66666,
	}

	handleMsgSysUnlockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test handlers from handlers_event.go with edge cases.

func TestHandleMsgMhfSetRestrictionEvent_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSetRestrictionEvent{
		AckHandle: 11111,
	}

	handleMsgMhfSetRestrictionEvent(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetRestrictionEvent_Empty(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetRestrictionEvent panicked: %v", r)
		}
	}()

	handleMsgMhfGetRestrictionEvent(session, nil)
}

// Test handlers from handlers_mercenary.go - legend dispatch (no DB).

func TestHandleMsgMhfLoadLegendDispatch_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadLegendDispatch{
		AckHandle: 22222,
	}

	handleMsgMhfLoadLegendDispatch(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test multiple handler invocations on the same session to verify session state is not corrupted.

func TestMultipleHandlersOnSameSession(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Call multiple handlers in sequence
	handleMsgSysPing(session, &mhfpacket.MsgSysPing{AckHandle: 1})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from Ping handler")
	}

	handleMsgSysTime(session, &mhfpacket.MsgSysTime{GetRemoteTime: true})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from Time handler")
	}

	handleMsgMhfRegisterEvent(session, &mhfpacket.MsgMhfRegisterEvent{AckHandle: 2, Unk2: 5, Unk4: 10})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from RegisterEvent handler")
	}

	handleMsgMhfReleaseEvent(session, &mhfpacket.MsgMhfReleaseEvent{AckHandle: 3})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from ReleaseEvent handler")
	}

	handleMsgMhfEnumerateEvent(session, &mhfpacket.MsgMhfEnumerateEvent{AckHandle: 4})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from EnumerateEvent handler")
	}

	handleMsgMhfSetCaAchievementHist(session, &mhfpacket.MsgMhfSetCaAchievementHist{AckHandle: 5})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from SetCaAchievementHist handler")
	}

	handleMsgMhfGetRengokuRankingRank(session, &mhfpacket.MsgMhfGetRengokuRankingRank{AckHandle: 6})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from GetRengokuRankingRank handler")
	}
}

// Test festa timestamp generation.

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

// Test trial struct from handlers_festa.go.

func TestTrialStruct(t *testing.T) {
	trial := Trial{
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

// Test prize struct from handlers_festa.go.

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

// Test CatDefinition struct from handlers_mercenary.go.

func TestCatDefinitionStruct(t *testing.T) {
	cat := CatDefinition{
		CatID:       42,
		CatName:     []byte("TestCat"),
		CurrentTask: 4,
		Personality: 2,
		Class:       1,
		Experience:  1500,
		WeaponType:  6,
		WeaponID:    100,
	}

	if cat.CatID != 42 {
		t.Errorf("CatID = %d, want 42", cat.CatID)
	}
	if cat.CurrentTask != 4 {
		t.Errorf("CurrentTask = %d, want 4", cat.CurrentTask)
	}
	if cat.Experience != 1500 {
		t.Errorf("Experience = %d, want 1500", cat.Experience)
	}
	if cat.WeaponType != 6 {
		t.Errorf("WeaponType = %d, want 6", cat.WeaponType)
	}
	if cat.WeaponID != 100 {
		t.Errorf("WeaponID = %d, want 100", cat.WeaponID)
	}
}

// Test RengokuScore struct default values.

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
