package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

func handleMsgMhfTransferItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfTransferItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumeratePrice(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumeratePrice)
	bf := byteframe.NewByteFrame()

	bf.WriteUint16(uint16(len(enumeratePriceLB)))
	for _, lb := range enumeratePriceLB {
		bf.WriteUint16(lb.Unk0)
		bf.WriteUint16(lb.Unk1)
		bf.WriteUint32(lb.Unk2)
	}
	bf.WriteUint16(uint16(len(enumeratePriceWanted)))
	for _, wanted := range enumeratePriceWanted {
		bf.WriteUint32(wanted.Unk0)
		bf.WriteUint32(wanted.Unk1)
		bf.WriteUint32(wanted.Unk2)
		bf.WriteUint16(wanted.Unk3)
		bf.WriteUint16(wanted.Unk4)
		bf.WriteUint16(wanted.Unk5)
		bf.WriteUint16(wanted.Unk6)
		bf.WriteUint16(wanted.Unk7)
		bf.WriteUint16(wanted.Unk8)
		bf.WriteUint16(wanted.Unk9)
	}
	bf.WriteUint8(uint8(len(enumeratePriceGZ)))
	for _, gz := range enumeratePriceGZ {
		bf.WriteUint16(gz.Unk0)
		bf.WriteUint16(gz.Gz)
		bf.WriteUint16(gz.Unk1)
		bf.WriteUint16(gz.Unk2)
		bf.WriteUint16(gz.MonID)
		bf.WriteUint16(gz.Unk3)
		bf.WriteUint8(gz.Unk4)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateOrder(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateOrder)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfGetExtraInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetExtraInfo)
	// TODO: response structure unknown; fail ACK prevents softlock without misleading client
	doAckBufFail(s, pkt.AckHandle, nil)
}

func userGetItems(s *Session) []mhfitem.MHFItemStack {
	var items []mhfitem.MHFItemStack
	data, err := s.server.userRepo.GetItemBox(s.userID)
	if err != nil {
		s.logger.Warn("Failed to load user item box", zap.Error(err))
	}
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			items = append(items, mhfitem.ReadWarehouseItem(box))
		}
	}
	return items
}

func handleMsgMhfEnumerateUnionItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateUnionItem)
	items := userGetItems(s)
	bf := byteframe.NewByteFrame()
	bf.WriteBytes(mhfitem.SerializeWarehouseItems(items))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateUnionItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateUnionItem)
	newStacks := mhfitem.DiffItemStacks(userGetItems(s), pkt.UpdatedItems)
	if err := s.server.userRepo.SetItemBox(s.userID, mhfitem.SerializeWarehouseItems(newStacks)); err != nil {
		s.logger.Error("Failed to update union item box", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetCogInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetCogInfo)
	// TODO: response structure unknown; fail ACK prevents softlock without misleading client
	doAckBufFail(s, pkt.AckHandle, nil)
}

func handleMsgMhfCheckWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckWeeklyStamp)
	if pkt.StampType != "hl" && pkt.StampType != "ex" {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 14))
		return
	}
	var total, redeemed, updated uint16
	lastCheck, err := s.server.stampRepo.GetChecked(s.charID, pkt.StampType)
	if err != nil {
		lastCheck = TimeAdjusted()
		if err := s.server.stampRepo.Init(s.charID, TimeAdjusted()); err != nil {
			s.logger.Error("Failed to insert stamps record", zap.Error(err))
		}
	} else {
		if err := s.server.stampRepo.SetChecked(s.charID, pkt.StampType, TimeAdjusted()); err != nil {
			s.logger.Error("Failed to update stamp check time", zap.Error(err))
		}
	}

	if lastCheck.Before(TimeWeekStart()) {
		if err := s.server.stampRepo.IncrementTotal(s.charID, pkt.StampType); err != nil {
			s.logger.Error("Failed to increment stamp total", zap.Error(err))
		}
		updated = 1
	}

	total, redeemed, err = s.server.stampRepo.GetTotals(s.charID, pkt.StampType)
	if err != nil {
		s.logger.Warn("Failed to get stamp totals", zap.Error(err))
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(total)
	bf.WriteUint16(redeemed)
	bf.WriteUint16(updated)
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	bf.WriteUint32(uint32(TimeWeekStart().Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfExchangeWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeWeeklyStamp)
	if pkt.StampType != "hl" && pkt.StampType != "ex" {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 12))
		return
	}
	var total, redeemed uint16
	var err error
	var tktStack mhfitem.MHFItemStack
	if pkt.ExchangeType == 10 { // Yearly Sub Ex
		if total, redeemed, err = s.server.stampRepo.ExchangeYearly(s.charID); err != nil {
			s.logger.Error("Failed to update yearly stamp exchange", zap.Error(err))
			doAckBufFail(s, pkt.AckHandle, nil)
			return
		}
		tktStack = mhfitem.MHFItemStack{Item: mhfitem.MHFItem{ItemID: 2210}, Quantity: 1}
	} else {
		if total, redeemed, err = s.server.stampRepo.Exchange(s.charID, pkt.StampType); err != nil {
			s.logger.Error("Failed to update stamp redemption", zap.Error(err))
			doAckBufFail(s, pkt.AckHandle, nil)
			return
		}
		if pkt.StampType == "hl" {
			tktStack = mhfitem.MHFItemStack{Item: mhfitem.MHFItem{ItemID: 1630}, Quantity: 5}
		} else {
			tktStack = mhfitem.MHFItemStack{Item: mhfitem.MHFItem{ItemID: 1631}, Quantity: 5}
		}
	}
	addWarehouseItem(s, tktStack)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(total)
	bf.WriteUint16(redeemed)
	bf.WriteUint16(0)
	bf.WriteUint16(tktStack.Item.ItemID)
	bf.WriteUint16(tktStack.Quantity)
	bf.WriteUint32(uint32(TimeWeekStart().Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfStampcardStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStampcardStamp)

	rewards := []struct {
		HR        uint16
		Item1     uint16
		Quantity1 uint16
		Item2     uint16
		Quantity2 uint16
	}{
		{0, 6164, 1, 6164, 2},
		{50, 6164, 2, 6164, 3},
		{100, 6164, 3, 5392, 1},
		{300, 5392, 1, 5392, 3},
		{999, 5392, 1, 5392, 4},
	}
	if s.server.erupeConfig.RealClientMode <= cfg.Z1 {
		for _, reward := range rewards {
			if pkt.HR >= reward.HR {
				pkt.Item1 = reward.Item1
				pkt.Quantity1 = reward.Quantity1
				pkt.Item2 = reward.Item2
				pkt.Quantity2 = reward.Quantity2
			}
		}
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint16(pkt.HR)
	if s.server.erupeConfig.RealClientMode >= cfg.G1 {
		bf.WriteUint16(pkt.GR)
	}
	var stamps, rewardTier, rewardUnk uint16
	reward := mhfitem.MHFItemStack{Item: mhfitem.MHFItem{}}
	stamps32, err := s.server.charRepo.AdjustInt(s.charID, "stampcard", int(pkt.Stamps))
	stamps = uint16(stamps32)
	if err != nil {
		s.logger.Error("Failed to update stampcard", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, nil)
		return
	}
	bf.WriteUint16(stamps - pkt.Stamps)
	bf.WriteUint16(stamps)

	if stamps/30 > (stamps-pkt.Stamps)/30 {
		rewardTier = 2
		rewardUnk = pkt.Reward2
		reward = mhfitem.MHFItemStack{Item: mhfitem.MHFItem{ItemID: pkt.Item2}, Quantity: pkt.Quantity2}
		addWarehouseItem(s, reward)
	} else if stamps/15 > (stamps-pkt.Stamps)/15 {
		rewardTier = 1
		rewardUnk = pkt.Reward1
		reward = mhfitem.MHFItemStack{Item: mhfitem.MHFItem{ItemID: pkt.Item1}, Quantity: pkt.Quantity1}
		addWarehouseItem(s, reward)
	}

	bf.WriteUint16(rewardTier)
	bf.WriteUint16(rewardUnk)
	bf.WriteUint16(reward.Item.ItemID)
	bf.WriteUint16(reward.Quantity)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfStampcardPrize(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented
