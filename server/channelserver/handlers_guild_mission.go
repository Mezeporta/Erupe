package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// GuildMission represents a guild mission entry.
type GuildMission struct {
	ID          uint32
	Unk         uint32
	Type        uint16
	Goal        uint16
	Quantity    uint16
	SkipTickets uint16
	GR          bool
	RewardType  uint16
	RewardLevel uint16
}

func handleMsgMhfGetGuildMissionList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildMissionList)
	bf := byteframe.NewByteFrame()
	missions := []GuildMission{
		{431201, 574, 1, 4761, 35, 1, false, 2, 1},
		{431202, 755, 0, 95, 12, 2, false, 3, 2},
		{431203, 746, 0, 95, 6, 1, false, 1, 1},
		{431204, 581, 0, 83, 16, 2, false, 4, 2},
		{431205, 694, 1, 4763, 25, 1, false, 2, 1},
		{431206, 988, 0, 27, 16, 1, false, 6, 1},
		{431207, 730, 1, 4768, 25, 1, false, 4, 1},
		{431208, 680, 1, 3567, 50, 2, false, 2, 2},
		{431209, 1109, 0, 34, 60, 2, false, 6, 2},
		{431210, 128, 1, 8921, 70, 2, false, 3, 2},
		{431211, 406, 0, 59, 10, 1, false, 1, 1},
		{431212, 1170, 0, 70, 90, 3, false, 6, 3},
		{431213, 164, 0, 38, 24, 2, false, 6, 2},
		{431214, 378, 1, 3556, 150, 3, false, 1, 3},
		{431215, 446, 0, 94, 20, 2, false, 4, 2},
	}
	for _, mission := range missions {
		bf.WriteUint32(mission.ID)
		bf.WriteUint32(mission.Unk)
		bf.WriteUint16(mission.Type)
		bf.WriteUint16(mission.Goal)
		bf.WriteUint16(mission.Quantity)
		bf.WriteUint16(mission.SkipTickets)
		bf.WriteBool(mission.GR)
		bf.WriteUint16(mission.RewardType)
		bf.WriteUint16(mission.RewardLevel)
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildMissionRecord(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildMissionRecord)

	const guildMissionRecordSize = 0x190
	// No guild mission records = empty buffer
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, guildMissionRecordSize))
}

func handleMsgMhfAddGuildMissionCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddGuildMissionCount)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfSetGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildMissionTarget)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCancelGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildMissionTarget)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
