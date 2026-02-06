package channelserver

import (
	"sync"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// Object represents a placeable object in a stage (e.g., ballista, bombs, traps).
//
// Objects are spawned by players during quests and can be interacted with by
// other players in the same stage. Each object has an owner, position, and
// unique ID for client-server synchronization.
type Object struct {
	sync.RWMutex         // Protects object state during updates
	id           uint32  // Unique object ID (see NextObjectID for ID generation)
	ownerCharID  uint32  // Character ID of the player who placed this object
	x, y, z      float32 // 3D position coordinates
}

// stageBinaryKey is a composite key for identifying a specific piece of stage binary data.
//
// Stage binary data is custom game state that the stage host (quest leader) sets
// and the server echoes to other clients. It's used for quest state, monster HP,
// environmental conditions, etc. The data is keyed by two ID bytes.
type stageBinaryKey struct {
	id0 uint8 // First binary data identifier
	id1 uint8 // Second binary data identifier
}

// Stage represents a game room/area where players interact.
//
// Stages are the core spatial concept in Monster Hunter Frontier. They represent:
//   - Town areas (Mezeporta, Pallone, etc.) - persistent, always exist
//   - Quest instances - created dynamically when a player starts a quest
//   - Private rooms - password-protected player gathering areas
//
// Stage Lifecycle:
//  1. Created via NewStage() or MSG_SYS_CREATE_STAGE packet
//  2. Players enter via MSG_SYS_ENTER_STAGE or MSG_SYS_MOVE_STAGE
//  3. Stage host manages state via binary data packets
//  4. Destroyed via MSG_SYS_STAGE_DESTRUCT when empty or quest completes
//
// Client Participation:
// There are two types of client participation:
//   - Active clients (in clients map): Currently in the stage, receive broadcasts
//   - Reserved slots (in reservedClientSlots): Quest participants who haven't
//     entered yet (e.g., loading screen, preparing). They hold a slot but don't
//     receive stage broadcasts until they fully enter.
//
// Thread Safety:
// Stage embeds sync.RWMutex. Use RLock for reads (broadcasts, queries) and
// Lock for writes (entering, leaving, state changes).
type Stage struct {
	sync.RWMutex // Protects all stage state during concurrent access

	// Stage identity
	id string // Stage ID string (e.g., "sl1Ns200p0a0u0" for Mezeporta)

	// Objects in the stage (ballistas, bombs, traps, etc.)
	objects     map[uint32]*Object // Active objects keyed by object ID
	objectIndex uint8              // Auto-incrementing index for object ID generation

	// Active participants
	clients map[*Session]uint32 // Sessions currently in stage -> their character ID

	// Reserved slots for quest participants
	// Map of charID -> ready status. These players have reserved a slot but
	// haven't fully entered yet (e.g., still loading, in preparation screen)
	reservedClientSlots map[uint32]bool // Character ID -> ready flag

	// Stage binary data
	// Raw binary blobs set by the stage host (quest leader) that track quest state.
	// The server stores and echoes this data to clients verbatim. Used for:
	//   - Monster HP and status
	//   - Environmental state (time remaining, weather)
	//   - Quest objectives and progress
	rawBinaryData map[stageBinaryKey][]byte // Binary state keyed by (id0, id1)

	// Stage settings
	host       *Session // Stage host (quest leader, room creator)
	maxPlayers uint16   // Maximum players allowed (default 4)
	password   string   // Password for private stages (empty if public)
}

// NewStage creates and initializes a new Stage with the given ID.
//
// The stage is created with:
//   - Empty client and reserved slot maps
//   - Empty object map with objectIndex starting at 0
//   - Empty binary data map
//   - Default max players set to 4 (standard quest party size)
//   - No password (public stage)
//
// For persistent town stages, this is called during server initialization.
// For dynamic quest stages, this is called when a player creates a quest.
func NewStage(ID string) *Stage {
	s := &Stage{
		id:                  ID,
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]bool),
		objects:             make(map[uint32]*Object),
		objectIndex:         0,
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		maxPlayers:          4,
	}
	return s
}

// BroadcastMHF sends a packet to all players currently in the stage.
//
// This method is used for stage-local events like:
//   - Player chat messages within the stage
//   - Monster state updates
//   - Object placement/removal notifications
//   - Quest events visible only to stage participants
//
// The packet is built individually for each client to support version-specific
// formatting. Packets are sent non-blocking (dropped if queue full).
//
// Reserved clients (those who haven't fully entered) do NOT receive broadcasts.
//
// Parameters:
//   - pkt: The MHFPacket to broadcast to stage participants
//   - ignoredSession: Optional session to exclude (typically the sender)
//
// Thread Safety: This method holds the stage lock during iteration.
func (s *Stage) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	s.Lock()
	defer s.Unlock()
	for session := range s.clients {
		if session == ignoredSession || session.clientContext == nil {
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

func (s *Stage) isCharInQuestByID(charID uint32) bool {
	if _, exists := s.reservedClientSlots[charID]; exists {
		return exists
	}

	return false
}

func (s *Stage) isQuest() bool {
	return len(s.reservedClientSlots) > 0
}

// NextObjectID generates the next available object ID for this stage.
//
// Object IDs have special constraints due to client limitations:
//   - Index 0 does not update position correctly (avoided)
//   - Index 127 does not update position correctly (avoided)
//   - Indices > 127 do not replicate correctly across clients (avoided)
//
// The ID is generated by packing bytes into a uint32 in a specific format
// expected by the client. The objectIndex cycles from 1-126 to stay within
// valid bounds.
//
// Thread Safety: Caller must hold stage lock when calling this method.
func (s *Stage) NextObjectID() uint32 {
	s.objectIndex = s.objectIndex + 1
	// Objects beyond 127 do not duplicate correctly
	// Indexes 0 and 127 does not update position correctly
	if s.objectIndex == 127 {
		s.objectIndex = 1
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	bf.WriteUint8(s.objectIndex)
	bf.WriteUint16(0)
	obj := uint32(bf.Data()[3]) | uint32(bf.Data()[2])<<8 | uint32(bf.Data()[1])<<16 | uint32(bf.Data()[0])<<24
	return obj
}
