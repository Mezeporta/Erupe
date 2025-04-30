package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"

	"sync"
)

// Semaphore holds Semaphore-specific information
type Semaphore struct {
	sync.RWMutex

	// Semaphore ID string
	name string

	id uint32

	// Map of session -> charID.
	// These are clients that are registered to the Semaphore
	clients map[*Session]uint32

	// Max Players for Semaphore
	maxPlayers uint16

	host *Session
}

// NewSemaphore creates a new Semaphore with intialized values
func NewSemaphore(s *Session, ID string, MaxPlayers uint16) *Semaphore {
	sema := &Semaphore{
		name:       ID,
		id:         s.GetSemaphoreID(),
		clients:    make(map[*Session]uint32),
		maxPlayers: MaxPlayers,
		host:       s,
	}
	return sema
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions in the Semaphore
func (s *Semaphore) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	// Broadcast the data.
	for session := range s.clients {
		if session == ignoredSession {
			continue
		}

		// Make the header
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(pkt.Opcode()))

		// Build the packet onto the byteframe.
		pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}
