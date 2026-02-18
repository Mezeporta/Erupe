package channelserver

import (
	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"math/bits"
	"time"

	"go.uber.org/zap"
)

func getGoocooData(s *Session, cid uint32) [][]byte {
	var goocoo []byte
	var goocoos [][]byte
	for i := 0; i < 5; i++ {
		err := s.server.db.QueryRow(fmt.Sprintf("SELECT goocoo%d FROM goocoo WHERE id=$1", i), cid).Scan(&goocoo)
		if err != nil {
			if _, err := s.server.db.Exec("INSERT INTO goocoo (id) VALUES ($1)", s.charID); err != nil {
				s.logger.Error("Failed to insert goocoo record", zap.Error(err))
			}
			return goocoos
		}
		if err == nil && goocoo != nil {
			goocoos = append(goocoos, goocoo)
		}
	}
	return goocoos
}

func handleMsgMhfEnumerateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuacot)
	bf := byteframe.NewByteFrame()
	goocoos := getGoocooData(s, s.charID)
	bf.WriteUint16(uint16(len(goocoos)))
	bf.WriteUint16(0)
	for _, goocoo := range goocoos {
		bf.WriteBytes(goocoo)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuacot)
	for _, goocoo := range pkt.Goocoos {
		if goocoo.Data1[0] == 0 {
			if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE goocoo SET goocoo%d=NULL WHERE id=$1", goocoo.Index), s.charID); err != nil {
				s.logger.Error("Failed to clear goocoo slot", zap.Error(err))
			}
		} else {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(goocoo.Index)
			for i := range goocoo.Data1 {
				bf.WriteInt16(goocoo.Data1[i])
			}
			for i := range goocoo.Data2 {
				bf.WriteUint32(goocoo.Data2[i])
			}
			bf.WriteUint8(uint8(len(goocoo.Name)))
			bf.WriteBytes(goocoo.Name)
			if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE goocoo SET goocoo%d=$1 WHERE id=$2", goocoo.Index), bf.Data(), s.charID); err != nil {
				s.logger.Error("Failed to update goocoo slot", zap.Error(err))
			}
			dumpSaveData(s, bf.Data(), fmt.Sprintf("goocoo-%d", goocoo.Index))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type Scenario struct {
	MainID uint32
	// 0 = Basic
	// 1 = Veteran
	// 3 = Other
	// 6 = Pallone
	// 7 = Diva
	CategoryID uint8
}

func handleMsgMhfInfoScenarioCounter(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoScenarioCounter)
	var scenarios []Scenario
	var scenario Scenario
	scenarioData, err := s.server.db.Queryx("SELECT scenario_id, category_id FROM scenario_counter")
	if err != nil {
		_ = scenarioData.Close()
		s.logger.Error("Failed to get scenario counter info from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for scenarioData.Next() {
		err = scenarioData.Scan(&scenario.MainID, &scenario.CategoryID)
		if err != nil {
			continue
		}
		scenarios = append(scenarios, scenario)
	}

	// Trim excess scenarios
	if len(scenarios) > 128 {
		scenarios = scenarios[:128]
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(scenarios)))
	for _, scenario := range scenarios {
		bf.WriteUint32(scenario.MainID)
		// If item exchange
		switch scenario.CategoryID {
		case 3, 6, 7:
			bf.WriteBool(true)
		default:
			bf.WriteBool(false)
		}
		bf.WriteUint8(scenario.CategoryID)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetEtcPoints(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEtcPoints)

	var dailyTime time.Time
	_ = s.server.db.QueryRow("SELECT COALESCE(daily_time, $2) FROM characters WHERE id = $1", s.charID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Scan(&dailyTime)
	if TimeAdjusted().After(dailyTime) {
		if _, err := s.server.db.Exec("UPDATE characters SET bonus_quests = 0, daily_quests = 0 WHERE id=$1", s.charID); err != nil {
			s.logger.Error("Failed to reset daily quests", zap.Error(err))
		}
	}

	var bonusQuests, dailyQuests, promoPoints uint32
	_ = s.server.db.QueryRow(`SELECT bonus_quests, daily_quests, promo_points FROM characters WHERE id = $1`, s.charID).Scan(&bonusQuests, &dailyQuests, &promoPoints)
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(3) // Maybe a count of uint32(s)?
	resp.WriteUint32(bonusQuests)
	resp.WriteUint32(dailyQuests)
	resp.WriteUint32(promoPoints)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateEtcPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateEtcPoint)

	var column string
	switch pkt.PointType {
	case 0:
		column = "bonus_quests"
	case 1:
		column = "daily_quests"
	case 2:
		column = "promo_points"
	}

	var value int16
	err := s.server.db.QueryRow(fmt.Sprintf(`SELECT %s FROM characters WHERE id = $1`, column), s.charID).Scan(&value)
	if err == nil {
		if value+pkt.Delta < 0 {
			if _, err := s.server.db.Exec(fmt.Sprintf(`UPDATE characters SET %s = 0 WHERE id = $1`, column), s.charID); err != nil {
				s.logger.Error("Failed to reset etc point", zap.Error(err))
			}
		} else {
			if _, err := s.server.db.Exec(fmt.Sprintf(`UPDATE characters SET %s = %s + $1 WHERE id = $2`, column, column), pkt.Delta, s.charID); err != nil {
				s.logger.Error("Failed to update etc point", zap.Error(err))
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUnreserveSrg(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUnreserveSrg)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfKickExportForce(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEarthStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthStatus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(TimeWeekStart().Unix())) // Start
	bf.WriteUint32(uint32(TimeWeekNext().Unix()))  // End
	bf.WriteInt32(s.server.erupeConfig.EarthStatus)
	bf.WriteInt32(s.server.erupeConfig.EarthID)
	for i, m := range s.server.erupeConfig.EarthMonsters {
		if _config.ErupeConfig.RealClientMode <= _config.G9 {
			if i == 3 {
				break
			}
		}
		if i == 4 {
			break
		}
		bf.WriteInt32(m)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistSpabiTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEarthValue(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthValue)
	type EarthValues struct {
		Value []uint32
	}

	var earthValues []EarthValues
	switch pkt.ReqType {
	case 1:
		earthValues = []EarthValues{
			{[]uint32{1, 312, 0, 0, 0, 0}},
			{[]uint32{2, 99, 0, 0, 0, 0}},
		}
	case 2:
		earthValues = []EarthValues{
			{[]uint32{1, 5771, 0, 0, 0, 0}},
			{[]uint32{2, 1847, 0, 0, 0, 0}},
		}
	case 3:
		earthValues = []EarthValues{
			{[]uint32{1001, 36, 0, 0, 0, 0}},
			{[]uint32{9001, 3, 0, 0, 0, 0}},
			{[]uint32{9002, 10, 300, 0, 0, 0}},
		}
	}

	var data []*byteframe.ByteFrame
	for _, i := range earthValues {
		bf := byteframe.NewByteFrame()
		for _, j := range i.Value {
			bf.WriteUint32(j)
		}
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfDebugPostValue(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRandFromTable(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRandFromTable)
	bf := byteframe.NewByteFrame()
	for i := uint16(0); i < pkt.Results; i++ {
		bf.WriteUint32(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetSenyuDailyCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetSenyuDailyCount)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type SeibattleTimetable struct {
	Start time.Time
	End   time.Time
}

type SeibattleKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type SeibattleCareer struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}

type SeibattleOpponent struct {
	Unk0 int32
	Unk1 int8
}

type SeibattleConventionResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
}

type SeibattleCharScore struct {
	Unk0 uint32
}

type SeibattleCurResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
}

type Seibattle struct {
	Timetable        []SeibattleTimetable
	KeyScore         []SeibattleKeyScore
	Career           []SeibattleCareer
	Opponent         []SeibattleOpponent
	ConventionResult []SeibattleConventionResult
	CharScore        []SeibattleCharScore
	CurResult        []SeibattleCurResult
}

func handleMsgMhfGetSeibattle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetSeibattle)
	var data []*byteframe.ByteFrame
	seibattle := Seibattle{
		Timetable: []SeibattleTimetable{
			{TimeMidnight(), TimeMidnight().Add(time.Hour * 8)},
			{TimeMidnight().Add(time.Hour * 8), TimeMidnight().Add(time.Hour * 16)},
			{TimeMidnight().Add(time.Hour * 16), TimeMidnight().Add(time.Hour * 24)},
		},
		KeyScore: []SeibattleKeyScore{
			{0, 0},
		},
		Career: []SeibattleCareer{
			{0, 0, 0},
		},
		Opponent: []SeibattleOpponent{
			{1, 1},
		},
		ConventionResult: []SeibattleConventionResult{
			{0, 0, 0, 0, 0},
		},
		CharScore: []SeibattleCharScore{
			{0},
		},
		CurResult: []SeibattleCurResult{
			{0, 0, 0, 0},
		},
	}

	switch pkt.Type {
	case 1:
		for _, timetable := range seibattle.Timetable {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(uint32(timetable.Start.Unix()))
			bf.WriteUint32(uint32(timetable.End.Unix()))
			data = append(data, bf)
		}
	case 3: // Key score?
		for _, keyScore := range seibattle.KeyScore {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(keyScore.Unk0)
			bf.WriteInt32(keyScore.Unk1)
			data = append(data, bf)
		}
	case 4: // Career?
		for _, career := range seibattle.Career {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(career.Unk0)
			bf.WriteUint16(career.Unk1)
			bf.WriteUint16(career.Unk2)
			data = append(data, bf)
		}
	case 5: // Opponent?
		for _, opponent := range seibattle.Opponent {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(opponent.Unk0)
			bf.WriteInt8(opponent.Unk1)
			data = append(data, bf)
		}
	case 6: // Convention result?
		for _, conventionResult := range seibattle.ConventionResult {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(conventionResult.Unk0)
			bf.WriteUint16(conventionResult.Unk1)
			bf.WriteUint16(conventionResult.Unk2)
			bf.WriteUint16(conventionResult.Unk3)
			bf.WriteUint16(conventionResult.Unk4)
			data = append(data, bf)
		}
	case 7: // Char score?
		for _, charScore := range seibattle.CharScore {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(charScore.Unk0)
			data = append(data, bf)
		}
	case 8: // Cur result?
		for _, curResult := range seibattle.CurResult {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(curResult.Unk0)
			bf.WriteUint16(curResult.Unk1)
			bf.WriteUint16(curResult.Unk2)
			bf.WriteUint16(curResult.Unk3)
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostSeibattle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostSeibattle)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetDailyMissionMaster(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func equipSkinHistSize() int {
	size := 3200
	if _config.ErupeConfig.RealClientMode <= _config.Z2 {
		size = 2560
	}
	if _config.ErupeConfig.RealClientMode <= _config.Z1 {
		size = 1280
	}
	return size
}

func handleMsgMhfGetEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEquipSkinHist)
	size := equipSkinHistSize()
	var data []byte
	err := s.server.db.QueryRow("SELECT COALESCE(skin_hist::bytea, $2::bytea) FROM characters WHERE id = $1", s.charID, make([]byte, size)).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load skin_hist", zap.Error(err))
		data = make([]byte, size)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfUpdateEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateEquipSkinHist)
	size := equipSkinHistSize()
	var data []byte
	err := s.server.db.QueryRow("SELECT COALESCE(skin_hist, $2) FROM characters WHERE id = $1", s.charID, make([]byte, size)).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to get skin_hist", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bit := int(pkt.ArmourID) - 10000
	startByte := (size / 5) * int(pkt.MogType)
	// psql set_bit could also work but I couldn't get it working
	byteInd := bit / 8
	bitInByte := bit % 8
	data[startByte+byteInd] |= bits.Reverse8(1 << uint(bitInByte))
	dumpSaveData(s, data, "skinhist")
	if _, err := s.server.db.Exec("UPDATE characters SET skin_hist=$1 WHERE id=$2", data, s.charID); err != nil {
		s.logger.Error("Failed to update skin history", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetUdShopCoin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdShopCoin)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUseUdShopCoin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEnhancedMinidata)
	// this looks to be the detailed chunk of information you can pull up on players in town
	var data []byte
	err := s.server.db.QueryRow("SELECT minidata FROM characters WHERE id = $1", pkt.CharID).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load minidata")
		data = make([]byte, 1)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetEnhancedMinidata)
	dumpSaveData(s, pkt.RawDataPayload, "minidata")
	_, err := s.server.db.Exec("UPDATE characters SET minidata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Error("Failed to save minidata", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetLobbyCrowd(s *Session, p mhfpacket.MHFPacket) {
	// this requests a specific server's population but seems to have been
	// broken at some point on live as every example response across multiple
	// servers sends back the exact same information?
	// It can be worried about later if we ever get to the point where there are
	// full servers to actually need to migrate people from and empty ones to
	pkt := p.(*mhfpacket.MsgMhfGetLobbyCrowd)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0x320))
}

type TrendWeapon struct {
	WeaponType uint8
	WeaponID   uint16
}

func handleMsgMhfGetTrendWeapon(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTrendWeapon)
	trendWeapons := [14][3]TrendWeapon{}
	for i := uint8(0); i < 14; i++ {
		rows, err := s.server.db.Query(`SELECT weapon_id FROM trend_weapons WHERE weapon_type=$1 ORDER BY count DESC LIMIT 3`, i)
		if err != nil {
			continue
		}
		j := 0
		for rows.Next() {
			trendWeapons[i][j].WeaponType = i
			_ = rows.Scan(&trendWeapons[i][j].WeaponID)
			j++
		}
	}

	x := uint8(0)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	for _, weaponType := range trendWeapons {
		for _, weapon := range weaponType {
			bf.WriteUint8(weapon.WeaponType)
			bf.WriteUint16(weapon.WeaponID)
			x++
		}
	}
	_, _ = bf.Seek(0, 0)
	bf.WriteUint8(x)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateUseTrendWeaponLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateUseTrendWeaponLog)
	if _, err := s.server.db.Exec(`INSERT INTO trend_weapons (weapon_id, weapon_type, count) VALUES ($1, $2, 1) ON CONFLICT (weapon_id) DO
		UPDATE SET count = trend_weapons.count+1`, pkt.WeaponID, pkt.WeaponType); err != nil {
		s.logger.Error("Failed to update trend weapon log", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
