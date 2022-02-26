package channelserver

import (
	"fmt"

	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgSysCreateSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateSemaphore)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x03, 0x00, 0x0d})
}

func handleMsgSysDeleteSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysDeleteSemaphore)

	s.semaphore.Lock()
	for id := range s.server.semaphore {
		delete(s.server.semaphore[id].reservedClientSlots, s.charID)
	}
	s.semaphore.Unlock()
}

func handleMsgSysCreateAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateAcquireSemaphore)
	SemaphoreID := pkt.SemaphoreID

	newSemaphore, gotNewStage := s.server.semaphore[SemaphoreID]

	fmt.Printf("Got reserve stage req, StageID: %v\n\n", SemaphoreID)
	if !gotNewStage {
		s.server.semaphoreLock.Lock()
		if SemaphoreID == "hs_l0u3B51J9k1" ||
			SemaphoreID == "hs_l0u3B51J9k2" ||
			SemaphoreID == "hs_l0u3B51J9k3" ||
			SemaphoreID == "hs_l0u3B51J9k4" ||
			SemaphoreID == "hs_l0u3B51J9k5" {
			s.server.semaphore[SemaphoreID] = NewSemaphore(SemaphoreID, 32)
		} else {
			s.server.semaphore[SemaphoreID] = NewSemaphore(SemaphoreID, 1)
		}
		newSemaphore = s.server.semaphore[SemaphoreID]
		s.server.semaphoreLock.Unlock()
	}

	newSemaphore.Lock()
	defer newSemaphore.Unlock()
	if _, exists := newSemaphore.reservedClientSlots[s.charID]; exists {
		s.logger.Info("IS ALREADY EXIST !")
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0F, 0x00, 0x1D})
	} else if uint16(len(newSemaphore.reservedClientSlots)) < newSemaphore.maxPlayers {
		newSemaphore.reservedClientSlots[s.charID] = nil
		s.Lock()
		s.semaphore = newSemaphore
		s.Unlock()
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0F, 0x00, 0x1D})
	} else {
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgSysAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysAcquireSemaphore)
}

func handleMsgSysReleaseSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysReleaseSemaphore)
	for _, session := range s.server.sessions {
		session.semaphore.Lock()
		for id := range session.server.semaphore {
			delete(s.server.semaphore[id].reservedClientSlots, s.charID)
		}
		session.semaphore.Unlock()
	}
	//data, _ := hex.DecodeString("000180e703000d443b37ff006d00131809000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010627426400a936a93600000100cf330600cc31cc31d431000025000000000000000000010218330600bd3cbd3cbd3c01032c280600ee3dee3da9360104f3300600d231a936a93601054a310600e23ae23ae23a00000d0000000000004d814c0000000003008501d723b7334001e7038b3fd437d516113505000000e7030001000002000203000000000000fafafafafafafafafafafafafafa000000000000ecb2000060da0000000000000000000000000000000000000000000000000000000000000000000000000000181818187e2d00003b31702d662d402e000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000000000")
	//doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgSysCheckSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCheckSemaphore)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
