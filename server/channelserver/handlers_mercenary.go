package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
	"io"
	"time"
)

func handleMsgMhfLoadPartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPartner)
	var data []byte
	err := s.server.db.QueryRow("SELECT partner FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		s.logger.Error("Failed to load partner", zap.Error(err))
		data = make([]byte, 9)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSavePartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePartner)
	dumpSaveData(s, pkt.RawDataPayload, "partner")
	_, err := s.server.db.Exec("UPDATE characters SET partner=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Error("Failed to save partner", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadLegendDispatch(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadLegendDispatch)
	bf := byteframe.NewByteFrame()
	legendDispatch := []struct {
		Unk       uint32
		Timestamp uint32
	}{
		{0, uint32(TimeMidnight().Add(-12 * time.Hour).Unix())},
		{0, uint32(TimeMidnight().Add(12 * time.Hour).Unix())},
		{0, uint32(TimeMidnight().Add(36 * time.Hour).Unix())},
	}
	bf.WriteUint8(uint8(len(legendDispatch)))
	for _, dispatch := range legendDispatch {
		bf.WriteUint32(dispatch.Unk)
		bf.WriteUint32(dispatch.Timestamp)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfLoadHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHunterNavi)
	naviLength := 552
	if s.server.erupeConfig.RealClientMode <= _config.G7 {
		naviLength = 280
	}
	var data []byte
	err := s.server.db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		s.logger.Error("Failed to load hunternavi", zap.Error(err))
		data = make([]byte, naviLength)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSaveHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveHunterNavi)
	if pkt.IsDataDiff {
		naviLength := 552
		if s.server.erupeConfig.RealClientMode <= _config.G7 {
			naviLength = 280
		}
		var data []byte
		// Load existing save
		err := s.server.db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Error("Failed to load hunternavi", zap.Error(err))
		}

		// Check if we actually had any hunternavi data, using a blank buffer if not.
		// This is requried as the client will try to send a diff after character creation without a prior MsgMhfSaveHunterNavi packet.
		if len(data) == 0 {
			data = make([]byte, naviLength)
		}

		// Perform diff and compress it to write back to db
		s.logger.Info("Diffing...")
		saveOutput := deltacomp.ApplyDataDiff(pkt.RawDataPayload, data)
		_, err = s.server.db.Exec("UPDATE characters SET hunternavi=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Error("Failed to save hunternavi", zap.Error(err))
		}
		s.logger.Info("Wrote recompressed hunternavi back to DB")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "hunternavi")
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET hunternavi=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Error("Failed to save hunternavi", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfMercenaryHuntdata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfMercenaryHuntdata)
	if pkt.Unk0 == 1 {
		// Format:
		// uint8 Hunts
		// struct Hunt
		//   uint32 HuntID
		//   uint32 MonID
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0))
	}
}

func handleMsgMhfEnumerateMercenaryLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateMercenaryLog)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	// Format:
	// struct Log
	//   uint32 Timestamp
	//   []byte Name (len 18)
	//   uint8 Unk
	//   uint8 Unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfCreateMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateMercenary)
	bf := byteframe.NewByteFrame()
	var nextID uint32
	_ = s.server.db.QueryRow("SELECT nextval('rasta_id_seq')").Scan(&nextID)
	s.server.db.Exec("UPDATE characters SET rasta_id=$1 WHERE id=$2", nextID, s.charID)
	bf.WriteUint32(nextID)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSaveMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMercenary)
	dumpSaveData(s, pkt.MercData, "mercenary")
	if len(pkt.MercData) > 0 {
		temp := byteframe.NewByteFrameFromBytes(pkt.MercData)
		s.server.db.Exec("UPDATE characters SET savemercenary=$1, rasta_id=$2 WHERE id=$3", pkt.MercData, temp.ReadUint32(), s.charID)
	}
	s.server.db.Exec("UPDATE characters SET gcp=$1, pact_id=$2 WHERE id=$3", pkt.GCP, pkt.PactMercID, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadMercenaryW(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryW)
	bf := byteframe.NewByteFrame()

	var pactID, cid uint32
	var name string
	s.server.db.QueryRow("SELECT pact_id FROM characters WHERE id=$1", s.charID).Scan(&pactID)
	if pactID > 0 {
		s.server.db.QueryRow("SELECT name, id FROM characters WHERE rasta_id = $1", pactID).Scan(&name, &cid)
		bf.WriteUint8(1) // numLends
		bf.WriteUint32(pactID)
		bf.WriteUint32(cid)
		bf.WriteBool(true) // Escort enabled
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
		bf.WriteUint32(uint32(TimeAdjusted().Add(time.Hour * 24 * 7).Unix()))
		bf.WriteBytes(stringsupport.PaddedString(name, 18, true))
	} else {
		bf.WriteUint8(0)
	}

	if pkt.Op != 2 && pkt.Op != 5 {
		var loans uint8
		temp := byteframe.NewByteFrame()
		rows, _ := s.server.db.Query("SELECT name, id, pact_id FROM characters WHERE pact_id=(SELECT rasta_id FROM characters WHERE id=$1)", s.charID)
		for rows.Next() {
			err := rows.Scan(&name, &cid, &pactID)
			if err != nil {
				continue
			}
			loans++
			temp.WriteUint32(pactID)
			temp.WriteUint32(cid)
			temp.WriteUint32(uint32(TimeAdjusted().Unix()))
			temp.WriteUint32(uint32(TimeAdjusted().Add(time.Hour * 24 * 7).Unix()))
			temp.WriteBytes(stringsupport.PaddedString(name, 18, true))
		}
		bf.WriteUint8(loans)
		bf.WriteBytes(temp.Data())

		if pkt.Op != 1 && pkt.Op != 4 {
			var data []byte
			var gcp uint32
			s.server.db.QueryRow("SELECT savemercenary FROM characters WHERE id=$1", s.charID).Scan(&data)
			s.server.db.QueryRow("SELECT COALESCE(gcp, 0) FROM characters WHERE id=$1", s.charID).Scan(&gcp)

			if len(data) == 0 {
				bf.WriteBool(false)
			} else {
				bf.WriteBool(true)
				bf.WriteBytes(data)
			}
			bf.WriteUint32(gcp)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadMercenaryM(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryM)
	var data []byte
	s.server.db.QueryRow("SELECT savemercenary FROM characters WHERE id = $1", pkt.CharID).Scan(&data)
	resp := byteframe.NewByteFrame()
	if len(data) == 0 {
		resp.WriteBool(false)
	} else {
		resp.WriteBytes(data)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfContractMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfContractMercenary)
	switch pkt.Op {
	case 0: // Form loan
		s.server.db.Exec("UPDATE characters SET pact_id=$1 WHERE id=$2", pkt.PactMercID, pkt.CID)
	case 1: // Cancel lend
		s.server.db.Exec("UPDATE characters SET pact_id=0 WHERE id=$1", s.charID)
	case 2: // Cancel loan
		s.server.db.Exec("UPDATE characters SET pact_id=0 WHERE id=$1", pkt.CID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadOtomoAirou)
	var data []byte
	err := s.server.db.QueryRow("SELECT otomoairou FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		s.logger.Error("Failed to load otomoairou", zap.Error(err))
		data = make([]byte, 10)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSaveOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveOtomoAirou)
	dumpSaveData(s, pkt.RawDataPayload, "otomoairou")
	decomp, err := nullcomp.Decompress(pkt.RawDataPayload[1:])
	if err != nil {
		s.logger.Error("Failed to decompress airou", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrameFromBytes(decomp)
	save := byteframe.NewByteFrame()
	var catsExist uint8
	save.WriteUint8(0)

	cats := bf.ReadUint8()
	for i := 0; i < int(cats); i++ {
		dataLen := bf.ReadUint32()
		catID := bf.ReadUint32()
		if catID == 0 {
			_ = s.server.db.QueryRow("SELECT nextval('airou_id_seq')").Scan(&catID)
		}
		exists := bf.ReadBool()
		data := bf.ReadBytes(uint(dataLen) - 5)
		if exists {
			catsExist++
			save.WriteUint32(dataLen)
			save.WriteUint32(catID)
			save.WriteBool(exists)
			save.WriteBytes(data)
		}
	}
	save.WriteBytes(bf.DataFromCurrent())
	save.Seek(0, 0)
	save.WriteUint8(catsExist)
	comp, err := nullcomp.Compress(save.Data())
	if err != nil {
		s.logger.Error("Failed to compress airou", zap.Error(err))
	} else {
		comp = append([]byte{0x01}, comp...)
		s.server.db.Exec("UPDATE characters SET otomoairou=$1 WHERE id=$2", comp, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateAiroulist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateAiroulist)
	resp := byteframe.NewByteFrame()
	airouList := getGuildAirouList(s)
	resp.WriteUint16(uint16(len(airouList)))
	resp.WriteUint16(uint16(len(airouList)))
	for _, cat := range airouList {
		resp.WriteUint32(cat.ID)
		resp.WriteBytes(cat.Name)
		resp.WriteUint32(cat.Experience)
		resp.WriteUint8(cat.Personality)
		resp.WriteUint8(cat.Class)
		resp.WriteUint8(cat.WeaponType)
		resp.WriteUint16(cat.WeaponID)
		resp.WriteUint32(0) // 32 bit unix timestamp, either time at which the cat stops being fatigued or the time at which it started
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

type Airou struct {
	ID          uint32
	Name        []byte
	Task        uint8
	Personality uint8
	Class       uint8
	Experience  uint32
	WeaponType  uint8
	WeaponID    uint16
}

func getGuildAirouList(s *Session) []Airou {
	var guildCats []Airou
	bannedCats := make(map[uint32]int)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil {
		return guildCats
	}
	rows, err := s.server.db.Query(`SELECT cats_used FROM guild_hunts gh
		INNER JOIN characters c ON gh.host_id = c.id WHERE c.id=$1
	`, s.charID)
	if err != nil {
		s.logger.Warn("Failed to get recently used airous", zap.Error(err))
		return guildCats
	}

	var csvTemp string
	var startTemp time.Time
	for rows.Next() {
		err = rows.Scan(&csvTemp, &startTemp)
		if err != nil {
			continue
		}
		if startTemp.Add(time.Second * time.Duration(s.server.erupeConfig.GameplayOptions.TreasureHuntPartnyaCooldown)).Before(TimeAdjusted()) {
			for i, j := range stringsupport.CSVElems(csvTemp) {
				bannedCats[uint32(j)] = i
			}
		}
	}

	rows, err = s.server.db.Query(`SELECT c.otomoairou FROM characters c
	INNER JOIN guild_characters gc ON gc.character_id = c.id
	WHERE gc.guild_id = $1 AND c.otomoairou IS NOT NULL
	ORDER BY c.id LIMIT 60`, guild.ID)
	if err != nil {
		s.logger.Warn("Selecting otomoairou based on guild failed", zap.Error(err))
		return guildCats
	}

	for rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil || len(data) == 0 {
			continue
		}
		// first byte has cat existence in general, can skip if 0
		if data[0] == 1 {
			decomp, err := nullcomp.Decompress(data[1:])
			if err != nil {
				s.logger.Warn("decomp failure", zap.Error(err))
				continue
			}
			bf := byteframe.NewByteFrameFromBytes(decomp)
			cats := GetAirouDetails(bf)
			for _, cat := range cats {
				_, exists := bannedCats[cat.ID]
				if cat.Task == 4 && !exists {
					guildCats = append(guildCats, cat)
				}
			}
		}
	}
	return guildCats
}

func GetAirouDetails(bf *byteframe.ByteFrame) []Airou {
	catCount := bf.ReadUint8()
	cats := make([]Airou, catCount)
	for x := 0; x < int(catCount); x++ {
		var catDef Airou
		// cat sometimes has additional bytes for whatever reason, gift items? timestamp?
		// until actual variance is known we can just seek to end based on start
		catDefLen := bf.ReadUint32()
		catStart, _ := bf.Seek(0, io.SeekCurrent)

		catDef.ID = bf.ReadUint32()
		bf.Seek(1, io.SeekCurrent)     // unknown value, probably a bool
		catDef.Name = bf.ReadBytes(18) // always 18 len, reads first null terminated string out of section and discards rest
		catDef.Task = bf.ReadUint8()
		bf.Seek(16, io.SeekCurrent) // appearance data and what is seemingly null bytes
		catDef.Personality = bf.ReadUint8()
		catDef.Class = bf.ReadUint8()
		bf.Seek(5, io.SeekCurrent)          // affection and colour sliders
		catDef.Experience = bf.ReadUint32() // raw cat rank points, doesn't have a rank
		bf.Seek(1, io.SeekCurrent)          // bool for weapon being equipped
		catDef.WeaponType = bf.ReadUint8()  // weapon type, presumably always 6 for melee?
		catDef.WeaponID = bf.ReadUint16()   // weapon id
		bf.Seek(catStart+int64(catDefLen), io.SeekStart)
		cats[x] = catDef
	}
	return cats
}
