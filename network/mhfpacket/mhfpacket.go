// Package mhfpacket provides Monster Hunter Frontier packet definitions and interfaces.
//
// This package contains:
//   - MHFPacket interface: The common interface all packets implement
//   - 400+ packet type definitions in msg_*.go files
//   - Packet parsing (client -> server) and building (server -> client) logic
//   - Opcode-to-packet-type mapping via FromOpcode()
//
// Packet Structure:
//
// MHF packets follow this wire format:
//
//	[2 bytes: Opcode][N bytes: Packet-specific data][2 bytes: Footer 0x00 0x10]
//
// Each packet type defines its own structure matching the binary format expected
// by the Monster Hunter Frontier client.
//
// Implementing a New Packet:
//
//  1. Create msg_mhf_your_packet.go with packet struct
//  2. Implement Parse() to read data from ByteFrame
//  3. Implement Build() to write data to ByteFrame
//  4. Implement Opcode() to return the packet's ID
//  5. Register in opcodeToPacketMap in opcode_mapping.go
//  6. Add handler in server/channelserver/handlers_table.go
//
// Example:
//
//	type MsgMhfYourPacket struct {
//	    AckHandle uint32 // Common field for request/response matching
//	    SomeField uint16
//	}
//
//	func (m *MsgMhfYourPacket) Opcode() network.PacketID {
//	    return network.MSG_MHF_YOUR_PACKET
//	}
//
//	func (m *MsgMhfYourPacket) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
//	    m.AckHandle = bf.ReadUint32()
//	    m.SomeField = bf.ReadUint16()
//	    return nil
//	}
//
//	func (m *MsgMhfYourPacket) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
//	    bf.WriteUint32(m.AckHandle)
//	    bf.WriteUint16(m.SomeField)
//	    return nil
//	}
package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// Parser is the interface for deserializing packets from wire format.
//
// The Parse method reads packet data from a ByteFrame (binary stream) and
// populates the packet struct's fields. It's called when a packet arrives
// from the client.
//
// Parameters:
//   - bf: ByteFrame positioned after the opcode (contains packet payload)
//   - ctx: Client context (version info, capabilities) for version-specific parsing
//
// Returns an error if the packet data is malformed or incomplete.
type Parser interface {
	Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error
}

// Builder is the interface for serializing packets to wire format.
//
// The Build method writes the packet struct's fields to a ByteFrame (binary stream)
// in the format expected by the client. It's called when sending a packet to the client.
//
// Parameters:
//   - bf: ByteFrame to write packet data to (opcode already written by caller)
//   - ctx: Client context (version info, capabilities) for version-specific building
//
// Returns an error if serialization fails.
type Builder interface {
	Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error
}

// Opcoder is the interface for identifying a packet's opcode.
//
// The Opcode method returns the unique packet identifier used for routing
// packets to their handlers and for packet logging/debugging.
type Opcoder interface {
	Opcode() network.PacketID
}

// MHFPacket is the unified interface that all Monster Hunter Frontier packets implement.
//
// Every packet type must be able to:
//   - Parse itself from binary data (Parser)
//   - Build itself into binary data (Builder)
//   - Identify its opcode (Opcoder)
//
// This interface allows the packet handling system to work generically across
// all packet types while maintaining type safety through type assertions in handlers.
type MHFPacket interface {
	Parser
	Builder
	Opcoder
}
