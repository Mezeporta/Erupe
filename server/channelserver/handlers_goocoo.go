package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"fmt"

	"go.uber.org/zap"
)

func getGoocooData(s *Session, cid uint32) [][]byte {
	var goocoo []byte
	var goocoos [][]byte
	for i := 0; i < 5; i++ {
		err := s.server.db.QueryRow(fmt.Sprintf("SELECT goocoo%d FROM goocoo WHERE id=$1", i), cid).Scan(&goocoo)
		if err != nil {
			if _, err := s.server.db.Exec("INSERT INTO goocoo (id) VALUES ($1)", s.charID); err != nil {
				s.logger.Error("Failed to insert goocoo record", zap.Error(err))
			}
			return goocoos
		}
		if err == nil && goocoo != nil {
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
		if goocoo.Data1[0] == 0 {
			if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE goocoo SET goocoo%d=NULL WHERE id=$1", goocoo.Index), s.charID); err != nil {
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
			if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE goocoo SET goocoo%d=$1 WHERE id=$2", goocoo.Index), bf.Data(), s.charID); err != nil {
				s.logger.Error("Failed to update goocoo slot", zap.Error(err))
			}
			dumpSaveData(s, bf.Data(), fmt.Sprintf("goocoo-%d", goocoo.Index))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
