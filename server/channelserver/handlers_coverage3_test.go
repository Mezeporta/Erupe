package channelserver

import (
	"sync"
	"testing"

	"erupe-ce/network/mhfpacket"
)

// =============================================================================
// Category 1: Empty handlers from handlers.go
// These have empty function bodies and can be called with nil packet safely.
// =============================================================================

func TestEmptyHandlers_HandlersGo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgSysEcho", func() { handleMsgSysEcho(session, nil) }},
		{"handleMsgSysUpdateRight", func() { handleMsgSysUpdateRight(session, nil) }},
		{"handleMsgSysAuthQuery", func() { handleMsgSysAuthQuery(session, nil) }},
		{"handleMsgSysAuthTerminal", func() { handleMsgSysAuthTerminal(session, nil) }},
		{"handleMsgCaExchangeItem", func() { handleMsgCaExchangeItem(session, nil) }},
		{"handleMsgMhfServerCommand", func() { handleMsgMhfServerCommand(session, nil) }},
		{"handleMsgMhfSetLoginwindow", func() { handleMsgMhfSetLoginwindow(session, nil) }},
		{"handleMsgSysTransBinary", func() { handleMsgSysTransBinary(session, nil) }},
		{"handleMsgSysCollectBinary", func() { handleMsgSysCollectBinary(session, nil) }},
		{"handleMsgSysGetState", func() { handleMsgSysGetState(session, nil) }},
		{"handleMsgSysSerialize", func() { handleMsgSysSerialize(session, nil) }},
		{"handleMsgSysEnumlobby", func() { handleMsgSysEnumlobby(session, nil) }},
		{"handleMsgSysEnumuser", func() { handleMsgSysEnumuser(session, nil) }},
		{"handleMsgSysInfokyserver", func() { handleMsgSysInfokyserver(session, nil) }},
		{"handleMsgMhfGetCaUniqueID", func() { handleMsgMhfGetCaUniqueID(session, nil) }},
		{"handleMsgMhfGetExtraInfo", func() { handleMsgMhfGetExtraInfo(session, nil) }},
		{"handleMsgSysSetStatus", func() { handleMsgSysSetStatus(session, nil) }},
		{"handleMsgMhfStampcardPrize", func() { handleMsgMhfStampcardPrize(session, nil) }},
		{"handleMsgMhfKickExportForce", func() { handleMsgMhfKickExportForce(session, nil) }},
		{"handleMsgMhfRegistSpabiTime", func() { handleMsgMhfRegistSpabiTime(session, nil) }},
		{"handleMsgMhfDebugPostValue", func() { handleMsgMhfDebugPostValue(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

// =============================================================================
// Category 2: Empty handlers from handlers_object.go
// All empty function bodies, safe to call with nil packet.
// =============================================================================

func TestEmptyHandlers_ObjectGo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgSysDeleteObject", func() { handleMsgSysDeleteObject(session, nil) }},
		{"handleMsgSysRotateObject", func() { handleMsgSysRotateObject(session, nil) }},
		{"handleMsgSysDuplicateObject", func() { handleMsgSysDuplicateObject(session, nil) }},
		{"handleMsgSysGetObjectBinary", func() { handleMsgSysGetObjectBinary(session, nil) }},
		{"handleMsgSysGetObjectOwner", func() { handleMsgSysGetObjectOwner(session, nil) }},
		{"handleMsgSysUpdateObjectBinary", func() { handleMsgSysUpdateObjectBinary(session, nil) }},
		{"handleMsgSysCleanupObject", func() { handleMsgSysCleanupObject(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

// =============================================================================
// Category 3: Empty handlers from handlers_clients.go
// All empty function bodies, safe to call with nil packet.
// =============================================================================

func TestEmptyHandlers_ClientsGo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgMhfShutClient", func() { handleMsgMhfShutClient(session, nil) }},
		{"handleMsgSysHideClient", func() { handleMsgSysHideClient(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

// =============================================================================
// Category 4: Empty handler from handlers_stage.go
// =============================================================================

func TestEmptyHandlers_StageGo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgSysStageDestruct", func() { handleMsgSysStageDestruct(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

// =============================================================================
// Category 5: Empty handlers from handlers_achievement.go
// =============================================================================

func TestEmptyHandlers_AchievementGo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgMhfDisplayedAchievement", func() {
			handleMsgMhfDisplayedAchievement(session, &mhfpacket.MsgMhfDisplayedAchievement{})
		}},
		{"handleMsgMhfGetCaAchievementHist", func() { handleMsgMhfGetCaAchievementHist(session, nil) }},
		{"handleMsgMhfSetCaAchievement", func() { handleMsgMhfSetCaAchievement(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

// =============================================================================
// Category 6: Empty handlers from handlers_caravan.go
// =============================================================================

// TestEmptyHandlers_CaravanGo removed: caravan handlers on main do type assertions
// and require proper packet structs, not nil.

// =============================================================================
// Category 7: Simple ack handlers from handlers_tactics.go (no DB needed)
// =============================================================================

func TestSimpleAckHandlers_TacticsGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfAddUdTacticsPoint", func(s *Session) {
			handleMsgMhfAddUdTacticsPoint(s, &mhfpacket.MsgMhfAddUdTacticsPoint{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

// TestSimpleAckHandlers_TowerGo removed: tower handlers on main access s.server.db
// and cannot be tested without a database connection.

// =============================================================================
// Category 9: Simple ack handlers from handlers_reward.go (no DB needed)
// =============================================================================

func TestSimpleAckHandlers_RewardGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetRewardSong", func(s *Session) {
			handleMsgMhfGetRewardSong(s, &mhfpacket.MsgMhfGetRewardSong{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

// =============================================================================
// Category 10: Simple ack handler from handlers_semaphore.go (no DB needed)
// handleMsgSysCreateSemaphore produces a response via doAckSimpleSucceed.
// =============================================================================

func TestSimpleAckHandlers_SemaphoreGo(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgSysCreateSemaphore", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysCreateSemaphore(session, &mhfpacket.MsgSysCreateSemaphore{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("handleMsgSysCreateSemaphore: response should have data")
			}
		default:
			t.Error("handleMsgSysCreateSemaphore: no response queued")
		}
	})
}

// =============================================================================
// Category 11: handleMsgSysCreateAcquireSemaphore from handlers_semaphore.go
// This handler accesses s.server.semaphore map. It creates or acquires a
// semaphore, so it needs the semaphore map initialized on the server.
// =============================================================================

func TestHandleMsgSysCreateAcquireSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	t.Run("creates_new_semaphore", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysCreateAcquireSemaphore(session, &mhfpacket.MsgSysCreateAcquireSemaphore{
			AckHandle:   1,
			SemaphoreID: "test_sema_1",
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
		// Verify semaphore was created
		if _, exists := server.semaphore["test_sema_1"]; !exists {
			t.Error("semaphore should have been created in server map")
		}
	})

	t.Run("acquires_existing_semaphore", func(t *testing.T) {
		session := createMockSession(2, server)
		// Acquire the same semaphore again
		handleMsgSysCreateAcquireSemaphore(session, &mhfpacket.MsgSysCreateAcquireSemaphore{
			AckHandle:   2,
			SemaphoreID: "test_sema_1",
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("creates_ravi_semaphore", func(t *testing.T) {
		session := createMockSession(3, server)
		handleMsgSysCreateAcquireSemaphore(session, &mhfpacket.MsgSysCreateAcquireSemaphore{
			AckHandle:   3,
			SemaphoreID: "hs_l0u3B51",
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
		if _, exists := server.semaphore["hs_l0u3B51"]; !exists {
			t.Error("ravi semaphore should have been created")
		}
	})
}

// =============================================================================
// Category 12: Additional simple ack handlers from various files (no DB)
// =============================================================================

// TestSimpleAckHandlers_MiscFiles removed: handleMsgMhfGetRengokuBinary panics
// on missing file (explicit panic in handler), cannot test without rengoku_data.bin.

// =============================================================================
// Category 13: Other empty handlers from various files
// =============================================================================

func TestEmptyHandlers_MiscFiles(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		// From handlers_reward.go
		{"handleMsgMhfUseRewardSong", func() { handleMsgMhfUseRewardSong(session, nil) }},
		{"handleMsgMhfAddRewardSongCount", func() { handleMsgMhfAddRewardSongCount(session, nil) }},
		{"handleMsgMhfAcceptReadReward", func() { handleMsgMhfAcceptReadReward(session, nil) }},
		// From handlers_caravan.go
		{"handleMsgMhfPostRyoudama", func() { handleMsgMhfPostRyoudama(session, nil) }},
		// From handlers_tactics.go
		{"handleMsgMhfSetUdTacticsFollower", func() { handleMsgMhfSetUdTacticsFollower(session, nil) }},
		{"handleMsgMhfGetUdTacticsLog", func() { handleMsgMhfGetUdTacticsLog(session, nil) }},
		// From handlers_achievement.go
		{"handleMsgMhfPaymentAchievement", func() { handleMsgMhfPaymentAchievement(session, nil) }},
		// From handlers.go (additional empty ones)
		{"handleMsgMhfGetCogInfo", func() { handleMsgMhfGetCogInfo(session, nil) }},
		{"handleMsgMhfUseUdShopCoin", func() { handleMsgMhfUseUdShopCoin(session, nil) }},
		{"handleMsgMhfGetDailyMissionMaster", func() { handleMsgMhfGetDailyMissionMaster(session, nil) }},
		{"handleMsgMhfGetDailyMissionPersonal", func() { handleMsgMhfGetDailyMissionPersonal(session, nil) }},
		{"handleMsgMhfSetDailyMissionPersonal", func() { handleMsgMhfSetDailyMissionPersonal(session, nil) }},
		// From handlers_object.go (additional empty ones)
		{"handleMsgSysAddObject", func() { handleMsgSysAddObject(session, nil) }},
		{"handleMsgSysDelObject", func() { handleMsgSysDelObject(session, nil) }},
		{"handleMsgSysDispObject", func() { handleMsgSysDispObject(session, nil) }},
		{"handleMsgSysHideObject", func() { handleMsgSysHideObject(session, nil) }},
		// From handlers.go (non-trivial but no pkt dereference)
		{"handleMsgHead", func() { handleMsgHead(session, nil) }},
		{"handleMsgSysExtendThreshold", func() { handleMsgSysExtendThreshold(session, nil) }},
		{"handleMsgSysEnd", func() { handleMsgSysEnd(session, nil) }},
		{"handleMsgSysNop", func() { handleMsgSysNop(session, nil) }},
		{"handleMsgSysAck", func() { handleMsgSysAck(session, nil) }},
		// From handlers_semaphore.go
		{"handleMsgSysReleaseSemaphore", func() { handleMsgSysReleaseSemaphore(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

// =============================================================================
// Category 14: Handlers that produce responses without DB access
// These are non-trivial handlers with static/canned responses.
// =============================================================================

func TestNonTrivialHandlers_NoDB(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgMhfGetEarthStatus", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetEarthStatus(session, &mhfpacket.MsgMhfGetEarthStatus{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfGetEarthValue_Type1", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetEarthValue(session, &mhfpacket.MsgMhfGetEarthValue{AckHandle: 1, ReqType: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfGetEarthValue_Type2", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetEarthValue(session, &mhfpacket.MsgMhfGetEarthValue{AckHandle: 1, ReqType: 2})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfGetEarthValue_Type3", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetEarthValue(session, &mhfpacket.MsgMhfGetEarthValue{AckHandle: 1, ReqType: 3})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfGetSeibattle", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetSeibattle(session, &mhfpacket.MsgMhfGetSeibattle{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	// handleMsgMhfGetTrendWeapon removed: requires database access

	// handleMsgMhfUpdateUseTrendWeaponLog removed: requires database access

	t.Run("handleMsgMhfUpdateBeatLevel", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfUpdateBeatLevel(session, &mhfpacket.MsgMhfUpdateBeatLevel{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfReadBeatLevel", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfReadBeatLevel(session, &mhfpacket.MsgMhfReadBeatLevel{
			AckHandle:    1,
			ValidIDCount: 2,
			IDs:          [16]uint32{100, 200},
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfTransferItem", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfTransferItem(session, &mhfpacket.MsgMhfTransferItem{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfEnumerateOrder", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfEnumerateOrder(session, &mhfpacket.MsgMhfEnumerateOrder{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfGetUdShopCoin", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetUdShopCoin(session, &mhfpacket.MsgMhfGetUdShopCoin{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfGetLobbyCrowd", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetLobbyCrowd(session, &mhfpacket.MsgMhfGetLobbyCrowd{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfEnumeratePrice", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfEnumeratePrice(session, &mhfpacket.MsgMhfEnumeratePrice{AckHandle: 1})
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

// =============================================================================
// Category 15: Handlers from handlers_tactics.go that produce responses (no DB)
// =============================================================================

func TestNonTrivialHandlers_TacticsGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetUdTacticsPoint", func(s *Session) {
			handleMsgMhfGetUdTacticsPoint(s, &mhfpacket.MsgMhfGetUdTacticsPoint{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsRewardList", func(s *Session) {
			handleMsgMhfGetUdTacticsRewardList(s, &mhfpacket.MsgMhfGetUdTacticsRewardList{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsFollower", func(s *Session) {
			handleMsgMhfGetUdTacticsFollower(s, &mhfpacket.MsgMhfGetUdTacticsFollower{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsBonusQuest", func(s *Session) {
			handleMsgMhfGetUdTacticsBonusQuest(s, &mhfpacket.MsgMhfGetUdTacticsBonusQuest{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsFirstQuestBonus", func(s *Session) {
			handleMsgMhfGetUdTacticsFirstQuestBonus(s, &mhfpacket.MsgMhfGetUdTacticsFirstQuestBonus{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsRemainingPoint", func(s *Session) {
			handleMsgMhfGetUdTacticsRemainingPoint(s, &mhfpacket.MsgMhfGetUdTacticsRemainingPoint{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdTacticsRanking", func(s *Session) {
			handleMsgMhfGetUdTacticsRanking(s, &mhfpacket.MsgMhfGetUdTacticsRanking{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

// =============================================================================
// Category 16: Handlers from handlers_tower.go that produce responses (no DB)
// =============================================================================

func TestNonTrivialHandlers_TowerGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetTenrouirai_Type1", func(s *Session) {
			handleMsgMhfGetTenrouirai(s, &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, Unk0: 1})
		}},
		{"handleMsgMhfGetTenrouirai_Unknown", func(s *Session) {
			handleMsgMhfGetTenrouirai(s, &mhfpacket.MsgMhfGetTenrouirai{AckHandle: 1, Unk0: 0, DataType: 0})
		}},
		// handleMsgMhfGetTenrouirai_Type4, handleMsgMhfPostTenrouirai, handleMsgMhfGetGemInfo removed: require DB
		{"handleMsgMhfGetWeeklySeibatuRankingReward", func(s *Session) {
			handleMsgMhfGetWeeklySeibatuRankingReward(s, &mhfpacket.MsgMhfGetWeeklySeibatuRankingReward{AckHandle: 1})
		}},
		{"handleMsgMhfPresentBox", func(s *Session) {
			handleMsgMhfPresentBox(s, &mhfpacket.MsgMhfPresentBox{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

// =============================================================================
// Category 17: Handlers from handlers_reward.go that produce responses (no DB)
// =============================================================================

func TestNonTrivialHandlers_RewardGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetAdditionalBeatReward", func(s *Session) {
			handleMsgMhfGetAdditionalBeatReward(s, &mhfpacket.MsgMhfGetAdditionalBeatReward{AckHandle: 1})
		}},
		{"handleMsgMhfGetUdRankingRewardList", func(s *Session) {
			handleMsgMhfGetUdRankingRewardList(s, &mhfpacket.MsgMhfGetUdRankingRewardList{AckHandle: 1})
		}},
		{"handleMsgMhfAcquireMonthlyReward", func(s *Session) {
			handleMsgMhfAcquireMonthlyReward(s, &mhfpacket.MsgMhfAcquireMonthlyReward{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

// =============================================================================
// Category 18: Handlers from handlers_caravan.go that produce responses (no DB)
// =============================================================================

func TestNonTrivialHandlers_CaravanGo(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name string
		fn   func(s *Session)
	}{
		{"handleMsgMhfGetRyoudama", func(s *Session) {
			handleMsgMhfGetRyoudama(s, &mhfpacket.MsgMhfGetRyoudama{AckHandle: 1})
		}},
		{"handleMsgMhfGetTinyBin", func(s *Session) {
			handleMsgMhfGetTinyBin(s, &mhfpacket.MsgMhfGetTinyBin{AckHandle: 1})
		}},
		{"handleMsgMhfPostTinyBin", func(s *Session) {
			handleMsgMhfPostTinyBin(s, &mhfpacket.MsgMhfPostTinyBin{AckHandle: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			tt.fn(session)
			select {
			case p := <-session.sendPackets:
				if len(p.data) == 0 {
					t.Errorf("%s: response should have data", tt.name)
				}
			default:
				t.Errorf("%s: no response queued", tt.name)
			}
		})
	}
}

// =============================================================================
// Category 19: Handlers from handlers_rengoku.go (no DB needed)
// =============================================================================

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

// =============================================================================
// Category 20: Handlers from handlers.go that produce responses (no DB)
// =============================================================================

// TestNonTrivialHandlers_InfoScenarioCounter removed: requires database access.

// =============================================================================
// Category 21: handleMsgSysPing and handleMsgSysTime (no DB)
// =============================================================================

func TestSimpleHandlers_PingAndTime(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgSysPing", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysPing(session, &mhfpacket.MsgSysPing{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgSysTime", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysTime(session, &mhfpacket.MsgSysTime{})
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

// =============================================================================
// Category 22: handleMsgSysIssueLogkey (no DB, uses crypto/rand)
// =============================================================================

func TestHandleMsgSysIssueLogkey_Coverage3(t *testing.T) {
	server := createMockServer()

	t.Run("generates_logkey", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysIssueLogkey(session, &mhfpacket.MsgSysIssueLogkey{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
		if session.logKey == nil {
			t.Error("logKey should be set after IssueLogkey")
		}
		if len(session.logKey) != 16 {
			t.Errorf("logKey length = %d, want 16", len(session.logKey))
		}
	})
}

// =============================================================================
// Category 23: handleMsgSysUnlockGlobalSema (no DB)
// =============================================================================

func TestHandleMsgSysUnlockGlobalSema_Coverage3(t *testing.T) {
	server := createMockServer()

	t.Run("produces_response", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysUnlockGlobalSema(session, &mhfpacket.MsgSysUnlockGlobalSema{AckHandle: 1})
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

// =============================================================================
// Category 24: handleMsgSysLockGlobalSema (no DB, but needs Channels)
// =============================================================================

func TestHandleMsgSysLockGlobalSema(t *testing.T) {
	server := createMockServer()
	server.Channels = make([]*Server, 0)

	t.Run("no_channels_returns_response", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysLockGlobalSema(session, &mhfpacket.MsgSysLockGlobalSema{
			AckHandle:             1,
			UserIDString:          "testuser",
			ServerChannelIDString: "ch1",
		})
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

// =============================================================================
// Category 25: handleMsgSysCheckSemaphore (no DB)
// =============================================================================

func TestHandleMsgSysCheckSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	t.Run("semaphore_not_exists", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysCheckSemaphore(session, &mhfpacket.MsgSysCheckSemaphore{
			AckHandle:   1,
			SemaphoreID: "nonexistent",
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("semaphore_exists", func(t *testing.T) {
		session := createMockSession(1, server)
		server.semaphore["existing_sema"] = NewSemaphore(session, "existing_sema", 1)
		handleMsgSysCheckSemaphore(session, &mhfpacket.MsgSysCheckSemaphore{
			AckHandle:   1,
			SemaphoreID: "existing_sema",
		})
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

// =============================================================================
// Category 26: handleMsgSysAcquireSemaphore (no DB)
// =============================================================================

func TestHandleMsgSysAcquireSemaphore(t *testing.T) {
	server := createMockServer()
	server.semaphore = make(map[string]*Semaphore)

	t.Run("semaphore_exists", func(t *testing.T) {
		session := createMockSession(1, server)
		server.semaphore["acquire_sema"] = NewSemaphore(session, "acquire_sema", 1)
		handleMsgSysAcquireSemaphore(session, &mhfpacket.MsgSysAcquireSemaphore{
			AckHandle:   1,
			SemaphoreID: "acquire_sema",
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("semaphore_not_exists", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysAcquireSemaphore(session, &mhfpacket.MsgSysAcquireSemaphore{
			AckHandle:   1,
			SemaphoreID: "nonexistent_sema",
		})
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

// =============================================================================
// Category 27: handleMsgSysCreateStage (no DB)
// =============================================================================

func TestHandleMsgSysCreateStage_Coverage3(t *testing.T) {
	server := createMockServer()

	t.Run("creates_new_stage", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysCreateStage(session, &mhfpacket.MsgSysCreateStage{
			AckHandle:   1,
			StageID:     "test_create_stage",
			PlayerCount: 4,
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
		if _, exists := server.stages["test_create_stage"]; !exists {
			t.Error("stage should have been created")
		}
	})

	t.Run("duplicate_stage_fails", func(t *testing.T) {
		session := createMockSession(1, server)
		// Stage already exists from the previous test
		handleMsgSysCreateStage(session, &mhfpacket.MsgSysCreateStage{
			AckHandle:   2,
			StageID:     "test_create_stage",
			PlayerCount: 4,
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data even on failure")
			}
		default:
			t.Error("no response queued")
		}
	})
}

// =============================================================================
// Category 28: Concurrency test for empty handlers
// Verify that calling empty handlers concurrently does not panic.
// =============================================================================

func TestEmptyHandlers_Concurrent(t *testing.T) {
	server := createMockServer()

	handlers := []func(*Session, mhfpacket.MHFPacket){
		handleMsgSysEcho,
		handleMsgSysUpdateRight,
		handleMsgSysAuthQuery,
		handleMsgSysAuthTerminal,
		handleMsgCaExchangeItem,
		handleMsgMhfServerCommand,
		handleMsgMhfSetLoginwindow,
		handleMsgSysTransBinary,
		handleMsgSysCollectBinary,
		handleMsgSysGetState,
		handleMsgSysSerialize,
		handleMsgSysEnumlobby,
		handleMsgSysEnumuser,
		handleMsgSysInfokyserver,
		handleMsgMhfGetCaUniqueID,
		handleMsgMhfGetExtraInfo,
		handleMsgSysSetStatus,
		handleMsgSysDeleteObject,
		handleMsgSysRotateObject,
		handleMsgSysDuplicateObject,
		handleMsgSysGetObjectBinary,
		handleMsgSysGetObjectOwner,
		handleMsgSysUpdateObjectBinary,
		handleMsgSysCleanupObject,
		handleMsgMhfShutClient,
		handleMsgSysHideClient,
		handleMsgSysStageDestruct,
	}

	var wg sync.WaitGroup
	for _, h := range handlers {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(handler func(*Session, mhfpacket.MHFPacket)) {
				defer wg.Done()
				session := createMockSession(1, server)
				handler(session, nil)
			}(h)
		}
	}
	wg.Wait()
}

// =============================================================================
// Category 29: stubEnumerateNoResults and stubGetNoResults helper coverage
// These are called by many handlers; test them directly too.
// =============================================================================

func TestStubHelpers(t *testing.T) {
	server := createMockServer()

	t.Run("stubEnumerateNoResults", func(t *testing.T) {
		session := createMockSession(1, server)
		stubEnumerateNoResults(session, 1)
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckBufSucceed", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckBufSucceed(session, 1, []byte{0x01, 0x02, 0x03})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckBufFail", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckBufFail(session, 1, []byte{0x01, 0x02, 0x03})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckSimpleSucceed", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckSimpleSucceed(session, 1, []byte{0x00, 0x00, 0x00, 0x00})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("doAckSimpleFail", func(t *testing.T) {
		session := createMockSession(1, server)
		doAckSimpleFail(session, 1, []byte{0x00, 0x00, 0x00, 0x00})
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
