package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	ps "erupe-ce/common/pascalstring"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"
)

func handleMsgMhfAcquireCafeItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireCafeItem)
	var netcafePoints uint32
	err := s.server.db.QueryRow("UPDATE characters SET netcafe_points = netcafe_points - $1 WHERE id = $2 RETURNING netcafe_points", pkt.PointCost, s.charID).Scan(&netcafePoints)
	if err != nil {
		s.logger.Error("Failed to get netcafe points from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(netcafePoints)
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateCafepoint)
	var netcafePoints uint32
	err := s.server.db.QueryRow("SELECT COALESCE(netcafe_points, 0) FROM characters WHERE id = $1", s.charID).Scan(&netcafePoints)
	if err != nil {
		s.logger.Error("Failed to get netcate points from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(netcafePoints)
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfCheckDailyCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckDailyCafepoint)

	midday := TimeMidnight().Add(12 * time.Hour)
	if TimeAdjusted().After(midday) {
		midday = midday.Add(24 * time.Hour)
	}

	// get time after which daily claiming would be valid from db
	var dailyTime time.Time
	err := s.server.db.QueryRow("SELECT COALESCE(daily_time, $2) FROM characters WHERE id = $1", s.charID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Scan(&dailyTime)
	if err != nil {
		s.logger.Error("Failed to get daily_time savedata from db", zap.Error(err))
	}

	var bondBonus, bonusQuests, dailyQuests uint32
	bf := byteframe.NewByteFrame()
	if midday.After(dailyTime) {
		addPointNetcafe(s, 5)
		bondBonus = 5 // Bond point bonus quests
		bonusQuests = s.server.erupeConfig.GameplayOptions.BonusQuestAllowance
		dailyQuests = s.server.erupeConfig.GameplayOptions.DailyQuestAllowance
		s.server.db.Exec("UPDATE characters SET daily_time=$1, bonus_quests = $2, daily_quests = $3 WHERE id=$4", midday, bonusQuests, dailyQuests, s.charID)
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

	var cafeReset time.Time
	err := s.server.db.QueryRow(`SELECT cafe_reset FROM characters WHERE id=$1`, s.charID).Scan(&cafeReset)
	if err != nil {
		cafeReset = TimeWeekNext()
		s.server.db.Exec(`UPDATE characters SET cafe_reset=$1 WHERE id=$2`, cafeReset, s.charID)
	}
	if TimeAdjusted().After(cafeReset) {
		cafeReset = TimeWeekNext()
		s.server.db.Exec(`UPDATE characters SET cafe_time=0, cafe_reset=$1 WHERE id=$2`, cafeReset, s.charID)
		s.server.db.Exec(`DELETE FROM cafe_accepted WHERE character_id=$1`, s.charID)
	}

	var cafeTime uint32
	err = s.server.db.QueryRow("SELECT cafe_time FROM characters WHERE id = $1", s.charID).Scan(&cafeTime)
	if err != nil {
		panic(err)
	}
	if mhfcourse.CourseExists(30, s.courses) {
		cafeTime = uint32(TimeAdjusted().Unix()) - uint32(s.sessionStart) + cafeTime
	}
	bf.WriteUint32(cafeTime)
	if _config.ErupeConfig.RealClientMode >= _config.ZZ {
		bf.WriteUint16(0)
		ps.Uint16(bf, fmt.Sprintf(s.server.i18n.cafe.reset, int(cafeReset.Month()), cafeReset.Day()), true)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

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
	rows, err := s.server.db.Queryx(`
	SELECT cb.id, time_req, item_type, item_id, quantity,
	(
		SELECT count(*)
		FROM cafe_accepted ca
		WHERE cb.id = ca.cafe_id AND ca.character_id = $1
	)::int::bool AS claimed
	FROM cafebonus cb ORDER BY id ASC;`, s.charID)
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
	rows, err := s.server.db.Queryx(`
	SELECT c.id, time_req, item_type, item_id, quantity
	FROM cafebonus c
	WHERE (
		SELECT count(*)
		FROM cafe_accepted ca
		WHERE c.id = ca.cafe_id AND ca.character_id = $1
	) < 1 AND (
		SELECT ch.cafe_time + $2
		FROM characters ch
		WHERE ch.id = $1 
	) >= time_req`, s.charID, TimeAdjusted().Unix()-s.sessionStart)
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
		bf.Seek(0, io.SeekStart)
		bf.WriteUint32(count)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
}

func handleMsgMhfPostCafeDurationBonusReceived(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostCafeDurationBonusReceived)
	var cafeBonus CafeBonus
	for _, cbID := range pkt.CafeBonusID {
		err := s.server.db.QueryRow(`
		SELECT cb.id, item_type, quantity FROM cafebonus cb WHERE cb.id=$1
		`, cbID).Scan(&cafeBonus.ID, &cafeBonus.ItemType, &cafeBonus.Quantity)
		if err == nil {
			if cafeBonus.ItemType == 17 {
				addPointNetcafe(s, int(cafeBonus.Quantity))
			}
		}
		s.server.db.Exec("INSERT INTO public.cafe_accepted VALUES ($1, $2)", cbID, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func addPointNetcafe(s *Session, p int) error {
	var points int
	err := s.server.db.QueryRow("SELECT netcafe_points FROM characters WHERE id = $1", s.charID).Scan(&points)
	if err != nil {
		return err
	}
	if points+p > s.server.erupeConfig.GameplayOptions.MaximumNP {
		points = s.server.erupeConfig.GameplayOptions.MaximumNP
	} else {
		points += p
	}
	s.server.db.Exec("UPDATE characters SET netcafe_points=$1 WHERE id=$2", points, s.charID)
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
	s.server.db.Exec("UPDATE characters SET boost_time=$1 WHERE id=$2", boostLimit, s.charID)
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
	var boostLimit time.Time
	err := s.server.db.QueryRow("SELECT boost_time FROM characters WHERE id=$1", s.charID).Scan(&boostLimit)
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
	var boostLimit time.Time
	err := s.server.db.QueryRow("SELECT boost_time FROM characters WHERE id=$1", s.charID).Scan(&boostLimit)
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
