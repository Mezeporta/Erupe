package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// ShopItem represents a shop item listing.
type ShopItem struct {
	ID           uint32 `db:"id"`
	ItemID       uint32 `db:"item_id"`
	Cost         uint32 `db:"cost"`
	Quantity     uint16 `db:"quantity"`
	MinHR        uint16 `db:"min_hr"`
	MinSR        uint16 `db:"min_sr"`
	MinGR        uint16 `db:"min_gr"`
	StoreLevel   uint8  `db:"store_level"`
	MaxQuantity  uint16 `db:"max_quantity"`
	UsedQuantity uint16 `db:"used_quantity"`
	RoadFloors   uint16 `db:"road_floors"`
	RoadFatalis  uint16 `db:"road_fatalis"`
}

func writeShopItems(bf *byteframe.ByteFrame, items []ShopItem, mode _config.Mode) {
	bf.WriteUint16(uint16(len(items)))
	bf.WriteUint16(uint16(len(items)))
	for _, item := range items {
		if mode >= _config.Z2 {
			bf.WriteUint32(item.ID)
		}
		bf.WriteUint32(item.ItemID)
		bf.WriteUint32(item.Cost)
		bf.WriteUint16(item.Quantity)
		bf.WriteUint16(item.MinHR)
		bf.WriteUint16(item.MinSR)
		if mode >= _config.Z2 {
			bf.WriteUint16(item.MinGR)
		}
		bf.WriteUint8(0) // Unk
		bf.WriteUint8(item.StoreLevel)
		if mode >= _config.Z2 {
			bf.WriteUint16(item.MaxQuantity)
			bf.WriteUint16(item.UsedQuantity)
		}
		if mode == _config.Z1 {
			bf.WriteUint8(uint8(item.RoadFloors))
			bf.WriteUint8(uint8(item.RoadFatalis))
		} else if mode >= _config.Z2 {
			bf.WriteUint16(item.RoadFloors)
			bf.WriteUint16(item.RoadFatalis)
		}
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
		// Fundamentally, gacha works completely differently, just hide it for now.
		if s.server.erupeConfig.RealClientMode <= _config.G7 {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}

		rows, err := s.server.db.Queryx("SELECT id, min_gr, min_hr, name, url_banner, url_feature, url_thumbnail, wide, recommended, gacha_type, hidden FROM gacha_shop")
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		bf := byteframe.NewByteFrame()
		var gacha Gacha
		var gachas []Gacha
		for rows.Next() {
			err = rows.StructScan(&gacha)
			if err == nil {
				gachas = append(gachas, gacha)
			}
		}
		bf.WriteUint16(uint16(len(gachas)))
		bf.WriteUint16(uint16(len(gachas)))
		for _, g := range gachas {
			bf.WriteUint32(g.ID)
			bf.WriteUint32(0) // Unknown rank restrictions
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint32(g.MinGR)
			bf.WriteUint32(g.MinHR)
			bf.WriteUint32(0) // only 0 in known packet
			ps.Uint8(bf, g.Name, true)
			ps.Uint8(bf, g.URLBanner, false)
			ps.Uint8(bf, g.URLFeature, false)
			if s.server.erupeConfig.RealClientMode >= _config.G10 {
				bf.WriteBool(g.Wide)
				ps.Uint8(bf, g.URLThumbnail, false)
			}
			if g.Recommended {
				bf.WriteUint16(2)
			} else {
				bf.WriteUint16(0)
			}
			bf.WriteUint8(g.GachaType)
			if s.server.erupeConfig.RealClientMode >= _config.G10 {
				bf.WriteBool(g.Hidden)
			}
		}
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	case 2: // Actual gacha
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(pkt.ShopID)
		var gachaType int
		_ = s.server.db.QueryRow(`SELECT gacha_type FROM gacha_shop WHERE id = $1`, pkt.ShopID).Scan(&gachaType)
		rows, err := s.server.db.Queryx(`SELECT entry_type, id, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points, COALESCE(name, '') AS name FROM gacha_entries WHERE gacha_id = $1 ORDER BY weight DESC`, pkt.ShopID)
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		var divisor float64
		_ = s.server.db.QueryRow(`SELECT COALESCE(SUM(weight) / 100000.0, 0) AS chance FROM gacha_entries WHERE gacha_id = $1`, pkt.ShopID).Scan(&divisor)

		var entry GachaEntry
		var entries []GachaEntry
		var item GachaItem
		for rows.Next() {
			err = rows.StructScan(&entry)
			if err == nil {
				entries = append(entries, entry)
			}
		}
		bf.WriteUint16(uint16(len(entries)))
		for _, ge := range entries {
			var items []GachaItem
			bf.WriteUint8(ge.EntryType)
			bf.WriteUint32(ge.ID)
			bf.WriteUint8(ge.ItemType)
			bf.WriteUint32(ge.ItemNumber)
			bf.WriteUint16(ge.ItemQuantity)
			if gachaType >= 4 { // If box
				bf.WriteUint16(1)
			} else {
				bf.WriteUint16(uint16(ge.Weight / divisor))
			}
			bf.WriteUint8(ge.Rarity)
			bf.WriteUint8(ge.Rolls)

			rows, err = s.server.db.Queryx(`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id=$1`, ge.ID)
			if err != nil {
				bf.WriteUint8(0)
			} else {
				for rows.Next() {
					err = rows.StructScan(&item)
					if err == nil {
						items = append(items, item)
					}
				}
				bf.WriteUint8(uint8(len(items)))
			}

			bf.WriteUint16(ge.FrontierPoints)
			bf.WriteUint8(ge.DailyLimit)
			if ge.EntryType < 10 {
				ps.Uint8(bf, ge.Name, true)
			} else {
				bf.WriteUint8(0)
			}
			for _, gi := range items {
				bf.WriteUint16(uint16(gi.ItemType))
				bf.WriteUint16(gi.ItemID)
				bf.WriteUint16(gi.Quantity)
			}
		}
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
		if len(items) > int(pkt.Limit) {
			items = items[:pkt.Limit]
		}
		writeShopItems(bf, items, s.server.erupeConfig.RealClientMode)
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
		if _, err := s.server.db.Exec(`INSERT INTO shop_items_bought (character_id, shop_item_id, bought)
			VALUES ($1,$2,$3) ON CONFLICT (character_id, shop_item_id)
			DO UPDATE SET bought = bought + $3
			WHERE EXCLUDED.character_id=$1 AND EXCLUDED.shop_item_id=$2
		`, s.charID, itemHash, buyCount); err != nil {
			s.logger.Error("Failed to update shop item purchase count", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

// FPointExchange represents a frontier point exchange entry.
type FPointExchange struct {
	ID       uint32 `db:"id"`
	ItemType uint8  `db:"item_type"`
	ItemID   uint16 `db:"item_id"`
	Quantity uint16 `db:"quantity"`
	FPoints  uint16 `db:"fpoints"`
	Buyable  bool   `db:"buyable"`
}

func handleMsgMhfExchangeFpoint2Item(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeFpoint2Item)
	var balance uint32
	var itemValue, quantity int
	if err := s.server.db.QueryRow("SELECT quantity, fpoints FROM fpoint_items WHERE id=$1", pkt.TradeID).Scan(&quantity, &itemValue); err != nil {
		s.logger.Error("Failed to read fpoint item cost", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	cost := (int(pkt.Quantity) * quantity) * itemValue
	if err := s.server.db.QueryRow("UPDATE users SET frontier_points=frontier_points::int - $1 WHERE id=$2 RETURNING frontier_points", cost, s.userID).Scan(&balance); err != nil {
		s.logger.Error("Failed to deduct frontier points", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(balance)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfExchangeItem2Fpoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeItem2Fpoint)
	var balance uint32
	var itemValue, quantity int
	if err := s.server.db.QueryRow("SELECT quantity, fpoints FROM fpoint_items WHERE id=$1", pkt.TradeID).Scan(&quantity, &itemValue); err != nil {
		s.logger.Error("Failed to read fpoint item value", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	cost := (int(pkt.Quantity) / quantity) * itemValue
	if err := s.server.db.QueryRow("UPDATE users SET frontier_points=COALESCE(frontier_points::int + $1, $1) WHERE id=$2 RETURNING frontier_points", cost, s.userID).Scan(&balance); err != nil {
		s.logger.Error("Failed to credit frontier points", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(balance)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetFpointExchangeList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFpointExchangeList)

	bf := byteframe.NewByteFrame()
	var exchange FPointExchange
	var exchanges []FPointExchange
	var buyables uint16
	rows, err := s.server.db.Queryx(`SELECT id, item_type, item_id, quantity, fpoints, buyable FROM fpoint_items ORDER BY buyable DESC`)
	if err == nil {
		for rows.Next() {
			err = rows.StructScan(&exchange)
			if err != nil {
				continue
			}
			if exchange.Buyable {
				buyables++
			}
			exchanges = append(exchanges, exchange)
		}
	}
	if s.server.erupeConfig.RealClientMode <= _config.Z2 {
		bf.WriteUint8(uint8(len(exchanges)))
		bf.WriteUint8(uint8(buyables))
	} else {
		bf.WriteUint16(uint16(len(exchanges)))
		bf.WriteUint16(buyables)
	}
	for _, e := range exchanges {
		bf.WriteUint32(e.ID)
		bf.WriteUint16(0)
		bf.WriteUint16(0)
		bf.WriteUint16(0)
		bf.WriteUint8(e.ItemType)
		bf.WriteUint16(e.ItemID)
		bf.WriteUint16(e.Quantity)
		bf.WriteUint16(e.FPoints)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
