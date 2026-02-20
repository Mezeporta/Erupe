package channelserver

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

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
	fp, gp, gt, _ := s.server.userRepo.GetGachaPoints(s.userID)
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

func spendGachaCoin(s *Session, quantity uint16) {
	gt, _ := s.server.userRepo.GetTrialCoins(s.userID)
	if quantity <= gt {
		if err := s.server.userRepo.DeductTrialCoins(s.userID, uint32(quantity)); err != nil {
			s.logger.Error("Failed to deduct gacha trial coins", zap.Error(err))
		}
	} else {
		if err := s.server.userRepo.DeductPremiumCoins(s.userID, uint32(quantity)); err != nil {
			s.logger.Error("Failed to deduct gacha premium coins", zap.Error(err))
		}
	}
}

func transactGacha(s *Session, gachaID uint32, rollID uint8) (int, error) {
	itemType, itemNumber, rolls, err := s.server.gachaRepo.GetEntryForTransaction(gachaID, rollID)
	if err != nil {
		return 0, err
	}
	switch itemType {
	/*
		valid types that need manual savedata manipulation:
		- Ryoudan Points
		- Bond Points
		- Image Change Points
		valid types that work (no additional code needed):
		- Tore Points
		- Festa Points
	*/
	case 17:
		_ = addPointNetcafe(s, int(itemNumber)*-1)
	case 19:
		fallthrough
	case 20:
		spendGachaCoin(s, itemNumber)
	case 21:
		if err := s.server.userRepo.DeductFrontierPoints(s.userID, uint32(itemNumber)); err != nil {
			s.logger.Error("Failed to deduct frontier points for gacha", zap.Error(err))
		}
	}
	return rolls, nil
}

func getGuaranteedItems(s *Session, gachaID uint32, rollID uint8) []GachaItem {
	rewards, _ := s.server.gachaRepo.GetGuaranteedItems(rollID, gachaID)
	return rewards
}

func addGachaItem(s *Session, items []GachaItem) {
	data, _ := s.server.charRepo.LoadColumn(s.charID, "gacha_items")
	if len(data) > 0 {
		numItems := int(data[0])
		data = data[1:]
		oldItem := byteframe.NewByteFrameFromBytes(data)
		for i := 0; i < numItems; i++ {
			items = append(items, GachaItem{
				ItemType: oldItem.ReadUint8(),
				ItemID:   oldItem.ReadUint16(),
				Quantity: oldItem.ReadUint16(),
			})
		}
	}
	newItem := byteframe.NewByteFrame()
	newItem.WriteUint8(uint8(len(items)))
	for i := range items {
		newItem.WriteUint8(items[i].ItemType)
		newItem.WriteUint16(items[i].ItemID)
		newItem.WriteUint16(items[i].Quantity)
	}
	if err := s.server.charRepo.SaveColumn(s.charID, "gacha_items", newItem.Data()); err != nil {
		s.logger.Error("Failed to update gacha items", zap.Error(err))
	}
}

func getRandomEntries(entries []GachaEntry, rolls int, isBox bool) ([]GachaEntry, error) {
	var chosen []GachaEntry
	var totalWeight float64
	for i := range entries {
		totalWeight += entries[i].Weight
	}
	for rolls != len(chosen) {

		if !isBox {
			result := rand.Float64() * totalWeight
			for _, entry := range entries {
				result -= entry.Weight
				if result < 0 {
					chosen = append(chosen, entry)
					break
				}
			}
		} else {
			result := rand.Intn(len(entries))
			chosen = append(chosen, entries[result])
			entries[result] = entries[len(entries)-1]
			entries = entries[:len(entries)-1]
		}
	}
	return chosen, nil
}

func handleMsgMhfReceiveGachaItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReceiveGachaItem)
	data, err := s.server.charRepo.LoadColumnWithDefault(s.charID, "gacha_items", []byte{0x00})
	if err != nil {
		data = []byte{0x00}
	}

	// I think there are still some edge cases where rewards can be nulled via overflow
	if data[0] > 36 || len(data) > 181 {
		resp := byteframe.NewByteFrame()
		resp.WriteUint8(36)
		resp.WriteBytes(data[1:181])
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	} else {
		doAckBufSucceed(s, pkt.AckHandle, data)
	}

	if !pkt.Freeze {
		if data[0] > 36 || len(data) > 181 {
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
	bf := byteframe.NewByteFrame()
	var rewards []GachaItem
	rolls, err := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	entries, err := s.server.gachaRepo.GetRewardPool(pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	rewardEntries, _ := getRandomEntries(entries, rolls, false)
	temp := byteframe.NewByteFrame()
	for i := range rewardEntries {
		entryItems, err := s.server.gachaRepo.GetItemsForEntry(rewardEntries[i].ID)
		if err != nil {
			continue
		}
		for _, reward := range entryItems {
			rewards = append(rewards, reward)
			temp.WriteUint8(reward.ItemType)
			temp.WriteUint16(reward.ItemID)
			temp.WriteUint16(reward.Quantity)
			temp.WriteUint8(rewardEntries[i].Rarity)
		}
	}

	bf.WriteUint8(uint8(len(rewards)))
	bf.WriteBytes(temp.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	addGachaItem(s, rewards)
}

func handleMsgMhfPlayStepupGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayStepupGacha)
	bf := byteframe.NewByteFrame()
	var rewards []GachaItem
	rolls, err := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	if err := s.server.userRepo.AddFrontierPointsFromGacha(s.userID, pkt.GachaID, pkt.RollType); err != nil {
		s.logger.Error("Failed to award stepup gacha frontier points", zap.Error(err))
	}
	if err := s.server.gachaRepo.DeleteStepup(pkt.GachaID, s.charID); err != nil {
		s.logger.Error("Failed to delete gacha stepup state", zap.Error(err))
	}
	if err := s.server.gachaRepo.InsertStepup(pkt.GachaID, pkt.RollType+1, s.charID); err != nil {
		s.logger.Error("Failed to insert gacha stepup state", zap.Error(err))
	}

	entries, err := s.server.gachaRepo.GetRewardPool(pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	guaranteedItems := getGuaranteedItems(s, pkt.GachaID, pkt.RollType)
	rewardEntries, _ := getRandomEntries(entries, rolls, false)
	temp := byteframe.NewByteFrame()
	for i := range rewardEntries {
		entryItems, err := s.server.gachaRepo.GetItemsForEntry(rewardEntries[i].ID)
		if err != nil {
			continue
		}
		for _, reward := range entryItems {
			rewards = append(rewards, reward)
			temp.WriteUint8(reward.ItemType)
			temp.WriteUint16(reward.ItemID)
			temp.WriteUint16(reward.Quantity)
			temp.WriteUint8(rewardEntries[i].Rarity)
		}
	}

	bf.WriteUint8(uint8(len(rewards) + len(guaranteedItems)))
	bf.WriteUint8(uint8(len(rewards)))
	for _, item := range guaranteedItems {
		bf.WriteUint8(item.ItemType)
		bf.WriteUint16(item.ItemID)
		bf.WriteUint16(item.Quantity)
		bf.WriteUint8(0)
	}
	bf.WriteBytes(temp.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	addGachaItem(s, rewards)
	addGachaItem(s, guaranteedItems)
}

func handleMsgMhfGetStepupStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetStepupStatus)

	// Compute the most recent noon boundary
	midday := TimeMidnight().Add(12 * time.Hour)
	if TimeAdjusted().Before(midday) {
		midday = midday.Add(-24 * time.Hour)
	}

	step, createdAt, err := s.server.gachaRepo.GetStepupWithTime(pkt.GachaID, s.charID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("Failed to get gacha stepup state", zap.Error(err))
	}
	// Reset stale stepup progress (created before the most recent noon)
	if err == nil && createdAt.Before(midday) {
		if err := s.server.gachaRepo.DeleteStepup(pkt.GachaID, s.charID); err != nil {
			s.logger.Error("Failed to reset stale gacha stepup", zap.Error(err))
		}
		step = 0
	} else if err == nil {
		// Only check for valid entry type if the stepup is fresh
		hasEntry, _ := s.server.gachaRepo.HasEntryType(pkt.GachaID, step)
		if !hasEntry {
			if err := s.server.gachaRepo.DeleteStepup(pkt.GachaID, s.charID); err != nil {
				s.logger.Error("Failed to reset gacha stepup state", zap.Error(err))
			}
			step = 0
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(step)
	bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoxGachaInfo)
	entryIDs, err := s.server.gachaRepo.GetBoxEntryIDs(pkt.GachaID, s.charID)
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
	bf := byteframe.NewByteFrame()
	var rewards []GachaItem
	rolls, err := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	entries, err := s.server.gachaRepo.GetRewardPool(pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	rewardEntries, _ := getRandomEntries(entries, rolls, true)
	for i := range rewardEntries {
		entryItems, err := s.server.gachaRepo.GetItemsForEntry(rewardEntries[i].ID)
		if err != nil {
			continue
		}
		if err := s.server.gachaRepo.InsertBoxEntry(pkt.GachaID, rewardEntries[i].ID, s.charID); err != nil {
			s.logger.Error("Failed to insert gacha box entry", zap.Error(err))
		}
		rewards = append(rewards, entryItems...)
	}
	bf.WriteUint8(uint8(len(rewards)))
	for _, r := range rewards {
		bf.WriteUint8(r.ItemType)
		bf.WriteUint16(r.ItemID)
		bf.WriteUint16(r.Quantity)
		bf.WriteUint8(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	addGachaItem(s, rewards)
}

func handleMsgMhfResetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfResetBoxGachaInfo)
	if err := s.server.gachaRepo.DeleteBoxEntries(pkt.GachaID, s.charID); err != nil {
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
