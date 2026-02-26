package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	cfg "erupe-ce/config"
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

func writeShopItems(bf *byteframe.ByteFrame, items []ShopItem, mode cfg.Mode) {
	bf.WriteUint16(uint16(len(items)))
	bf.WriteUint16(uint16(len(items)))
	for _, item := range items {
		if mode >= cfg.Z2 {
			bf.WriteUint32(item.ID)
		}
		bf.WriteUint32(item.ItemID)
		bf.WriteUint32(item.Cost)
		bf.WriteUint16(item.Quantity)
		bf.WriteUint16(item.MinHR)
		bf.WriteUint16(item.MinSR)
		if mode >= cfg.Z2 {
			bf.WriteUint16(item.MinGR)
		}
		bf.WriteUint8(0) // Unk
		bf.WriteUint8(item.StoreLevel)
		if mode >= cfg.Z2 {
			bf.WriteUint16(item.MaxQuantity)
			bf.WriteUint16(item.UsedQuantity)
		}
		if mode == cfg.Z1 {
			bf.WriteUint8(uint8(item.RoadFloors))
			bf.WriteUint8(uint8(item.RoadFatalis))
		} else if mode >= cfg.Z2 {
			bf.WriteUint16(item.RoadFloors)
			bf.WriteUint16(item.RoadFatalis)
		}
	}
}

func getShopItems(s *Session, shopType uint8, shopID uint32) []ShopItem {
	items, err := s.server.shopRepo.GetShopItems(shopType, shopID, s.charID)
	if err != nil {
		return nil
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
		if s.server.erupeConfig.RealClientMode < cfg.G1 {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}

		gachas, err := s.server.gachaRepo.ListShop()
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(len(gachas)))
		bf.WriteUint16(uint16(len(gachas)))
		for _, g := range gachas {
			if s.server.erupeConfig.RealClientMode >= cfg.GG{ 
				//Before GG, there was no data for G1, so there was no data for G1 except for ID and name
				//But the difference between G2 and G3 still needs to be tested, and the data for G1 and GG are already clear
				bf.WriteUint32(g.ID)
				bf.WriteUint32(0) // Unknown rank restrictions
				bf.WriteUint32(0)
				bf.WriteUint32(0)
				bf.WriteUint32(0)
				bf.WriteUint32(g.MinGR)
				bf.WriteUint32(g.MinHR)
				bf.WriteUint32(0) // only 0 in known packet
			}
			ps.Uint8(bf, g.Name, true)
			if s.server.erupeConfig.RealClientMode <= cfg.GG{  { //For versions less than or equal to GG, each message sent to the name ends
				continue
			}
			ps.Uint8(bf, g.URLBanner, false)
			ps.Uint8(bf, g.URLFeature, false)
			if s.server.erupeConfig.RealClientMode >= cfg.G10 {
				bf.WriteBool(g.Wide)
				ps.Uint8(bf, g.URLThumbnail, false)
			}
			if g.Recommended {
				bf.WriteUint16(2)
			} else {
				bf.WriteUint16(0)
			}
			bf.WriteUint8(g.GachaType)
			if s.server.erupeConfig.RealClientMode >= cfg.G10 {
				bf.WriteBool(g.Hidden)
			}
		}
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	case 2: // Actual gacha
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(pkt.ShopID)
		gachaType, err := s.server.gachaRepo.GetShopType(pkt.ShopID)
		if err != nil {
			s.logger.Error("Failed to get gacha shop type", zap.Error(err))
		}
		entries, err := s.server.gachaRepo.GetAllEntries(pkt.ShopID)
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		divisor, err := s.server.gachaRepo.GetWeightDivisor(pkt.ShopID)
		if err != nil {
			s.logger.Error("Failed to get gacha weight divisor", zap.Error(err))
		}
		bf.WriteUint16(uint16(len(entries)))
		for _, ge := range entries {
			var items []GachaItem
			if s.server.erupeConfig.RealClientMode <= cfg.GG {
				// If you need to configure the optional material list among the three options,Configure directly in gacha_detries,The same Entry Type can be merged and displayed in GG,In addition, the prizes are also directly configured in the gacha-entries table,
				// MHFG1~GG does not use the gacha_items table throughout the entire process, which meets the lottery function of MHFG with a more single function
				// In addition, the MHFG function itself is relatively simple,Example of lottery configuration for G1~GG:
				// 	eg: gachaname:test
				// 	entry: itemgroup: group1:(choose one of the two) ITEM_1_ID:7 COUNT:1  ITEM_1_ID:8 COUNT:2group2:ITEM_1_ID:9 COUNT:3 ; reward:reward1: ITEM_ID:1 COUNT:4 weight:10% reward1: ITEM_ID:2 COUNT:5 weight:90%
				// 	table:gacha_shop |1|0|0|test|null|null|null|f|f|3|f|
				// 	table:gacha_entries
				// 	|1|1|0|7|7|1|0|0|0|0|0|null|
				// 	|4|1|0|7|8|2|0|0|0|0|0|null|
				// 	|5|1|1|7|9|3|0|0|0|0|0|null|
				// 	|8|1|100|7|1|4|1000|0|0|0|0|null|
				// 	|9|1|100|7|2|5|9000|0|0|0|0|null|
				bf.WriteUint8(ge.EntryType)     
				bf.WriteUint32(ge.ID)           
				bf.WriteUint8(ge.ItemType)      
				bf.WriteUint32(ge.ItemNumber)   
				bf.WriteUint16(ge.ItemQuantity) 
				var weightPr uint16
				if gachaType >= 4 { // If box   
					weightPr = 1
				} else {
					weightPr = uint16(ge.Weight / divisor)
				}
				bf.WriteUint16(weightPr)
				bf.WriteUint8(0)
				continue
			}
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

			items, err := s.server.gachaRepo.GetItemsForEntry(ge.ID)
			if err != nil {
				bf.WriteUint8(0)
			} else {
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
		if err := s.server.shopRepo.RecordPurchase(s.charID, itemHash, buyCount); err != nil {
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
	quantity, itemValue, err := s.server.shopRepo.GetFpointItem(pkt.TradeID)
	if err != nil {
		s.logger.Error("Failed to read fpoint item cost", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	cost := (int(pkt.Quantity) * quantity) * itemValue
	balance, err := s.server.userRepo.AdjustFrontierPointsDeduct(s.userID, cost)
	if err != nil {
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
	quantity, itemValue, err := s.server.shopRepo.GetFpointItem(pkt.TradeID)
	if err != nil {
		s.logger.Error("Failed to read fpoint item value", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	cost := (int(pkt.Quantity) / quantity) * itemValue
	balance, err := s.server.userRepo.AdjustFrontierPointsCredit(s.userID, cost)
	if err != nil {
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
	exchanges, err := s.server.shopRepo.GetFpointExchangeList()
	if err != nil {
		s.logger.Error("Failed to get fpoint exchange list", zap.Error(err))
	}
	var buyables uint16
	for _, e := range exchanges {
		if e.Buyable {
			buyables++
		}
	}
	if s.server.erupeConfig.RealClientMode <= cfg.Z2 {
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
