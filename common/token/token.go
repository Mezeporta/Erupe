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

// NewSafeRand creates a SafeRand seeded with the current time.
func NewSafeRand() *SafeRand {
	return &SafeRand{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Intn returns a non-negative pseudo-random int in [0,n). It is safe for
// concurrent use.
func (sr *SafeRand) Intn(n int) int {
	sr.mu.Lock()
	v := sr.rng.Intn(n)
	sr.mu.Unlock()
	return v
}

// Uint32 returns a pseudo-random uint32. It is safe for concurrent use.
func (sr *SafeRand) Uint32() uint32 {
	sr.mu.Lock()
	v := sr.rng.Uint32()
	sr.mu.Unlock()
	return v
}

// RNG is the global concurrency-safe random number generator used throughout
// the server for generating warehouse IDs, session tokens, and other values.
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
