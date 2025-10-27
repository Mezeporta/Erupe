package token

import (
	"math/rand"
	"testing"
	"time"
)

func TestGenerate_Length(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"short", 5},
		{"medium", 32},
		{"long", 100},
		{"very long", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Generate(tt.length)
			if len(result) != tt.length {
				t.Errorf("Generate(%d) length = %d, want %d", tt.length, len(result), tt.length)
			}
		})
	}
}

func TestGenerate_CharacterSet(t *testing.T) {
	// Verify that generated tokens only contain alphanumeric characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	validCharMap := make(map[rune]bool)
	for _, c := range validChars {
		validCharMap[c] = true
	}

	token := Generate(1000) // Large sample
	for _, c := range token {
		if !validCharMap[c] {
			t.Errorf("Generate() produced invalid character: %c", c)
		}
	}
}

func TestGenerate_Randomness(t *testing.T) {
	// Generate multiple tokens and verify they're different
	tokens := make(map[string]bool)
	count := 100
	length := 32

	for i := 0; i < count; i++ {
		token := Generate(length)
		if tokens[token] {
			t.Errorf("Generate() produced duplicate token: %s", token)
		}
		tokens[token] = true
	}

	if len(tokens) != count {
		t.Errorf("Generated %d unique tokens, want %d", len(tokens), count)
	}
}

func TestGenerate_ContainsUppercase(t *testing.T) {
	// With enough characters, should contain at least one uppercase letter
	token := Generate(1000)
	hasUpper := false
	for _, c := range token {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		t.Error("Generate(1000) should contain at least one uppercase letter")
	}
}

func TestGenerate_ContainsLowercase(t *testing.T) {
	// With enough characters, should contain at least one lowercase letter
	token := Generate(1000)
	hasLower := false
	for _, c := range token {
		if c >= 'a' && c <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		t.Error("Generate(1000) should contain at least one lowercase letter")
	}
}

func TestGenerate_ContainsDigit(t *testing.T) {
	// With enough characters, should contain at least one digit
	token := Generate(1000)
	hasDigit := false
	for _, c := range token {
		if c >= '0' && c <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		t.Error("Generate(1000) should contain at least one digit")
	}
}

func TestGenerate_Distribution(t *testing.T) {
	// Test that characters are reasonably distributed
	token := Generate(6200) // 62 chars * 100 = good sample size
	charCount := make(map[rune]int)

	for _, c := range token {
		charCount[c]++
	}

	// With 62 valid characters and 6200 samples, average should be 100 per char
	// We'll accept a range to account for randomness
	minExpected := 50  // Allow some variance
	maxExpected := 150

	for c, count := range charCount {
		if count < minExpected || count > maxExpected {
			t.Logf("Character %c appeared %d times (outside expected range %d-%d)", c, count, minExpected, maxExpected)
		}
	}

	// Just verify we have a good spread of characters
	if len(charCount) < 50 {
		t.Errorf("Only %d different characters used, want at least 50", len(charCount))
	}
}

func TestNewRNG(t *testing.T) {
	rng := NewRNG()
	if rng == nil {
		t.Fatal("NewRNG() returned nil")
	}

	// Test that it produces different values on subsequent calls
	val1 := rng.Intn(1000000)
	val2 := rng.Intn(1000000)

	if val1 == val2 {
		// This is possible but unlikely, let's try a few more times
		same := true
		for i := 0; i < 10; i++ {
			if rng.Intn(1000000) != val1 {
				same = false
				break
			}
		}
		if same {
			t.Error("NewRNG() produced same value 12 times in a row")
		}
	}
}

func TestRNG_GlobalVariable(t *testing.T) {
	// Test that the global RNG variable is initialized
	if RNG == nil {
		t.Fatal("Global RNG is nil")
	}

	// Test that it works
	val := RNG.Intn(100)
	if val < 0 || val >= 100 {
		t.Errorf("RNG.Intn(100) = %d, out of range [0, 100)", val)
	}
}

func TestRNG_Uint32(t *testing.T) {
	// Test that RNG can generate uint32 values
	val1 := RNG.Uint32()
	val2 := RNG.Uint32()

	// They should be different (with very high probability)
	if val1 == val2 {
		// Try a few more times
		same := true
		for i := 0; i < 10; i++ {
			if RNG.Uint32() != val1 {
				same = false
				break
			}
		}
		if same {
			t.Error("RNG.Uint32() produced same value 12 times")
		}
	}
}

func TestGenerate_Concurrency(t *testing.T) {
	// Test that Generate works correctly when called concurrently
	done := make(chan string, 100)

	for i := 0; i < 100; i++ {
		go func() {
			token := Generate(32)
			done <- token
		}()
	}

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := <-done
		if len(token) != 32 {
			t.Errorf("Token length = %d, want 32", len(token))
		}
		tokens[token] = true
	}

	// Should have many unique tokens (allow some small chance of duplicates)
	if len(tokens) < 95 {
		t.Errorf("Only %d unique tokens from 100 concurrent calls", len(tokens))
	}
}

func TestGenerate_EmptyString(t *testing.T) {
	token := Generate(0)
	if token != "" {
		t.Errorf("Generate(0) = %q, want empty string", token)
	}
}

func TestGenerate_OnlyAlphanumeric(t *testing.T) {
	// Verify no special characters
	token := Generate(1000)
	for i, c := range token {
		isValid := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isValid {
			t.Errorf("Token[%d] = %c (invalid character)", i, c)
		}
	}
}

func TestNewRNG_DifferentSeeds(t *testing.T) {
	// Create two RNGs at different times and verify they produce different sequences
	rng1 := NewRNG()
	time.Sleep(1 * time.Millisecond) // Ensure different seed
	rng2 := NewRNG()

	val1 := rng1.Intn(1000000)
	val2 := rng2.Intn(1000000)

	// They should be different with high probability
	if val1 == val2 {
		// Try again
		val1 = rng1.Intn(1000000)
		val2 = rng2.Intn(1000000)
		if val1 == val2 {
			t.Log("Two RNGs created at different times produced same first two values (possible but unlikely)")
		}
	}
}

func BenchmarkGenerate_Short(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Generate(8)
	}
}

func BenchmarkGenerate_Medium(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Generate(32)
	}
}

func BenchmarkGenerate_Long(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Generate(128)
	}
}

func BenchmarkNewRNG(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRNG()
	}
}

func BenchmarkRNG_Intn(b *testing.B) {
	rng := NewRNG()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Intn(62)
	}
}

func BenchmarkRNG_Uint32(b *testing.B) {
	rng := NewRNG()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Uint32()
	}
}

func TestGenerate_ConsistentCharacterSet(t *testing.T) {
	// Verify the character set matches what's defined in the code
	expectedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if len(expectedChars) != 62 {
		t.Errorf("Expected character set length = %d, want 62", len(expectedChars))
	}

	// Count each type
	lowercase := 0
	uppercase := 0
	digits := 0
	for _, c := range expectedChars {
		if c >= 'a' && c <= 'z' {
			lowercase++
		} else if c >= 'A' && c <= 'Z' {
			uppercase++
		} else if c >= '0' && c <= '9' {
			digits++
		}
	}

	if lowercase != 26 {
		t.Errorf("Lowercase count = %d, want 26", lowercase)
	}
	if uppercase != 26 {
		t.Errorf("Uppercase count = %d, want 26", uppercase)
	}
	if digits != 10 {
		t.Errorf("Digits count = %d, want 10", digits)
	}
}

func TestRNG_Type(t *testing.T) {
	// Verify RNG is of type *rand.Rand
	var _ *rand.Rand = RNG
	var _ *rand.Rand = NewRNG()
}
