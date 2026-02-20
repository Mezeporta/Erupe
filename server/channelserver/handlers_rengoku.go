package channelserver

import (
	"encoding/binary"
	ps "erupe-ce/common/pascalstring"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"

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
		var existing []byte
		if err := s.server.db.QueryRow("SELECT rengokudata FROM characters WHERE id=$1", s.charID).Scan(&existing); err == nil {
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
		var existing []byte
		if err := s.server.db.QueryRow("SELECT rengokudata FROM characters WHERE id=$1", s.charID).Scan(&existing); err == nil {
			if len(existing) >= 4 && binary.BigEndian.Uint32(existing[:4]) != 0 {
				s.logger.Warn("Refusing to overwrite valid rengoku data with empty sentinel",
					zap.Uint32("charID", s.charID))
				doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
				return
			}
		}
	}

	_, err := s.server.db.Exec("UPDATE characters SET rengokudata=$1 WHERE id=$2", saveData, s.charID)
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
	var t int
	err = s.server.db.QueryRow("SELECT character_id FROM rengoku_score WHERE character_id=$1", s.charID).Scan(&t)
	if err != nil {
		if _, err := s.server.db.Exec("INSERT INTO rengoku_score (character_id) VALUES ($1)", s.charID); err != nil {
			s.logger.Error("Failed to insert rengoku score", zap.Error(err))
		}
	}
	if _, err := s.server.db.Exec("UPDATE rengoku_score SET max_stages_mp=$1, max_points_mp=$2, max_stages_sp=$3, max_points_sp=$4 WHERE character_id=$5", maxStageMp, maxScoreMp, maxStageSp, maxScoreSp, s.charID); err != nil {
		s.logger.Error("Failed to update rengoku score", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadRengokuData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadRengokuData)
	var data []byte
	err := s.server.db.QueryRow("SELECT rengokudata FROM characters WHERE id = $1", s.charID).Scan(&data)
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

const rengokuScoreQuery = `, c.name FROM rengoku_score rs
LEFT JOIN characters c ON c.id = rs.character_id
LEFT JOIN guild_characters gc ON gc.character_id = rs.character_id `

// RengokuScore represents a Rengoku (Hunting Road) ranking score.
type RengokuScore struct {
	Name  string `db:"name"`
	Score uint32 `db:"score"`
}

func handleMsgMhfEnumerateRengokuRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRengokuRanking)

	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	isApplicant, _ := guild.HasApplicationForCharID(s, s.charID)
	if isApplicant {
		guild = nil
	}

	if pkt.Leaderboard == 2 || pkt.Leaderboard == 3 || pkt.Leaderboard == 6 || pkt.Leaderboard == 7 {
		if guild == nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 11))
			return
		}
	}

	var score RengokuScore
	var selfExist bool
	i := uint32(1)
	bf := byteframe.NewByteFrame()
	scoreData := byteframe.NewByteFrame()

	var rows *sqlx.Rows
	var err error
	switch pkt.Leaderboard {
	case 0:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_stages_mp AS score %s ORDER BY max_stages_mp DESC", rengokuScoreQuery))
	case 1:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_points_mp AS score %s ORDER BY max_points_mp DESC", rengokuScoreQuery))
	case 2:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_stages_mp AS score %s WHERE guild_id=$1 ORDER BY max_stages_mp DESC", rengokuScoreQuery), guild.ID)
	case 3:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_points_mp AS score %s WHERE guild_id=$1 ORDER BY max_points_mp DESC", rengokuScoreQuery), guild.ID)
	case 4:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_stages_sp AS score %s ORDER BY max_stages_sp DESC", rengokuScoreQuery))
	case 5:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_points_sp AS score %s ORDER BY max_points_sp DESC", rengokuScoreQuery))
	case 6:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_stages_sp AS score %s WHERE guild_id=$1 ORDER BY max_stages_sp DESC", rengokuScoreQuery), guild.ID)
	case 7:
		rows, err = s.server.db.Queryx(fmt.Sprintf("SELECT max_points_sp AS score %s WHERE guild_id=$1 ORDER BY max_points_sp DESC", rengokuScoreQuery), guild.ID)
	}
	if err != nil {
		s.logger.Error("Failed to query rengoku ranking", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 11))
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		_ = rows.StructScan(&score)
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
