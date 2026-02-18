package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"math/rand"

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
	var fp, gp, gt uint32
	_ = s.server.db.QueryRow("SELECT COALESCE(frontier_points, 0), COALESCE(gacha_premium, 0), COALESCE(gacha_trial, 0) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)", s.charID).Scan(&fp, &gp, &gt)
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(gp)
	resp.WriteUint32(gt)
	resp.WriteUint32(fp)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUseGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseGachaPoint)
	if pkt.TrialCoins > 0 {
		if _, err := s.server.db.Exec(`UPDATE users u SET gacha_trial=gacha_trial-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, pkt.TrialCoins, s.charID); err != nil {
			s.logger.Error("Failed to deduct gacha trial coins", zap.Error(err))
		}
	}
	if pkt.PremiumCoins > 0 {
		if _, err := s.server.db.Exec(`UPDATE users u SET gacha_premium=gacha_premium-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, pkt.PremiumCoins, s.charID); err != nil {
			s.logger.Error("Failed to deduct gacha premium coins", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func spendGachaCoin(s *Session, quantity uint16) {
	var gt uint16
	_ = s.server.db.QueryRow(`SELECT COALESCE(gacha_trial, 0) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.charID).Scan(&gt)
	if quantity <= gt {
		if _, err := s.server.db.Exec(`UPDATE users u SET gacha_trial=gacha_trial-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, quantity, s.charID); err != nil {
			s.logger.Error("Failed to deduct gacha trial coins", zap.Error(err))
		}
	} else {
		if _, err := s.server.db.Exec(`UPDATE users u SET gacha_premium=gacha_premium-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, quantity, s.charID); err != nil {
			s.logger.Error("Failed to deduct gacha premium coins", zap.Error(err))
		}
	}
}

func transactGacha(s *Session, gachaID uint32, rollID uint8) (int, error) {
	var itemType uint8
	var itemNumber uint16
	var rolls int
	err := s.server.db.QueryRowx(`SELECT item_type, item_number, rolls FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2`, gachaID, rollID).Scan(&itemType, &itemNumber, &rolls)
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
		if _, err := s.server.db.Exec("UPDATE users u SET frontier_points=frontier_points-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", itemNumber, s.charID); err != nil {
			s.logger.Error("Failed to deduct frontier points for gacha", zap.Error(err))
		}
	}
	return rolls, nil
}

func getGuaranteedItems(s *Session, gachaID uint32, rollID uint8) []GachaItem {
	var rewards []GachaItem
	var reward GachaItem
	items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = (SELECT id FROM gacha_entries WHERE entry_type = $1 AND gacha_id = $2)`, rollID, gachaID)
	if err == nil {
		for items.Next() {
			_ = items.StructScan(&reward)
			rewards = append(rewards, reward)
		}
	}
	return rewards
}

func addGachaItem(s *Session, items []GachaItem) {
	var data []byte
	_ = s.server.db.QueryRow(`SELECT gacha_items FROM characters WHERE id = $1`, s.charID).Scan(&data)
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
	if _, err := s.server.db.Exec(`UPDATE characters SET gacha_items = $1 WHERE id = $2`, newItem.Data(), s.charID); err != nil {
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
	var data []byte
	err := s.server.db.QueryRow("SELECT COALESCE(gacha_items, $2) FROM characters WHERE id = $1", s.charID, []byte{0x00}).Scan(&data)
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
			if _, err := s.server.db.Exec("UPDATE characters SET gacha_items = $1 WHERE id = $2", update.Data(), s.charID); err != nil {
				s.logger.Error("Failed to update gacha items overflow", zap.Error(err))
			}
		} else {
			if _, err := s.server.db.Exec("UPDATE characters SET gacha_items = null WHERE id = $1", s.charID); err != nil {
				s.logger.Error("Failed to clear gacha items", zap.Error(err))
			}
		}
	}
}

func handleMsgMhfPlayNormalGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayNormalGacha)
	bf := byteframe.NewByteFrame()
	var entries []GachaEntry
	var entry GachaEntry
	var rewards []GachaItem
	var reward GachaItem
	rolls, err := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	rows, err := s.server.db.Queryx(`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`, pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for rows.Next() {
		err = rows.StructScan(&entry)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	rewardEntries, _ := getRandomEntries(entries, rolls, false)
	temp := byteframe.NewByteFrame()
	for i := range rewardEntries {
		rows, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`, rewardEntries[i].ID)
		if err != nil {
			continue
		}
		for rows.Next() {
			err = rows.StructScan(&reward)
			if err != nil {
				continue
			}
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
	var entries []GachaEntry
	var entry GachaEntry
	var rewards []GachaItem
	var reward GachaItem
	rolls, err := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	if _, err := s.server.db.Exec("UPDATE users u SET frontier_points=frontier_points+(SELECT frontier_points FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2) WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$3)", pkt.GachaID, pkt.RollType, s.charID); err != nil {
		s.logger.Error("Failed to award stepup gacha frontier points", zap.Error(err))
	}
	if _, err := s.server.db.Exec(`DELETE FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID); err != nil {
		s.logger.Error("Failed to delete gacha stepup state", zap.Error(err))
	}
	if _, err := s.server.db.Exec(`INSERT INTO gacha_stepup (gacha_id, step, character_id) VALUES ($1, $2, $3)`, pkt.GachaID, pkt.RollType+1, s.charID); err != nil {
		s.logger.Error("Failed to insert gacha stepup state", zap.Error(err))
	}

	rows, err := s.server.db.Queryx(`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`, pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for rows.Next() {
		err = rows.StructScan(&entry)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	guaranteedItems := getGuaranteedItems(s, pkt.GachaID, pkt.RollType)
	rewardEntries, _ := getRandomEntries(entries, rolls, false)
	temp := byteframe.NewByteFrame()
	for i := range rewardEntries {
		rows, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`, rewardEntries[i].ID)
		if err != nil {
			continue
		}
		for rows.Next() {
			err = rows.StructScan(&reward)
			if err != nil {
				continue
			}
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
	// TODO: Reset daily (noon)
	var step uint8
	_ = s.server.db.QueryRow(`SELECT step FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID).Scan(&step)
	var stepCheck int
	_ = s.server.db.QueryRow(`SELECT COUNT(1) FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2`, pkt.GachaID, step).Scan(&stepCheck)
	if stepCheck == 0 {
		if _, err := s.server.db.Exec(`DELETE FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID); err != nil {
			s.logger.Error("Failed to reset gacha stepup state", zap.Error(err))
		}
		step = 0
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(step)
	bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoxGachaInfo)
	entries, err := s.server.db.Queryx(`SELECT entry_id FROM gacha_box WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	var entryIDs []uint32
	for entries.Next() {
		var entryID uint32
		_ = entries.Scan(&entryID)
		entryIDs = append(entryIDs, entryID)
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
	var entries []GachaEntry
	var entry GachaEntry
	var rewards []GachaItem
	var reward GachaItem
	rolls, err := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	rows, err := s.server.db.Queryx(`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`, pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for rows.Next() {
		err = rows.StructScan(&entry)
		if err == nil {
			entries = append(entries, entry)
		}
	}
	rewardEntries, _ := getRandomEntries(entries, rolls, true)
	for i := range rewardEntries {
		items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`, rewardEntries[i].ID)
		if err != nil {
			continue
		}
		if _, err := s.server.db.Exec(`INSERT INTO gacha_box (gacha_id, entry_id, character_id) VALUES ($1, $2, $3)`, pkt.GachaID, rewardEntries[i].ID, s.charID); err != nil {
			s.logger.Error("Failed to insert gacha box entry", zap.Error(err))
		}
		for items.Next() {
			err = items.StructScan(&reward)
			if err == nil {
				rewards = append(rewards, reward)
			}
		}
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
	if _, err := s.server.db.Exec("DELETE FROM gacha_box WHERE gacha_id = $1 AND character_id = $2", pkt.GachaID, s.charID); err != nil {
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
