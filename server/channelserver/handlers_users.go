package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgSysInsertUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetUserBinary)
	if pkt.BinaryType < 1 || pkt.BinaryType > 5 {
		s.logger.Warn("Invalid BinaryType", zap.Uint8("type", pkt.BinaryType))
		return
	}
	s.server.userBinary.Set(s.charID, pkt.BinaryType, pkt.RawDataPayload)

	s.server.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.charID,
		BinaryType: pkt.BinaryType,
	}, s)
}

func handleMsgSysGetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetUserBinary)

	data, ok := s.server.userBinary.Get(pkt.CharID, pkt.BinaryType)

	if !ok {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, data)
	}
}

func handleMsgSysNotifyUserBinary(s *Session, p mhfpacket.MHFPacket) {}
