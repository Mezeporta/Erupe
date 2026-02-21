package channelserver

import (
	"encoding/binary"
	ps "erupe-ce/common/pascalstring"
	"os"
	"path/filepath"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// Rengoku save blob layout offsets
const (
	rengokuSkillSlotsStart  = 0x1B
	rengokuSkillSlotsEnd    = 0x21
	rengokuSkillValuesStart = 0x2E
	rengokuSkillValuesEnd   = 0x3A
	rengokuPointsStart      = 0x3B
	rengokuPointsEnd        = 0x47
	rengokuMaxStageMpOffset = 71
	rengokuMinPayloadSize   = 91
	rengokuMaxPayloadSize   = 4096
)

// rengokuSkillsZeroed checks if the skill slot IDs (offsets 0x1B-0x20) and
// equipped skill values (offsets 0x2E-0x39) are all zero in a rengoku save blob.
func rengokuSkillsZeroed(data []byte) bool {
	if len(data) < rengokuSkillValuesEnd {
		return true
	}
	for _, b := range data[rengokuSkillSlotsStart:rengokuSkillSlotsEnd] {
		if b != 0 {
			return false
		}
	}
	for _, b := range data[rengokuSkillValuesStart:rengokuSkillValuesEnd] {
		if b != 0 {
			return false
		}
	}
	return true
}

// rengokuHasPoints checks if any skill point allocation (offsets 0x3B-0x46) is nonzero.
func rengokuHasPoints(data []byte) bool {
	if len(data) < rengokuPointsEnd {
		return false
	}
	for _, b := range data[rengokuPointsStart:rengokuPointsEnd] {
		if b != 0 {
			return true
		}
	}
	return false
}

// rengokuMergeSkills copies skill slot IDs (0x1B-0x20) and equipped skill
// values (0x2E-0x39) from existing data into the incoming save payload,
// preserving the skills that the client failed to populate due to a race
// condition during area transitions (see issue #85).
func rengokuMergeSkills(dst, src []byte) {
	copy(dst[rengokuSkillSlotsStart:rengokuSkillSlotsEnd], src[rengokuSkillSlotsStart:rengokuSkillSlotsEnd])
	copy(dst[rengokuSkillValuesStart:rengokuSkillValuesEnd], src[rengokuSkillValuesStart:rengokuSkillValuesEnd])
}

func handleMsgMhfSaveRengokuData(s *Session, p mhfpacket.MHFPacket) {
	// Saved every floor on road, holds values such as floors progressed, points etc.
	// Can be safely handled by the client.
	pkt := p.(*mhfpacket.MsgMhfSaveRengokuData)
	if len(pkt.RawDataPayload) < rengokuMinPayloadSize || len(pkt.RawDataPayload) > rengokuMaxPayloadSize {
		s.logger.Warn("Rengoku payload size out of range", zap.Int("len", len(pkt.RawDataPayload)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	dumpSaveData(s, pkt.RawDataPayload, "rengoku")

	saveData := pkt.RawDataPayload

	// Guard against a client race condition (issue #85): the Sky Corridor init
	// path triggers a rengoku save BEFORE the load response has been parsed into
	// the character data area. This produces a save with zeroed skill fields but
	// preserved point totals. Detect this pattern and merge existing skill data.
	if len(saveData) >= rengokuPointsEnd && rengokuSkillsZeroed(saveData) && rengokuHasPoints(saveData) {
		existing, err := s.server.charRepo.LoadColumn(s.charID, "rengokudata")
		if err == nil {
			if len(existing) >= rengokuPointsEnd && !rengokuSkillsZeroed(existing) {
				s.logger.Info("Rengoku save has zeroed skills with invested points, preserving existing skills",
					zap.Uint32("charID", s.charID))
				merged := make([]byte, len(saveData))
				copy(merged, saveData)
				rengokuMergeSkills(merged, existing)
				saveData = merged
			}
		}
	}

	// Also reject saves where the sentinel is 0 (no data) if valid data already exists.
	if len(saveData) >= 4 && binary.BigEndian.Uint32(saveData[:4]) == 0 {
		existing, err := s.server.charRepo.LoadColumn(s.charID, "rengokudata")
		if err == nil {
			if len(existing) >= 4 && binary.BigEndian.Uint32(existing[:4]) != 0 {
				s.logger.Warn("Refusing to overwrite valid rengoku data with empty sentinel",
					zap.Uint32("charID", s.charID))
				doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
				return
			}
		}
	}

	err := s.server.charRepo.SaveColumn(s.charID, "rengokudata", saveData)
	if err != nil {
		s.logger.Error("Failed to save rengokudata", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrameFromBytes(saveData)
	_, _ = bf.Seek(rengokuMaxStageMpOffset, 0)
	maxStageMp := bf.ReadUint32()
	maxScoreMp := bf.ReadUint32()
	_, _ = bf.Seek(4, 1)
	maxStageSp := bf.ReadUint32()
	maxScoreSp := bf.ReadUint32()
	if err := s.server.rengokuRepo.UpsertScore(s.charID, maxStageMp, maxScoreMp, maxStageSp, maxScoreSp); err != nil {
		s.logger.Error("Failed to upsert rengoku score", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadRengokuData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadRengokuData)
	data, err := s.server.charRepo.LoadColumn(s.charID, "rengokudata")
	if err != nil {
		s.logger.Error("Failed to load rengokudata", zap.Error(err),
			zap.Uint32("charID", s.charID))
	}
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint16(0)
		resp.WriteUint32(0)
		resp.WriteUint16(0)
		resp.WriteUint16(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0) // an extra 4 bytes were missing based on pcaps

		resp.WriteUint8(3) // Count of next 3
		resp.WriteUint16(0)
		resp.WriteUint16(0)
		resp.WriteUint16(0)

		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		resp.WriteUint8(3) // Count of next 3
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		resp.WriteUint8(3) // Count of next 3
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

func handleMsgMhfGetRengokuBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRengokuBinary)
	// a (massively out of date) version resides in the game's /dat/ folder or up to date can be pulled from packets
	data, err := os.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "rengoku_data.bin"))
	if err != nil {
		s.logger.Error("Failed to read rengoku_data.bin", zap.Error(err))
		doAckBufFail(s, pkt.AckHandle, nil)
		return
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

// RengokuScore represents a Rengoku (Hunting Road) ranking score.
type RengokuScore struct {
	Name  string `db:"name"`
	Score uint32 `db:"score"`
}

func handleMsgMhfEnumerateRengokuRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRengokuRanking)

	guild, _ := s.server.guildRepo.GetByCharID(s.charID)
	var isApplicant bool
	if guild != nil {
		var appErr error
		isApplicant, appErr = s.server.guildRepo.HasApplication(guild.ID, s.charID)
		if appErr != nil {
			s.logger.Warn("Failed to check guild application status", zap.Error(appErr))
		}
	}
	if isApplicant {
		guild = nil
	}

	if pkt.Leaderboard == 2 || pkt.Leaderboard == 3 || pkt.Leaderboard == 6 || pkt.Leaderboard == 7 {
		if guild == nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 11))
			return
		}
	}

	var selfExist bool
	i := uint32(1)
	bf := byteframe.NewByteFrame()
	scoreData := byteframe.NewByteFrame()

	var guildID uint32
	if guild != nil {
		guildID = guild.ID
	}
	scores, err := s.server.rengokuRepo.GetRanking(pkt.Leaderboard, guildID)
	if err != nil {
		s.logger.Error("Failed to query rengoku ranking", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 11))
		return
	}

	for _, score := range scores {
		if score.Name == s.Name {
			bf.WriteUint32(i)
			bf.WriteUint32(score.Score)
			ps.Uint8(bf, s.Name, true)
			ps.Uint8(bf, "", false)
			selfExist = true
		}
		if i > 100 {
			i++
			continue
		}
		scoreData.WriteUint32(i)
		scoreData.WriteUint32(score.Score)
		ps.Uint8(scoreData, score.Name, true)
		ps.Uint8(scoreData, "", false)
		i++
	}

	if !selfExist {
		bf.WriteBytes(make([]byte, 10))
	}
	bf.WriteUint8(uint8(i) - 1)
	bf.WriteBytes(scoreData.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRengokuRankingRank(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRengokuRankingRank)
	// What is this for?
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // Max stage overall MP rank
	bf.WriteUint32(0) // Max RdP overall MP rank
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
