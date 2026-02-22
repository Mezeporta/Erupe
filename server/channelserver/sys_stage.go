package channelserver

import (
	"sync"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// StageMap is a concurrent-safe map of stage ID → *Stage backed by sync.Map.
// It replaces the former stagesLock + map[string]*Stage pattern, eliminating
// read contention entirely (reads are lock-free) and allowing concurrent
// writes to disjoint keys.
type StageMap struct {
	m sync.Map
}

// Get returns the stage for the given ID, or (nil, false) if not found.
func (sm *StageMap) Get(id string) (*Stage, bool) {
	v, ok := sm.m.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Stage), true
}

// GetOrCreate atomically returns the existing stage for id, or creates and
// stores a new one. The second return value is true when a new stage was created.
func (sm *StageMap) GetOrCreate(id string) (*Stage, bool) {
	newStage := NewStage(id)
	v, loaded := sm.m.LoadOrStore(id, newStage)
	return v.(*Stage), !loaded // created == !loaded
}

// StoreIfAbsent stores the stage only if the key does not already exist.
// Returns true if the store succeeded (key was absent).
func (sm *StageMap) StoreIfAbsent(id string, stage *Stage) bool {
	_, loaded := sm.m.LoadOrStore(id, stage)
	return !loaded
}

// Store unconditionally sets the stage for the given ID.
func (sm *StageMap) Store(id string, stage *Stage) {
	sm.m.Store(id, stage)
}

// Delete removes the stage with the given ID.
func (sm *StageMap) Delete(id string) {
	sm.m.Delete(id)
}

// Range iterates over all stages. The callback receives each (id, stage) pair
// and should return true to continue iteration or false to stop.
// It is safe to call Delete during iteration.
func (sm *StageMap) Range(fn func(id string, stage *Stage) bool) {
	sm.m.Range(func(key, value any) bool {
		return fn(key.(string), value.(*Stage))
	})
}

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
