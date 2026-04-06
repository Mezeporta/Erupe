package channelserver

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfGetUdTacticsPoint(s *Session, p mhfpacket.MHFPacket) {
	// Diva defense interception points
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsPoint)

	pointsMap, err := s.server.divaRepo.GetCharacterInterceptionPoints(s.charID)
	if err != nil {
		s.logger.Warn("Failed to get interception points", zap.Uint32("charID", s.charID), zap.Error(err))
		pointsMap = map[string]int{}
	}

	// Build per-quest list and compute total.
	type questEntry struct {
		questFileID int
		points      int
	}
	var entries []questEntry
	var total int
	for k, pts := range pointsMap {
		qid, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		entries = append(entries, questEntry{qid, pts})
		total += pts
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(total))
	bf.WriteUint32(uint32(len(entries)))
	for _, e := range entries {
		bf.WriteUint32(uint32(e.questFileID))
		bf.WriteUint32(uint32(e.points))
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// udTacticsQuestFileIDs is the allowed range of interception quest file IDs.
const (
	udTacticsQuestMin = 58079
	udTacticsQuestMax = 58083
)

func handleMsgMhfAddUdTacticsPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddUdTacticsPoint)
	questFileID := int(pkt.QuestID)
	points := int(pkt.TacticsPoints)

	if questFileID < udTacticsQuestMin || questFileID > udTacticsQuestMax {
		s.logger.Warn("AddUdTacticsPoint: quest file ID out of range",
			zap.Int("questFileID", questFileID),
			zap.String("range", fmt.Sprintf("%d-%d", udTacticsQuestMin, udTacticsQuestMax)))
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		return
	}

	if points > 0 {
		if err := s.server.divaRepo.AddInterceptionPoints(s.charID, questFileID, points); err != nil {
			s.logger.Warn("Failed to add interception points",
				zap.Uint32("charID", s.charID),
				zap.Int("questFileID", questFileID),
				zap.Int("points", points),
				zap.Error(err))
		}
	}

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func writeDivaPrizeList(bf *byteframe.ByteFrame, prizes []DivaPrize) {
	bf.WriteUint32(uint32(len(prizes)))
	for _, p := range prizes {
		bf.WriteUint32(uint32(p.PointsReq))
		bf.WriteUint16(uint16(p.ItemType))
		bf.WriteUint16(uint16(p.ItemID))
		bf.WriteUint16(uint16(p.Quantity))
		if p.GR {
			bf.WriteUint8(1)
		} else {
			bf.WriteUint8(0)
		}
		if p.Repeatable {
			bf.WriteUint8(1)
		} else {
			bf.WriteUint8(0)
		}
	}
}

func handleMsgMhfGetUdTacticsRewardList(s *Session, p mhfpacket.MHFPacket) {
	// Diva defense interception reward list
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRewardList)

	personal, err := s.server.divaRepo.GetPersonalPrizes()
	if err != nil {
		s.logger.Warn("Failed to get personal prizes", zap.Error(err))
	}
	guild, err := s.server.divaRepo.GetGuildPrizes()
	if err != nil {
		s.logger.Warn("Failed to get guild prizes", zap.Error(err))
	}

	bf := byteframe.NewByteFrame()
	writeDivaPrizeList(bf, personal)
	writeDivaPrizeList(bf, guild)

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsFollower)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdTacticsBonusQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsBonusQuest)
	// Temporary canned response
	data, _ := hex.DecodeString("14E2F55DCBFE505DCC1A7003E8E2C55DCC6ED05DCC8AF00258E2CE5DCCDF505DCCFB700279E3075DCD4FD05DCD6BF0041AE2F15DCDC0505DCDDC700258E2C45DCE30D05DCE4CF00258E2F55DCEA1505DCEBD7003E8E2C25DCF11D05DCF2DF00258E2CE5DCF82505DCF9E700279E3075DCFF2D05DD00EF0041AE2CE5DD063505DD07F700279E2F35DD0D3D05DD0EFF0028AE2C35DD144505DD160700258E2F05DD1B4D05DD1D0F00258E2CE5DD225505DD241700279E2F55DD295D05DD2B1F003E8E2F25DD306505DD3227002EEE2CA5DD376D05DD392F00258E3075DD3E7505DD40370041AE2F55DD457D05DD473F003E82027313220686F757273273A3A696E74657276616C29202B2027313220686F757273273A3A696E74657276616C2047524F5550204259206D6170204F52444552204259206D61703B2000C7312B000032")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

// udTacticsFirstQuestBonuses are the static first-quest bonus point values.
var udTacticsFirstQuestBonuses = []uint32{1500, 2000, 2500, 3500, 4500}

func handleMsgMhfGetUdTacticsFirstQuestBonus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsFirstQuestBonus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(len(udTacticsFirstQuestBonuses)))
	for i, bonus := range udTacticsFirstQuestBonuses {
		bf.WriteUint32(bonus)
		bf.WriteUint32(uint32(i))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsRemainingPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRemainingPoint)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // Points until Special Guild Hall earned
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRanking)
	// Temporary canned response
	data, _ := hex.DecodeString("00000515000005150000CEB4000003CE000003CE0000CEB44D49444E494748542D414E47454C0000000000000000000000")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgMhfGetUdTacticsLog(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented
