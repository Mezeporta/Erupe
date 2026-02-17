package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"io"
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

type Achievement struct {
	Level     uint8
	Value     uint32
	NextValue uint16
	Required  uint32
	Updated   bool
	Progress  uint32
	Trophy    uint8
}

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
				ach.Trophy = 0x40
			case 7:
				ach.NextValue = 20
				ach.Trophy = 0x60
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
	ach.Trophy = 0x7F
	ach.Progress = ach.Required
	return ach
}

func handleMsgMhfGetAchievement(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAchievement)

	var exists int
	err := s.server.db.QueryRow("SELECT id FROM achievements WHERE id=$1", pkt.CharID).Scan(&exists)
	if err != nil {
		s.server.db.Exec("INSERT INTO achievements (id) VALUES ($1)", pkt.CharID)
	}

	var scores [33]int32
	err = s.server.db.QueryRow("SELECT * FROM achievements WHERE id=$1", pkt.CharID).Scan(&scores[0],
		&scores[0], &scores[1], &scores[2], &scores[3], &scores[4], &scores[5], &scores[6], &scores[7], &scores[8],
		&scores[9], &scores[10], &scores[11], &scores[12], &scores[13], &scores[14], &scores[15], &scores[16],
		&scores[17], &scores[18], &scores[19], &scores[20], &scores[21], &scores[22], &scores[23], &scores[24],
		&scores[25], &scores[26], &scores[27], &scores[28], &scores[29], &scores[30], &scores[31], &scores[32])
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 20))
		return
	}

	resp := byteframe.NewByteFrame()
	var points uint32
	resp.WriteBytes(make([]byte, 16))
	resp.WriteBytes([]byte{0x02, 0x00, 0x00}) // Unk

	var id uint8
	entries := uint8(33)
	resp.WriteUint8(entries) // Entry count
	for id = 0; id < entries; id++ {
		achData := GetAchData(id, scores[id])
		points += achData.Value
		resp.WriteUint8(id)
		resp.WriteUint8(achData.Level)
		resp.WriteUint16(achData.NextValue)
		resp.WriteUint32(achData.Required)
		resp.WriteBool(false) // TODO: Notify on rank increase since last checked, see MhfDisplayedAchievement
		resp.WriteUint8(achData.Trophy)
		/* Trophy bitfield
		0000 0000
		abcd efgh
		B - Bronze (0x40)
		B-C - Silver (0x60)
		B-H - Gold (0x7F)
		*/
		resp.WriteUint16(0) // Unk
		resp.WriteUint32(achData.Progress)
	}
	_, _ = resp.Seek(0, io.SeekStart)
	resp.WriteUint32(points)
	resp.WriteUint32(points)
	resp.WriteUint32(points)
	resp.WriteUint32(points)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfSetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetCaAchievementHist)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfResetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddAchievement(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddAchievement)

	var exists int
	err := s.server.db.QueryRow("SELECT id FROM achievements WHERE id=$1", s.charID).Scan(&exists)
	if err != nil {
		s.server.db.Exec("INSERT INTO achievements (id) VALUES ($1)", s.charID)
	}

	s.server.db.Exec(fmt.Sprintf("UPDATE achievements SET ach%d=ach%d+1 WHERE id=$1", pkt.AchievementID, pkt.AchievementID), s.charID)
}

func handleMsgMhfPaymentAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfDisplayedAchievement(s *Session, p mhfpacket.MHFPacket) {
	// This is how you would figure out if the rank-up notification needs to occur
}

func handleMsgMhfGetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetCaAchievement(s *Session, p mhfpacket.MHFPacket) {}
