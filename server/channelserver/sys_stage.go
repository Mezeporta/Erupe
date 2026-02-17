package channelserver

import (
	"sync"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// Object holds infomation about a specific object.
type Object struct {
	sync.RWMutex
	id          uint32
	ownerCharID uint32
	x, y, z     float32
}

// stageBinaryKey is a struct used as a map key for identifying a stage binary part.
type stageBinaryKey struct {
	id0 uint8
	id1 uint8
}

// Stage holds stage-specific information
type Stage struct {
	sync.RWMutex

	// Stage ID string
	id string

	// Objects
	objects     map[uint32]*Object
	objectIndex uint8

	// Map of session -> charID.
	// These are clients that are CURRENTLY in the stage
	clients map[*Session]uint32

	// Map of charID -> bool, key represents whether they are ready
	// These are clients that aren't in the stage, but have reserved a slot (for quests, etc).
	reservedClientSlots map[uint32]bool

	// These are raw binary blobs that the stage owner sets,
	// other clients expect the server to echo them back in the exact same format.
	rawBinaryData map[stageBinaryKey][]byte

	host       *Session
	maxPlayers uint16
	password   string
	locked     bool
}

// NewStage creates a new stage with intialized values.
func NewStage(ID string) *Stage {
	s := &Stage{
		id:                  ID,
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]bool),
		objects:             make(map[uint32]*Object),
		objectIndex:         0,
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		maxPlayers:          127,
	}
	return s
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions in the stage.
func (s *Stage) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	s.Lock()
	defer s.Unlock()
	for session := range s.clients {
		if session == ignoredSession {
			continue
		}

		// Make the header
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(pkt.Opcode()))

		// Build the packet onto the byteframe.
		_ = pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}
