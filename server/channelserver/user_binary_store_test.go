package channelserver

import (
	"sync"
	"testing"
)

func TestUserBinaryStore_GetMiss(t *testing.T) {
	s := NewUserBinaryStore()
	_, ok := s.Get(1, 1)
	if ok {
		t.Error("expected miss for unknown key")
	}
}

func TestUserBinaryStore_SetGet(t *testing.T) {
	s := NewUserBinaryStore()
	data := []byte{0x01, 0x02, 0x03}
	s.Set(100, 3, data)

	got, ok := s.Get(100, 3)
	if !ok {
		t.Fatal("expected hit")
	}
	if len(got) != 3 || got[0] != 0x01 {
		t.Errorf("got %v, want [1 2 3]", got)
	}
}

func TestUserBinaryStore_DifferentIndexes(t *testing.T) {
	s := NewUserBinaryStore()
	s.Set(1, 1, []byte{0xAA})
	s.Set(1, 2, []byte{0xBB})

	got1, _ := s.Get(1, 1)
	got2, _ := s.Get(1, 2)
	if got1[0] != 0xAA || got2[0] != 0xBB {
		t.Error("different indexes should store independent data")
	}
}

func TestUserBinaryStore_Delete(t *testing.T) {
	s := NewUserBinaryStore()
	s.Set(1, 3, []byte{0x01})
	s.Delete(1, 3)

	_, ok := s.Get(1, 3)
	if ok {
		t.Error("expected miss after delete")
	}
}

func TestUserBinaryStore_DeleteNonExistent(t *testing.T) {
	s := NewUserBinaryStore()
	s.Delete(999, 1) // should not panic
}

func TestUserBinaryStore_GetCopy(t *testing.T) {
	s := NewUserBinaryStore()
	s.Set(1, 3, []byte{0x01, 0x02})

	cp := s.GetCopy(1, 3)
	if cp[0] != 0x01 || cp[1] != 0x02 {
		t.Fatal("copy data mismatch")
	}

	// Mutating the copy must not affect the store
	cp[0] = 0xFF
	orig, _ := s.Get(1, 3)
	if orig[0] == 0xFF {
		t.Error("GetCopy returned a reference, not a copy")
	}
}

func TestUserBinaryStore_GetCopyMiss(t *testing.T) {
	s := NewUserBinaryStore()
	cp := s.GetCopy(999, 1)
	if cp != nil {
		t.Error("expected nil for missing key")
	}
}

func TestUserBinaryStore_ConcurrentAccess(t *testing.T) {
	s := NewUserBinaryStore()
	var wg sync.WaitGroup
	for i := uint32(0); i < 100; i++ {
		wg.Add(3)
		charID := i
		go func() {
			defer wg.Done()
			s.Set(charID, 1, []byte{byte(charID)})
		}()
		go func() {
			defer wg.Done()
			s.Get(charID, 1)
		}()
		go func() {
			defer wg.Done()
			s.GetCopy(charID, 1)
		}()
	}
	wg.Wait()
}
