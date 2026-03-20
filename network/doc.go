// Package network defines the encrypted TCP transport layer for MHF client
// connections. It provides Blowfish-based packet encryption/decryption via
// [CryptConn], packet header parsing, and the [PacketID] enumeration of all
// ~400 message types in the MHF protocol.
package network
