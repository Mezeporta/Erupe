package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"fmt"

	"go.uber.org/zap"
)

func getGoocooData(s *Session, cid uint32) [][]byte {
	var goocoos [][]byte
	for i := uint32(0); i < 5; i++ {
		goocoo, err := s.server.goocooRepo.GetSlot(cid, i)
		if err != nil {
			if err := s.server.goocooRepo.EnsureExists(s.charID); err != nil {
				s.logger.Error("Failed to insert goocoo record", zap.Error(err))
			}
			return goocoos
		}
		if goocoo != nil {
			goocoos = append(goocoos, goocoo)
		}
	}
	return goocoos
}

func handleMsgMhfEnumerateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuacot)
	bf := byteframe.NewByteFrame()
	goocoos := getGoocooData(s, s.charID)
	bf.WriteUint16(uint16(len(goocoos)))
	bf.WriteUint16(0)
	for _, goocoo := range goocoos {
		bf.WriteBytes(goocoo)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuacot)
	for _, goocoo := range pkt.Goocoos {
		if goocoo.Index > 4 {
			continue
		}
		if goocoo.Data1[0] == 0 {
			if err := s.server.goocooRepo.ClearSlot(s.charID, goocoo.Index); err != nil {
				s.logger.Error("Failed to clear goocoo slot", zap.Error(err))
			}
		} else {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(goocoo.Index)
			for i := range goocoo.Data1 {
				bf.WriteInt16(goocoo.Data1[i])
			}
			for i := range goocoo.Data2 {
				bf.WriteUint32(goocoo.Data2[i])
			}
			bf.WriteUint8(uint8(len(goocoo.Name)))
			bf.WriteBytes(goocoo.Name)
			if err := s.server.goocooRepo.SaveSlot(s.charID, goocoo.Index, bf.Data()); err != nil {
				s.logger.Error("Failed to update goocoo slot", zap.Error(err))
			}
			dumpSaveData(s, bf.Data(), fmt.Sprintf("goocoo-%d", goocoo.Index))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
