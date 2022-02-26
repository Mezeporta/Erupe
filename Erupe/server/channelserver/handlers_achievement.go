package channelserver

import (
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgMhfGetAchievement(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAchievement)

	achievementStruct := []struct {
		ID   uint8  // Main ID
		Unk0 uint8  // always FF
		Unk1 uint16 // 0x05 0x00
		Unk2 uint32 // 0x01 0x0A 0x05 0x00
	}{
		{ID: 0, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 1, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 2, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 3, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 4, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 5, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 6, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 7, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 8, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 9, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 10, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 11, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 12, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 13, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 14, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 15, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 16, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 17, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 18, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 19, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 20, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 21, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 22, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 23, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 24, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 25, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 26, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 27, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 28, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 29, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 30, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 31, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 32, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 33, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 34, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 35, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 36, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 37, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 38, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 39, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 40, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 41, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 42, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 43, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 44, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 45, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 46, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 47, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 48, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 49, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 50, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 51, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 52, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 53, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 54, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 55, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 56, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 57, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 58, Unk0: 0xFF, Unk1: 0, Unk2: 0},
		{ID: 59, Unk0: 0xFF, Unk1: 0, Unk2: 0},
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(achievementStruct))) // Entry count
	for _, entry := range achievementStruct {
		resp.WriteUint8(entry.ID)
		resp.WriteUint8(entry.Unk0)
		resp.WriteUint16(entry.Unk1)
		resp.WriteUint32(entry.Unk2)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfSetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetCaAchievementHist)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfResetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPaymentAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfDisplayedAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetCaAchievement(s *Session, p mhfpacket.MHFPacket) {}
