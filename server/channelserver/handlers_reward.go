package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetAdditionalBeatReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAdditionalBeatReward)
	// Actual response in packet captures are all just giant batches of null bytes
	// I'm assuming this is because it used to be tied to an actual event and
	// they never bothered killing off the packet when they made it static
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0x104))
}

func handleMsgMhfGetUdRankingRewardList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdRankingRewardList)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // Len
	// Format
	// uint8 Unk
	// uint16 Unk
	// uint16 Unk
	// uint8 Unk
	// uint32 Unk
	// uint32 Unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRewardSong(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRewardSong)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)           // No error
	bf.WriteUint8(0)           // usage count - common
	bf.WriteUint32(0)          // Prayer ID
	bf.WriteUint32(0xFFFFFFFF) // Prayer end
	for i := 1; i < 5; i++ {
		bf.WriteUint8(0)   // No error
		bf.WriteUint8(i)   // ColorID
		bf.WriteUint8(0)   // usage count - Only Bead of Storms
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUseRewardSong(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseRewardSong)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0})
}

func handleMsgMhfAddRewardSongCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddRewardSongCount)
        //Only Bead of Storms add usage count
	doAckBufSucceed(s, pkt.AckHandle, []byte{0})
}

func handleMsgMhfAcquireMonthlyReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyReward)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfAcceptReadReward(s *Session, p mhfpacket.MHFPacket) {}
