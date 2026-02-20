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
	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	adventures, err := s.server.guildRepo.ListAdventures(guild.ID)
	if err != nil {
		s.logger.Error("Failed to get guild adventures from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	temp := byteframe.NewByteFrame()
	for _, adventureData := range adventures {
		temp.WriteUint32(adventureData.ID)
		temp.WriteUint32(adventureData.Destination)
		temp.WriteUint32(adventureData.Charge)
		temp.WriteUint32(adventureData.Depart)
		temp.WriteUint32(adventureData.Return)
		temp.WriteBool(stringsupport.CSVContains(adventureData.CollectedBy, int(s.charID)))
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(adventures)))
	bf.WriteBytes(temp.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildAdventure)
	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	if err := s.server.guildRepo.CreateAdventure(guild.ID, pkt.Destination, TimeAdjusted().Unix(), TimeAdjusted().Add(6*time.Hour).Unix()); err != nil {
		s.logger.Error("Failed to register guild adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildAdventure)
	if err := s.server.guildRepo.CollectAdventure(pkt.ID, s.charID); err != nil {
		s.logger.Error("Failed to collect adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfChargeGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfChargeGuildAdventure)
	if err := s.server.guildRepo.ChargeAdventure(pkt.ID, pkt.Amount); err != nil {
		s.logger.Error("Failed to charge guild adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfRegistGuildAdventureDiva(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildAdventureDiva)
	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	if err := s.server.guildRepo.CreateAdventureWithCharge(guild.ID, pkt.Destination, pkt.Charge, TimeAdjusted().Unix(), TimeAdjusted().Add(1*time.Hour).Unix()); err != nil {
		s.logger.Error("Failed to register guild adventure", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
