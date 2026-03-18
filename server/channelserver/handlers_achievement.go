package channelserver

import (
	"io"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// Achievement trophy tier thresholds (bitfield values)
const (
	AchievementTrophyBronze = uint8(0x40)
	AchievementTrophySilver = uint8(0x60)
	AchievementTrophyGold   = uint8(0x7F)
)

var achievementCurves = [][]int32{
	// 0: HR weapon use, Class use, Tore dailies
	{5, 15, 30, 50, 100, 150, 200, 300},
	// 1: Weapon collector, G wep enhances
	{1, 5, 10, 15, 30, 50, 75, 100},
	// 2: Festa wins
	{1, 2, 3, 4, 5, 6, 7, 8},
	// 3: GR weapon use, Sigil crafts
	{10, 50, 100, 200, 350, 500, 750, 999},
}

var achievementCurveMap = map[uint8][]int32{
	0: achievementCurves[0], 1: achievementCurves[0], 2: achievementCurves[0], 3: achievementCurves[0],
	4: achievementCurves[0], 5: achievementCurves[0], 6: achievementCurves[0], 7: achievementCurves[1],
	8: achievementCurves[2], 9: achievementCurves[0], 10: achievementCurves[0], 11: achievementCurves[0],
	12: achievementCurves[0], 13: achievementCurves[0], 14: achievementCurves[0], 15: achievementCurves[0],
	16: achievementCurves[3], 17: achievementCurves[3], 18: achievementCurves[3], 19: achievementCurves[3],
	20: achievementCurves[3], 21: achievementCurves[3], 22: achievementCurves[3], 23: achievementCurves[3],
	24: achievementCurves[3], 25: achievementCurves[3], 26: achievementCurves[3], 27: achievementCurves[1],
	28: achievementCurves[1], 29: achievementCurves[3], 30: achievementCurves[3], 31: achievementCurves[3],
	32: achievementCurves[3],
}

// Achievement represents computed achievement data for a character.
type Achievement struct {
	Level     uint8
	Value     uint32
	NextValue uint16
	Required  uint32
	Updated   bool
	Progress  uint32
	Trophy    uint8
}

// GetAchData computes achievement level and progress from a raw score.
func GetAchData(id uint8, score int32) Achievement {
	curve := achievementCurveMap[id]
	var ach Achievement
	for i, v := range curve {
		temp := score - v
		if temp < 0 {
			ach.Progress = uint32(score)
			ach.Required = uint32(curve[i])
			switch ach.Level {
			case 0:
				ach.NextValue = 5
			case 1, 2, 3:
				ach.NextValue = 10
			case 4, 5:
				ach.NextValue = 15
			case 6:
				ach.NextValue = 15
				ach.Trophy = AchievementTrophyBronze
			case 7:
				ach.NextValue = 20
				ach.Trophy = AchievementTrophySilver
			}
			return ach
		} else {
			score = temp
			ach.Level++
			switch ach.Level {
			case 1:
				ach.Value += 5
			case 2, 3, 4:
				ach.Value += 10
			case 5, 6, 7:
				ach.Value += 15
			case 8:
				ach.Value += 20
			}
		}
	}
	ach.Required = uint32(curve[7])
	ach.Trophy = AchievementTrophyGold
	ach.Progress = ach.Required
	return ach
}

func handleMsgMhfGetAchievement(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAchievement)

	summary, err := s.server.achievementService.GetAll(pkt.CharID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 20))
		return
	}

	resp := byteframe.NewByteFrame()
	resp.WriteBytes(make([]byte, 16))
	resp.WriteBytes([]byte{0x02, 0x00, 0x00}) // Unk

	resp.WriteUint8(achievementEntryCount)
	for id := uint8(0); id < achievementEntryCount; id++ {
		ach := summary.Achievements[id]
		resp.WriteUint8(id)
		resp.WriteUint8(ach.Level)
		resp.WriteUint16(ach.NextValue)
		resp.WriteUint32(ach.Required)
		resp.WriteBool(summary.Notify[id])
		resp.WriteUint8(ach.Trophy)
		/* Trophy bitfield
		0000 0000
		abcd efgh
		B - Bronze (0x40)
		B-C - Silver (0x60)
		B-H - Gold (0x7F)
		*/
		resp.WriteUint16(0) // Unk
		resp.WriteUint32(ach.Progress)
	}
	_, _ = resp.Seek(0, io.SeekStart)
	resp.WriteUint32(summary.Points)
	resp.WriteUint32(summary.Points)
	resp.WriteUint32(summary.Points)
	resp.WriteUint32(summary.Points)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfSetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetCaAchievementHist)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfResetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddAchievement(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddAchievement)

	if err := s.server.achievementService.Increment(s.charID, pkt.AchievementID); err != nil {
		s.logger.Warn("Failed to increment achievement", zap.Error(err))
	}
}

func handleMsgMhfPaymentAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfDisplayedAchievement(s *Session, p mhfpacket.MHFPacket) {
	if s.server.achievementService == nil {
		return
	}
	if err := s.server.achievementService.MarkDisplayed(s.charID); err != nil {
		s.logger.Warn("Failed to mark achievements as displayed", zap.Error(err))
	}
}

func handleMsgMhfGetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetCaAchievement(s *Session, p mhfpacket.MHFPacket) {}
