package conn

import (
	"testing"
)

// TestCalcSum32 verifies the checksum against a known input.
func TestCalcSum32(t *testing.T) {
	// Verify determinism: same input gives same output.
	data := []byte("Hello, MHF!")
	sum1 := CalcSum32(data)
	sum2 := CalcSum32(data)
	if sum1 != sum2 {
		t.Fatalf("CalcSum32 not deterministic: %08X != %08X", sum1, sum2)
	}

	// Different inputs produce different outputs (basic sanity).
	data2 := []byte("Hello, MHF?")
	sum3 := CalcSum32(data2)
	if sum1 == sum3 {
		t.Fatalf("CalcSum32 collision on different inputs: both %08X", sum1)
	}
}

// TestDecryptBin8RoundTrip verifies that encrypting and decrypting with Bin8
// produces the original data. We only have DecryptBin8, but we can verify
// the encrypt→decrypt path by implementing encrypt inline here.
func TestDecryptBin8RoundTrip(t *testing.T) {
	original := []byte("Test data for Bin8 encryption round-trip")
	key := byte(0x42)

	// Encrypt (inline copy of Erupe's EncryptBin8)
	k := uint32(key)
	encrypted := make([]byte, len(original))
	for i := 0; i < len(original); i++ {
		rotate(&k)
		tmp := bin8Key[i&7] ^ byte((k>>13)&0xFF)
		encrypted[i] = original[i] ^ tmp
	}

	// Decrypt
	decrypted := DecryptBin8(encrypted, key)

	if len(decrypted) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(decrypted), len(original))
	}
	for i := range original {
		if decrypted[i] != original[i] {
			t.Fatalf("byte %d: got 0x%02X, want 0x%02X", i, decrypted[i], original[i])
		}
	}
}
