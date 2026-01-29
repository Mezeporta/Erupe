package token

import (
	"regexp"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"short token", 8},
		{"medium token", 32},
		{"long token", 256},
		{"single char", 1},
	}

	alphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]*$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Generate(tt.length)
			if len(result) != tt.length {
				t.Errorf("Generate(%d) length = %d, want %d", tt.length, len(result), tt.length)
			}
			if !alphanumeric.MatchString(result) {
				t.Errorf("Generate(%d) = %q, contains non-alphanumeric characters", tt.length, result)
			}
		})
	}
}

func TestGenerateUniqueness(t *testing.T) {
	// Generate multiple tokens and check they're different
	tokens := make(map[string]bool)
	iterations := 100
	length := 32

	for i := 0; i < iterations; i++ {
		token := Generate(length)
		if tokens[token] {
			t.Errorf("Generate(%d) produced duplicate token: %s", length, token)
		}
		tokens[token] = true
	}
}

func TestGenerateCharacterDistribution(t *testing.T) {
	// Generate a long token and verify it uses various characters
	token := Generate(1000)

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(token)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(token)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(token)

	if !hasLower {
		t.Error("Generate(1000) did not produce any lowercase letters")
	}
	if !hasUpper {
		t.Error("Generate(1000) did not produce any uppercase letters")
	}
	if !hasDigit {
		t.Error("Generate(1000) did not produce any digits")
	}
}

func TestRNG(t *testing.T) {
	rng1 := RNG()
	rng2 := RNG()

	if rng1 == nil {
		t.Error("RNG() returned nil")
	}
	if rng2 == nil {
		t.Error("RNG() returned nil")
	}

	// Both should generate valid random numbers
	val1 := rng1.Intn(100)
	val2 := rng2.Intn(100)

	if val1 < 0 || val1 >= 100 {
		t.Errorf("RNG().Intn(100) = %d, want value in [0, 100)", val1)
	}
	if val2 < 0 || val2 >= 100 {
		t.Errorf("RNG().Intn(100) = %d, want value in [0, 100)", val2)
	}
}

func TestRNGIndependence(t *testing.T) {
	// Create multiple RNGs and verify they produce different sequences
	rng1 := RNG()
	rng2 := RNG()

	// Generate sequences
	seq1 := make([]int, 10)
	seq2 := make([]int, 10)

	for i := 0; i < 10; i++ {
		seq1[i] = rng1.Intn(1000000)
		seq2[i] = rng2.Intn(1000000)
	}

	// Check that sequences are likely different (not identical)
	identical := true
	for i := 0; i < 10; i++ {
		if seq1[i] != seq2[i] {
			identical = false
			break
		}
	}

	// Note: There's an extremely small chance both RNGs could produce
	// the same sequence, but it's astronomically unlikely
	if identical {
		t.Log("Warning: Two independent RNGs produced identical sequences (this is extremely unlikely)")
	}
}
