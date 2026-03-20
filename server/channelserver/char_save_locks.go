package channelserver

import "sync"

// CharacterLocks provides per-character mutexes to serialize save operations.
// This prevents concurrent saves for the same character from racing, which
// could defeat corruption detection (e.g. house tier snapshot vs. write).
//
// The underlying sync.Map grows lazily — entries are created on first access
// and never removed (character IDs are bounded and reused across sessions).
type CharacterLocks struct {
	m sync.Map // map[uint32]*sync.Mutex
}

// Lock acquires the mutex for the given character and returns an unlock function.
// Usage:
//
//	unlock := s.server.charSaveLocks.Lock(charID)
//	defer unlock()
func (cl *CharacterLocks) Lock(charID uint32) func() {
	val, _ := cl.m.LoadOrStore(charID, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}
