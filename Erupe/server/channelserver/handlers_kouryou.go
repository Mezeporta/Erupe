package channelserver

import (
	"github.com/Solenataris/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
)

func handleMsgMhfAddKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	// hunting with both ranks maxed gets you these
	pkt := p.(*mhfpacket.MsgMhfAddKouryouPoint)
	var points int
	err := s.server.db.QueryRow("UPDATE characters SET kouryou_point=COALESCE(kouryou_point + $1, $1) WHERE id=$2 RETURNING kouryou_point", pkt.KouryouPoints, s.charID).Scan(&points)
	if err != nil {
		s.logger.Fatal("Failed to update KouryouPoint in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKouryouPoint)
	var points int
	err := s.server.db.QueryRow("SELECT COALESCE(kouryou_point, 0) FROM characters WHERE id = $1", s.charID).Scan(&points)
	if err != nil {
		s.logger.Fatal("Failed to get kouryou_point savedata from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfExchangeKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	// spent at the guildmaster, 10000 a roll
	var points int
	pkt := p.(*mhfpacket.MsgMhfExchangeKouryouPoint)
	err := s.server.db.QueryRow("UPDATE characters SET kouryou_point=kouryou_point - $1 WHERE id=$2 RETURNING kouryou_point", pkt.KouryouPoints, s.charID).Scan(&points)
	if err != nil {
		s.logger.Fatal("Failed to update platemyset savedata in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}
