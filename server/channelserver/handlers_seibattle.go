package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"time"
)

// SeibattleTimetable represents a seibattle schedule entry.
type SeibattleTimetable struct {
	Start time.Time
	End   time.Time
}

// SeibattleKeyScore represents a seibattle key score.
type SeibattleKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

// SeibattleCareer represents seibattle career stats.
type SeibattleCareer struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}

// SeibattleOpponent represents seibattle opponent data.
type SeibattleOpponent struct {
	Unk0 int32
	Unk1 int8
}

// SeibattleConventionResult represents a seibattle convention result.
type SeibattleConventionResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
}

// SeibattleCharScore represents a seibattle per-character score.
type SeibattleCharScore struct {
	Unk0 uint32
}

// SeibattleCurResult represents a seibattle current result.
type SeibattleCurResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
}

// Seibattle represents complete seibattle data.
type Seibattle struct {
	Timetable        []SeibattleTimetable
	KeyScore         []SeibattleKeyScore
	Career           []SeibattleCareer
	Opponent         []SeibattleOpponent
	ConventionResult []SeibattleConventionResult
	CharScore        []SeibattleCharScore
	CurResult        []SeibattleCurResult
}

func handleMsgMhfGetSeibattle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetSeibattle)
	var data []*byteframe.ByteFrame
	seibattle := Seibattle{
		Timetable: []SeibattleTimetable{
			{TimeMidnight(), TimeMidnight().Add(time.Hour * 8)},
			{TimeMidnight().Add(time.Hour * 8), TimeMidnight().Add(time.Hour * 16)},
			{TimeMidnight().Add(time.Hour * 16), TimeMidnight().Add(time.Hour * 24)},
		},
		KeyScore: []SeibattleKeyScore{
			{0, 0},
		},
		Career: []SeibattleCareer{
			{0, 0, 0},
		},
		Opponent: []SeibattleOpponent{
			{1, 1},
		},
		ConventionResult: []SeibattleConventionResult{
			{0, 0, 0, 0, 0},
		},
		CharScore: []SeibattleCharScore{
			{0},
		},
		CurResult: []SeibattleCurResult{
			{0, 0, 0, 0},
		},
	}

	switch pkt.Type {
	case 1:
		for _, timetable := range seibattle.Timetable {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(uint32(timetable.Start.Unix()))
			bf.WriteUint32(uint32(timetable.End.Unix()))
			data = append(data, bf)
		}
	case 3: // Key score?
		for _, keyScore := range seibattle.KeyScore {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(keyScore.Unk0)
			bf.WriteInt32(keyScore.Unk1)
			data = append(data, bf)
		}
	case 4: // Career?
		for _, career := range seibattle.Career {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(career.Unk0)
			bf.WriteUint16(career.Unk1)
			bf.WriteUint16(career.Unk2)
			data = append(data, bf)
		}
	case 5: // Opponent?
		for _, opponent := range seibattle.Opponent {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(opponent.Unk0)
			bf.WriteInt8(opponent.Unk1)
			data = append(data, bf)
		}
	case 6: // Convention result?
		for _, conventionResult := range seibattle.ConventionResult {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(conventionResult.Unk0)
			bf.WriteUint16(conventionResult.Unk1)
			bf.WriteUint16(conventionResult.Unk2)
			bf.WriteUint16(conventionResult.Unk3)
			bf.WriteUint16(conventionResult.Unk4)
			data = append(data, bf)
		}
	case 7: // Char score?
		for _, charScore := range seibattle.CharScore {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(charScore.Unk0)
			data = append(data, bf)
		}
	case 8: // Cur result?
		for _, curResult := range seibattle.CurResult {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(curResult.Unk0)
			bf.WriteUint16(curResult.Unk1)
			bf.WriteUint16(curResult.Unk2)
			bf.WriteUint16(curResult.Unk3)
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostSeibattle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostSeibattle)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetBreakSeibatuLevelReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBreakSeibatuLevelReward)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// WeeklySeibatuRankingReward represents a weekly seibattle ranking reward.
type WeeklySeibatuRankingReward struct {
	Unk0 int32
	Unk1 int32
	Unk2 uint32
	Unk3 int32
	Unk4 int32
	Unk5 int32
}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySeibatuRankingReward)
	var data []*byteframe.ByteFrame
	weeklySeibatuRankingRewards := []WeeklySeibatuRankingReward{
		{0, 0, 0, 0, 0, 0},
	}
	for _, reward := range weeklySeibatuRankingRewards {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(reward.Unk0)
		bf.WriteInt32(reward.Unk1)
		bf.WriteUint32(reward.Unk2)
		bf.WriteInt32(reward.Unk3)
		bf.WriteInt32(reward.Unk4)
		bf.WriteInt32(reward.Unk5)
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfGetFixedSeibatuRankingTable(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFixedSeibatuRankingTable)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteBytes(make([]byte, 32))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevel(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevel)

	// This response is fixed and will never change on JP,
	// but I've left it dynamic for possible other client differences.
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(pkt.ValidIDCount); i++ {
		resp.WriteUint32(pkt.IDs[i])
		resp.WriteUint32(1)
		resp.WriteUint32(1)
		resp.WriteUint32(1)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfReadLastWeekBeatRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadLastWeekBeatRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateBeatLevel(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateBeatLevel)

	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadBeatLevelAllRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevelAllRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)

	for i := 0; i < 100; i++ {
		bf.WriteUint32(0)
		bf.WriteUint32(0)
		bf.WriteBytes(make([]byte, 32))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevelMyRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevelMyRanking)
	bf := byteframe.NewByteFrame()
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
