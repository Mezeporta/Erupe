package channelserver

import "github.com/Solenataris/Erupe/network/mhfpacket"

func handleMsgMhfEnumerateGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildTresure)

	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfRegistGuildTresure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireGuildTresure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateGuildTresureReport(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTresureSouvenir)

	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 6))
}

func handleMsgMhfAcquireGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {}
