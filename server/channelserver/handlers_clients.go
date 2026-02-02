package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgSysEnumerateClient(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateClient)

	s.server.stagesLock.RLock()
	stage, ok := s.server.stages[pkt.StageID]
	if !ok {
		s.server.stagesLock.RUnlock()
		s.logger.Warn("Can't enumerate clients for stage that doesn't exist!", zap.String("stageID", pkt.StageID))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	s.server.stagesLock.RUnlock()

	// Read-lock the stage and make the response with all of the charID's in the stage.
	resp := byteframe.NewByteFrame()
	stage.RLock()
	var clients []uint32
	switch pkt.Get {
	case 0: // All
		for _, cid := range stage.clients {
			clients = append(clients, cid)
		}
	case 1: // Not ready
		for cid, ready := range stage.reservedClientSlots {
			if !ready {
				clients = append(clients, cid)
			}
		}
	case 2: // Ready
		for cid, ready := range stage.reservedClientSlots {
			if ready {
				clients = append(clients, cid)
			}
		}
	}
	resp.WriteUint16(uint16(len(clients)))
	for _, cid := range clients {
		resp.WriteUint32(cid)
	}
	stage.RUnlock()

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	s.logger.Debug("MsgSysEnumerateClient Done!")
}

func handleMsgMhfListMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMember)

	var csv string
	var count uint32
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Blacklist count
	err := s.server.db.QueryRow("SELECT blocked FROM characters WHERE id=$1", s.charID).Scan(&csv)
	if err != nil {
		s.logger.Error("failed to get blocked list", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		return
	}
	cids := stringsupport.CSVElems(csv)
	for _, cid := range cids {
		var name string
		err = s.server.db.QueryRow("SELECT name FROM characters WHERE id=$1", cid).Scan(&name)
		if err != nil {
			continue
		}
		count++
		resp.WriteUint32(uint32(cid))
		resp.WriteUint32(16)
		resp.WriteBytes(stringsupport.PaddedString(name, 16, true))
	}
	resp.Seek(0, 0)
	resp.WriteUint32(count)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfOprMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprMember)
	var csv string
	if pkt.Blacklist {
		err := s.server.db.QueryRow("SELECT blocked FROM characters WHERE id=$1", s.charID).Scan(&csv)
		if err != nil {
			s.logger.Error("failed to get blocked list for operation", zap.Error(err), zap.Uint32("charID", s.charID))
			doAckSimpleFail(s, pkt.AckHandle, nil)
			return
		}
		if pkt.Operation {
			csv = stringsupport.CSVRemove(csv, int(pkt.CharID))
		} else {
			csv = stringsupport.CSVAdd(csv, int(pkt.CharID))
		}
		s.server.db.Exec("UPDATE characters SET blocked=$1 WHERE id=$2", csv, s.charID)
	} else { // Friendlist
		err := s.server.db.QueryRow("SELECT friends FROM characters WHERE id=$1", s.charID).Scan(&csv)
		if err != nil {
			s.logger.Error("failed to get friends list for operation", zap.Error(err), zap.Uint32("charID", s.charID))
			doAckSimpleFail(s, pkt.AckHandle, nil)
			return
		}
		if pkt.Operation {
			csv = stringsupport.CSVRemove(csv, int(pkt.CharID))
		} else {
			csv = stringsupport.CSVAdd(csv, int(pkt.CharID))
		}
		s.server.db.Exec("UPDATE characters SET friends=$1 WHERE id=$2", csv, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfShutClient(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysHideClient(s *Session, p mhfpacket.MHFPacket) {}
