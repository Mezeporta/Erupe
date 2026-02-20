package channelserver

import (
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// GuildMeal represents a guild cooking meal entry.
type GuildMeal struct {
	ID        uint32    `db:"id"`
	MealID    uint32    `db:"meal_id"`
	Level     uint32    `db:"level"`
	CreatedAt time.Time `db:"created_at"`
}

func handleMsgMhfLoadGuildCooking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadGuildCooking)
	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	allMeals, err := s.server.guildRepo.ListMeals(guild.ID)
	if err != nil {
		s.logger.Error("Failed to get guild meals from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	}
	var meals []*GuildMeal
	for _, meal := range allMeals {
		if meal.CreatedAt.Add(60 * time.Minute).After(TimeAdjusted()) {
			meals = append(meals, meal)
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(meals)))
	for _, meal := range meals {
		bf.WriteUint32(meal.ID)
		bf.WriteUint32(meal.MealID)
		bf.WriteUint32(meal.Level)
		bf.WriteUint32(uint32(meal.CreatedAt.Unix()))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistGuildCooking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildCooking)
	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	startTime := TimeAdjusted().Add(time.Duration(s.server.erupeConfig.GameplayOptions.ClanMealDuration-3600) * time.Second)
	if pkt.OverwriteID != 0 {
		if err := s.server.guildRepo.UpdateMeal(pkt.OverwriteID, uint32(pkt.MealID), uint32(pkt.Success), startTime); err != nil {
			s.logger.Error("Failed to update guild meal", zap.Error(err))
		}
	} else {
		id, err := s.server.guildRepo.CreateMeal(guild.ID, uint32(pkt.MealID), uint32(pkt.Success), startTime)
		if err != nil {
			s.logger.Error("Failed to insert guild meal", zap.Error(err))
			doAckBufFail(s, pkt.AckHandle, nil)
			return
		}
		pkt.OverwriteID = id
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(1)
	bf.WriteUint32(pkt.OverwriteID)
	bf.WriteUint32(uint32(pkt.MealID))
	bf.WriteUint32(uint32(pkt.Success))
	bf.WriteUint32(uint32(startTime.Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildWeeklyBonusMaster(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildWeeklyBonusMaster)

	// Values taken from brand new guild capture
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 40))
}
func handleMsgMhfGetGuildWeeklyBonusActiveCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildWeeklyBonusActiveCount)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(60) // Active count
	bf.WriteUint8(60) // Current active count
	bf.WriteUint8(0)  // New active count
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGuildHuntdata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGuildHuntdata)
	bf := byteframe.NewByteFrame()
	switch pkt.Operation {
	case 0: // Acquire
		if err := s.server.guildRepo.ClaimHuntBox(s.charID, TimeAdjusted()); err != nil {
			s.logger.Error("Failed to update guild hunt box claimed time", zap.Error(err))
		}
	case 1: // Enumerate
		bf.WriteUint8(0) // Entries
		kills, err := s.server.guildRepo.ListGuildKills(pkt.GuildID, s.charID)
		if err == nil {
			var count uint8
			for _, kill := range kills {
				if count == 255 {
					break
				}
				count++
				bf.WriteUint32(kill.ID)
				bf.WriteUint32(kill.Monster)
			}
			_, _ = bf.Seek(0, 0)
			bf.WriteUint8(count)
		}
	case 2: // Check
		guild, err := s.server.guildRepo.GetByCharID(s.charID)
		if err == nil {
			count, err := s.server.guildRepo.CountGuildKills(guild.ID, s.charID)
			if err == nil && count > 0 {
				bf.WriteBool(true)
			} else {
				bf.WriteBool(false)
			}
		} else {
			bf.WriteBool(false)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAddGuildWeeklyBonusExceptionalUser(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddGuildWeeklyBonusExceptionalUser)
	// TODO: record pkt.NumUsers to DB
	// must use addition
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
