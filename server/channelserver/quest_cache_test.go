package channelserver

import (
	"sync"
	"testing"
	"time"
)

func TestQuestCache_GetMiss(t *testing.T) {
	c := NewQuestCache(60)
	_, ok := c.Get(999, "jp")
	if ok {
		t.Error("expected cache miss for unknown quest ID")
	}
}

func TestQuestCache_PutGet(t *testing.T) {
	c := NewQuestCache(60)
	data := []byte{0xDE, 0xAD}
	c.Put(1, "jp", data)

	got, ok := c.Get(1, "jp")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 2 || got[0] != 0xDE || got[1] != 0xAD {
		t.Errorf("got %v, want [0xDE 0xAD]", got)
	}
}

func TestQuestCache_Expiry(t *testing.T) {
	c := NewQuestCache(0) // TTL=0 disables caching
	c.Put(1, "jp", []byte{0x01})

	_, ok := c.Get(1, "jp")
	if ok {
		t.Error("expected cache miss when TTL is 0")
	}
}

func TestQuestCache_ExpiryElapsed(t *testing.T) {
	c := &QuestCache{
		data:   make(map[questCacheKey][]byte),
		expiry: make(map[questCacheKey]time.Time),
		ttl:    50 * time.Millisecond,
	}
	c.Put(1, "jp", []byte{0x01})

	// Should hit immediately
	if _, ok := c.Get(1, "jp"); !ok {
		t.Fatal("expected cache hit before expiry")
	}

	time.Sleep(60 * time.Millisecond)

	// Should miss after expiry
	if _, ok := c.Get(1, "jp"); ok {
		t.Error("expected cache miss after expiry")
	}
}

// TestQuestCache_LangIsolation verifies that entries for different languages
// of the same quest ID are stored independently (phase B of #188).
func TestQuestCache_LangIsolation(t *testing.T) {
	c := NewQuestCache(60)
	c.Put(1, "jp", []byte{0x01})
	c.Put(1, "en", []byte{0x02})

	if got, ok := c.Get(1, "jp"); !ok || got[0] != 0x01 {
		t.Errorf("jp variant: got %v ok=%v, want [0x01] true", got, ok)
	}
	if got, ok := c.Get(1, "en"); !ok || got[0] != 0x02 {
		t.Errorf("en variant: got %v ok=%v, want [0x02] true", got, ok)
	}
	// Unset language variant should miss.
	if _, ok := c.Get(1, "fr"); ok {
		t.Error("fr variant should miss when not populated")
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
			c.Put(id, "jp", []byte{byte(id)})
		}()
		go func() {
			defer wg.Done()
			c.Get(id, "jp")
		}()
	}
	wg.Wait()
}
