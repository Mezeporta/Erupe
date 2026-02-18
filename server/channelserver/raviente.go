package channelserver

import (
	"strings"
	"sync"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// Raviente holds shared state for the Raviente siege event.
type Raviente struct {
	sync.Mutex
	id       uint16
	register []uint32
	state    []uint32
	support  []uint32
}

func (s *Server) resetRaviente() {
	for _, semaphore := range s.semaphore {
		if strings.HasPrefix(semaphore.name, "hs_l0") {
			return
		}
	}
	s.logger.Debug("All Raviente Semaphores empty, resetting")
	s.raviente.id = s.raviente.id + 1
	s.raviente.register = make([]uint32, 30)
	s.raviente.state = make([]uint32, 30)
	s.raviente.support = make([]uint32, 30)
}

func (s *Server) GetRaviMultiplier() float64 {
	raviSema := s.getRaviSemaphore()
	if raviSema != nil {
		var minPlayers int
		if s.raviente.register[9] > 8 {
			minPlayers = 24
		} else {
			minPlayers = 4
		}
		if len(raviSema.clients) > minPlayers {
			return 1
		}
		return float64(minPlayers / len(raviSema.clients))
	}
	return 0
}

func (s *Server) UpdateRavi(semaID uint32, index uint8, value uint32, update bool) (uint32, uint32) {
	var prev uint32
	var dest *[]uint32
	switch semaID {
	case 0x40000:
		switch index {
		case 17, 28: // Ignore res and poison
			break
		default:
			value = uint32(float64(value) * s.GetRaviMultiplier())
		}
		dest = &s.raviente.state
	case 0x50000:
		dest = &s.raviente.support
	case 0x60000:
		dest = &s.raviente.register
	default:
		return 0, 0
	}
	if update {
		(*dest)[index] += value
	} else {
		(*dest)[index] = value
	}
	return prev, (*dest)[index]
}

func (s *Server) BroadcastRaviente(ip uint32, port uint16, stage []byte, _type uint8) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint16(0)    // Unk
	bf.WriteUint16(0x43) // Data len
	bf.WriteUint16(3)    // Unk len
	var text string
	switch _type {
	case 2:
		text = s.i18n.raviente.berserk
	case 3:
		text = s.i18n.raviente.extreme
	case 4:
		text = s.i18n.raviente.extremeLimited
	case 5:
		text = s.i18n.raviente.berserkSmall
	default:
		s.logger.Error("Unk raviente type", zap.Uint8("_type", _type))
	}
	ps.Uint16(bf, text, true)
	bf.WriteBytes([]byte{0x5F, 0x53, 0x00})
	bf.WriteUint32(ip)   // IP address
	bf.WriteUint16(port) // Port
	bf.WriteUint16(0)    // Unk
	bf.WriteBytes(stage)
	s.WorldcastMHF(&mhfpacket.MsgSysCastedBinary{
		BroadcastType:  BroadcastTypeServer,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}, nil, s)
}

func (s *Server) getRaviSemaphore() *Semaphore {
	for _, semaphore := range s.semaphore {
		if strings.HasPrefix(semaphore.name, "hs_l0") && strings.HasSuffix(semaphore.name, "3") {
			return semaphore
		}
	}
	return nil
}
