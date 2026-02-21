package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"time"

	"go.uber.org/zap"
)

// Distribution represents an item distribution event.
type Distribution struct {
	ID              uint32    `db:"id"`
	Deadline        time.Time `db:"deadline"`
	Rights          uint32    `db:"rights"`
	TimesAcceptable uint16    `db:"times_acceptable"`
	TimesAccepted   uint16    `db:"times_accepted"`
	MinHR           int16     `db:"min_hr"`
	MaxHR           int16     `db:"max_hr"`
	MinSR           int16     `db:"min_sr"`
	MaxSR           int16     `db:"max_sr"`
	MinGR           int16     `db:"min_gr"`
	MaxGR           int16     `db:"max_gr"`
	EventName       string    `db:"event_name"`
	Description     string    `db:"description"`
	Selection       bool      `db:"selection"`
}

func handleMsgMhfEnumerateDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateDistItem)

	bf := byteframe.NewByteFrame()
	itemDists, _ := s.server.distRepo.List(s.charID, pkt.DistType)

	bf.WriteUint16(uint16(len(itemDists)))
	for _, dist := range itemDists {
		bf.WriteUint32(dist.ID)
		bf.WriteUint32(uint32(dist.Deadline.Unix()))
		bf.WriteUint32(dist.Rights)
		bf.WriteUint16(dist.TimesAcceptable)
		bf.WriteUint16(dist.TimesAccepted)
		if s.server.erupeConfig.RealClientMode >= cfg.G9 {
			bf.WriteUint16(0) // Unk
		}
		bf.WriteInt16(dist.MinHR)
		bf.WriteInt16(dist.MaxHR)
		bf.WriteInt16(dist.MinSR)
		bf.WriteInt16(dist.MaxSR)
		bf.WriteInt16(dist.MinGR)
		bf.WriteInt16(dist.MaxGR)
		if s.server.erupeConfig.RealClientMode >= cfg.G7 {
			bf.WriteUint8(0) // Unk
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G6 {
			bf.WriteUint16(0) // Unk
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G8 {
			if dist.Selection {
				bf.WriteUint8(2) // Selection
			} else {
				bf.WriteUint8(0)
			}
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G7 {
			bf.WriteUint16(0) // Unk
			bf.WriteUint16(0) // Unk
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G10 {
			bf.WriteUint8(0) // Unk
		}
		ps.Uint8(bf, dist.EventName, true)
		k := 6
		if s.server.erupeConfig.RealClientMode >= cfg.G8 {
			k = 13
		}
		for i := 0; i < 6; i++ {
			for j := 0; j < k; j++ {
				bf.WriteUint8(0)
				bf.WriteUint32(0)
			}
		}
		if s.server.erupeConfig.RealClientMode >= cfg.Z2 {
			i := uint8(0)
			bf.WriteUint8(i)
			if i <= 10 {
				for j := uint8(0); j < i; j++ {
					bf.WriteUint32(0)
					bf.WriteUint32(0)
					bf.WriteUint32(0)
				}
			}
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// DistributionItem represents a single item in a distribution.
type DistributionItem struct {
	ItemType uint8  `db:"item_type"`
	ID       uint32 `db:"id"`
	ItemID   uint32 `db:"item_id"`
	Quantity uint32 `db:"quantity"`
}

func handleMsgMhfApplyDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyDistItem)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(pkt.DistributionID)
	distItems, _ := s.server.distRepo.GetItems(pkt.DistributionID)
	bf.WriteUint16(uint16(len(distItems)))
	for _, item := range distItems {
		bf.WriteUint8(item.ItemType)
		bf.WriteUint32(item.ItemID)
		bf.WriteUint32(item.Quantity)
		if s.server.erupeConfig.RealClientMode >= cfg.G8 {
			bf.WriteUint32(item.ID)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireDistItem)
	if pkt.DistributionID > 0 {
		err := s.server.distRepo.RecordAccepted(pkt.DistributionID, s.charID)
		if err == nil {
			distItems, _ := s.server.distRepo.GetItems(pkt.DistributionID)
			for _, item := range distItems {
				switch item.ItemType {
				case 17:
					_ = addPointNetcafe(s, int(item.Quantity))
				case 19:
					if err := s.server.userRepo.AddPremiumCoins(s.userID, item.Quantity); err != nil {
						s.logger.Error("Failed to update gacha premium", zap.Error(err))
					}
				case 20:
					if err := s.server.userRepo.AddTrialCoins(s.userID, item.Quantity); err != nil {
						s.logger.Error("Failed to update gacha trial", zap.Error(err))
					}
				case 21:
					if err := s.server.userRepo.AddFrontierPoints(s.userID, item.Quantity); err != nil {
						s.logger.Error("Failed to update frontier points", zap.Error(err))
					}
				case 23:
					saveData, err := GetCharacterSaveData(s, s.charID)
					if err == nil {
						saveData.RP += uint16(item.Quantity)
						saveData.Save(s)
					}
				}
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetDistDescription(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetDistDescription)
	desc, err := s.server.distRepo.GetDescription(pkt.DistributionID)
	if err != nil {
		s.logger.Error("Error parsing item distribution description", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrame()
	ps.Uint16(bf, desc, true)
	ps.Uint16(bf, "", false)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
