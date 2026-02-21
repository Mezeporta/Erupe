package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	ps "erupe-ce/common/pascalstring"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"
)

func handleMsgMhfAcquireCafeItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireCafeItem)
	netcafePoints, err := adjustCharacterInt(s, "netcafe_points", -int(pkt.PointCost))
	if err != nil {
		s.logger.Error("Failed to deduct netcafe points", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(netcafePoints))
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateCafepoint)
	netcafePoints, err := readCharacterInt(s, "netcafe_points")
	if err != nil {
		s.logger.Error("Failed to get netcafe points", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(netcafePoints))
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfCheckDailyCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckDailyCafepoint)

	midday := TimeMidnight().Add(12 * time.Hour)
	if TimeAdjusted().After(midday) {
		midday = midday.Add(24 * time.Hour)
	}

	// get time after which daily claiming would be valid from db
	dailyTime, err := s.server.charRepo.ReadTime(s.charID, "daily_time", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		s.logger.Error("Failed to get daily_time savedata from db", zap.Error(err))
	}

	var bondBonus, bonusQuests, dailyQuests uint32
	bf := byteframe.NewByteFrame()
	if midday.After(dailyTime) {
		_ = addPointNetcafe(s, 5)
		bondBonus = 5 // Bond point bonus quests
		bonusQuests = s.server.erupeConfig.GameplayOptions.BonusQuestAllowance
		dailyQuests = s.server.erupeConfig.GameplayOptions.DailyQuestAllowance
		if err := s.server.charRepo.UpdateDailyCafe(s.charID, midday, bonusQuests, dailyQuests); err != nil {
			s.logger.Error("Failed to update daily cafe data", zap.Error(err))
		}
		bf.WriteBool(true) // Success?
	} else {
		bf.WriteBool(false)
	}
	bf.WriteUint32(bondBonus)
	bf.WriteUint32(bonusQuests)
	bf.WriteUint32(dailyQuests)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetCafeDuration(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetCafeDuration)
	bf := byteframe.NewByteFrame()

	cafeReset, err := s.server.charRepo.ReadTime(s.charID, "cafe_reset", time.Time{})
	if err != nil {
		cafeReset = TimeWeekNext()
		if err := s.server.charRepo.SaveTime(s.charID, "cafe_reset", cafeReset); err != nil {
			s.logger.Error("Failed to set cafe reset time", zap.Error(err))
		}
	}
	if TimeAdjusted().After(cafeReset) {
		cafeReset = TimeWeekNext()
		if err := s.server.charRepo.ResetCafeTime(s.charID, cafeReset); err != nil {
			s.logger.Error("Failed to reset cafe time", zap.Error(err))
		}
		if err := s.server.cafeRepo.ResetAccepted(s.charID); err != nil {
			s.logger.Error("Failed to delete accepted cafe bonuses", zap.Error(err))
		}
	}

	cafeTime, err := readCharacterInt(s, "cafe_time")
	if err != nil {
		s.logger.Error("Failed to get cafe time", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if mhfcourse.CourseExists(30, s.courses) {
		cafeTime = int(TimeAdjusted().Unix()) - int(s.sessionStart) + cafeTime
	}
	bf.WriteUint32(uint32(cafeTime))
	if s.server.erupeConfig.RealClientMode >= cfg.ZZ {
		bf.WriteUint16(0)
		ps.Uint16(bf, fmt.Sprintf(s.server.i18n.cafe.reset, int(cafeReset.Month()), cafeReset.Day()), true)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// CafeBonus represents a cafe duration bonus reward entry.
type CafeBonus struct {
	ID       uint32 `db:"id"`
	TimeReq  uint32 `db:"time_req"`
	ItemType uint32 `db:"item_type"`
	ItemID   uint32 `db:"item_id"`
	Quantity uint32 `db:"quantity"`
	Claimed  bool   `db:"claimed"`
}

func handleMsgMhfGetCafeDurationBonusInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetCafeDurationBonusInfo)
	bf := byteframe.NewByteFrame()

	var count uint32
	rows, err := s.server.cafeRepo.GetBonuses(s.charID)
	if err != nil {
		s.logger.Error("Error getting cafebonus", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	} else {
		for rows.Next() {
			count++
			cafeBonus := &CafeBonus{}
			err = rows.StructScan(&cafeBonus)
			if err != nil {
				s.logger.Error("Error scanning cafebonus", zap.Error(err))
			}
			bf.WriteUint32(cafeBonus.TimeReq)
			bf.WriteUint32(cafeBonus.ItemType)
			bf.WriteUint32(cafeBonus.ItemID)
			bf.WriteUint32(cafeBonus.Quantity)
			bf.WriteBool(cafeBonus.Claimed)
		}
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0)
		resp.WriteUint32(uint32(TimeAdjusted().Unix()))
		resp.WriteUint32(count)
		resp.WriteBytes(bf.Data())
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

func handleMsgMhfReceiveCafeDurationBonus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReceiveCafeDurationBonus)
	bf := byteframe.NewByteFrame()
	var count uint32
	bf.WriteUint32(0)
	rows, err := s.server.cafeRepo.GetClaimable(s.charID, TimeAdjusted().Unix()-s.sessionStart)
	if err != nil || !mhfcourse.CourseExists(30, s.courses) {
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		for rows.Next() {
			cafeBonus := &CafeBonus{}
			err = rows.StructScan(cafeBonus)
			if err != nil {
				continue
			}
			count++
			bf.WriteUint32(cafeBonus.ID)
			bf.WriteUint32(cafeBonus.ItemType)
			bf.WriteUint32(cafeBonus.ItemID)
			bf.WriteUint32(cafeBonus.Quantity)
		}
		_, _ = bf.Seek(0, io.SeekStart)
		bf.WriteUint32(count)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
}

func handleMsgMhfPostCafeDurationBonusReceived(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostCafeDurationBonusReceived)
	for _, cbID := range pkt.CafeBonusID {
		itemType, quantity, err := s.server.cafeRepo.GetBonusItem(cbID)
		if err == nil {
			if itemType == 17 {
				_ = addPointNetcafe(s, int(quantity))
			}
		}
		if err := s.server.cafeRepo.AcceptBonus(cbID, s.charID); err != nil {
			s.logger.Error("Failed to insert accepted cafe bonus", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func addPointNetcafe(s *Session, p int) error {
	points, err := readCharacterInt(s, "netcafe_points")
	if err != nil {
		return err
	}
	points = min(points+p, s.server.erupeConfig.GameplayOptions.MaximumNP)
	if err := s.server.charRepo.SaveInt(s.charID, "netcafe_points", points); err != nil {
		s.logger.Error("Failed to update netcafe points", zap.Error(err))
	}
	return nil
}

func handleMsgMhfStartBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStartBoostTime)
	bf := byteframe.NewByteFrame()
	boostLimit := TimeAdjusted().Add(time.Duration(s.server.erupeConfig.GameplayOptions.BoostTimeDuration) * time.Second)
	if s.server.erupeConfig.GameplayOptions.DisableBoostTime {
		bf.WriteUint32(0)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}
	if err := s.server.charRepo.SaveTime(s.charID, "boost_time", boostLimit); err != nil {
		s.logger.Error("Failed to update boost time", zap.Error(err))
	}
	bf.WriteUint32(uint32(boostLimit.Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTime)
	doAckBufSucceed(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfGetBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTimeLimit)
	bf := byteframe.NewByteFrame()
	boostLimit, err := s.server.charRepo.ReadTime(s.charID, "boost_time", time.Time{})
	if err != nil {
		bf.WriteUint32(0)
	} else {
		bf.WriteUint32(uint32(boostLimit.Unix()))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetBoostRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostRight)
	boostLimit, err := s.server.charRepo.ReadTime(s.charID, "boost_time", time.Time{})
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		return
	}
	if boostLimit.After(TimeAdjusted()) {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x02})
	}
}

func handleMsgMhfPostBoostTimeQuestReturn(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTimeQuestReturn)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfPostBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTime)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfPostBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTimeLimit)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
