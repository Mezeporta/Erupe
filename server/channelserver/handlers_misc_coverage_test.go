package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"testing"
)

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
