package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"

	"sync"
)

// Semaphore is a multiplayer coordination mechanism for quests and events.
//
// Despite the name, Semaphore is NOT an OS synchronization primitive (like sync.Semaphore).
// Instead, it's a game-specific resource lock that coordinates multiplayer activities where:
//   - Players must acquire a semaphore before participating
//   - A limited number of participants are allowed (maxPlayers)
//   - The semaphore tracks both active and reserved participants
//
// Use Cases:
//   - Quest coordination: Ensures quest party size limits are enforced
//   - Event coordination: Raviente, VS Tournament, Diva Defense
//   - Global resources: Prevents multiple groups from starting conflicting events
//
// Semaphore vs Stage:
//   - Stages are spatial (game rooms, areas). Players in a stage can see each other.
//   - Semaphores are logical (coordination locks). Players in a semaphore are
//     participating in the same activity but may be in different stages.
//
// Example: Raviente Event
//   - Players acquire the Raviente semaphore to register for the event
//   - Multiple quest stages exist (preparation, phase 1, phase 2, carving)
//   - All participants share the same semaphore across different stages
//   - The semaphore enforces the 32-player limit across all stages
//
// Thread Safety:
// Semaphore embeds sync.RWMutex. Use RLock for reads and Lock for writes.
type Semaphore struct {
	sync.RWMutex // Protects semaphore state during concurrent access

	// Semaphore identity
	id_semaphore string // Semaphore ID string (identifies the resource/activity)
	id           uint32 // Numeric ID for client communication (auto-generated, starts at 7)

	// Active participants
	clients map[*Session]uint32 // Sessions actively using this semaphore -> character ID

	// Reserved slots
	// Players who have acquired the semaphore but may not be actively in the stage yet.
	// The value is always nil; only the key (charID) matters. This is a set implementation.
	reservedClientSlots map[uint32]interface{} // Character ID -> nil (set of reserved IDs)

	// Capacity
	maxPlayers uint16 // Maximum concurrent participants (e.g., 4 for quests, 32 for Raviente)
}

// NewSemaphore creates and initializes a new Semaphore for coordinating an activity.
//
// The semaphore is assigned an auto-incrementing ID from the server's semaphoreIndex.
// IDs 0-6 are reserved, so the first semaphore gets ID 7.
//
// Parameters:
//   - s: The server (used to generate unique semaphore ID)
//   - ID: Semaphore ID string (identifies the activity/resource)
//   - MaxPlayers: Maximum participants allowed
//
// Returns a new Semaphore ready for client acquisition.
func NewSemaphore(s *Server, ID string, MaxPlayers uint16) *Semaphore {
	sema := &Semaphore{
		id_semaphore:        ID,
		id:                  s.NextSemaphoreID(),
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]interface{}),
		maxPlayers:          MaxPlayers,
	}
	return sema
}

func (s *Semaphore) BroadcastRavi(pkt mhfpacket.MHFPacket) {
	// Broadcast the data.
	for session := range s.clients {

		// Make the header
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(pkt.Opcode()))

		// Build the packet onto the byteframe.
		pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}

// BroadcastMHF sends a packet to all active participants in the semaphore.
//
// This is used for event-wide announcements that all participants need to see,
// regardless of which stage they're currently in. Examples:
//   - Raviente phase changes
//   - Tournament updates
//   - Event completion notifications
//
// Only active clients (in the clients map) receive broadcasts. Reserved clients
// who haven't fully joined yet are excluded.
//
// Parameters:
//   - pkt: The MHFPacket to broadcast to all participants
//   - ignoredSession: Optional session to exclude from broadcast
//
// Thread Safety: Caller should hold semaphore lock when iterating clients.
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
