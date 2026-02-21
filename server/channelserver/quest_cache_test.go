package channelserver

import (
	"sync"
	"testing"
	"time"
)

func TestQuestCache_GetMiss(t *testing.T) {
	c := NewQuestCache(60)
	_, ok := c.Get(999)
	if ok {
		t.Error("expected cache miss for unknown quest ID")
	}
}

func TestQuestCache_PutGet(t *testing.T) {
	c := NewQuestCache(60)
	data := []byte{0xDE, 0xAD}
	c.Put(1, data)

	got, ok := c.Get(1)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 2 || got[0] != 0xDE || got[1] != 0xAD {
		t.Errorf("got %v, want [0xDE 0xAD]", got)
	}
}

func TestQuestCache_Expiry(t *testing.T) {
	c := NewQuestCache(0) // TTL=0 disables caching
	c.Put(1, []byte{0x01})

	_, ok := c.Get(1)
	if ok {
		t.Error("expected cache miss when TTL is 0")
	}
}

func TestQuestCache_ExpiryElapsed(t *testing.T) {
	c := &QuestCache{
		data:   make(map[int][]byte),
		expiry: make(map[int]time.Time),
		ttl:    50 * time.Millisecond,
	}
	c.Put(1, []byte{0x01})

	// Should hit immediately
	if _, ok := c.Get(1); !ok {
		t.Fatal("expected cache hit before expiry")
	}

	time.Sleep(60 * time.Millisecond)

	// Should miss after expiry
	if _, ok := c.Get(1); ok {
		t.Error("expected cache miss after expiry")
	}
}

func TestQuestCache_ConcurrentAccess(t *testing.T) {
	c := NewQuestCache(60)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		id := i
		go func() {
			defer wg.Done()
			c.Put(id, []byte{byte(id)})
		}()
		go func() {
			defer wg.Done()
			c.Get(id)
		}()
	}
	wg.Wait()
}
