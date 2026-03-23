package channelserver

import (
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

// Test handlers with simple responses

func TestHandleMsgMhfGetEarthStatus(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthStatus{
		AckHandle: 12345,
	}

	handleMsgMhfGetEarthStatus(session, pkt)

	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_Type1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   1,
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_Type2(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   2,
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_Type3(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   3,
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_UnknownType(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   99, // Unknown type
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Should still return a response (empty values)
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevel(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevel{
		AckHandle:    12345,
		ValidIDCount: 2,
		IDs:          [16]uint32{1, 2},
	}

	handleMsgMhfReadBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevel_NoIDs(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevel{
		AckHandle:    12345,
		ValidIDCount: 0,
		IDs:          [16]uint32{},
	}

	handleMsgMhfReadBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateBeatLevel(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateBeatLevel{
		AckHandle: 12345,
	}

	handleMsgMhfUpdateBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test empty handlers don't panic

func TestHandleMsgMhfStampcardPrize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfStampcardPrize panicked: %v", r)
		}
	}()

	handleMsgMhfStampcardPrize(session, nil)
}

func TestHandleMsgMhfUnreserveSrg(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUnreserveSrg{
		AckHandle: 12345,
	}

	handleMsgMhfUnreserveSrg(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevelAllRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevelAllRanking{
		AckHandle: 12345,
	}

	handleMsgMhfReadBeatLevelAllRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevelMyRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevelMyRanking{
		AckHandle: 12345,
	}

	handleMsgMhfReadBeatLevelMyRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadLastWeekBeatRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadLastWeekBeatRanking{
		AckHandle: 12345,
	}

	handleMsgMhfReadLastWeekBeatRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetFixedSeibatuRankingTable(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetFixedSeibatuRankingTable{
		AckHandle: 12345,
	}

	handleMsgMhfGetFixedSeibatuRankingTable(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfKickExportForce(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfKickExportForce panicked: %v", r)
		}
	}()

	handleMsgMhfKickExportForce(session, nil)
}

func TestHandleMsgMhfRegistSpabiTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfRegistSpabiTime panicked: %v", r)
		}
	}()

	handleMsgMhfRegistSpabiTime(session, nil)
}

func TestHandleMsgMhfDebugPostValue(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfDebugPostValue panicked: %v", r)
		}
	}()

	handleMsgMhfDebugPostValue(session, nil)
}

func TestHandleMsgMhfGetCogInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetCogInfo{AckHandle: 1}
	handleMsgMhfGetCogInfo(session, pkt)
}

// Additional handler tests for coverage

func TestHandleMsgMhfGetNotice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetNotice{
		AckHandle: 12345,
	}

	handleMsgMhfGetNotice(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostNotice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostNotice{
		AckHandle: 12345,
	}

	handleMsgMhfPostNotice(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetRandFromTable(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRandFromTable{
		AckHandle: 12345,
		Results:   3,
	}

	handleMsgMhfGetRandFromTable(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetSenyuDailyCount(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetSenyuDailyCount{
		AckHandle: 12345,
	}

	handleMsgMhfGetSenyuDailyCount(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetSeibattle(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetSeibattle{
		AckHandle: 12345,
	}

	handleMsgMhfGetSeibattle(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostSeibattle(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostSeibattle{
		AckHandle: 12345,
	}

	handleMsgMhfPostSeibattle(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetDailyMissionMaster(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetDailyMissionMaster panicked: %v", r)
		}
	}()

	handleMsgMhfGetDailyMissionMaster(session, nil)
}

func TestHandleMsgMhfGetDailyMissionPersonal(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetDailyMissionPersonal panicked: %v", r)
		}
	}()

	handleMsgMhfGetDailyMissionPersonal(session, nil)
}

func TestHandleMsgMhfSetDailyMissionPersonal(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfSetDailyMissionPersonal panicked: %v", r)
		}
	}()

	handleMsgMhfSetDailyMissionPersonal(session, nil)
}

func TestHandleMsgMhfGetUdShopCoin(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdShopCoin{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdShopCoin(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUseUdShopCoin(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUseUdShopCoin{AckHandle: 1}
	handleMsgMhfUseUdShopCoin(session, pkt)
}

func TestHandleMsgMhfGetLobbyCrowd(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetLobbyCrowd{
		AckHandle: 12345,
	}

	handleMsgMhfGetLobbyCrowd(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// =============================================================================
// equipSkinHistSize: pure function, tests all 3 config branches
// =============================================================================

func TestEquipSkinHistSize_Default(t *testing.T) {
	got := equipSkinHistSize(cfg.ZZ)
	if got != 3200 {
		t.Errorf("equipSkinHistSize(ZZ) = %d, want 3200", got)
	}
}

func TestEquipSkinHistSize_Z2(t *testing.T) {
	got := equipSkinHistSize(cfg.Z2)
	if got != 2560 {
		t.Errorf("equipSkinHistSize(Z2) = %d, want 2560", got)
	}
}

func TestEquipSkinHistSize_Z1(t *testing.T) {
	got := equipSkinHistSize(cfg.Z1)
	if got != 1280 {
		t.Errorf("equipSkinHistSize(Z1) = %d, want 1280", got)
	}
}

func TestEquipSkinHistSize_OlderMode(t *testing.T) {
	got := equipSkinHistSize(cfg.G1)
	if got != 1280 {
		t.Errorf("equipSkinHistSize(G1) = %d, want 1280", got)
	}
}

// Distribution struct tests
func TestDistributionStruct(t *testing.T) {
	dist := Distribution{
		ID:              1,
		MinHR:           1,
		MaxHR:           999,
		MinSR:           0,
		MaxSR:           999,
		MinGR:           0,
		MaxGR:           999,
		TimesAcceptable: 1,
		TimesAccepted:   0,
		EventName:       "Test Event",
		Description:     "Test Description",
		Selection:       false,
	}

	if dist.ID != 1 {
		t.Errorf("ID = %d, want 1", dist.ID)
	}
	if dist.EventName != "Test Event" {
		t.Errorf("EventName = %s, want Test Event", dist.EventName)
	}
}

func TestDistributionItemStruct(t *testing.T) {
	item := DistributionItem{
		ItemType: 1,
		ID:       100,
		ItemID:   1234,
		Quantity: 10,
	}

	if item.ItemType != 1 {
		t.Errorf("ItemType = %d, want 1", item.ItemType)
	}
	if item.ItemID != 1234 {
		t.Errorf("ItemID = %d, want 1234", item.ItemID)
	}
}

// --- Enhanced Minidata tests (in-memory store, no DB) ---

func TestHandleMsgMhfGetEnhancedMinidata_NotFound(t *testing.T) {
	srv := createMockServer()
	srv.minidata = NewMinidataStore()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetEnhancedMinidata{AckHandle: 1, CharID: 999}
	handleMsgMhfGetEnhancedMinidata(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetEnhancedMinidata_Found(t *testing.T) {
	srv := createMockServer()
	srv.minidata = NewMinidataStore()
	srv.minidata.Set(42, []byte{0xDE, 0xAD, 0xBE, 0xEF})
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetEnhancedMinidata{AckHandle: 1, CharID: 42}
	handleMsgMhfGetEnhancedMinidata(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfSetEnhancedMinidata(t *testing.T) {
	srv := createMockServer()
	srv.minidata = NewMinidataStore()
	s := createMockSession(100, srv)

	payload := []byte{0x01, 0x02, 0x03}
	pkt := &mhfpacket.MsgMhfSetEnhancedMinidata{AckHandle: 1, RawDataPayload: payload}
	handleMsgMhfSetEnhancedMinidata(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}

	data, ok := srv.minidata.Get(100)
	if !ok {
		t.Fatal("Minidata not stored")
	}
	if len(data) != 3 || data[0] != 0x01 {
		t.Errorf("Unexpected stored data: %v", data)
	}
}

// --- Trend Weapon tests ---

func TestHandleMsgMhfGetTrendWeapon_Empty(t *testing.T) {
	srv := createMockServer()
	srv.miscRepo = &mockMiscRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetTrendWeapon{AckHandle: 1}
	handleMsgMhfGetTrendWeapon(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfGetTrendWeapon_WithWeapons(t *testing.T) {
	srv := createMockServer()
	srv.miscRepo = &mockMiscRepo{
		trendWeapons: []uint16{100, 200, 300},
	}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetTrendWeapon{AckHandle: 1}
	handleMsgMhfGetTrendWeapon(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateUseTrendWeaponLog(t *testing.T) {
	srv := createMockServer()
	srv.miscRepo = &mockMiscRepo{}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateUseTrendWeaponLog{AckHandle: 1, WeaponType: 3, WeaponID: 500}
	handleMsgMhfUpdateUseTrendWeaponLog(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}
}

// --- Etc Points tests ---

func TestHandleMsgMhfGetEtcPoints(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.etcBonusQuests = 100
	charRepo.etcDailyQuests = 50
	charRepo.etcPromoPoints = 25
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetEtcPoints{AckHandle: 1}
	handleMsgMhfGetEtcPoints(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateEtcPoint(t *testing.T) {
	tests := []struct {
		name      string
		pointType uint8
		delta     int16
		column    string
	}{
		{"bonus_quests", 0, 5, "bonus_quests"},
		{"daily_quests", 1, 3, "daily_quests"},
		{"promo_points", 2, 1, "promo_points"},
		{"invalid_type", 99, 1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := createMockServer()
			charRepo := newMockCharacterRepo()
			srv.charRepo = charRepo
			s := createMockSession(100, srv)

			pkt := &mhfpacket.MsgMhfUpdateEtcPoint{
				AckHandle: 1,
				PointType: tt.pointType,
				Delta:     tt.delta,
			}
			handleMsgMhfUpdateEtcPoint(s, pkt)

			select {
			case <-s.sendPackets:
			default:
				t.Fatal("No response packet queued")
			}

			if tt.column != "" {
				val := charRepo.ints[tt.column]
				if val != int(tt.delta) {
					t.Errorf("Expected %s=%d, got %d", tt.column, tt.delta, val)
				}
			}
		})
	}
}

func TestHandleMsgMhfUpdateEtcPoint_NegativeDelta(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["bonus_quests"] = 10
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateEtcPoint{AckHandle: 1, PointType: 0, Delta: -5}
	handleMsgMhfUpdateEtcPoint(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}

	if charRepo.ints["bonus_quests"] != 5 {
		t.Errorf("Expected bonus_quests=5, got %d", charRepo.ints["bonus_quests"])
	}
}

func TestHandleMsgMhfUpdateEtcPoint_ClampToZero(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["bonus_quests"] = 3
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateEtcPoint{AckHandle: 1, PointType: 0, Delta: -10}
	handleMsgMhfUpdateEtcPoint(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}

	if charRepo.ints["bonus_quests"] != 0 {
		t.Errorf("Expected bonus_quests=0, got %d", charRepo.ints["bonus_quests"])
	}
}

// --- Equip Skin History tests ---

func TestHandleMsgMhfGetEquipSkinHist(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetEquipSkinHist{AckHandle: 1}
	handleMsgMhfGetEquipSkinHist(s, pkt)

	select {
	case p := <-s.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Expected non-empty response")
		}
	default:
		t.Fatal("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateEquipSkinHist_Valid(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateEquipSkinHist{AckHandle: 1, ArmourID: 10001, MogType: 0}
	handleMsgMhfUpdateEquipSkinHist(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}

	if _, ok := charRepo.columns["skin_hist"]; !ok {
		t.Error("Expected skin_hist to be saved")
	}
}

func TestHandleMsgMhfUpdateEquipSkinHist_LowArmourID(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateEquipSkinHist{AckHandle: 1, ArmourID: 5000, MogType: 0}
	handleMsgMhfUpdateEquipSkinHist(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}

	if _, ok := charRepo.columns["skin_hist"]; ok {
		t.Error("Expected skin_hist NOT to be saved for low ArmourID")
	}
}

func TestHandleMsgMhfUpdateEquipSkinHist_HighMogType(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateEquipSkinHist{AckHandle: 1, ArmourID: 10001, MogType: 5}
	handleMsgMhfUpdateEquipSkinHist(s, pkt)

	select {
	case <-s.sendPackets:
	default:
		t.Fatal("No response packet queued")
	}

	if _, ok := charRepo.columns["skin_hist"]; ok {
		t.Error("Expected skin_hist NOT to be saved for high MogType")
	}
}

// Tests consolidated from handlers_coverage3_test.go

func TestNonTrivialHandlers_NoDB_Misc(t *testing.T) {
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
}

func TestEmptyHandlers_MiscFiles_Misc(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{

		{"handleMsgMhfGetDailyMissionMaster", func() { handleMsgMhfGetDailyMissionMaster(session, nil) }},
		{"handleMsgMhfGetDailyMissionPersonal", func() { handleMsgMhfGetDailyMissionPersonal(session, nil) }},
		{"handleMsgMhfSetDailyMissionPersonal", func() { handleMsgMhfSetDailyMissionPersonal(session, nil) }},
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
