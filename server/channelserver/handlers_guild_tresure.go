package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"time"

	"go.uber.org/zap"
)

// TreasureHunt represents a guild treasure hunt entry.
type TreasureHunt struct {
	HuntID      uint32    `db:"id"`
	HostID      uint32    `db:"host_id"`
	Destination uint32    `db:"destination"`
	Level       uint32    `db:"level"`
	Start       time.Time `db:"start"`
	Acquired    bool      `db:"acquired"`
	Collected   bool      `db:"collected"`
	HuntData    []byte    `db:"hunt_data"`
	Hunters     uint32    `db:"hunters"`
	Claimed     bool      `db:"claimed"`
}

func handleMsgMhfEnumerateGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildTresure)
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil || guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var hunts []TreasureHunt

	switch pkt.MaxHunts {
	case 1:
		hunt, err := s.server.guildRepo.GetPendingHunt(s.charID)
		if err == nil && hunt != nil {
			hunts = append(hunts, *hunt)
		}
	case 30:
		guildHunts, err := s.server.guildRepo.ListGuildHunts(guild.ID, s.charID)
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		for _, hunt := range guildHunts {
			if hunt.Start.Add(time.Second * time.Duration(s.server.erupeConfig.GameplayOptions.TreasureHuntExpiry)).After(TimeAdjusted()) {
				hunts = append(hunts, *hunt)
			}
		}
		if len(hunts) > 30 {
			hunts = hunts[:30]
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(hunts)))
	bf.WriteUint16(uint16(len(hunts)))
	for _, h := range hunts {
		bf.WriteUint32(h.HuntID)
		bf.WriteUint32(h.Destination)
		bf.WriteUint32(h.Level)
		bf.WriteUint32(h.Hunters)
		bf.WriteUint32(uint32(h.Start.Unix()))
		bf.WriteBool(h.Collected)
		bf.WriteBool(h.Claimed)
		bf.WriteBytes(h.HuntData)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildTresure)
	bf := byteframe.NewByteFrameFromBytes(pkt.Data)
	huntData := byteframe.NewByteFrame()
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	guildCats := getGuildAirouList(s)
	destination := bf.ReadUint32()
	level := bf.ReadUint32()
	huntData.WriteUint32(s.charID)
	huntData.WriteBytes(stringsupport.PaddedString(s.Name, 18, true))
	catsUsed := ""
	for i := 0; i < 5; i++ {
		catID := bf.ReadUint32()
		huntData.WriteUint32(catID)
		if catID > 0 {
			catsUsed = stringsupport.CSVAdd(catsUsed, int(catID))
			for _, cat := range guildCats {
				if cat.ID == catID {
					huntData.WriteBytes(cat.Name)
					break
				}
			}
			huntData.WriteBytes(bf.ReadBytes(9))
		}
	}
	if err := s.server.guildRepo.CreateHunt(guild.ID, s.charID, destination, level, huntData.Data(), catsUsed); err != nil {
		s.logger.Error("Failed to register guild treasure hunt", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresure)
	if err := s.server.guildRepo.AcquireHunt(pkt.HuntID); err != nil {
		s.logger.Error("Failed to acquire guild treasure hunt", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfOperateGuildTresureReport(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildTresureReport)
	switch pkt.State {
	case 0: // Report registration
		if err := s.server.guildRepo.RegisterHuntReport(pkt.HuntID, s.charID); err != nil {
			s.logger.Error("Failed to register treasure hunt report", zap.Error(err))
		}
	case 1: // Collected by hunter
		if err := s.server.guildRepo.CollectHunt(pkt.HuntID); err != nil {
			s.logger.Error("Failed to collect treasure hunt", zap.Error(err))
		}
	case 2: // Claim treasure
		if err := s.server.guildRepo.ClaimHuntReward(pkt.HuntID, s.charID); err != nil {
			s.logger.Error("Failed to claim treasure hunt reward", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// TreasureSouvenir represents a guild treasure souvenir entry.
type TreasureSouvenir struct {
	Destination uint32
	Quantity    uint32
}

func handleMsgMhfGetGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTresureSouvenir)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	souvenirs := []TreasureSouvenir{}
	bf.WriteUint16(uint16(len(souvenirs)))
	for _, souvenir := range souvenirs {
		bf.WriteUint32(souvenir.Destination)
		bf.WriteUint32(souvenir.Quantity)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresureSouvenir)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
