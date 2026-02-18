package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	"erupe-ce/common/mhfmon"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func handleMsgMhfTransferItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfTransferItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumeratePrice(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumeratePrice)
	bf := byteframe.NewByteFrame()
	var lbPrices []struct {
		Unk0 uint16
		Unk1 uint16
		Unk2 uint32
	}
	var wantedList []struct {
		Unk0 uint32
		Unk1 uint32
		Unk2 uint32
		Unk3 uint16
		Unk4 uint16
		Unk5 uint16
		Unk6 uint16
		Unk7 uint16
		Unk8 uint16
		Unk9 uint16
	}
	gzPrices := []struct {
		Unk0  uint16
		Gz    uint16
		Unk1  uint16
		Unk2  uint16
		MonID uint16
		Unk3  uint16
		Unk4  uint8
	}{
		{0, 1000, 0, 0, mhfmon.Pokaradon, 100, 1},
		{0, 800, 0, 0, mhfmon.YianKutKu, 100, 1},
		{0, 800, 0, 0, mhfmon.DaimyoHermitaur, 100, 1},
		{0, 1100, 0, 0, mhfmon.Farunokku, 100, 1},
		{0, 900, 0, 0, mhfmon.Congalala, 100, 1},
		{0, 900, 0, 0, mhfmon.Gypceros, 100, 1},
		{0, 1300, 0, 0, mhfmon.Hyujikiki, 100, 1},
		{0, 1000, 0, 0, mhfmon.Basarios, 100, 1},
		{0, 1000, 0, 0, mhfmon.Rathian, 100, 1},
		{0, 800, 0, 0, mhfmon.ShogunCeanataur, 100, 1},
		{0, 1400, 0, 0, mhfmon.Midogaron, 100, 1},
		{0, 900, 0, 0, mhfmon.Blangonga, 100, 1},
		{0, 1100, 0, 0, mhfmon.Rathalos, 100, 1},
		{0, 1000, 0, 0, mhfmon.Khezu, 100, 1},
		{0, 1600, 0, 0, mhfmon.Giaorugu, 100, 1},
		{0, 1100, 0, 0, mhfmon.Gravios, 100, 1},
		{0, 1400, 0, 0, mhfmon.Tigrex, 100, 1},
		{0, 1000, 0, 0, mhfmon.Pariapuria, 100, 1},
		{0, 1700, 0, 0, mhfmon.Anorupatisu, 100, 1},
		{0, 1500, 0, 0, mhfmon.Lavasioth, 100, 1},
		{0, 1500, 0, 0, mhfmon.Espinas, 100, 1},
		{0, 1600, 0, 0, mhfmon.Rajang, 100, 1},
		{0, 1800, 0, 0, mhfmon.Rebidiora, 100, 1},
		{0, 1100, 0, 0, mhfmon.YianGaruga, 100, 1},
		{0, 1500, 0, 0, mhfmon.AqraVashimu, 100, 1},
		{0, 1600, 0, 0, mhfmon.Gurenzeburu, 100, 1},
		{0, 1500, 0, 0, mhfmon.Dyuragaua, 100, 1},
		{0, 1300, 0, 0, mhfmon.Gougarf, 100, 1},
		{0, 1000, 0, 0, mhfmon.Shantien, 100, 1},
		{0, 1800, 0, 0, mhfmon.Disufiroa, 100, 1},
		{0, 600, 0, 0, mhfmon.Velocidrome, 100, 1},
		{0, 600, 0, 0, mhfmon.Gendrome, 100, 1},
		{0, 700, 0, 0, mhfmon.Iodrome, 100, 1},
		{0, 1700, 0, 0, mhfmon.Baruragaru, 100, 1},
		{0, 800, 0, 0, mhfmon.Cephadrome, 100, 1},
		{0, 1000, 0, 0, mhfmon.Plesioth, 100, 1},
		{0, 1800, 0, 0, mhfmon.Zerureusu, 100, 1},
		{0, 1100, 0, 0, mhfmon.Diablos, 100, 1},
		{0, 1600, 0, 0, mhfmon.Berukyurosu, 100, 1},
		{0, 2000, 0, 0, mhfmon.Fatalis, 100, 1},
		{0, 1500, 0, 0, mhfmon.BlackGravios, 100, 1},
		{0, 1600, 0, 0, mhfmon.GoldRathian, 100, 1},
		{0, 1900, 0, 0, mhfmon.Meraginasu, 100, 1},
		{0, 700, 0, 0, mhfmon.Bulldrome, 100, 1},
		{0, 900, 0, 0, mhfmon.NonoOrugaron, 100, 1},
		{0, 1600, 0, 0, mhfmon.KamuOrugaron, 100, 1},
		{0, 1700, 0, 0, mhfmon.Forokururu, 100, 1},
		{0, 1900, 0, 0, mhfmon.Diorex, 100, 1},
		{0, 1500, 0, 0, mhfmon.AqraJebia, 100, 1},
		{0, 1600, 0, 0, mhfmon.SilverRathalos, 100, 1},
		{0, 2400, 0, 0, mhfmon.CrimsonFatalis, 100, 1},
		{0, 2000, 0, 0, mhfmon.Inagami, 100, 1},
		{0, 2100, 0, 0, mhfmon.GarubaDaora, 100, 1},
		{0, 900, 0, 0, mhfmon.Monoblos, 100, 1},
		{0, 1000, 0, 0, mhfmon.RedKhezu, 100, 1},
		{0, 900, 0, 0, mhfmon.Hypnocatrice, 100, 1},
		{0, 1700, 0, 0, mhfmon.PearlEspinas, 100, 1},
		{0, 900, 0, 0, mhfmon.PurpleGypceros, 100, 1},
		{0, 1800, 0, 0, mhfmon.Poborubarumu, 100, 1},
		{0, 1900, 0, 0, mhfmon.Lunastra, 100, 1},
		{0, 1600, 0, 0, mhfmon.Kuarusepusu, 100, 1},
		{0, 1100, 0, 0, mhfmon.PinkRathian, 100, 1},
		{0, 1200, 0, 0, mhfmon.AzureRathalos, 100, 1},
		{0, 1800, 0, 0, mhfmon.Varusaburosu, 100, 1},
		{0, 1000, 0, 0, mhfmon.Gogomoa, 100, 1},
		{0, 1600, 0, 0, mhfmon.BurningEspinas, 100, 1},
		{0, 2000, 0, 0, mhfmon.Harudomerugu, 100, 1},
		{0, 1800, 0, 0, mhfmon.Akantor, 100, 1},
		{0, 900, 0, 0, mhfmon.BrightHypnoc, 100, 1},
		{0, 2200, 0, 0, mhfmon.Gureadomosu, 100, 1},
		{0, 1200, 0, 0, mhfmon.GreenPlesioth, 100, 1},
		{0, 2400, 0, 0, mhfmon.Zinogre, 100, 1},
		{0, 1900, 0, 0, mhfmon.Gasurabazura, 100, 1},
		{0, 1300, 0, 0, mhfmon.Abiorugu, 100, 1},
		{0, 1200, 0, 0, mhfmon.BlackDiablos, 100, 1},
		{0, 1000, 0, 0, mhfmon.WhiteMonoblos, 100, 1},
		{0, 3000, 0, 0, mhfmon.Deviljho, 100, 1},
		{0, 2300, 0, 0, mhfmon.YamaKurai, 100, 1},
		{0, 2800, 0, 0, mhfmon.Brachydios, 100, 1},
		{0, 1700, 0, 0, mhfmon.Toridcless, 100, 1},
		{0, 1100, 0, 0, mhfmon.WhiteHypnoc, 100, 1},
		{0, 1500, 0, 0, mhfmon.RedLavasioth, 100, 1},
		{0, 2200, 0, 0, mhfmon.Barioth, 100, 1},
		{0, 1800, 0, 0, mhfmon.Odibatorasu, 100, 1},
		{0, 1600, 0, 0, mhfmon.Doragyurosu, 100, 1},
		{0, 900, 0, 0, mhfmon.BlueYianKutKu, 100, 1},
		{0, 2300, 0, 0, mhfmon.ToaTesukatora, 100, 1},
		{0, 2000, 0, 0, mhfmon.Uragaan, 100, 1},
		{0, 1900, 0, 0, mhfmon.Teostra, 100, 1},
		{0, 1700, 0, 0, mhfmon.Chameleos, 100, 1},
		{0, 1800, 0, 0, mhfmon.KushalaDaora, 100, 1},
		{0, 2100, 0, 0, mhfmon.Nargacuga, 100, 1},
		{0, 2600, 0, 0, mhfmon.Guanzorumu, 100, 1},
		{0, 1900, 0, 0, mhfmon.Kirin, 100, 1},
		{0, 2000, 0, 0, mhfmon.Rukodiora, 100, 1},
		{0, 2700, 0, 0, mhfmon.StygianZinogre, 100, 1},
		{0, 2200, 0, 0, mhfmon.Voljang, 100, 1},
		{0, 1800, 0, 0, mhfmon.Zenaserisu, 100, 1},
		{0, 3100, 0, 0, mhfmon.GoreMagala, 100, 1},
		{0, 3200, 0, 0, mhfmon.ShagaruMagala, 100, 1},
		{0, 3500, 0, 0, mhfmon.Eruzerion, 100, 1},
		{0, 3200, 0, 0, mhfmon.Amatsu, 100, 1},
	}

	bf.WriteUint16(uint16(len(lbPrices)))
	for _, lb := range lbPrices {
		bf.WriteUint16(lb.Unk0)
		bf.WriteUint16(lb.Unk1)
		bf.WriteUint32(lb.Unk2)
	}
	bf.WriteUint16(uint16(len(wantedList)))
	for _, wanted := range wantedList {
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
	bf.WriteUint8(uint8(len(gzPrices)))
	for _, gz := range gzPrices {
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

func handleMsgMhfGetExtraInfo(s *Session, p mhfpacket.MHFPacket) {}

func userGetItems(s *Session) []mhfitem.MHFItemStack {
	var data []byte
	var items []mhfitem.MHFItemStack
	_ = s.server.db.QueryRow(`SELECT item_box FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.charID).Scan(&data)
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
	if _, err := s.server.db.Exec(`UPDATE users u SET item_box=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, mhfitem.SerializeWarehouseItems(newStacks), s.charID); err != nil {
		s.logger.Error("Failed to update union item box", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetCogInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCheckWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckWeeklyStamp)
	if pkt.StampType != "hl" && pkt.StampType != "ex" {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 14))
		return
	}
	var total, redeemed, updated uint16
	var lastCheck time.Time
	err := s.server.db.QueryRow(fmt.Sprintf("SELECT %s_checked FROM stamps WHERE character_id=$1", pkt.StampType), s.charID).Scan(&lastCheck)
	if err != nil {
		lastCheck = TimeAdjusted()
		if _, err := s.server.db.Exec("INSERT INTO stamps (character_id, hl_checked, ex_checked) VALUES ($1, $2, $2)", s.charID, TimeAdjusted()); err != nil {
			s.logger.Error("Failed to insert stamps record", zap.Error(err))
		}
	} else {
		if _, err := s.server.db.Exec(fmt.Sprintf(`UPDATE stamps SET %s_checked=$1 WHERE character_id=$2`, pkt.StampType), TimeAdjusted(), s.charID); err != nil {
			s.logger.Error("Failed to update stamp check time", zap.Error(err))
		}
	}

	if lastCheck.Before(TimeWeekStart()) {
		if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE stamps SET %s_total=%s_total+1 WHERE character_id=$1", pkt.StampType, pkt.StampType), s.charID); err != nil {
			s.logger.Error("Failed to increment stamp total", zap.Error(err))
		}
		updated = 1
	}

	_ = s.server.db.QueryRow(fmt.Sprintf("SELECT %s_total, %s_redeemed FROM stamps WHERE character_id=$1", pkt.StampType, pkt.StampType), s.charID).Scan(&total, &redeemed)
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
	var tktStack mhfitem.MHFItemStack
	if pkt.ExchangeType == 10 { // Yearly Sub Ex
		_ = s.server.db.QueryRow("UPDATE stamps SET hl_total=hl_total-48, hl_redeemed=hl_redeemed-48 WHERE character_id=$1 RETURNING hl_total, hl_redeemed", s.charID).Scan(&total, &redeemed)
		tktStack = mhfitem.MHFItemStack{Item: mhfitem.MHFItem{ItemID: 2210}, Quantity: 1}
	} else {
		_ = s.server.db.QueryRow(fmt.Sprintf("UPDATE stamps SET %s_redeemed=%s_redeemed+8 WHERE character_id=$1 RETURNING %s_total, %s_redeemed", pkt.StampType, pkt.StampType, pkt.StampType, pkt.StampType), s.charID).Scan(&total, &redeemed)
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
	if _config.ErupeConfig.RealClientMode <= _config.Z1 {
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
	if _config.ErupeConfig.RealClientMode >= _config.G1 {
		bf.WriteUint16(pkt.GR)
	}
	var stamps, rewardTier, rewardUnk uint16
	reward := mhfitem.MHFItemStack{Item: mhfitem.MHFItem{}}
	_ = s.server.db.QueryRow(`UPDATE characters SET stampcard = stampcard + $1 WHERE id = $2 RETURNING stampcard`, pkt.Stamps, s.charID).Scan(&stamps)
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

func handleMsgMhfStampcardPrize(s *Session, p mhfpacket.MHFPacket) {}
