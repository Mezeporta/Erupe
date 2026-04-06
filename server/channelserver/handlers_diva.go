package channelserver

import (
	"encoding/hex"
	"erupe-ce/common/stringsupport"
	cfg "erupe-ce/config"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// Diva Defense event duration constants (all values in seconds)
const (
	divaPhaseDuration = 601200      // 6d 23h = first song phase
	divaInterlude     = 3900        // 65 min = gap between phases
	divaWeekDuration  = secsPerWeek // 7 days = subsequent phase length
	divaTotalLifespan = 2977200     // ~34.5 days = full event window
)

func cleanupDiva(s *Session) {
	if err := s.server.divaRepo.DeleteEvents(); err != nil {
		s.logger.Error("Failed to delete diva events", zap.Error(err))
	}
	if err := s.server.divaRepo.CleanupBeads(); err != nil {
		s.logger.Error("Failed to cleanup diva beads", zap.Error(err))
	}
}

func generateDivaTimestamps(s *Session, start uint32, debug bool) []uint32 {
	timestamps := make([]uint32, 6)
	midnight := TimeMidnight()
	if debug && start <= 3 {
		midnight := uint32(midnight.Unix())
		switch start {
		case 1:
			timestamps[0] = midnight
			timestamps[1] = timestamps[0] + divaPhaseDuration
			timestamps[2] = timestamps[1] + divaInterlude
			timestamps[3] = timestamps[1] + divaWeekDuration
			timestamps[4] = timestamps[3] + divaInterlude
			timestamps[5] = timestamps[3] + divaWeekDuration
		case 2:
			timestamps[0] = midnight - (divaPhaseDuration + divaInterlude)
			timestamps[1] = midnight - divaInterlude
			timestamps[2] = midnight
			timestamps[3] = timestamps[1] + divaWeekDuration
			timestamps[4] = timestamps[3] + divaInterlude
			timestamps[5] = timestamps[3] + divaWeekDuration
		case 3:
			timestamps[0] = midnight - (divaPhaseDuration + divaInterlude + divaWeekDuration + divaInterlude)
			timestamps[1] = midnight - (divaWeekDuration + divaInterlude)
			timestamps[2] = midnight - divaWeekDuration
			timestamps[3] = midnight - divaInterlude
			timestamps[4] = midnight
			timestamps[5] = timestamps[3] + divaWeekDuration
		}
		return timestamps
	}
	if start == 0 || TimeAdjusted().Unix() > int64(start)+divaTotalLifespan {
		cleanupDiva(s)
		// Generate a new diva defense, starting midnight tomorrow
		start = uint32(midnight.Add(24 * time.Hour).Unix())
		if err := s.server.divaRepo.InsertEvent(start); err != nil {
			s.logger.Error("Failed to insert diva event", zap.Error(err))
		}
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + divaPhaseDuration
	timestamps[2] = timestamps[1] + divaInterlude
	timestamps[3] = timestamps[1] + divaWeekDuration
	timestamps[4] = timestamps[3] + divaInterlude
	timestamps[5] = timestamps[3] + divaWeekDuration
	return timestamps
}

func handleMsgMhfGetUdSchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSchedule)
	bf := byteframe.NewByteFrame()

	const divaIDSentinel = uint32(0xCAFEBEEF)
	id, start := divaIDSentinel, uint32(0)
	events, err := s.server.divaRepo.GetEvents()
	if err != nil {
		s.logger.Error("Failed to query diva schedule", zap.Error(err))
	} else if len(events) > 0 {
		last := events[len(events)-1]
		id = last.ID
		start = last.StartTime
	}

	var timestamps []uint32
	if s.server.erupeConfig.DebugOptions.DivaOverride >= 0 {
		if s.server.erupeConfig.DebugOptions.DivaOverride == 0 {
			if s.server.erupeConfig.RealClientMode >= cfg.Z2 {
				doAckBufSucceed(s, pkt.AckHandle, make([]byte, 36))
			} else {
				doAckBufSucceed(s, pkt.AckHandle, make([]byte, 32))
			}
			return
		}
		timestamps = generateDivaTimestamps(s, uint32(s.server.erupeConfig.DebugOptions.DivaOverride), true)
	} else {
		timestamps = generateDivaTimestamps(s, start, false)
	}

	if s.server.erupeConfig.RealClientMode >= cfg.Z2 {
		bf.WriteUint32(id)
	}
	for i := range timestamps {
		bf.WriteUint32(timestamps[i])
	}

	bf.WriteUint16(0x19) // Unk 00011001
	bf.WriteUint16(0x2D) // Unk 00101101
	bf.WriteUint16(0x02) // Unk 00000010
	bf.WriteUint16(0x02) // Unk 00000010

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdInfo)
	// Message that appears on the Diva Defense NPC and triggers the green exclamation mark
	udInfos := []struct {
		Text      string
		StartTime time.Time
		EndTime   time.Time
	}{}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udInfos)))
	for _, udInfo := range udInfos {
		resp.WriteBytes(stringsupport.PaddedString(udInfo.Text, 1024, true))
		resp.WriteUint32(uint32(udInfo.StartTime.Unix()))
		resp.WriteUint32(uint32(udInfo.EndTime.Unix()))
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

// defaultBeadTypes are used when the database has no bead rows configured.
var defaultBeadTypes = []int{1, 3, 4, 8}

func handleMsgMhfGetKijuInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKijuInfo)

	// RE-confirmed entry layout (546 bytes each):
	//   +0x000 char[32]  name
	//   +0x020 char[512] description
	//   +0x220 u8        color_id  (slot index, 1-based)
	//   +0x221 u8        bead_type (effect ID)
	// Response: u8 count + count × 546 bytes.
	beadTypes, err := s.server.divaRepo.GetBeads()
	if err != nil || len(beadTypes) == 0 {
		beadTypes = defaultBeadTypes
	}

	lang := getLangStrings(s.server)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(beadTypes)))
	for i, bt := range beadTypes {
		name, desc := lang.beadName(bt), lang.beadDescription(bt)
		bf.WriteBytes(stringsupport.PaddedString(name, 32, true))
		bf.WriteBytes(stringsupport.PaddedString(desc, 512, true))
		bf.WriteUint8(uint8(i + 1)) // color_id: slot 1..N
		bf.WriteUint8(uint8(bt))    // bead_type: effect ID
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSetKiju(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetKiju)
	beadIndex := int(pkt.Unk1)
	expiry := TimeAdjusted().Add(24 * time.Hour)
	if err := s.server.divaRepo.AssignBead(s.charID, beadIndex, expiry); err != nil {
		s.logger.Warn("Failed to assign bead",
			zap.Uint32("charID", s.charID),
			zap.Int("beadIndex", beadIndex),
			zap.Error(err))
	} else {
		s.currentBeadIndex = beadIndex
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00})
}

func handleMsgMhfAddUdPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddUdPoint)

	// Find the current diva event to associate points with.
	eventID := uint32(0)
	if s.server.divaRepo != nil {
		events, err := s.server.divaRepo.GetEvents()
		if err == nil && len(events) > 0 {
			eventID = events[len(events)-1].ID
		}
	}

	if eventID != 0 && s.charID != 0 && (pkt.QuestPoints > 0 || pkt.BonusPoints > 0) {
		if err := s.server.divaRepo.AddPoints(s.charID, eventID, pkt.QuestPoints, pkt.BonusPoints); err != nil {
			s.logger.Warn("Failed to add diva points",
				zap.Uint32("charID", s.charID),
				zap.Uint32("questPoints", pkt.QuestPoints),
				zap.Uint32("bonusPoints", pkt.BonusPoints),
				zap.Error(err))
		}
		if s.currentBeadIndex >= 0 {
			total := int(pkt.QuestPoints) + int(pkt.BonusPoints)
			if total > 0 {
				if err := s.server.divaRepo.AddBeadPoints(s.charID, s.currentBeadIndex, total); err != nil {
					s.logger.Warn("Failed to add bead points",
						zap.Uint32("charID", s.charID),
						zap.Int("beadIndex", s.currentBeadIndex),
						zap.Error(err))
				}
			}
		}
	}

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdMyPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMyPoint)

	// RE confirms: no count prefix. Client hardcodes exactly 8 loop iterations.
	// Per-entry stride is 18 bytes:
	//   +0x00 u8  bead_index
	//   +0x01 u32 points
	//   +0x05 u32 points_dupe  (same value as points)
	//   +0x09 u8  unk1         (half-period: 0=first 12h, 1=second 12h)
	//   +0x0A u32 unk2
	//   +0x0E u32 unk3
	// Total: 8 × 18 = 144 bytes.
	beadPoints, err := s.server.divaRepo.GetCharacterBeadPoints(s.charID)
	if err != nil {
		s.logger.Warn("Failed to get bead points", zap.Uint32("charID", s.charID), zap.Error(err))
		beadPoints = map[int]int{}
	}
	activeBead := uint8(0)
	if s.currentBeadIndex >= 0 {
		activeBead = uint8(s.currentBeadIndex)
	}
	pts := uint32(0)
	if s.currentBeadIndex >= 0 {
		if p, ok := beadPoints[s.currentBeadIndex]; ok {
			pts = uint32(p)
		}
	}
	bf := byteframe.NewByteFrame()
	for i := 0; i < 8; i++ {
		bf.WriteUint8(activeBead)
		bf.WriteUint32(pts)
		bf.WriteUint32(pts)         // points_dupe
		bf.WriteUint8(uint8(i % 2)) // unk1: 0=first half, 1=second half
		bf.WriteUint32(0)           // unk2
		bf.WriteUint32(0)           // unk3
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// udMilestones are the global contribution milestones for Diva Defense.
// RE confirms: 64 × u64 target_values + 64 × u8 target_types + u64 total = ~585 bytes.
// Slots 0–12 are populated; slots 13–63 are zero.
var udMilestones = []uint64{
	500000, 1000000, 2000000, 3000000, 5000000, 7000000, 10000000,
	15000000, 20000000, 30000000, 50000000, 70000000, 100000000,
}

func handleMsgMhfGetUdTotalPointInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTotalPointInfo)

	total, err := s.server.divaRepo.GetTotalBeadPoints()
	if err != nil {
		s.logger.Warn("Failed to get total bead points", zap.Error(err))
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // error = success
	// 64 × u64 target_values (big-endian)
	for i := 0; i < 64; i++ {
		var v uint64
		if i < len(udMilestones) {
			v = udMilestones[i]
		}
		bf.WriteUint64(v)
	}
	// 64 × u8 target_types (0 = global)
	for i := 0; i < 64; i++ {
		bf.WriteUint8(0)
	}
	// u64 total_souls
	bf.WriteUint64(uint64(total))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdSelectedColorInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSelectedColorInfo)

	// RE confirms: exactly 9 bytes = u8 error + u8[8] winning colors.
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // error = success
	for day := 0; day < 8; day++ {
		topBead, err := s.server.divaRepo.GetTopBeadPerDay(day)
		if err != nil {
			topBead = 0
		}
		bf.WriteUint8(uint8(topBead))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdMonsterPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMonsterPoint)

	monsterPoints := []struct {
		MID    uint8
		Points uint16
	}{
		{MID: 0x01, Points: 0x3C}, // em1 Rathian
		{MID: 0x02, Points: 0x5A}, // em2 Fatalis
		{MID: 0x06, Points: 0x14}, // em6 Yian Kut-Ku
		{MID: 0x07, Points: 0x50}, // em7 Lao-Shan Lung
		{MID: 0x08, Points: 0x28}, // em8 Cephadrome
		{MID: 0x0B, Points: 0x3C}, // em11 Rathalos
		{MID: 0x0E, Points: 0x3C}, // em14 Diablos
		{MID: 0x0F, Points: 0x46}, // em15 Khezu
		{MID: 0x11, Points: 0x46}, // em17 Gravios
		{MID: 0x14, Points: 0x28}, // em20 Gypceros
		{MID: 0x15, Points: 0x3C}, // em21 Plesioth
		{MID: 0x16, Points: 0x32}, // em22 Basarios
		{MID: 0x1A, Points: 0x32}, // em26 Monoblos
		{MID: 0x1B, Points: 0x0A}, // em27 Velocidrome
		{MID: 0x1C, Points: 0x0A}, // em28 Gendrome
		{MID: 0x1F, Points: 0x0A}, // em31 Iodrome
		{MID: 0x21, Points: 0x50}, // em33 Kirin
		{MID: 0x24, Points: 0x64}, // em36 Crimson Fatalis
		{MID: 0x25, Points: 0x3C}, // em37 Pink Rathian
		{MID: 0x26, Points: 0x1E}, // em38 Blue Yian Kut-Ku
		{MID: 0x27, Points: 0x28}, // em39 Purple Gypceros
		{MID: 0x28, Points: 0x50}, // em40 Yian Garuga
		{MID: 0x29, Points: 0x5A}, // em41 Silver Rathalos
		{MID: 0x2A, Points: 0x50}, // em42 Gold Rathian
		{MID: 0x2B, Points: 0x3C}, // em43 Black Diablos
		{MID: 0x2C, Points: 0x3C}, // em44 White Monoblos
		{MID: 0x2D, Points: 0x46}, // em45 Red Khezu
		{MID: 0x2E, Points: 0x3C}, // em46 Green Plesioth
		{MID: 0x2F, Points: 0x50}, // em47 Black Gravios
		{MID: 0x30, Points: 0x1E}, // em48 Daimyo Hermitaur
		{MID: 0x31, Points: 0x3C}, // em49 Azure Rathalos
		{MID: 0x32, Points: 0x50}, // em50 Ashen Lao-Shan Lung
		{MID: 0x33, Points: 0x3C}, // em51 Blangonga
		{MID: 0x34, Points: 0x28}, // em52 Congalala
		{MID: 0x35, Points: 0x50}, // em53 Rajang
		{MID: 0x36, Points: 0x6E}, // em54 Kushala Daora
		{MID: 0x37, Points: 0x50}, // em55 Shen Gaoren
		{MID: 0x3A, Points: 0x50}, // em58 Yama Tsukami
		{MID: 0x3B, Points: 0x6E}, // em59 Chameleos
		{MID: 0x40, Points: 0x64}, // em64 Lunastra
		{MID: 0x41, Points: 0x6E}, // em65 Teostra
		{MID: 0x43, Points: 0x28}, // em67 Shogun Ceanataur
		{MID: 0x44, Points: 0x0A}, // em68 Bulldrome
		{MID: 0x47, Points: 0x6E}, // em71 White Fatalis
		{MID: 0x4A, Points: 0xFA}, // em74 Hypnocatrice
		{MID: 0x4B, Points: 0xFA}, // em75 Lavasioth
		{MID: 0x4C, Points: 0x46}, // em76 Tigrex
		{MID: 0x4D, Points: 0x64}, // em77 Akantor
		{MID: 0x4E, Points: 0xFA}, // em78 Bright Hypnoc
		{MID: 0x4F, Points: 0xFA}, // em79 Lavasioth Subspecies
		{MID: 0x50, Points: 0xFA}, // em80 Espinas
		{MID: 0x51, Points: 0xFA}, // em81 Orange Espinas
		{MID: 0x52, Points: 0xFA}, // em82 White Hypnoc
		{MID: 0x53, Points: 0xFA}, // em83 Akura Vashimu
		{MID: 0x54, Points: 0xFA}, // em84 Akura Jebia
		{MID: 0x55, Points: 0xFA}, // em85 Berukyurosu
		{MID: 0x59, Points: 0xFA}, // em89 Pariapuria
		{MID: 0x5A, Points: 0xFA}, // em90 White Espinas
		{MID: 0x5B, Points: 0xFA}, // em91 Kamu Orugaron
		{MID: 0x5C, Points: 0xFA}, // em92 Nono Orugaron
		{MID: 0x5E, Points: 0xFA}, // em94 Dyuragaua
		{MID: 0x5F, Points: 0xFA}, // em95 Doragyurosu
		{MID: 0x60, Points: 0xFA}, // em96 Gurenzeburu
		{MID: 0x63, Points: 0xFA}, // em99 Rukodiora
		{MID: 0x65, Points: 0xFA}, // em101 Gogomoa
		{MID: 0x67, Points: 0xFA}, // em103 Taikun Zamuza
		{MID: 0x68, Points: 0xFA}, // em104 Abiorugu
		{MID: 0x69, Points: 0xFA}, // em105 Kuarusepusu
		{MID: 0x6A, Points: 0xFA}, // em106 Odibatorasu
		{MID: 0x6B, Points: 0xFA}, // em107 Disufiroa
		{MID: 0x6C, Points: 0xFA}, // em108 Rebidiora
		{MID: 0x6D, Points: 0xFA}, // em109 Anorupatisu
		{MID: 0x6E, Points: 0xFA}, // em110 Hyujikiki
		{MID: 0x6F, Points: 0xFA}, // em111 Midogaron
		{MID: 0x70, Points: 0xFA}, // em112 Giaorugu
		{MID: 0x72, Points: 0xFA}, // em114 Farunokku
		{MID: 0x73, Points: 0xFA}, // em115 Pokaradon
		{MID: 0x74, Points: 0xFA}, // em116 Shantien
		{MID: 0x77, Points: 0xFA}, // em119 Goruganosu
		{MID: 0x78, Points: 0xFA}, // em120 Aruganosu
		{MID: 0x79, Points: 0xFA}, // em121 Baruragaru
		{MID: 0x7A, Points: 0xFA}, // em122 Zerureusu
		{MID: 0x7B, Points: 0xFA}, // em123 Gougarf
		{MID: 0x7D, Points: 0xFA}, // em125 Forokururu
		{MID: 0x7E, Points: 0xFA}, // em126 Meraginasu
		{MID: 0x7F, Points: 0xFA}, // em127 Diorekkusu
		{MID: 0x80, Points: 0xFA}, // em128 Garuba Daora
		{MID: 0x81, Points: 0xFA}, // em129 Inagami
		{MID: 0x82, Points: 0xFA}, // em130 Varusaburosu
		{MID: 0x83, Points: 0xFA}, // em131 Poborubarumu
		{MID: 0x8B, Points: 0xFA}, // em139 Gureadomosu
		{MID: 0x8C, Points: 0xFA}, // em140 Harudomerugu
		{MID: 0x8D, Points: 0xFA}, // em141 Toridcless
		{MID: 0x8E, Points: 0xFA}, // em142 Gasurabazura
		{MID: 0x90, Points: 0xFA}, // em144 Yama Kurai
		{MID: 0x92, Points: 0x78}, // em146 Zinogre
		{MID: 0x93, Points: 0x78}, // em147 Deviljho
		{MID: 0x94, Points: 0x78}, // em148 Brachydios
		{MID: 0x96, Points: 0xFA}, // em150 Toa Tesukatora
		{MID: 0x97, Points: 0x78}, // em151 Barioth
		{MID: 0x98, Points: 0x78}, // em152 Uragaan
		{MID: 0x99, Points: 0x78}, // em153 Stygian Zinogre
		{MID: 0x9A, Points: 0xFA}, // em154 Guanzorumu
		{MID: 0x9E, Points: 0xFA}, // em158 Voljang
		{MID: 0x9F, Points: 0x78}, // em159 Nargacuga
		{MID: 0xA0, Points: 0xFA}, // em160 Keoaruboru
		{MID: 0xA1, Points: 0xFA}, // em161 Zenaserisu
		{MID: 0xA2, Points: 0x78}, // em162 Gore Magala
		{MID: 0xA4, Points: 0x78}, // em164 Shagaru Magala
		{MID: 0xA5, Points: 0x78}, // em165 Amatsu
		{MID: 0xA6, Points: 0xFA}, // em166 Elzelion
		{MID: 0xA9, Points: 0x78}, // em169 Seregios
		{MID: 0xAA, Points: 0xFA}, // em170 Bogabadorumu
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(monsterPoints)))
	for _, mp := range monsterPoints {
		resp.WriteUint8(mp.MID)
		resp.WriteUint16(mp.Points)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdDailyPresentList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdDailyPresentList)
	// DailyPresentList: u16 count + count × 15-byte entries.
	// Entry: u8 rank_type, u16 rank_from, u16 rank_to, u8 item_type,
	//        u16 _pad0(skip), u16 item_id, u16 _pad1(skip), u16 quantity, u8 unk.
	// Padding at +6 and +10 is NOT read by the client.
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // count = 0 (no entries configured)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdNormaPresentList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdNormaPresentList)
	// NormaPresentList: u16 count + count × 19-byte entries.
	// Same layout as DailyPresent (+0x00..+0x0D), plus:
	//   +0x0E u32 points_required (norma threshold)
	//   +0x12 u8  bead_type (BeadType that unlocks this tier)
	// Padding at +6 and +10 NOT read.
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // count = 0 (no entries configured)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireUdItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireUdItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdRanking)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdMyRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMyRanking)
	// Temporary canned response
	data, _ := hex.DecodeString("00000515000005150000CEB4000003CE000003CE0000CEB44D49444E494748542D414E47454C0000000000000000000000")
	doAckBufSucceed(s, pkt.AckHandle, data)
}
