package channelserver

import (
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

type GuildMeal struct {
	ID        uint32    `db:"id"`
	MealID    uint32    `db:"meal_id"`
	Level     uint32    `db:"level"`
	CreatedAt time.Time `db:"created_at"`
}

func handleMsgMhfLoadGuildCooking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadGuildCooking)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	data, err := s.server.db.Queryx("SELECT id, meal_id, level, created_at FROM guild_meals WHERE guild_id = $1", guild.ID)
	if err != nil {
		s.logger.Error("Failed to get guild meals from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	}
	var meals []GuildMeal
	var temp GuildMeal
	for data.Next() {
		err = data.StructScan(&temp)
		if err != nil {
			continue
		}
		if temp.CreatedAt.Add(60 * time.Minute).After(TimeAdjusted()) {
			meals = append(meals, temp)
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
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	startTime := TimeAdjusted().Add(time.Duration(s.server.erupeConfig.GameplayOptions.ClanMealDuration-3600) * time.Second)
	if pkt.OverwriteID != 0 {
		if _, err := s.server.db.Exec("UPDATE guild_meals SET meal_id = $1, level = $2, created_at = $3 WHERE id = $4", pkt.MealID, pkt.Success, startTime, pkt.OverwriteID); err != nil {
			s.logger.Error("Failed to update guild meal", zap.Error(err))
		}
	} else {
		_ = s.server.db.QueryRow("INSERT INTO guild_meals (guild_id, meal_id, level, created_at) VALUES ($1, $2, $3, $4) RETURNING id", guild.ID, pkt.MealID, pkt.Success, startTime).Scan(&pkt.OverwriteID)
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
		if _, err := s.server.db.Exec(`UPDATE guild_characters SET box_claimed=$1 WHERE character_id=$2`, TimeAdjusted(), s.charID); err != nil {
			s.logger.Error("Failed to update guild hunt box claimed time", zap.Error(err))
		}
	case 1: // Enumerate
		bf.WriteUint8(0) // Entries
		rows, err := s.server.db.Query(`SELECT kl.id, kl.monster FROM kill_logs kl
			INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
			WHERE gc.guild_id=$1
			AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
		`, pkt.GuildID, s.charID)
		if err == nil {
			var count uint8
			var huntID, monID uint32
			for rows.Next() {
				err = rows.Scan(&huntID, &monID)
				if err != nil {
					continue
				}
				if count == 255 {
					_ = rows.Close()
					break
				}
				count++
				bf.WriteUint32(huntID)
				bf.WriteUint32(monID)
			}
			_, _ = bf.Seek(0, 0)
			bf.WriteUint8(count)
		}
	case 2: // Check
		guild, err := GetGuildInfoByCharacterId(s, s.charID)
		if err == nil {
			var count uint8
			err = s.server.db.QueryRow(`SELECT COUNT(*) FROM kill_logs kl
				INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
				WHERE gc.guild_id=$1
				AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
			`, guild.ID, s.charID).Scan(&count)
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
