package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"time"
)

// RyoudamaReward represents a caravan (Ryoudama) reward entry.
type RyoudamaReward struct {
	Unk0 uint8
	Unk1 uint8
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
	Unk5 uint16
}

// RyoudamaKeyScore represents a caravan key score entry.
type RyoudamaKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

// RyoudamaCharInfo represents per-character caravan info.
type RyoudamaCharInfo struct {
	CID  uint32
	Unk0 int32
	Name string
}

// RyoudamaBoostInfo represents caravan boost status.
type RyoudamaBoostInfo struct {
	Start time.Time
	End   time.Time
}

// Ryoudama represents complete caravan data.
type Ryoudama struct {
	Reward    []RyoudamaReward
	KeyScore  []RyoudamaKeyScore
	CharInfo  []RyoudamaCharInfo
	BoostInfo []RyoudamaBoostInfo
	Score     []int32
}

func handleMsgMhfGetRyoudama(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRyoudama)
	var data []*byteframe.ByteFrame
	ryoudama := Ryoudama{Score: []int32{0}}
	switch pkt.Request2 {
	case 4:
		for _, score := range ryoudama.Score {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(score)
			data = append(data, bf)
		}
	case 5:
		for _, info := range ryoudama.CharInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(info.CID)
			bf.WriteInt32(info.Unk0)
			bf.WriteBytes(stringsupport.PaddedString(info.Name, 14, true))
			data = append(data, bf)
		}
	case 6:
		for _, info := range ryoudama.BoostInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(uint32(info.Start.Unix()))
			bf.WriteUint32(uint32(info.End.Unix()))
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTinyBin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTinyBin)
	// requested after conquest quests
	doAckBufSucceed(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfPostTinyBin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTinyBin)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCaravanMyScore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyScore)
	var data []*byteframe.ByteFrame
	/*
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfCaravanRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanRanking)
	var data []*byteframe.ByteFrame
	/* RYOUDAN
	bf.WriteInt32(1)
	bf.WriteUint32(2)
	bf.WriteBytes(stringsupport.PaddedString("Test", 26, true))
	*/

	/* PERSONAL
	bf.WriteInt32(1)
	bf.WriteBytes(stringsupport.PaddedString("Test", 14, true))
	*/
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfCaravanMyRank(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyRank)
	var data []*byteframe.ByteFrame
	/*
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/
	doAckEarthSucceed(s, pkt.AckHandle, data)
}
