// Package nullcomp implements null-byte run-length compression used by the MHF
// client for save data. The format uses a "cmp 20110113" header and encodes
// runs of zero bytes as a (0x00, count) pair.
package nullcomp
