// Package crypto implements the symmetric substitution-cipher used by Monster
// Hunter Frontier to encrypt and decrypt TCP packet bodies. The algorithm uses
// a 256-byte S-box with a rolling derived key and produces three integrity
// checksums alongside the ciphertext.
package crypto
