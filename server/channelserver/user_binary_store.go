package channelserver

import "sync"

// userBinaryPartID is the composite key for a user binary part.
type userBinaryPartID struct {
	charID uint32
	index  uint8
}

// UserBinaryStore is a thread-safe store for per-character binary data parts.
type UserBinaryStore struct {
	mu   sync.RWMutex
	data map[userBinaryPartID][]byte
}

// NewUserBinaryStore creates an empty UserBinaryStore.
func NewUserBinaryStore() *UserBinaryStore {
	return &UserBinaryStore{data: make(map[userBinaryPartID][]byte)}
}

// Get returns the binary data for the given character and index.
func (s *UserBinaryStore) Get(charID uint32, index uint8) ([]byte, bool) {
	s.mu.RLock()
	data, ok := s.data[userBinaryPartID{charID: charID, index: index}]
	s.mu.RUnlock()
	return data, ok
}

// GetCopy returns a copy of the binary data, safe for use after the lock is released.
func (s *UserBinaryStore) GetCopy(charID uint32, index uint8) []byte {
	s.mu.RLock()
	src := s.data[userBinaryPartID{charID: charID, index: index}]
	if len(src) == 0 {
		s.mu.RUnlock()
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	s.mu.RUnlock()
	return dst
}

// Set stores binary data for the given character and index.
func (s *UserBinaryStore) Set(charID uint32, index uint8, data []byte) {
	s.mu.Lock()
	s.data[userBinaryPartID{charID: charID, index: index}] = data
	s.mu.Unlock()
}

// Delete removes binary data for the given character and index.
func (s *UserBinaryStore) Delete(charID uint32, index uint8) {
	s.mu.Lock()
	delete(s.data, userBinaryPartID{charID: charID, index: index})
	s.mu.Unlock()
}
