package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// Gacha represents a gacha lottery definition.
type Gacha struct {
	ID           uint32 `db:"id"`
	MinGR        uint32 `db:"min_gr"`
	MinHR        uint32 `db:"min_hr"`
	Name         string `db:"name"`
	URLBanner    string `db:"url_banner"`
	URLFeature   string `db:"url_feature"`
	URLThumbnail string `db:"url_thumbnail"`
	Wide         bool   `db:"wide"`
	Recommended  bool   `db:"recommended"`
	GachaType    uint8  `db:"gacha_type"`
	Hidden       bool   `db:"hidden"`
}

// GachaEntry represents a gacha entry (step/box).
type GachaEntry struct {
	EntryType      uint8   `db:"entry_type"`
	ID             uint32  `db:"id"`
	ItemType       uint8   `db:"item_type"`
	ItemNumber     uint32  `db:"item_number"`
	ItemQuantity   uint16  `db:"item_quantity"`
	Weight         float64 `db:"weight"`
	Rarity         uint8   `db:"rarity"`
	Rolls          uint8   `db:"rolls"`
	FrontierPoints uint16  `db:"frontier_points"`
	DailyLimit     uint8   `db:"daily_limit"`
	Name           string  `db:"name"`
}

// GachaItem represents a single item in a gacha pool.
type GachaItem struct {
	ItemType uint8  `db:"item_type"`
	ItemID   uint16 `db:"item_id"`
	Quantity uint16 `db:"quantity"`
}

func handleMsgMhfGetGachaPlayHistory(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGachaPlayHistory)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(1)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGachaPoint)
	fp, gp, gt, err := s.server.userRepo.GetGachaPoints(s.userID)
	if err != nil {
		s.logger.Error("Failed to get gacha points", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 12))
		return
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(gp)
	resp.WriteUint32(gt)
	resp.WriteUint32(fp)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUseGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseGachaPoint)
	if pkt.TrialCoins > 0 {
		if err := s.server.userRepo.DeductTrialCoins(s.userID, pkt.TrialCoins); err != nil {
			s.logger.Error("Failed to deduct gacha trial coins", zap.Error(err))
		}
	}
	if pkt.PremiumCoins > 0 {
		if err := s.server.userRepo.DeductPremiumCoins(s.userID, pkt.PremiumCoins); err != nil {
			s.logger.Error("Failed to deduct gacha premium coins", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfReceiveGachaItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReceiveGachaItem)
	data, err := s.server.charRepo.LoadColumnWithDefault(s.charID, "gacha_items", []byte{0x00})
	if err != nil || len(data) == 0 {
		// The client requires at least one byte (the item count) or it
		// treats the response as malformed and crashes (see #175).
		data = []byte{0x00}
	}

	// Handle overflow: the client can only display 36 items (36 * 5 + 1 count byte = 181 bytes).
	// If there are more, send the first 36 and keep the rest for next time.
	isOverflow := len(data) > 181 && data[0] > 36
	if isOverflow {
		resp := byteframe.NewByteFrame()
		resp.WriteUint8(36)
		resp.WriteBytes(data[1:181])
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	} else {
		doAckBufSucceed(s, pkt.AckHandle, data)
	}

	if !pkt.Freeze {
		if isOverflow {
			update := byteframe.NewByteFrame()
			update.WriteUint8(uint8(len(data[181:]) / 5))
			update.WriteBytes(data[181:])
			if err := s.server.charRepo.SaveColumn(s.charID, "gacha_items", update.Data()); err != nil {
				s.logger.Error("Failed to update gacha items overflow", zap.Error(err))
			}
		} else {
			if err := s.server.charRepo.SaveColumn(s.charID, "gacha_items", nil); err != nil {
				s.logger.Error("Failed to clear gacha items", zap.Error(err))
			}
		}
	}
}

func handleMsgMhfPlayNormalGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayNormalGacha)

	result, err := s.server.gachaService.PlayNormalGacha(s.userID, s.charID, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(result.Rewards)))
	for _, r := range result.Rewards {
		bf.WriteUint8(r.ItemType)
		bf.WriteUint16(r.ItemID)
		bf.WriteUint16(r.Quantity)
		bf.WriteUint8(r.Rarity)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfPlayStepupGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayStepupGacha)

	result, err := s.server.gachaService.PlayStepupGacha(s.userID, s.charID, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(result.RandomRewards) + len(result.GuaranteedRewards)))
	bf.WriteUint8(uint8(len(result.RandomRewards)))
	for _, item := range result.GuaranteedRewards {
		bf.WriteUint8(item.ItemType)
		bf.WriteUint16(item.ItemID)
		bf.WriteUint16(item.Quantity)
		bf.WriteUint8(item.Rarity)
	}
	for _, r := range result.RandomRewards {
		bf.WriteUint8(r.ItemType)
		bf.WriteUint16(r.ItemID)
		bf.WriteUint16(r.Quantity)
		bf.WriteUint8(r.Rarity)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetStepupStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetStepupStatus)

	status, err := s.server.gachaService.GetStepupStatus(pkt.GachaID, s.charID, TimeAdjusted())
	if err != nil {
		s.logger.Error("Failed to get stepup status", zap.Error(err))
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(status.Step)
	bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoxGachaInfo)

	entryIDs, err := s.server.gachaService.GetBoxInfo(pkt.GachaID, s.charID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(entryIDs)))
	for i := range entryIDs {
		bf.WriteUint32(entryIDs[i])
		bf.WriteBool(true)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfPlayBoxGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayBoxGacha)

	result, err := s.server.gachaService.PlayBoxGacha(s.userID, s.charID, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(result.Rewards)))
	for _, r := range result.Rewards {
		bf.WriteUint8(r.ItemType)
		bf.WriteUint16(r.ItemID)
		bf.WriteUint16(r.Quantity)
		bf.WriteUint8(r.Rarity)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfResetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfResetBoxGachaInfo)
	if err := s.server.gachaService.ResetBox(pkt.GachaID, s.charID); err != nil {
		s.logger.Error("Failed to reset gacha box", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfPlayFreeGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayFreeGacha)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}
