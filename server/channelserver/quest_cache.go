package channelserver

import (
	"sync"
	"time"
)

// QuestCache is a thread-safe, expiring cache for parsed quest file data.
type QuestCache struct {
	mu     sync.RWMutex
	data   map[int][]byte
	expiry map[int]time.Time
	ttl    time.Duration
}

// NewQuestCache creates a QuestCache with the given TTL in seconds.
// A TTL of 0 disables caching (Get always misses).
func NewQuestCache(ttlSeconds int) *QuestCache {
	return &QuestCache{
		data:   make(map[int][]byte),
		expiry: make(map[int]time.Time),
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

// Get returns cached quest data if it exists and has not expired.
func (c *QuestCache) Get(questID int) ([]byte, bool) {
	if c.ttl <= 0 {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	b, ok := c.data[questID]
	if !ok {
		return nil, false
	}
	if time.Now().After(c.expiry[questID]) {
		return nil, false
	}
	return b, true
}

// Put stores quest data in the cache with the configured TTL.
func (c *QuestCache) Put(questID int, b []byte) {
	c.mu.Lock()
	c.data[questID] = b
	c.expiry[questID] = time.Now().Add(c.ttl)
	c.mu.Unlock()
}
