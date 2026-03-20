// Package mhfpacket defines the struct representations and binary
// serialization for every MHF network packet (~400 message types). Each
// packet implements the [MHFPacket] interface (Parse, Build, Opcode).
//
// # Unk Fields
//
// Fields named Unk0, Unk1, … UnkN (or simply Unk) are protocol fields
// whose purpose has not yet been reverse-engineered. They are parsed and
// round-tripped faithfully but their semantic meaning is unknown. As
// fields are identified through protocol research or client
// decompilation, they should be renamed to descriptive names. The same
// convention applies to Unk fields in handler and repository code
// throughout the channelserver package.
package mhfpacket
