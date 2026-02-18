package channelserver

import (
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// GuildAdventure represents a guild adventure expedition.
type GuildAdventure struct {
	ID          uint32 `db:"id"`
	Destination uint32 `db:"destination"`
	Charge      uint32 `db:"charge"`
	Depart      uint32 `db:"depart"`
	Return      uint32 `db:"return"`
	CollectedBy string `db:"collected_by"`
}

func handleMsgMhfLoadGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadGuildAdventure)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	data, err := s.server.db.Queryx("SELECT id, destination, charge, depart, return, collected_by FROM guild_adventures WHERE guild_id = $1", guild.ID)
	if err != nil {
		s.logger.Error("Failed to get guild adventures from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	temp := byteframe.NewByteFrame()
	count := 0
	for data.Next() {
		count++
		adventureData := &GuildAdventure{}
		err = data.StructScan(&adventureData)
		if err != nil {
			continue
		}
		temp.WriteUint32(adventureData.ID)
		temp.WriteUint32(adventureData.Destination)
		temp.WriteUint32(adventureData.Charge)
		temp.WriteUint32(adventureData.Depart)
		temp.WriteUint32(adventureData.Return)
		temp.WriteBool(stringsupport.CSVContains(adventureData.CollectedBy, int(s.charID)))
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(count))
	bf.WriteBytes(temp.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildAdventure)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	_, err := s.server.db.Exec("INSERT INTO guild_adventures (guild_id, destination, depart, return) VALUES ($1, $2, $3, $4)", guild.ID, pkt.Destination, TimeAdjusted().Unix(), TimeAdjusted().Add(6*time.Hour).Unix())
	if err != nil {
		s.logger.Error("Failed to register guild adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildAdventure)
	var collectedBy string
	err := s.server.db.QueryRow("SELECT collected_by FROM guild_adventures WHERE id = $1", pkt.ID).Scan(&collectedBy)
	if err != nil {
		s.logger.Error("Error parsing adventure collected by", zap.Error(err))
	} else {
		collectedBy = stringsupport.CSVAdd(collectedBy, int(s.charID))
		_, err := s.server.db.Exec("UPDATE guild_adventures SET collected_by = $1 WHERE id = $2", collectedBy, pkt.ID)
		if err != nil {
			s.logger.Error("Failed to collect adventure in db", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfChargeGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfChargeGuildAdventure)
	_, err := s.server.db.Exec("UPDATE guild_adventures SET charge = charge + $1 WHERE id = $2", pkt.Amount, pkt.ID)
	if err != nil {
		s.logger.Error("Failed to charge guild adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfRegistGuildAdventureDiva(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildAdventureDiva)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	_, err := s.server.db.Exec("INSERT INTO guild_adventures (guild_id, destination, charge, depart, return) VALUES ($1, $2, $3, $4, $5)", guild.ID, pkt.Destination, pkt.Charge, TimeAdjusted().Unix(), TimeAdjusted().Add(1*time.Hour).Unix())
	if err != nil {
		s.logger.Error("Failed to register guild adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
