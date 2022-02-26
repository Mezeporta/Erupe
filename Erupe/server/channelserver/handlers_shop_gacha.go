package channelserver

import (
	"encoding/hex"
	"time"

	//"github.com/Solenataris/Erupe/common/stringsupport"
	"github.com/Solenataris/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"github.com/lib/pq"
	"github.com/sachaos/lottery"
	"go.uber.org/zap"
)

func handleMsgMhfEnumerateShop(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateShop)
	// SHOP TYPES:
	// 01 = Running Gachas, 02 = actual gacha, 04 = N Points, 05 = GCP, 07 = Item to GCP, 08 = Diva Defense, 10 = Hunter's Road

	// GACHA FORMAT:
	// int32: gacha id

	// STORE FORMAT:
	// Int16: total item count
	// Int16: total item count

	// ITEM FORMAT:
	// int32: Unique item hash for tracking purchases
	// int16: padding?
	// int16: Item ID
	// int16: padding?
	// int16: GCP returns
	// int16: Number traded at once
	// int16: HR or SR Requirement
	// int16: Whichever of the above it isn't
	// int16: GR Requirement
	// int16: Store level requirement
	// int16: Maximum quantity purchasable
	// int16: Unk
	// int16: Road floors cleared requirement
	// int16: Road White Fatalis weekly kills
	if pkt.ShopType == 2 {
		shopEntries, err := s.server.db.Query("SELECT entryType, itemhash, currType, currNumber, currQuant, percentage, rarityIcon, rollsCount, itemCount, dailyLimit, itemType, itemId, quantity FROM gacha_shop_items WHERE shophash=$1", pkt.ShopID)
		if err != nil {
			panic(err)
		}
		var entryType, currType, rarityIcon, rollsCount, itemCount, dailyLimit byte
		var currQuant, currNumber, percentage uint16
		var itemhash uint32
		var itemType, itemId, quantity pq.Int64Array
		var entryCount int
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(pkt.ShopID)
		resp.WriteUint16(0) // total defs
		for shopEntries.Next() {
			err = shopEntries.Scan(&entryType, &itemhash, &currType, &currNumber, &currQuant, &percentage, &rarityIcon, &rollsCount, &itemCount, &dailyLimit, (*pq.Int64Array)(&itemType), (*pq.Int64Array)(&itemId), (*pq.Int64Array)(&quantity))
			if err != nil {
				panic(err)
			}
			resp.WriteUint8(entryType)
			resp.WriteUint32(itemhash)
			resp.WriteUint8(currType)
			resp.WriteUint16(0)          // unk, always 0 in existing packets
			resp.WriteUint16(currNumber) // it's either item ID or quantity for gacha coins
			resp.WriteUint16(currQuant)  // only for item ID
			resp.WriteUint16(percentage)
			resp.WriteUint8(rarityIcon)
			resp.WriteUint8(rollsCount)
			resp.WriteUint8(itemCount)
			resp.WriteUint16(0) // unk, always 0 in existing packets
			resp.WriteUint8(dailyLimit)
			resp.WriteUint8(0) // unk, always 0 in existing packets
			for i := 0; i < int(itemCount); i++ {
				resp.WriteUint16(uint16(itemType[i])) // unk, always 0 in existing packets
				resp.WriteUint16(uint16(itemId[i]))   // unk, always 0 in existing packets
				resp.WriteUint16(uint16(quantity[i])) // unk, always 0 in existing packets
			}
			entryCount++
		}
		if entryCount == 0 {
			doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
		resp.Seek(4, 0)
		resp.WriteUint16(uint16(entryCount))
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	} else if pkt.ShopType == 1 {
		gachaCount := 0
		shopEntries, err := s.server.db.Query("SELECT hash, reqGR, reqHR, gachaName, gachaLink0, gachaLink1, COALESCE(gachaLink2, ''), extraIcon, gachaType, hideFlag FROM gacha_shop")
		if err != nil {
			panic(err)
		}
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0)
		var gachaName, gachaLink0, gachaLink1, gachaLink2 string
		var hash, reqGR, reqHR, extraIcon, gachaType int
		var hideFlag bool
		for shopEntries.Next() {
			err = shopEntries.Scan(&hash, &reqGR, &reqHR, &gachaName, &gachaLink0, &gachaLink1, &gachaLink2, &extraIcon, &gachaType, &hideFlag)
			if err != nil {
				panic(err)
			}
			resp.WriteUint32(uint32(hash))
			resp.WriteUint32(0) // only 0 in known packets
			resp.WriteUint32(0) // all of these seem to trigger the 'rank restriction'
			resp.WriteUint32(0) // message so they are presumably placeholders for a
			resp.WriteUint32(0) // Z Rank or similar that never turned up?
			resp.WriteUint32(uint32(reqGR))
			resp.WriteUint32(uint32(reqHR))
			resp.WriteUint32(0) // only 0 in known packet
			stringBytes := append([]byte(gachaName), 0x00)
			resp.WriteUint8(byte(len(stringBytes)))
			resp.WriteBytes(stringBytes)
			stringBytes = append([]byte(gachaLink0), 0x00)
			resp.WriteUint8(byte(len(stringBytes)))
			resp.WriteBytes(stringBytes)
			stringBytes = append([]byte(gachaLink1), 0x00)
			resp.WriteUint8(byte(len(stringBytes)))
			resp.WriteBytes(stringBytes)
			stringBytes = append([]byte(gachaLink2), 0x00)
			resp.WriteBool(hideFlag)
			resp.WriteUint8(uint8(len(stringBytes)))
			resp.WriteBytes(stringBytes)
			resp.WriteUint16(uint16(extraIcon))
			resp.WriteUint16(uint16(gachaType))
			gachaCount++
		}
		resp.Seek(0, 0)
		resp.WriteUint16(uint16(gachaCount))
		resp.WriteUint16(uint16(gachaCount))
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())

	} else if pkt.ShopType == 7 {
		// GCP conversion store
		if pkt.ShopID == 0 {
			// Items to GCP exchange. Gou Tickets, Shiten Tickets, GP Tickets
			data, _ := hex.DecodeString("000300033a9186fb000033860000000a000100000000000000000000000000000000097fdb1c0000067e0000000a0001000000000000000000000000000000001374db29000027c300000064000100000000000000000000000000000000")
			doAckBufSucceed(s, pkt.AckHandle, data)
		} else {
			doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		}
	} else if pkt.ShopType == 8 {
		// Dive Defense sections
		// 00 = normal level limited exchange store, 05 = GCP skill store, 07 = limited quantity exchange
		if pkt.ShopID == 5 {
			// diva defense skill level limited store
			data, _ := hex.DecodeString("001f001f2c9365c1000000010000001e000a0000000000000000000a0000000000001979f1c2000000020000003c000a0000000000000000000a0000000000003e5197df000000030000003c000a0000000000000000000a000000000000219337c0000000040000001e000a0000000000000000000a00000000000009b24c9d000000140000001e000a0000000000000000000a0000000000001f1d496e000000150000001e000a0000000000000000000a0000000000003b918fcb000000160000003c000a0000000000000000000a0000000000000b7fd81c000000170000003c000a0000000000000000000a0000000000001374f239000000180000003c000a0000000000000000000a00000000000026950cba0000001c0000003c000a0000000000000000000a0000000000003797eae70000001d0000003c000a012b000000000000000a00000000000015758ad8000000050000003c00000000000000000000000a0000000000003c7035050000000600000050000a0000000000000001000a00000000000024f3b5560000000700000050000a0000000000000001000a00000000000000b600330000000800000050000a0000000000000001000a0000000000002efdce840000001900000050000a0000000000000001000a0000000000002d9365f10000001a00000050000a0000000000000001000a0000000000001979f3420000001f00000050000a012b000000000001000a0000000000003f5397cf0000002000000050000a012b000000000001000a000000000000319337c00000002100000050000a012b000000000001000a00000000000008b04cbd0000000900000064000a0000000000000002000a0000000000000b1d4b6e0000000a00000064000a0000000000000002000a0000000000003b918feb0000000b00000064000a0000000000000002000a0000000000001b7fd81c0000000c00000064000a0000000000000002000a0000000000001276f2290000000d00000064000a0000000000000002000a00000000000022950cba0000000e000000c8000a0000000000000002000a0000000000003697ead70000000f000001f4000a0000000000000003000a00000000000005758a5800000010000003e8000a0000000000000003000a0000000000003c7035250000001b000001f4000a0000000000010003000a00000000000034f3b5d60000001e00000064000a012b000000000003000a00000000000000b600030000002200000064000a0000000000010003000a000000000000")
			doAckBufSucceed(s, pkt.AckHandle, data)
		} else {
			doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		}
         } else {
		shopEntries, err := s.server.db.Query("SELECT itemhash,itemID,Points,TradeQuantity,rankReqLow,rankReqHigh,rankReqG,storeLevelReq,maximumQuantity,boughtQuantity,roadFloorsRequired,weeklyFatalisKills FROM normal_shop_items WHERE shoptype=$1 AND shopid=$2", pkt.ShopType, pkt.ShopID)
		if err != nil {
			panic(err)
		}
		var ItemHash, entryCount int
		var itemID, Points, TradeQuantity, rankReqLow, rankReqHigh, rankReqG, storeLevelReq, maximumQuantity, boughtQuantity, roadFloorsRequired, weeklyFatalisKills, charQuantity uint16
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0) // total defs
		for shopEntries.Next() {
			err = shopEntries.Scan(&ItemHash, &itemID, &Points, &TradeQuantity, &rankReqLow, &rankReqHigh, &rankReqG, &storeLevelReq, &maximumQuantity, &boughtQuantity, &roadFloorsRequired, &weeklyFatalisKills)
			if err != nil {
				panic(err)
			}
			resp.WriteUint32(uint32(ItemHash))
			resp.WriteUint16(0) // unk, always 0 in existing packets
			resp.WriteUint16(itemID)
			resp.WriteUint16(0)             // unk, always 0 in existing packets
			resp.WriteUint16(Points)        // it's either item ID or quantity for gacha coins
			resp.WriteUint16(TradeQuantity) // only for item ID
			resp.WriteUint16(rankReqLow)
			resp.WriteUint16(rankReqHigh)
			resp.WriteUint16(rankReqG)
			resp.WriteUint16(storeLevelReq)
			resp.WriteUint16(maximumQuantity)
			if maximumQuantity > 0 {
				err = s.server.db.QueryRow("SELECT COALESCE(usedquantity,0) FROM shop_item_state WHERE itemhash=$1 AND char_id=$2", ItemHash, s.charID).Scan(&charQuantity)
				if err != nil {
					resp.WriteUint16(0)
				} else {
					resp.WriteUint16(charQuantity)
				}
			} else {
				resp.WriteUint16(boughtQuantity)
			}
			resp.WriteUint16(roadFloorsRequired)
			resp.WriteUint16(weeklyFatalisKills)
			entryCount++
		}
		if entryCount == 0 {
			doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
		resp.Seek(0, 0)
		resp.WriteUint16(uint16(entryCount))
		resp.WriteUint16(uint16(entryCount))
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

func handleMsgMhfAcquireExchangeShop(s *Session, p mhfpacket.MHFPacket) {
	// writing out to an editable shop enumeration
	pkt := p.(*mhfpacket.MsgMhfAcquireExchangeShop)
	if pkt.DataSize == 10 {
		bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
		_ = bf.ReadUint16() // unk, always 1 in examples
		itemHash := bf.ReadUint32()
		buyCount := bf.ReadUint32()
		_, err := s.server.db.Exec(`INSERT INTO shop_item_state (char_id, itemhash, usedquantity)
  														 VALUES ($1,$2,$3) ON CONFLICT (char_id, itemhash)
  														 DO UPDATE SET usedquantity = shop_item_state.usedquantity + $3
  														 WHERE EXCLUDED.char_id=$1 AND EXCLUDED.itemhash=$2`, s.charID, itemHash, buyCount)
		if err != nil {
			s.logger.Fatal("Failed to update shop_item_state in db", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetGachaPlayHistory(s *Session, p mhfpacket.MHFPacket) {
	// returns number of times the gacha was played, will need persistent db stuff
	pkt := p.(*mhfpacket.MsgMhfGetGachaPlayHistory)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x0A})
}

func handleMsgMhfGetGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGachaPoint)
	var fp, gp, gt uint32
	_ = s.server.db.QueryRow("SELECT COALESCE(frontier_points, 0), COALESCE(gacha_prem, 0), COALESCE(gacha_trial,0) FROM characters WHERE id=$1", s.charID).Scan(&fp, &gp, &gt)
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(gp) // Real Gacha Points?
	resp.WriteUint32(gt) // Trial Gacha Point?
	resp.WriteUint32(fp) // Frontier Points?

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

type gachaItem struct {
	itemhash   uint32
	percentage uint16
	rarityIcon byte
	itemCount  byte
	itemType   pq.Int64Array
	itemId     pq.Int64Array
	quantity   pq.Int64Array
}

func (i gachaItem) Weight() int {
	return int(i.percentage)
}

func handleMsgMhfPlayNormalGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayNormalGacha)
	// needs to query db for input gacha and return a result or number of results
	// uint8 number of results
	// uint8 item type
	// uint16 item id
	// uint16 quantity

	var currType, rarityIcon, rollsCount, itemCount byte
	var currQuant, currNumber, percentage uint16
	var itemhash uint32
	var itemType, itemId, quantity pq.Int64Array
	var items []lottery.Weighter
	// get info for updating data and calculating costs
	err := s.server.db.QueryRow("SELECT currType, currNumber, currQuant, rollsCount FROM gacha_shop_items WHERE shophash=$1 AND entryType=$2", pkt.GachaHash, pkt.RollType).Scan(&currType, &currNumber, &currQuant, &rollsCount)
	if err != nil {
		panic(err)
	}
	// get existing items in storage if any
	var data []byte
	_ = s.server.db.QueryRow("SELECT gacha_items FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		data = []byte{0x00}
	}
	// get gacha items and iterate through them for gacha roll
	shopEntries, err := s.server.db.Query("SELECT itemhash, percentage, rarityIcon, itemCount, itemType, itemId, quantity FROM gacha_shop_items WHERE shophash=$1 AND entryType=100", pkt.GachaHash)
	if err != nil {
		panic(err)
	}
	for shopEntries.Next() {
		err = shopEntries.Scan(&itemhash, &percentage, &rarityIcon, &itemCount, (*pq.Int64Array)(&itemType), (*pq.Int64Array)(&itemId), (*pq.Int64Array)(&quantity))
		if err != nil {
			panic(err)
		}
		items = append(items, &gachaItem{itemhash: itemhash, percentage: percentage, rarityIcon: rarityIcon, itemCount: itemCount, itemType: itemType, itemId: itemId, quantity: quantity})
	}
	// execute rolls, build response and update database
	results := byte(0)
	resp := byteframe.NewByteFrame()
	dbUpdate := byteframe.NewByteFrame()
	resp.WriteUint8(0) // results go here later
	l := lottery.NewDefaultLottery()
	for x := 0; x < int(rollsCount); x++ {
		ind := l.Draw(items)
		results += items[ind].(*gachaItem).itemCount
		for y := 0; y < int(items[ind].(*gachaItem).itemCount); y++ {
			// items in storage don't get rarity
			dbUpdate.WriteUint8(uint8(items[ind].(*gachaItem).itemType[y]))
			dbUpdate.WriteUint16(uint16(items[ind].(*gachaItem).itemId[y]))
			dbUpdate.WriteUint16(uint16(items[ind].(*gachaItem).quantity[y]))
			data = append(data, dbUpdate.Data()...)
			dbUpdate.Seek(0, 0)
			// response needs all item info and the rarity
			resp.WriteBytes(dbUpdate.Data())
			resp.WriteUint8(uint8(items[ind].(*gachaItem).rarityIcon))
		}
	}
	resp.Seek(0, 0)
	resp.WriteUint8(uint8(results))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())

	// add claimables to DB
	data[0] = data[0] + results
	_, err = s.server.db.Exec("UPDATE characters SET gacha_items = $1 WHERE id = $2", data, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update minidata in db", zap.Error(err))
	}
	// deduct gacha coins if relevant, items are handled fine by the standard savedata packet immediately afterwards
	if currType == 19 {
		_, err = s.server.db.Exec("UPDATE characters SET gacha_trial = CASE WHEN (gacha_trial > $1) then gacha_trial - $1 else gacha_trial end, gacha_prem = CASE WHEN NOT (gacha_trial > $1) then gacha_prem - $1 else gacha_prem end WHERE id=$2", currNumber, s.charID)
	}
	if err != nil {
		s.logger.Fatal("Failed to update gacha_items in db", zap.Error(err))
	}
}

func handleMsgMhfUseGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	// should write to database when that's set up
	pkt := p.(*mhfpacket.MsgMhfUseGachaPoint)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfExchangeFpoint2Item(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeFpoint2Item)

	var itemValue, quant int
	_ = s.server.db.QueryRow("SELECT quant, itemValue FROM fpoint_items WHERE hash=$1", pkt.ItemHash).Scan(&quant, &itemValue)
	itemCost := (int(pkt.Quantity) * quant) * itemValue

	// also update frontierpoints entry in database
	_, err := s.server.db.Exec("UPDATE characters SET frontier_points=frontier_points::int - $1 WHERE id=$2", itemCost, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update minidata in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfExchangeItem2Fpoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeItem2Fpoint)

	var itemValue, quant int
	_ = s.server.db.QueryRow("SELECT quant, itemValue FROM fpoint_items WHERE hash=$1", pkt.ItemHash).Scan(&quant, &itemValue)
	itemCost := (int(pkt.Quantity) / quant) * itemValue
	// also update frontierpoints entry in database
	_, err := s.server.db.Exec("UPDATE characters SET frontier_points=COALESCE(frontier_points::int + $1, $1) WHERE id=$2", itemCost, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update minidata in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetFpointExchangeList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFpointExchangeList)
	//absurd, probably lists every single item to trade to FP?

	var buyables int
	var sellables int

	buyRows, err := s.server.db.Query("SELECT hash,itemType,itemID,quant,itemValue FROM fpoint_items WHERE tradeType=0")
	if err != nil {
		panic(err)
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	var hash, itemType, itemID, quant, itemValue int
	for buyRows.Next() {
		err = buyRows.Scan(&hash, &itemType, &itemID, &quant, &itemValue)
		if err != nil {
			panic("Error in fpoint_items")
		}
		resp.WriteUint32(uint32(hash))
		resp.WriteUint32(0) // this and following only 0 in known packets
		resp.WriteUint16(0)
		resp.WriteUint8(byte(itemType))
		resp.WriteUint16(uint16(itemID))
		resp.WriteUint16(uint16(quant))
		resp.WriteUint16(uint16(itemValue))
		buyables++
	}

	sellRows, err := s.server.db.Query("SELECT hash,itemType,itemID,quant,itemValue FROM fpoint_items WHERE tradeType=1")
	if err != nil {
		panic(err)
	}
	for sellRows.Next() {
		err = sellRows.Scan(&hash, &itemType, &itemID, &quant, &itemValue)
		if err != nil {
			panic("Error in fpoint_items")
		}
		resp.WriteUint32(uint32(hash))
		resp.WriteUint32(0) // this and following only 0 in known packets
		resp.WriteUint16(0)
		resp.WriteUint8(byte(itemType))
		resp.WriteUint16(uint16(itemID))
		resp.WriteUint16(uint16(quant))
		resp.WriteUint16(uint16(itemValue))
		sellables++
	}
	resp.Seek(0, 0)
	resp.WriteUint16(uint16(sellables))
	resp.WriteUint16(uint16(buyables))

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfPlayStepupGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayStepupGacha)
	results := byte(0)
	stepResults := byte(0)
	resp := byteframe.NewByteFrame()
	rollFrame := byteframe.NewByteFrame()
	stepFrame := byteframe.NewByteFrame()
	stepData := []byte{}
	var currType, rarityIcon, rollsCount, itemCount byte
	var currQuant, currNumber, percentage uint16
	var itemhash uint32
	var itemType, itemId, quantity pq.Int64Array
	var items []lottery.Weighter
	// get info for updating data and calculating costs
	err := s.server.db.QueryRow("SELECT currType, currNumber, currQuant, rollsCount, itemCount, itemType, itemId, quantity FROM gacha_shop_items WHERE shophash=$1 AND entryType=$2", pkt.GachaHash, pkt.RollType).Scan(&currType, &currNumber, &currQuant, &rollsCount, &itemCount, (*pq.Int64Array)(&itemType), (*pq.Int64Array)(&itemId), (*pq.Int64Array)(&quantity))
	if err != nil {
		panic(err)
	}
	// get existing items in storage if any
	var data []byte
	_ = s.server.db.QueryRow("SELECT gacha_items FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		data = []byte{0x00}
	}
	// roll definition includes items with step up gachas that are appended last
	for x := 0; x < int(itemCount); x++ {
		stepFrame.WriteUint8(uint8(itemType[x]))
		stepFrame.WriteUint16(uint16(itemId[x]))
		stepFrame.WriteUint16(uint16(quantity[x]))
		stepData = append(stepData, stepFrame.Data()...)
		stepFrame.WriteUint8(0) // rarity still defined
		stepResults++
	}
	// get gacha items and iterate through them for gacha roll
	shopEntries, err := s.server.db.Query("SELECT itemhash, percentage, rarityIcon, itemCount, itemType, itemId, quantity FROM gacha_shop_items WHERE shophash=$1 AND entryType=100", pkt.GachaHash)
	if err != nil {
		panic(err)
	}
	for shopEntries.Next() {
		err = shopEntries.Scan(&itemhash, &percentage, &rarityIcon, &itemCount, (*pq.Int64Array)(&itemType), (*pq.Int64Array)(&itemId), (*pq.Int64Array)(&quantity))
		if err != nil {
			panic(err)
		}
		items = append(items, &gachaItem{itemhash: itemhash, percentage: percentage, rarityIcon: rarityIcon, itemCount: itemCount, itemType: itemType, itemId: itemId, quantity: quantity})
	}
	// execute rolls, build response and update database
	resp.WriteUint16(0) // results count goes here later
	l := lottery.NewDefaultLottery()
	for x := 0; x < int(rollsCount); x++ {
		ind := l.Draw(items)
		results += items[ind].(*gachaItem).itemCount
		for y := 0; y < int(items[ind].(*gachaItem).itemCount); y++ {
			// items in storage don't get rarity
			rollFrame.WriteUint8(uint8(items[ind].(*gachaItem).itemType[y]))
			rollFrame.WriteUint16(uint16(items[ind].(*gachaItem).itemId[y]))
			rollFrame.WriteUint16(uint16(items[ind].(*gachaItem).quantity[y]))
			data = append(data, rollFrame.Data()...)
			rollFrame.Seek(0, 0)
			// response needs all item info and the rarity
			resp.WriteBytes(rollFrame.Data())
			resp.WriteUint8(uint8(items[ind].(*gachaItem).rarityIcon))
		}
	}
	resp.WriteBytes(stepFrame.Data())
	resp.Seek(0, 0)
	resp.WriteUint8(uint8(results + stepResults))
	resp.WriteUint8(uint8(results))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())

	// add claimables to DB
	data = append(data, stepData...)
	data[0] = data[0] + results + stepResults
	_, err = s.server.db.Exec("UPDATE characters SET gacha_items = $1 WHERE id = $2", data, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update gacha_items in db", zap.Error(err))
	}
	// deduct gacha coins if relevant, items are handled fine by the standard savedata packet immediately afterwards
	// reduce real if trial don't cover cost
	if currType == 19 {
		_, err = s.server.db.Exec(`UPDATE characters
																	SET gacha_trial = CASE WHEN (gacha_trial > $1) then gacha_trial - $1 else gacha_trial end,
																	gacha_prem = CASE WHEN NOT (gacha_trial > $1) then gacha_prem - $1 else gacha_prem end
																	WHERE id=$2`, currNumber, s.charID)

	}
	if err != nil {
		s.logger.Fatal("Failed to update gacha_items in db", zap.Error(err))
	}
	// update step progression
	_, err = s.server.db.Exec("UPDATE stepup_state SET step_progression = $1 WHERE char_id = $2", pkt.RollType+1, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update step_progression in db", zap.Error(err))
	}

}

func handleMsgMhfReceiveGachaItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReceiveGachaItem)
	// persistent for claimable items on cat
	var data []byte
	err := s.server.db.QueryRow("SELECT COALESCE(gacha_items, $2) FROM characters WHERE id = $1", s.charID, []byte{0x00}).Scan(&data)
	if err != nil {
		panic("Failed to get gacha_items")
	}
	// limit of 36 items are returned
	if data[0] > 36 {
		outData := make([]byte, 181)
		copy(outData, data[0:181])
		outData[0] = byte(36)
		saveData := append(data[:1], data[181:]...)
		saveData[0] = saveData[0] - 36
		doAckBufSucceed(s, pkt.AckHandle, outData)
		if pkt.Unk0 != 0x2401 {
			_, err := s.server.db.Exec("UPDATE characters SET gacha_items = $2 WHERE id = $1", s.charID, saveData)
			if err != nil {
				s.logger.Fatal("Failed to update gacha_items in db", zap.Error(err))
			}
		}
	} else {
		doAckBufSucceed(s, pkt.AckHandle, data)
		if pkt.Unk0 != 0x2401 {
			_, err := s.server.db.Exec("UPDATE characters SET gacha_items = null WHERE id = $1", s.charID)
			if err != nil {
				s.logger.Fatal("Failed to update gacha_items in db", zap.Error(err))
			}
		}
	}
}

func handleMsgMhfGetStepupStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetStepupStatus)
	// get the reset time from db
	var step_progression int
	var step_time time.Time
	err := s.server.db.QueryRow(`SELECT COALESCE(step_progression, 0), COALESCE(step_time, $1) FROM stepup_state WHERE char_id = $2 AND shophash = $3`, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), s.charID, pkt.GachaHash).Scan(&step_progression, &step_time)
	if err != nil {
		s.logger.Fatal("Failed to Select coalesce in db", zap.Error(err))
	}

	// calculate next midday
	var t = time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	year, month, day := t.Date()
	midday := time.Date(year, month, day, 12, 0, 0, 0, t.Location())
	if t.After(midday) {
		midday = midday.Add(24 * time.Hour)
	}
	// after midday or not set
	if t.After(step_time) {
		step_progression = 0
	}
	_, err = s.server.db.Exec(`INSERT INTO stepup_state (shophash, step_progression, step_time, char_id)
														 VALUES ($1,$2,$3,$4) ON CONFLICT (shophash, char_id)
														 DO UPDATE SET step_progression=$2, step_time=$3
														 WHERE EXCLUDED.char_id=$4 AND EXCLUDED.shophash=$1`, pkt.GachaHash, step_progression, midday, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(step_progression))
	resp.WriteUint32(uint32(time.Now().In(time.FixedZone("UTC+9", 9*60*60)).Unix()))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfPlayFreeGacha(s *Session, p mhfpacket.MHFPacket) {
	// not sure this is used anywhere, free gachas use the MSG_MHF_PLAY_NORMAL_GACHA method in captures
}

func handleMsgMhfGetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoxGachaInfo)
	count := 0
	var used_itemhash pq.Int64Array
	// pull array of used values
	// single sized respone with 0x00 is a valid with no items present
	_ = s.server.db.QueryRow("SELECT used_itemhash FROM lucky_box_state WHERE shophash=$1 AND char_id=$2", pkt.GachaHash, s.charID).Scan((*pq.Int64Array)(&used_itemhash))
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(0)
	for ind := range used_itemhash {
		resp.WriteUint32(uint32(used_itemhash[ind]))
		resp.WriteUint8(1)
		count++
	}
	resp.Seek(0, 0)
	resp.WriteUint8(uint8(count))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfPlayBoxGacha(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPlayBoxGacha)
	// needs to query db for input gacha and return a result or number of results
	// uint8 number of results
	// uint8 item type
	// uint16 item id
	// uint16 quantity

	var currType, rarityIcon, rollsCount, itemCount byte
	var currQuant, currNumber, percentage uint16
	var itemhash uint32
	var itemType, itemId, quantity, usedItemHash pq.Int64Array
	var items []lottery.Weighter
	// get info for updating data and calculating costs
	err := s.server.db.QueryRow("SELECT currType, currNumber, currQuant, rollsCount FROM gacha_shop_items WHERE shophash=$1 AND entryType=$2", pkt.GachaHash, pkt.RollType).Scan(&currType, &currNumber, &currQuant, &rollsCount)
	if err != nil {
		panic(err)
	}
	// get existing items in storage if any
	var data []byte
	_ = s.server.db.QueryRow("SELECT gacha_items FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		data = []byte{0x00}
	}
	// get gacha items and iterate through them for gacha roll
	shopEntries, err := s.server.db.Query(`SELECT itemhash, percentage, rarityIcon, itemCount, itemType, itemId, quantity
																				 FROM gacha_shop_items
																				 WHERE shophash=$1 AND entryType=100
																				 EXCEPT ALL SELECT itemhash, percentage, rarityIcon, itemCount, itemType, itemId, quantity
																				 FROM gacha_shop_items gsi JOIN lucky_box_state lbs ON gsi.itemhash = ANY(lbs.used_itemhash)
																				 WHERE lbs.char_id=$2`, pkt.GachaHash, s.charID)
	if err != nil {
		panic(err)
	}
	for shopEntries.Next() {
		err = shopEntries.Scan(&itemhash, &percentage, &rarityIcon, &itemCount, (*pq.Int64Array)(&itemType), (*pq.Int64Array)(&itemId), (*pq.Int64Array)(&quantity))
		if err != nil {
			panic(err)
		}
		items = append(items, &gachaItem{itemhash: itemhash, percentage: percentage, rarityIcon: rarityIcon, itemCount: itemCount, itemType: itemType, itemId: itemId, quantity: quantity})
	}
	// execute rolls, build response and update database
	results := byte(0)
	resp := byteframe.NewByteFrame()
	dbUpdate := byteframe.NewByteFrame()
	resp.WriteUint8(0) // results go here later
	l := lottery.NewDefaultLottery()
	for x := 0; x < int(rollsCount); x++ {
		ind := l.Draw(items)
		results += items[ind].(*gachaItem).itemCount
		for y := 0; y < int(items[ind].(*gachaItem).itemCount); y++ {
			// items in storage don't get rarity
			dbUpdate.WriteUint8(uint8(items[ind].(*gachaItem).itemType[y]))
			dbUpdate.WriteUint16(uint16(items[ind].(*gachaItem).itemId[y]))
			dbUpdate.WriteUint16(uint16(items[ind].(*gachaItem).quantity[y]))
			data = append(data, dbUpdate.Data()...)
			dbUpdate.Seek(0, 0)
			// response needs all item info and the rarity
			resp.WriteBytes(dbUpdate.Data())
			resp.WriteUint8(uint8(items[ind].(*gachaItem).rarityIcon))

			usedItemHash = append(usedItemHash, int64(items[ind].(*gachaItem).itemhash))
		}
		// remove rolled
		items = append(items[:ind], items[ind+1:]...)
	}
	resp.Seek(0, 0)
	resp.WriteUint8(uint8(results))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())

	// add claimables to DB
	data[0] = data[0] + results
	_, err = s.server.db.Exec("UPDATE characters SET gacha_items = $1 WHERE id = $2", data, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update gacha_items in db", zap.Error(err))
	}
	// update lucky_box_state
	_, err = s.server.db.Exec(`	INSERT INTO lucky_box_state (char_id, shophash, used_itemhash)
								VALUES ($1,$2,$3) ON CONFLICT (char_id, shophash)
								DO UPDATE SET used_itemhash = COALESCE(lucky_box_state.used_itemhash::int[] || $3::int[], $3::int[])
								WHERE EXCLUDED.char_id=$1 AND EXCLUDED.shophash=$2`, s.charID, pkt.GachaHash, usedItemHash)
	if err != nil {
		s.logger.Fatal("Failed to update lucky box state in db", zap.Error(err))
	}
	// deduct gacha coins if relevant, items are handled fine by the standard savedata packet immediately afterwards
	if currType == 19 {
		_, err = s.server.db.Exec(`	UPDATE characters
									SET gacha_trial = CASE WHEN (gacha_trial > $1) then gacha_trial - $1 else gacha_trial end, gacha_prem = CASE WHEN NOT (gacha_trial > $1) then gacha_prem - $1 else gacha_prem end
									WHERE id=$2`, currNumber, s.charID)
	}
	if err != nil {
		s.logger.Fatal("Failed to update gacha_trial in db", zap.Error(err))
	}
}

func handleMsgMhfResetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfResetBoxGachaInfo)
	_, err := s.server.db.Exec("DELETE FROM lucky_box_state WHERE shophash=$1 AND char_id=$2", pkt.GachaHash, s.charID)
	if err != nil {
		panic(err)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
