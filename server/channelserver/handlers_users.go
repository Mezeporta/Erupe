package channelserver

import (
	"fmt"

	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgSysInsertUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetUserBinary)
	s.server.userBinaryPartsLock.Lock()
	s.server.userBinaryParts[userBinaryPartID{charID: s.charID, index: pkt.BinaryType}] = pkt.RawDataPayload
	s.server.userBinaryPartsLock.Unlock()

	var exists []byte
	err := s.server.db.QueryRow("SELECT type2 FROM user_binary WHERE id=$1", s.charID).Scan(&exists)
	if err != nil {
		if _, err := s.server.db.Exec("INSERT INTO user_binary (id) VALUES ($1)", s.charID); err != nil {
			s.logger.Error("Failed to insert user binary", zap.Error(err))
		}
	}

	if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE user_binary SET type%d=$1 WHERE id=$2", pkt.BinaryType), pkt.RawDataPayload, s.charID); err != nil {
		s.logger.Error("Failed to update user binary", zap.Error(err))
	}

	msg := &mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.charID,
		BinaryType: pkt.BinaryType,
	}

	s.server.BroadcastMHF(msg, s)
}

func handleMsgSysGetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetUserBinary)

	// Try to get the data.
	s.server.userBinaryPartsLock.RLock()
	defer s.server.userBinaryPartsLock.RUnlock()
	data, ok := s.server.userBinaryParts[userBinaryPartID{charID: pkt.CharID, index: pkt.BinaryType}]

	// If we can't get the real data, try to get it from the database.
	if !ok {
		err := s.server.db.QueryRow(fmt.Sprintf("SELECT type%d FROM user_binary WHERE id=$1", pkt.BinaryType), pkt.CharID).Scan(&data)
		if err != nil {
			doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		} else {
			doAckBufSucceed(s, pkt.AckHandle, data)
		}
	} else {
		doAckBufSucceed(s, pkt.AckHandle, data)
	}
}

func handleMsgSysNotifyUserBinary(s *Session, p mhfpacket.MHFPacket) {}
