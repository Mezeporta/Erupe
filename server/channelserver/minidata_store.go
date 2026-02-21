package channelserver

import "sync"

// MinidataStore is a thread-safe store for per-character enhanced minidata.
type MinidataStore struct {
	mu   sync.RWMutex
	data map[uint32][]byte
}

// NewMinidataStore creates an empty MinidataStore.
func NewMinidataStore() *MinidataStore {
	return &MinidataStore{data: make(map[uint32][]byte)}
}

// Get returns the minidata for the given character ID.
func (s *MinidataStore) Get(charID uint32) ([]byte, bool) {
	s.mu.RLock()
	data, ok := s.data[charID]
	s.mu.RUnlock()
	return data, ok
}

// Set stores minidata for the given character ID.
func (s *MinidataStore) Set(charID uint32, data []byte) {
	s.mu.Lock()
	s.data[charID] = data
	s.mu.Unlock()
}
