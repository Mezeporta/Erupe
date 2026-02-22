package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgSysEnumerateClient(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateClient)

	stage, ok := s.server.stages.Get(pkt.StageID)
	if !ok {
		s.logger.Warn("Can't enumerate clients for stage that doesn't exist!", zap.String("stageID", pkt.StageID))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	// Read-lock the stage and make the response with all of the charID's in the stage.
	resp := byteframe.NewByteFrame()
	stage.RLock()
	var clients []uint32
	switch pkt.Get {
	case 0: // All
		for _, cid := range stage.clients {
			clients = append(clients, cid)
		}
		for cid := range stage.reservedClientSlots {
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

	var count uint32
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Blacklist count
	csv, err := s.server.charRepo.ReadString(s.charID, "blocked")
	if err == nil {
		cids := stringsupport.CSVElems(csv)
		for _, cid := range cids {
			name, err := s.server.charRepo.GetName(uint32(cid))
			if err != nil {
				continue
			}
			count++
			resp.WriteUint32(uint32(cid))
			resp.WriteUint32(16)
			resp.WriteBytes(stringsupport.PaddedString(name, 16, true))
		}
	}
	_, _ = resp.Seek(0, 0)
	resp.WriteUint32(count)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfOprMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprMember)
	for _, cid := range pkt.CharIDs {
		if pkt.Blacklist {
			csv, err := s.server.charRepo.ReadString(s.charID, "blocked")
			if err == nil {
				if pkt.Operation {
					csv = stringsupport.CSVRemove(csv, int(cid))
				} else {
					csv = stringsupport.CSVAdd(csv, int(cid))
				}
				if err := s.server.charRepo.SaveString(s.charID, "blocked", csv); err != nil {
					s.logger.Error("Failed to update blocked list", zap.Error(err))
				}
			}
		} else { // Friendlist
			csv, err := s.server.charRepo.ReadString(s.charID, "friends")
			if err == nil {
				if pkt.Operation {
					csv = stringsupport.CSVRemove(csv, int(cid))
				} else {
					csv = stringsupport.CSVAdd(csv, int(cid))
				}
				if err := s.server.charRepo.SaveString(s.charID, "friends", csv); err != nil {
					s.logger.Error("Failed to update friends list", zap.Error(err))
				}
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfShutClient(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysHideClient(s *Session, p mhfpacket.MHFPacket) {}
