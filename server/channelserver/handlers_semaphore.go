package channelserver

import (
	"erupe-ce/common/byteframe"
	"go.uber.org/zap"
	"strconv"
	"strings"

	"erupe-ce/network/mhfpacket"
)

func removeSessionFromSemaphore(s *Session) {
	s.server.semaphoreLock.Lock()
	for _, semaphore := range s.server.semaphore {
		if _, exists := semaphore.clients[s]; exists {
			delete(semaphore.clients, s)
		}
	}
	s.server.semaphoreLock.Unlock()
}

func handleMsgSysCreateSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateSemaphore)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x03, 0x00, 0x0d})
}

func destructEmptySemaphores(s *Session) {
	s.server.semaphoreLock.Lock()
	for id, sema := range s.server.semaphore {
		if len(sema.clients) == 0 {
			delete(s.server.semaphore, id)
			if strings.HasPrefix(id, "hs_l0") {
				s.server.resetRaviente()
			}
			s.logger.Debug("Destructed semaphore", zap.String("sema.name", id))
		}
	}
	s.server.semaphoreLock.Unlock()
}

func handleMsgSysDeleteSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysDeleteSemaphore)
	destructEmptySemaphores(s)
	s.server.semaphoreLock.Lock()
	for id, sema := range s.server.semaphore {
		if sema.id == pkt.SemaphoreID {
			for session := range sema.clients {
				if s == session {
					delete(sema.clients, s)
				}
			}
			if len(sema.clients) == 0 {
				delete(s.server.semaphore, id)
				if strings.HasPrefix(id, "hs_l0") {
					s.server.resetRaviente()
				}
				s.logger.Debug("Destructed semaphore", zap.String("sema.name", id))
			}
		}
	}
	s.server.semaphoreLock.Unlock()
}

func handleMsgSysCreateAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateAcquireSemaphore)
	SemaphoreID := pkt.SemaphoreID

	if s.server.HasSemaphore(s) {
		s.semaphoreMode = !s.semaphoreMode
	}
	if s.semaphoreMode {
		s.semaphoreID[1]++
	} else {
		s.semaphoreID[0]++
	}

	newSemaphore, exists := s.server.semaphore[SemaphoreID]
	if !exists {
		s.server.semaphoreLock.Lock()
		if strings.HasPrefix(SemaphoreID, "hs_l0") {
			suffix, _ := strconv.Atoi(pkt.SemaphoreID[len(pkt.SemaphoreID)-1:])
			s.server.semaphore[SemaphoreID] = &Semaphore{
				name:       pkt.SemaphoreID,
				id:         uint32((suffix + 1) * 0x10000),
				clients:    make(map[*Session]uint32),
				maxPlayers: 127,
			}
		} else {
			s.server.semaphore[SemaphoreID] = NewSemaphore(s, SemaphoreID, 1)
		}
		newSemaphore = s.server.semaphore[SemaphoreID]
		s.server.semaphoreLock.Unlock()
	}

	newSemaphore.Lock()
	defer newSemaphore.Unlock()
	bf := byteframe.NewByteFrame()
	if _, exists := newSemaphore.clients[s]; exists {
		bf.WriteUint32(newSemaphore.id)
	} else if uint16(len(newSemaphore.clients)) < newSemaphore.maxPlayers {
		newSemaphore.clients[s] = s.charID
		s.Lock()
		s.semaphore = newSemaphore
		s.Unlock()
		bf.WriteUint32(newSemaphore.id)
	} else {
		bf.WriteUint32(0)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgSysAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysAcquireSemaphore)
	if sema, exists := s.server.semaphore[pkt.SemaphoreID]; exists {
		sema.host = s
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(sema.id)
		doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgSysReleaseSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysReleaseSemaphore)
}

func handleMsgSysCheckSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCheckSemaphore)
	resp := []byte{0x00, 0x00, 0x00, 0x00}
	s.server.semaphoreLock.Lock()
	if _, exists := s.server.semaphore[pkt.SemaphoreID]; exists {
		resp = []byte{0x00, 0x00, 0x00, 0x01}
	}
	s.server.semaphoreLock.Unlock()
	doAckSimpleSucceed(s, pkt.AckHandle, resp)
}
