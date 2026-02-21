package channelserver

import (
	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"math/bits"
	"time"

	"go.uber.org/zap"
)

func handleMsgMhfGetEtcPoints(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEtcPoints)

	dailyTime, _ := s.server.charRepo.ReadTime(s.charID, "daily_time", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if TimeAdjusted().After(dailyTime) {
		if err := s.server.charRepo.ResetDailyQuests(s.charID); err != nil {
			s.logger.Error("Failed to reset daily quests", zap.Error(err))
		}
	}

	bonusQuests, dailyQuests, promoPoints, err := s.server.charRepo.ReadEtcPoints(s.charID)
	if err != nil {
		s.logger.Error("Failed to get etc points", zap.Error(err))
	}
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
	default:
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	value, err := readCharacterInt(s, column)
	if err == nil {
		newVal := max(value+int(pkt.Delta), 0)
		if err := s.server.charRepo.SaveInt(s.charID, column, newVal); err != nil {
			s.logger.Error("Failed to update etc point", zap.Error(err))
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
		if s.server.erupeConfig.RealClientMode <= cfg.G9 {
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

func handleMsgMhfGetDailyMissionMaster(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

// Equip skin history buffer sizes per game version
const (
	skinHistSizeZZ = 3200 // ZZ and newer
	skinHistSizeZ2 = 2560 // Z2 and older
	skinHistSizeZ1 = 1280 // Z1 and older
)

func equipSkinHistSize(mode cfg.Mode) int {
	size := skinHistSizeZZ
	if mode <= cfg.Z2 {
		size = skinHistSizeZ2
	}
	if mode <= cfg.Z1 {
		size = skinHistSizeZ1
	}
	return size
}

func handleMsgMhfGetEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEquipSkinHist)
	size := equipSkinHistSize(s.server.erupeConfig.RealClientMode)
	loadCharacterData(s, pkt.AckHandle, "skin_hist", make([]byte, size))
}

func handleMsgMhfUpdateEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateEquipSkinHist)
	size := equipSkinHistSize(s.server.erupeConfig.RealClientMode)
	data, err := s.server.charRepo.LoadColumnWithDefault(s.charID, "skin_hist", make([]byte, size))
	if err != nil {
		s.logger.Error("Failed to get skin_hist", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if pkt.ArmourID < 10000 || pkt.MogType > 4 {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bit := int(pkt.ArmourID) - 10000
	sectionSize := size / 5
	if bit/8 >= sectionSize {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	startByte := sectionSize * int(pkt.MogType)
	byteInd := bit / 8
	bitInByte := bit % 8
	data[startByte+byteInd] |= bits.Reverse8(1 << uint(bitInByte))
	dumpSaveData(s, data, "skinhist")
	if err := s.server.charRepo.SaveColumn(s.charID, "skin_hist", data); err != nil {
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

	s.server.minidataLock.RLock()
	data, ok := s.server.minidataParts[pkt.CharID]
	s.server.minidataLock.RUnlock()

	if !ok {
		data = make([]byte, 1)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetEnhancedMinidata)
	dumpSaveData(s, pkt.RawDataPayload, "minidata")

	s.server.minidataLock.Lock()
	s.server.minidataParts[s.charID] = pkt.RawDataPayload
	s.server.minidataLock.Unlock()

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetLobbyCrowd(s *Session, p mhfpacket.MHFPacket) {
	// this requests a specific server's population but seems to have been
	// broken at some point on live as every example response across multiple
	// servers sends back the exact same information?
	// It can be worried about later if we ever get to the point where there are
	// full servers to actually need to migrate people from and empty ones to
	pkt := p.(*mhfpacket.MsgMhfGetLobbyCrowd)
	const lobbyCrowdResponseSize = 0x320
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, lobbyCrowdResponseSize))
}

// TrendWeapon represents trending weapon usage data.
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
