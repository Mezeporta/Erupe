package token

import (
	"math/rand"
	"sync"
	"time"
)

// SafeRand is a concurrency-safe wrapper around *rand.Rand.
type SafeRand struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func NewSafeRand() *SafeRand {
	return &SafeRand{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (sr *SafeRand) Intn(n int) int {
	sr.mu.Lock()
	v := sr.rng.Intn(n)
	sr.mu.Unlock()
	return v
}

func (sr *SafeRand) Uint32() uint32 {
	sr.mu.Lock()
	v := sr.rng.Uint32()
	sr.mu.Unlock()
	return v
}

var RNG = NewSafeRand()

// Generate returns an alphanumeric token of specified length
func Generate(length int) string {
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = chars[RNG.Intn(len(chars))]
	}
	return string(b)
}
