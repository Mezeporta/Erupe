package channelserver

import (
	"database/sql"
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"io"
	"time"

	"go.uber.org/zap"
)

func handleMsgMhfLoadPartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPartner)
	loadCharacterData(s, pkt.AckHandle, "partner", make([]byte, 9))
}

func handleMsgMhfSavePartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePartner)
	saveCharacterData(s, pkt.AckHandle, "partner", pkt.RawDataPayload, 65536)
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

// Hunter Navi buffer sizes per game version
const (
	hunterNaviSizeG8 = 552 // G8+ navi buffer size
	hunterNaviSizeG7 = 280 // G7 and older navi buffer size
)

func handleMsgMhfLoadHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHunterNavi)
	naviLength := hunterNaviSizeG8
	if s.server.erupeConfig.RealClientMode <= cfg.G7 {
		naviLength = hunterNaviSizeG7
	}
	loadCharacterData(s, pkt.AckHandle, "hunternavi", make([]byte, naviLength))
}

func handleMsgMhfSaveHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveHunterNavi)
	if len(pkt.RawDataPayload) > 4096 {
		s.logger.Warn("HunterNavi payload too large", zap.Int("len", len(pkt.RawDataPayload)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	saveStart := time.Now()

	s.logger.Debug("Hunter Navi save request",
		zap.Uint32("charID", s.charID),
		zap.Bool("is_diff", pkt.IsDataDiff),
		zap.Int("data_size", len(pkt.RawDataPayload)),
	)

	var dataSize int
	if pkt.IsDataDiff {
		naviLength := hunterNaviSizeG8
		if s.server.erupeConfig.RealClientMode <= cfg.G7 {
			naviLength = hunterNaviSizeG7
		}
		// Load existing save
		data, err := s.server.charRepo.LoadColumn(s.charID, "hunternavi")
		if err != nil {
			s.logger.Error("Failed to load hunternavi",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
			)
		}

		// Check if we actually had any hunternavi data, using a blank buffer if not.
		// This is requried as the client will try to send a diff after character creation without a prior MsgMhfSaveHunterNavi packet.
		if len(data) == 0 {
			data = make([]byte, naviLength)
		}

		// Perform diff and compress it to write back to db
		s.logger.Debug("Applying Hunter Navi diff",
			zap.Uint32("charID", s.charID),
			zap.Int("base_size", len(data)),
			zap.Int("diff_size", len(pkt.RawDataPayload)),
		)
		saveOutput := deltacomp.ApplyDataDiff(pkt.RawDataPayload, data)
		dataSize = len(saveOutput)

		err = s.server.charRepo.SaveColumn(s.charID, "hunternavi", saveOutput)
		if err != nil {
			s.logger.Error("Failed to save hunternavi",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
				zap.Int("data_size", dataSize),
			)
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "hunternavi")
		dataSize = len(pkt.RawDataPayload)

		// simply update database, no extra processing
		err := s.server.charRepo.SaveColumn(s.charID, "hunternavi", pkt.RawDataPayload)
		if err != nil {
			s.logger.Error("Failed to save hunternavi",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
				zap.Int("data_size", dataSize),
			)
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
	}

	saveDuration := time.Since(saveStart)
	s.logger.Info("Hunter Navi saved successfully",
		zap.Uint32("charID", s.charID),
		zap.Bool("was_diff", pkt.IsDataDiff),
		zap.Int("data_size", dataSize),
		zap.Duration("duration", saveDuration),
	)

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfMercenaryHuntdata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfMercenaryHuntdata)

	if pkt.RequestType != 1 {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0))
		return
	}

	rastaID, err := s.server.charRepo.ReadInt(s.charID, "rasta_id")
	if err != nil {
		s.logger.Error("failed to fetch rasta_id for character", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	defaultClaim := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	lastKillLog, lastClaim, err := s.server.mercenaryRepo.GetLastMercenaryReward(uint32(rastaID), defaultClaim)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("failed to fetch last mercenary claim", zap.Error(err), zap.Uint32("rastaID", uint32(rastaID)))
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
			return
		}
		// No previous claim; treat as first claim.
		lastKillLog = 0
		lastClaim = defaultClaim
	}

	// Same boundary logic as the daily reset: current period's noon cutoff,
	// rolled forward a day if we're already past today's noon.
	midday := TimeMidnight().Add(12 * time.Hour)
	if TimeAdjusted().After(midday) {
		midday = midday.Add(24 * time.Hour)
	}

	if !midday.After(lastClaim) {
		s.logger.Info("Woops Too Soon")
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	logs, err := s.server.mercenaryRepo.GetRecentKillLogs(uint32(rastaID), lastKillLog, 5)
	if err != nil {
		s.logger.Error("failed to fetch kill logs for mercenary huntdata", zap.Error(err), zap.Uint32("rastaID", uint32(rastaID)))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	if len(logs) == 0 {
		s.logger.Info("Woops No new Hunts")
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	var maxID uint32
	for _, log := range logs {
		if log.ID > maxID {
			maxID = log.ID
		}
	}

	if err := s.server.mercenaryRepo.InsertMercenaryReward(uint32(rastaID), maxID, midday); err != nil {
		s.logger.Error("failed to insert mercenary claim", zap.Error(err), zap.Uint32("rastaID", uint32(rastaID)), zap.Uint32("killLogID", maxID))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(logs)))
	for _, log := range logs {
		bf.WriteUint32(log.ID)
		bf.WriteUint32(log.Monster)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateMercenaryLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateMercenaryLog)
	bf := byteframe.NewByteFrame()

	logs, err := s.server.mercenaryRepo.GetMercenaryLogs(s.charID)
	if err != nil {
		s.logger.Error("failed to fetch mercenary logs", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}

	bf.WriteUint32(uint32(len(logs)))
	for _, l := range logs {
		bf.WriteUint32(uint32(l.Date.Unix()))
		bf.WriteBytes(stringsupport.PaddedString(l.Name, 18, true))

		state := []byte{0, byte(l.State)}
		if l.IsInitiator {
			state[0] = 1
		}
		bf.WriteBytes(state)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfCreateMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateMercenary)
	nextID, err := s.server.mercenaryRepo.NextRastaID()
	if err != nil {
		s.logger.Error("Failed to get next rasta ID", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	if err := s.server.charRepo.SaveInt(s.charID, "rasta_id", int(nextID)); err != nil {
		s.logger.Error("Failed to set rasta ID", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(nextID)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSaveMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMercenary)

	if len(pkt.MercData) > 0x8B {
		s.logger.Warn("Mercenary payload too large", zap.Int("len", len(pkt.MercData)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	dumpSaveData(s, pkt.MercData, "mercenary")

	if len(pkt.MercData) == 0x8B {
		if pkt.RastaMercID == 0 {
			s.logger.Warn("Mercenary save with rasta_id=0, preserving existing value",
				zap.Uint32("charID", s.charID))
		}
		if err := s.server.charRepo.SaveMercenary(s.charID, pkt.MercData, pkt.RastaMercID); err != nil {
			s.logger.Error("Failed to save mercenary data", zap.Error(err))
		}
	} else {
		s.logger.Warn("Mercenary payload too small", zap.Int("len", len(pkt.MercData)))
	}

	var gcp int
	gcp, gcpErr := readCharacterInt(s, "gcp")
	if gcpErr != nil {
		s.logger.Warn("Failed to read gcp", zap.Error(gcpErr))
	} else if pkt.GCP > uint32(gcp) {
		difference := pkt.GCP - uint32(gcp)
		// s.logger.Info("GCP increased",
		// 	zap.Uint32("old_gcp", uint32(gcp)),
		// 	zap.Uint32("new_gcp", pkt.GCP),
		// 	zap.Uint32("difference", difference),
		// )

		if pkt.PactMercID > 0 {
			ownerCharID, _, _, ownerErr := s.server.charRepo.FindByRastaID(int(pkt.PactMercID), s.charID)
			if ownerErr != nil {
				s.logger.Warn("Failed to find pact owner for GCP credit",
					zap.Error(ownerErr), zap.Uint32("pactMercID", pkt.PactMercID))
			} else if _, err := s.server.charRepo.AdjustInt(ownerCharID, "gcp", int(difference)); err != nil {
				s.logger.Warn("Failed to credit GCP to pact owner",
					zap.Error(err), zap.Uint32("ownerCharID", ownerCharID))
			}
		}
	}

	if err := s.server.charRepo.UpdateGCPAndPact(s.charID, pkt.GCP, pkt.PactMercID); err != nil {
		s.logger.Error("Failed to update GCP and pact ID", zap.Error(err))
	}

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadMercenaryW(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryW)
	bf := byteframe.NewByteFrame()

	pactID, pactErr := readCharacterInt(s, "pact_id")
	if pactErr != nil {
		s.logger.Warn("Failed to read pact_id", zap.Error(pactErr))
	}

	var cid uint32
	var name string
	if pactID > 0 {
		var findErr error
		var contractDate sql.NullTime
		cid, name, contractDate, findErr = s.server.charRepo.FindByRastaID(pactID, s.charID)
		if findErr != nil {
			s.logger.Warn("Failed to find character by rasta ID", zap.Error(findErr))
		}

		start := TimeAdjusted()
		if contractDate.Valid {
			start = contractDate.Time
		}
		expiry := start.Add(7 * 24 * time.Hour)

		bf.WriteUint8(1) // numLends
		bf.WriteUint32(uint32(pactID))
		bf.WriteUint32(cid)
		bf.WriteBool(true) // Escort enabled
		bf.WriteUint32(uint32(start.Unix()))
		bf.WriteUint32(uint32(expiry.Unix()))
		bf.WriteBytes(stringsupport.PaddedString(name, 18, true))
	} else {
		bf.WriteUint8(0)
	}

	if pkt.Op != 2 && pkt.Op != 5 {
		loans, err := s.server.mercenaryRepo.GetMercenaryLoans(s.charID)
		if err != nil {
			s.logger.Error("Failed to query mercenary loans", zap.Error(err))
		}

		bf.WriteUint8(uint8(len(loans)))
		for _, loan := range loans {
			start := TimeAdjusted()
			if loan.ContractDate.Valid {
				start = loan.ContractDate.Time
			}
			expiry := start.Add(7 * 24 * time.Hour)

			bf.WriteUint32(uint32(loan.PactID))
			bf.WriteUint32(loan.CharID)
			bf.WriteUint32(uint32(start.Unix()))
			bf.WriteUint32(uint32(expiry.Unix()))
			bf.WriteBytes(stringsupport.PaddedString(loan.Name, 18, true))
		}

		if pkt.Op != 1 && pkt.Op != 4 {
			data, dataErr := s.server.charRepo.LoadColumn(s.charID, "savemercenary")
			if dataErr != nil {
				s.logger.Warn("Failed to load savemercenary", zap.Error(dataErr))
			}
			rastaid, rastaErr := readCharacterInt(s, "rasta_id")
			if rastaErr != nil {
				s.logger.Warn("Failed to read rasta_id", zap.Error(rastaErr))
			}
			gcp, gcpErr := readCharacterInt(s, "gcp")
			if gcpErr != nil {
				s.logger.Warn("Failed to read gcp", zap.Error(gcpErr))
			}

			if len(data) == 0 {
				bf.WriteBool(false)
			} else {
				bf.WriteBool(true)
				bf.WriteUint32(uint32(rastaid))
				bf.WriteUint32(s.charID)
				bf.WriteBool(true) // TODO: Load escort toggle state from DB.
				bf.WriteBytes(data)
			}
			bf.WriteUint32(uint32(gcp))
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// This function load contracted mercenary data (triggered by mercenaryW if theres some contract)
func handleMsgMhfReadMercenaryM(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryM)
	data, err := s.server.charRepo.LoadColumn(pkt.CharID, "savemercenary")
	if err != nil {
		s.logger.Warn("Failed to load savemercenary for other character", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	if len(data) == 0 {
		resp.WriteBool(false)
	} else {
		//mercenary header data
		resp.WriteUint32(uint32(pkt.MercID))
		resp.WriteUint32(uint32(pkt.CharID))
		//hard code for now, should be in database to check rasta escort state (follow/not)
		resp.WriteBool(true)
		//mercenary savedata
		resp.WriteBytes(data)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfContractMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfContractMercenary)
	now := TimeAdjusted()

	switch pkt.Op {
	case 0: // Form loan
		if err := s.server.charRepo.SetPact(pkt.CID, pkt.PactMercID, now); err != nil {
			s.logger.Error("Failed to form mercenary loan",
				zap.Uint32("charID", pkt.CID),
				zap.Uint32("pactMercID", pkt.PactMercID),
				zap.Error(err),
			)
		}
		if err := s.server.mercenaryRepo.LogMercenaryEvent(s.charID, pkt.CID, now, true, 0); err != nil {
			s.logger.Error("Failed to log mercenary event",
				zap.Uint32("playerID", s.charID),
				zap.Uint32("mercenaryID", pkt.CID),
				zap.Int("state", 0),
				zap.Error(err),
			)
		}
		if err := s.server.mercenaryRepo.LogMercenaryEvent(pkt.CID, s.charID, now, false, 0); err != nil {
			s.logger.Error("Failed to log mercenary event",
				zap.Uint32("playerID", pkt.CID),
				zap.Uint32("mercenaryID", s.charID),
				zap.Int("state", 0),
				zap.Error(err),
			)
		}

	case 1: // Cancel lend
		if err := s.server.charRepo.ClearPact(s.charID); err != nil {
			s.logger.Error("Failed to cancel mercenary lend",
				zap.Uint32("charID", s.charID),
				zap.Error(err),
			)
		}
		if err := s.server.mercenaryRepo.LogMercenaryEvent(pkt.CID, s.charID, now, true, 4); err != nil {
			s.logger.Error("Failed to log mercenary event",
				zap.Uint32("playerID", pkt.CID),
				zap.Uint32("mercenaryID", s.charID),
				zap.Int("state", 4),
				zap.Error(err),
			)
		}
		if err := s.server.mercenaryRepo.LogMercenaryEvent(s.charID, pkt.CID, now, false, 4); err != nil {
			s.logger.Error("Failed to log mercenary event",
				zap.Uint32("playerID", s.charID),
				zap.Uint32("mercenaryID", pkt.CID),
				zap.Int("state", 4),
				zap.Error(err),
			)
		}

	case 2: // Cancel loan - player terminate the backup contract
		if err := s.server.charRepo.ClearPact(pkt.CID); err != nil {
			s.logger.Error("Failed to cancel mercenary loan",
				zap.Uint32("charID", pkt.CID),
				zap.Error(err),
			)
		}
		if err := s.server.mercenaryRepo.LogMercenaryEvent(pkt.CID, s.charID, now, false, 3); err != nil {
			s.logger.Error("Failed to log mercenary event",
				zap.Uint32("playerID", pkt.CID),
				zap.Uint32("mercenaryID", s.charID),
				zap.Int("state", 3),
				zap.Error(err),
			)
		}
		if err := s.server.mercenaryRepo.LogMercenaryEvent(s.charID, pkt.CID, now, true, 3); err != nil {
			s.logger.Error("Failed to log mercenary event",
				zap.Uint32("playerID", s.charID),
				zap.Uint32("mercenaryID", pkt.CID),
				zap.Int("state", 3),
				zap.Error(err),
			)
		}
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadOtomoAirou)
	loadCharacterData(s, pkt.AckHandle, "otomoairou", make([]byte, 10))
}

func handleMsgMhfSaveOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveOtomoAirou)
	if len(pkt.RawDataPayload) < 2 {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
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
			catID, err = s.server.mercenaryRepo.NextAirouID()
			if err != nil {
				s.logger.Error("Failed to get next airou ID", zap.Error(err))
				doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
				return
			}
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
	_, _ = save.Seek(0, 0)
	save.WriteUint8(catsExist)
	comp, err := nullcomp.Compress(save.Data())
	if err != nil {
		s.logger.Error("Failed to compress airou", zap.Error(err))
	} else {
		comp = append([]byte{0x01}, comp...)
		if err := s.server.charRepo.SaveColumn(s.charID, "otomoairou", comp); err != nil {
			s.logger.Error("Failed to save otomoairou", zap.Error(err))
		}
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

// Airou represents Airou (felyne companion) data.
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
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil {
		return guildCats
	}
	usages, err := s.server.mercenaryRepo.GetGuildHuntCatsUsed(s.charID)
	if err != nil {
		s.logger.Warn("Failed to get recently used airous", zap.Error(err))
		return guildCats
	}

	for _, usage := range usages {
		if usage.Start.Add(time.Second * time.Duration(s.server.erupeConfig.GameplayOptions.TreasureHuntPartnyaCooldown)).Before(TimeAdjusted()) {
			for i, j := range stringsupport.CSVElems(usage.CatsUsed) {
				bannedCats[uint32(j)] = i
			}
		}
	}

	airouData, err := s.server.mercenaryRepo.GetGuildAirou(guild.ID)
	if err != nil {
		s.logger.Warn("Selecting otomoairou based on guild failed", zap.Error(err))
		return guildCats
	}

	for _, data := range airouData {
		if len(data) == 0 {
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

// GetAirouDetails parses Airou data from a ByteFrame.
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
		_, _ = bf.Seek(1, io.SeekCurrent) // unknown value, probably a bool
		catDef.Name = bf.ReadBytes(18)    // always 18 len, reads first null terminated string out of section and discards rest
		catDef.Task = bf.ReadUint8()
		_, _ = bf.Seek(16, io.SeekCurrent) // appearance data and what is seemingly null bytes
		catDef.Personality = bf.ReadUint8()
		catDef.Class = bf.ReadUint8()
		_, _ = bf.Seek(5, io.SeekCurrent)   // affection and colour sliders
		catDef.Experience = bf.ReadUint32() // raw cat rank points, doesn't have a rank
		_, _ = bf.Seek(1, io.SeekCurrent)   // bool for weapon being equipped
		catDef.WeaponType = bf.ReadUint8()  // weapon type, presumably always 6 for melee?
		catDef.WeaponID = bf.ReadUint16()   // weapon id
		_, _ = bf.Seek(catStart+int64(catDefLen), io.SeekStart)
		cats[x] = catDef
	}
	return cats
}
