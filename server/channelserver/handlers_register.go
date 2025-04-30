package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"strings"
)

func handleMsgMhfRegisterEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegisterEvent)
	bf := byteframe.NewByteFrame()
	// Some kind of check if there's already a session
	if pkt.Unk1 && s.server.getRaviSemaphore() == nil {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf.WriteUint8(uint8(pkt.WorldID))
	bf.WriteUint8(uint8(pkt.LandID))
	bf.WriteUint16(s.server.raviente.id)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReleaseEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReleaseEvent)

	// Do this ack manually because it uses a non-(0|1) error code
	/*
		_ACK_SUCCESS = 0
		_ACK_ERROR = 1

		_ACK_EINPROGRESS = 16
		_ACK_ENOENT = 17
		_ACK_ENOSPC = 18
		_ACK_ETIMEOUT = 19

		_ACK_EINVALID = 64
		_ACK_EFAILED = 65
		_ACK_ENOMEM = 66
		_ACK_ENOTEXIT = 67
		_ACK_ENOTREADY = 68
		_ACK_EALREADY = 69
		_ACK_DISABLE_WORK = 71
	*/
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        pkt.AckHandle,
		IsBufferResponse: false,
		ErrorCode:        0x41,
		AckData:          []byte{0x00, 0x00, 0x00, 0x00},
	})
}

type RaviUpdate struct {
	Op   uint8
	Dest uint8
	Data uint32
}

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysOperateRegister)

	var raviUpdates []RaviUpdate
	var raviUpdate RaviUpdate
	// Strip null terminator
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload[:len(pkt.RawDataPayload)-1])
	for i := len(pkt.RawDataPayload) / 6; i > 0; i-- {
		raviUpdate.Op = bf.ReadUint8()
		raviUpdate.Dest = bf.ReadUint8()
		raviUpdate.Data = bf.ReadUint32()
		raviUpdates = append(raviUpdates, raviUpdate)
	}
	bf = byteframe.NewByteFrame()

	var _old, _new uint32
	s.server.raviente.Lock()
	for _, update := range raviUpdates {
		switch update.Op {
		case 2:
			_old, _new = s.server.UpdateRavi(pkt.SemaphoreID, update.Dest, update.Data, true)
		case 13, 14:
			_old, _new = s.server.UpdateRavi(pkt.SemaphoreID, update.Dest, update.Data, false)
		}
		bf.WriteUint8(1)
		bf.WriteUint8(update.Dest)
		bf.WriteUint32(_old)
		bf.WriteUint32(_new)
	}
	s.server.raviente.Unlock()
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())

	if s.server.erupeConfig.GameplayOptions.LowLatencyRaviente {
		s.notifyRavi()
	}
}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLoadRegister)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	bf.WriteUint8(pkt.Values)
	for i := uint8(0); i < pkt.Values; i++ {
		switch pkt.RegisterID {
		case 0x40000:
			bf.WriteUint32(s.server.raviente.state[i])
		case 0x50000:
			bf.WriteUint32(s.server.raviente.support[i])
		case 0x60000:
			bf.WriteUint32(s.server.raviente.register[i])
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func (s *Session) notifyRavi() {
	sema := s.server.getRaviSemaphore()
	if sema == nil {
		return
	}
	var temp mhfpacket.MHFPacket
	raviNotif := byteframe.NewByteFrame()
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: 0x40000}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: 0x50000}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: 0x60000}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	raviNotif.WriteUint16(0x0010) // End it.
	if s.server.erupeConfig.GameplayOptions.LowLatencyRaviente {
		for session := range sema.clients {
			session.QueueSendNonBlocking(raviNotif.Data())
		}
	} else {
		for session := range sema.clients {
			if session.charID == s.charID {
				session.QueueSendNonBlocking(raviNotif.Data())
			}
		}
	}
}

func (s *Server) getRaviSemaphore() *Semaphore {
	for _, semaphore := range s.semaphore {
		if strings.HasPrefix(semaphore.name, "hs_l0") && strings.HasSuffix(semaphore.name, "3") {
			return semaphore
		}
	}
	return nil
}

func handleMsgSysNotifyRegister(s *Session, p mhfpacket.MHFPacket) {}
