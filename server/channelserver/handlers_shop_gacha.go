package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"math/rand"
)

type ShopItem struct {
	ID           uint32 `db:"id"`
	ItemID       uint16 `db:"item_id"`
	Cost         uint32 `db:"cost"`
	Quantity     uint16 `db:"quantity"`
	MinHR        uint16 `db:"min_hr"`
	MinSR        uint16 `db:"min_sr"`
	MinGR        uint16 `db:"min_gr"`
	StoreLevel   uint16 `db:"store_level"`
	MaxQuantity  uint16 `db:"max_quantity"`
	UsedQuantity uint16 `db:"used_quantity"`
	RoadFloors   uint16 `db:"road_floors"`
	RoadFatalis  uint16 `db:"road_fatalis"`
}

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

type GachaEntry struct {
	EntryType      uint8   `db:"entry_type"`
	ID             uint32  `db:"id"`
	ItemType       uint8   `db:"item_type"`
	ItemNumber     uint16  `db:"item_number"`
	ItemQuantity   uint16  `db:"item_quantity"`
	Weight         float64 `db:"weight"`
	Rarity         uint8   `db:"rarity"`
	Rolls          uint8   `db:"rolls"`
	FrontierPoints uint16  `db:"frontier_points"`
	DailyLimit     uint8   `db:"daily_limit"`
}

type GachaItem struct {
	ItemType uint8  `db:"item_type"`
	ItemID   uint16 `db:"item_id"`
	Quantity uint16 `db:"quantity"`
}

func writeShopItems(bf *byteframe.ByteFrame, items []ShopItem) {
	bf.WriteUint16(uint16(len(items)))
	bf.WriteUint16(uint16(len(items)))
	for _, item := range items {
		bf.WriteUint32(item.ID)
		bf.WriteUint16(0)
		bf.WriteUint16(item.ItemID)
		bf.WriteUint32(item.Cost)
		bf.WriteUint16(item.Quantity)
		bf.WriteUint16(item.MinHR)
		bf.WriteUint16(item.MinSR)
		bf.WriteUint16(item.MinGR)
		bf.WriteUint16(item.StoreLevel)
		bf.WriteUint16(item.MaxQuantity)
		bf.WriteUint16(item.UsedQuantity)
		bf.WriteUint16(item.RoadFloors)
		bf.WriteUint16(item.RoadFatalis)
	}
}

func getShopItems(s *Session, shopType uint8, shopID uint32) []ShopItem {
	var items []ShopItem
	var temp ShopItem
	rows, err := s.server.db.Queryx(`SELECT id, item_id, cost, quantity, min_hr, min_sr, min_gr, store_level, max_quantity,
       		COALESCE((SELECT bought FROM shop_items_bought WHERE shop_item_id=si.id AND character_id=$3), 0) as used_quantity,
       		road_floors, road_fatalis FROM shop_items si WHERE shop_type=$1 AND shop_id=$2
       		`, shopType, shopID, s.charID)
	if err == nil {
		for rows.Next() {
			err = rows.StructScan(&temp)
			if err != nil {
				continue
			}
			items = append(items, temp)
		}
	}
	return items
}

func handleMsgMhfEnumerateShop(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateShop)
	// Generic Shop IDs
	// 0: basic item
	// 1: gatherables
	// 2: hr1-4 materials
	// 3: hr5-7 materials
	// 4: decos
	// 5: other item
	// 6: g mats
	// 7: limited item
	// 8: special item
	switch pkt.ShopType {
	case 1: // Running gachas
		var count uint16
		shopEntries, err := s.server.db.Queryx("SELECT id, min_gr, min_hr, name, url_banner, url_feature, url_thumbnail, wide, recommended, gacha_type, hidden FROM gacha_shop")
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0)
		var gacha Gacha
		for shopEntries.Next() {
			err = shopEntries.StructScan(&gacha)
			if err != nil {
				continue
			}
			resp.WriteUint32(gacha.ID)
			resp.WriteBytes(make([]byte, 16)) // Rank restriction
			resp.WriteUint32(gacha.MinGR)
			resp.WriteUint32(gacha.MinHR)
			resp.WriteUint32(0) // only 0 in known packet
			ps.Uint8(resp, gacha.Name, true)
			ps.Uint8(resp, gacha.URLBanner, false)
			ps.Uint8(resp, gacha.URLFeature, false)
			resp.WriteBool(gacha.Wide)
			ps.Uint8(resp, gacha.URLThumbnail, false)
			resp.WriteUint8(0) // Unk
			if gacha.Recommended {
				resp.WriteUint8(2)
			} else {
				resp.WriteUint8(0)
			}
			resp.WriteUint8(gacha.GachaType)
			resp.WriteBool(gacha.Hidden)
			count++
		}
		resp.Seek(0, 0)
		resp.WriteUint16(count)
		resp.WriteUint16(count)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	case 2: // Actual gacha
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(pkt.ShopID)
		var gachaType int
		s.server.db.QueryRow(`SELECT gacha_type FROM gacha_shop WHERE id = $1`, pkt.ShopID).Scan(&gachaType)
		entries, err := s.server.db.Queryx(`SELECT entry_type, id, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points FROM gacha_entries WHERE gacha_id = $1 ORDER BY weight DESC`, pkt.ShopID)
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		var divisor float64
		s.server.db.QueryRow(`SELECT COALESCE(SUM(weight) / 100000.0, 0) AS chance FROM gacha_entries WHERE gacha_id = $1`, pkt.ShopID).Scan(&divisor)
		var entryCount uint16
		bf.WriteUint16(0)
		gachaEntry := GachaEntry{}
		gachaItem := GachaItem{}
		for entries.Next() {
			entryCount++
			entries.StructScan(&gachaEntry)
			bf.WriteUint8(gachaEntry.EntryType)
			bf.WriteUint32(gachaEntry.ID)
			bf.WriteUint8(gachaEntry.ItemType)
			bf.WriteUint16(0)
			bf.WriteUint16(gachaEntry.ItemNumber)
			bf.WriteUint16(gachaEntry.ItemQuantity)
			if gachaType >= 4 { // If box
				bf.WriteUint16(1)
			} else {
				bf.WriteUint16(uint16(gachaEntry.Weight / divisor))
			}
			bf.WriteUint8(gachaEntry.Rarity)
			bf.WriteUint8(gachaEntry.Rolls)

			var itemCount uint8
			temp := byteframe.NewByteFrame()
			items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id=$1`, gachaEntry.ID)
			if err != nil {
				bf.WriteUint8(0)
			} else {
				for items.Next() {
					itemCount++
					items.StructScan(&gachaItem)
					temp.WriteUint16(uint16(gachaItem.ItemType))
					temp.WriteUint16(gachaItem.ItemID)
					temp.WriteUint16(gachaItem.Quantity)
				}
				bf.WriteUint8(itemCount)
			}

			bf.WriteUint16(gachaEntry.FrontierPoints)
			bf.WriteUint8(gachaEntry.DailyLimit)
			bf.WriteUint8(0)
			bf.WriteBytes(temp.Data())
		}
		bf.Seek(4, 0)
		bf.WriteUint16(entryCount)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	case 3: // Hunting Festival Exchange
		fallthrough
	case 4: // N Points, 0-6
		fallthrough
	case 5: // GCP->Item, 0-6
		fallthrough
	case 6: // Gacha coin->Item
		fallthrough
	case 7: // Item->GCP
		fallthrough
	case 8: // Diva
		fallthrough
	case 9: // Diva song shop
		fallthrough
	case 10: // Item shop, 0-8
		bf := byteframe.NewByteFrame()
		items := getShopItems(s, pkt.ShopType, pkt.ShopID)
		writeShopItems(bf, items)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
}

func handleMsgMhfAcquireExchangeShop(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireExchangeShop)
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	exchanges := int(bf.ReadUint16())
	for i := 0; i < exchanges; i++ {
		itemHash := bf.ReadUint32()
		if itemHash == 0 {
			continue
		}
		buyCount := bf.ReadUint32()
		s.server.db.Exec(`INSERT INTO shop_items_bought (character_id, shop_item_id, bought)
			VALUES ($1,$2,$3) ON CONFLICT (character_id, shop_item_id)
			DO UPDATE SET bought = bought + $3
			WHERE EXCLUDED.character_id=$1 AND EXCLUDED.shop_item_id=$2
		`, s.charID, itemHash, buyCount)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
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
	s.server.db.QueryRow("SELECT COALESCE(frontier_points, 0), COALESCE(gacha_premium, 0), COALESCE(gacha_trial, 0) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)", s.charID).Scan(&fp, &gp, &gt)
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(gp)
	resp.WriteUint32(gt)
	resp.WriteUint32(fp)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUseGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseGachaPoint)
	if pkt.TrialCoins > 0 {
		s.server.db.Exec(`UPDATE users u SET gacha_trial=gacha_trial-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, pkt.TrialCoins, s.charID)
	}
	if pkt.PremiumCoins > 0 {
		s.server.db.Exec(`UPDATE users u SET gacha_premium=gacha_premium-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, pkt.PremiumCoins, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func spendGachaCoin(s *Session, quantity uint16) {
	var gt uint16
	s.server.db.QueryRow(`SELECT COALESCE(gacha_trial, 0) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.charID).Scan(&gt)
	if quantity <= gt {
		s.server.db.Exec(`UPDATE users u SET gacha_trial=gacha_trial-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, quantity, s.charID)
	} else {
		s.server.db.Exec(`UPDATE users u SET gacha_premium=gacha_premium-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, quantity, s.charID)
	}
}

func transactGacha(s *Session, gachaID uint32, rollID uint8) (error, int) {
	var itemType uint8
	var itemNumber uint16
	var rolls int
	err := s.server.db.QueryRowx(`SELECT item_type, item_number, rolls FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2`, gachaID, rollID).Scan(&itemType, &itemNumber, &rolls)
	if err != nil {
		return err, 0
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
		s.server.db.Exec("UPDATE users u SET frontier_points=frontier_points-$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", itemNumber, s.charID)
	}
	return nil, rolls
}

func getGuaranteedItems(s *Session, gachaID uint32, rollID uint8) []GachaItem {
	var rewards []GachaItem
	var reward GachaItem
	items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = (SELECT id FROM gacha_entries WHERE entry_type = $1 AND gacha_id = $2)`, rollID, gachaID)
	if err == nil {
		for items.Next() {
			items.StructScan(&reward)
			rewards = append(rewards, reward)
		}
	}
	return rewards
}

func addGachaItem(s *Session, items []GachaItem) {
	var data []byte
	s.server.db.QueryRow(`SELECT gacha_items FROM characters WHERE id = $1`, s.charID).Scan(&data)
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
	s.server.db.Exec(`UPDATE characters SET gacha_items = $1 WHERE id = $2`, newItem.Data(), s.charID)
}

func getRandomEntries(entries []GachaEntry, rolls int, isBox bool) ([]GachaEntry, error) {
	var chosen []GachaEntry
	if len(entries) == 0 || rolls <= 0 {
		return chosen, nil
	}
	var totalWeight float64
	for i := range entries {
		totalWeight += entries[i].Weight
	}
	if !isBox && totalWeight <= 0 {
		return chosen, nil
	}
	for {
		if rolls == len(chosen) {
			break
		}
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
			if len(entries) == 0 {
				break
			}
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
			s.server.db.Exec("UPDATE characters SET gacha_items = $1 WHERE id = $2", update.Data(), s.charID)
		} else {
			s.server.db.Exec("UPDATE characters SET gacha_items = null WHERE id = $1", s.charID)
		}
	}
}

func handleMsgMhfPlayNormalGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayNormalGacha)
	bf := byteframe.NewByteFrame()
	var gachaEntries []GachaEntry
	var entry GachaEntry
	var rewards []GachaItem
	var reward GachaItem
	err, rolls := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	temp := byteframe.NewByteFrame()
	entries, err := s.server.db.Queryx(`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`, pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for entries.Next() {
		entries.StructScan(&entry)
		gachaEntries = append(gachaEntries, entry)
	}
	rewardEntries, err := getRandomEntries(gachaEntries, rolls, false)
	for i := range rewardEntries {
		items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`, rewardEntries[i].ID)
		if err != nil {
			continue
		}
		for items.Next() {
			items.StructScan(&reward)
			rewards = append(rewards, reward)
			temp.WriteUint8(reward.ItemType)
			temp.WriteUint16(reward.ItemID)
			temp.WriteUint16(reward.Quantity)
			temp.WriteUint8(entry.Rarity)
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
	var gachaEntries []GachaEntry
	var entry GachaEntry
	var rewards []GachaItem
	var reward GachaItem
	err, rolls := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	s.server.db.Exec("UPDATE users u SET frontier_points=frontier_points+(SELECT frontier_points FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2) WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$3)", pkt.GachaID, pkt.RollType, s.charID)
	s.server.db.Exec(`DELETE FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID)
	s.server.db.Exec(`INSERT INTO gacha_stepup (gacha_id, step, character_id) VALUES ($1, $2, $3)`, pkt.GachaID, pkt.RollType+1, s.charID)
	temp := byteframe.NewByteFrame()
	guaranteedItems := getGuaranteedItems(s, pkt.GachaID, pkt.RollType)
	for _, item := range guaranteedItems {
		temp.WriteUint8(item.ItemType)
		temp.WriteUint16(item.ItemID)
		temp.WriteUint16(item.Quantity)
		temp.WriteUint8(0)
	}
	entries, err := s.server.db.Queryx(`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`, pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for entries.Next() {
		entries.StructScan(&entry)
		gachaEntries = append(gachaEntries, entry)
	}
	rewardEntries, err := getRandomEntries(gachaEntries, rolls, false)
	for i := range rewardEntries {
		items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`, rewardEntries[i].ID)
		if err != nil {
			continue
		}
		for items.Next() {
			items.StructScan(&reward)
			rewards = append(rewards, reward)
			temp.WriteUint8(reward.ItemType)
			temp.WriteUint16(reward.ItemID)
			temp.WriteUint16(reward.Quantity)
			temp.WriteUint8(entry.Rarity)
		}
	}
	bf.WriteUint8(uint8(len(rewards) + len(guaranteedItems)))
	bf.WriteUint8(uint8(len(rewards)))
	bf.WriteBytes(temp.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	addGachaItem(s, rewards)
	addGachaItem(s, guaranteedItems)
}

func handleMsgMhfGetStepupStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetStepupStatus)
	// TODO: Reset daily (noon)
	var step uint8
	s.server.db.QueryRow(`SELECT step FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID).Scan(&step)
	var stepCheck int
	s.server.db.QueryRow(`SELECT COUNT(1) FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2`, pkt.GachaID, step).Scan(&stepCheck)
	if stepCheck == 0 {
		s.server.db.Exec(`DELETE FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`, pkt.GachaID, s.charID)
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
		entries.Scan(&entryID)
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
	var gachaEntries []GachaEntry
	var entry GachaEntry
	var rewards []GachaItem
	var reward GachaItem
	err, rolls := transactGacha(s, pkt.GachaID, pkt.RollType)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	temp := byteframe.NewByteFrame()
	entries, err := s.server.db.Queryx(`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`, pkt.GachaID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for entries.Next() {
		entries.StructScan(&entry)
		gachaEntries = append(gachaEntries, entry)
	}
	rewardEntries, err := getRandomEntries(gachaEntries, rolls, true)
	for i := range rewardEntries {
		items, err := s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`, rewardEntries[i].ID)
		if err != nil {
			continue
		}
		s.server.db.Exec(`INSERT INTO gacha_box (gacha_id, entry_id, character_id) VALUES ($1, $2, $3)`, pkt.GachaID, rewardEntries[i].ID, s.charID)
		for items.Next() {
			items.StructScan(&reward)
			rewards = append(rewards, reward)
			temp.WriteUint8(reward.ItemType)
			temp.WriteUint16(reward.ItemID)
			temp.WriteUint16(reward.Quantity)
			temp.WriteUint8(0)
		}
	}
	bf.WriteUint8(uint8(len(rewards)))
	bf.WriteBytes(temp.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	addGachaItem(s, rewards)
}

func handleMsgMhfResetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfResetBoxGachaInfo)
	s.server.db.Exec("DELETE FROM gacha_box WHERE gacha_id = $1 AND character_id = $2", pkt.GachaID, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfExchangeFpoint2Item(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeFpoint2Item)
	var balance uint32
	var itemValue, quantity int
	s.server.db.QueryRow("SELECT quantity, fpoints FROM fpoint_items WHERE id=$1", pkt.TradeID).Scan(&quantity, &itemValue)
	cost := (int(pkt.Quantity) * quantity) * itemValue
	s.server.db.QueryRow("UPDATE users u SET frontier_points=frontier_points::int - $1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2) RETURNING frontier_points", cost, s.charID).Scan(&balance)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(balance)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfExchangeItem2Fpoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeItem2Fpoint)
	var balance uint32
	var itemValue, quantity int
	s.server.db.QueryRow("SELECT quantity, fpoints FROM fpoint_items WHERE id=$1", pkt.TradeID).Scan(&quantity, &itemValue)
	cost := (int(pkt.Quantity) / quantity) * itemValue
	s.server.db.QueryRow("UPDATE users u SET frontier_points=COALESCE(frontier_points::int + $1, $1) WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2) RETURNING frontier_points", cost, s.charID).Scan(&balance)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(balance)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetFpointExchangeList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFpointExchangeList)
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	var buyables, sellables uint16
	var id uint32
	var itemType uint8
	var itemID, quantity, fPoints uint16

	buyRows, err := s.server.db.Query("SELECT id,item_type,item_id,quantity,fpoints FROM fpoint_items WHERE trade_type=0")
	if err == nil {
		for buyRows.Next() {
			err = buyRows.Scan(&id, &itemType, &itemID, &quantity, &fPoints)
			if err != nil {
				continue
			}
			resp.WriteUint32(id)
			resp.WriteUint16(0)
			resp.WriteUint16(0)
			resp.WriteUint16(0)
			resp.WriteUint8(itemType)
			resp.WriteUint16(itemID)
			resp.WriteUint16(quantity)
			resp.WriteUint16(fPoints)
			buyables++
		}
	}

	sellRows, err := s.server.db.Query("SELECT id,item_type,item_id,quantity,fpoints FROM fpoint_items WHERE trade_type=1")
	if err == nil {
		for sellRows.Next() {
			err = sellRows.Scan(&id, &itemType, &itemID, &quantity, &fPoints)
			if err != nil {
				continue
			}
			resp.WriteUint32(id)
			resp.WriteUint16(0)
			resp.WriteUint16(0)
			resp.WriteUint16(0)
			resp.WriteUint8(itemType)
			resp.WriteUint16(itemID)
			resp.WriteUint16(quantity)
			resp.WriteUint16(fPoints)
			sellables++
		}
	}
	resp.Seek(0, 0)
	resp.WriteUint16(buyables)
	resp.WriteUint16(sellables)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfPlayFreeGacha(s *Session, p mhfpacket.MHFPacket) {
	// not sure this is used anywhere, free gachas use the MSG_MHF_PLAY_NORMAL_GACHA method in captures
}
