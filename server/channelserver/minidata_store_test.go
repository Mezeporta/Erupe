package channelserver

import (
	"sync"
	"testing"
)

func TestMinidataStore_GetMiss(t *testing.T) {
	s := NewMinidataStore()
	_, ok := s.Get(1)
	if ok {
		t.Error("expected miss for unknown charID")
	}
}

func TestMinidataStore_SetGet(t *testing.T) {
	s := NewMinidataStore()
	data := []byte{0xAA, 0xBB}
	s.Set(42, data)

	got, ok := s.Get(42)
	if !ok {
		t.Fatal("expected hit")
	}
	if len(got) != 2 || got[0] != 0xAA {
		t.Errorf("got %v, want [0xAA 0xBB]", got)
	}
}

func TestMinidataStore_Overwrite(t *testing.T) {
	s := NewMinidataStore()
	s.Set(1, []byte{0x01})
	s.Set(1, []byte{0x02})

	got, _ := s.Get(1)
	if got[0] != 0x02 {
		t.Error("overwrite should replace previous value")
	}
}

func TestMinidataStore_ConcurrentAccess(t *testing.T) {
	s := NewMinidataStore()
	var wg sync.WaitGroup
	for i := uint32(0); i < 100; i++ {
		wg.Add(2)
		charID := i
		go func() {
			defer wg.Done()
			s.Set(charID, []byte{byte(charID)})
		}()
		go func() {
			defer wg.Done()
			s.Get(charID)
		}()
	}
	wg.Wait()
}
