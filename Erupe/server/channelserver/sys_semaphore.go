package channelserver

import (
	"sync"
)

// Stage holds stage-specific information
type Semaphore struct {
	sync.RWMutex

	// Stage ID string
	id_semaphore string

	// Map of charID -> interface{}, only the key is used, value is always nil.
	reservedClientSlots map[uint32]interface{}

	// Max Players for Semaphore
	maxPlayers uint16
}

// NewStage creates a new stage with intialized values.
func NewSemaphore(ID string, MaxPlayers uint16) *Semaphore {
	s := &Semaphore{
		id_semaphore:        ID,
		reservedClientSlots: make(map[uint32]interface{}),
		maxPlayers:          MaxPlayers,
	}
	return s
}
