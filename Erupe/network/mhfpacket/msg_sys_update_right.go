package mhfpacket

import (
	"errors"

	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
)

/*
00 58 // Opcode

00 00 00 00
00 00 00 4e

00 04 // Count
00 00 // Skipped(padding?)

00 01  00 00  00 00 00 00
00 02  00 00  5d fa 14 c0
00 03  00 00  5d fa 14 c0
00 06  00 00  5d e7 05 10

00 00 // Count of some buf up to 0x800 bytes following it.

00 10 // Trailer
*/

// ClientRight represents a right that the client has.
type ClientRight struct {
	ID        uint16
	Unk0      uint16
	Timestamp uint32
}

// MsgSysUpdateRight represents the MSG_SYS_UPDATE_RIGHT
type MsgSysUpdateRight struct {
	ClientRespAckHandle uint32 // If non-0, requests the client to send back a MSG_SYS_ACK packet with this value.
	Unk1                uint32
	Rights              []ClientRight
	UnkSize             uint16 // Count of some buf up to 0x800 bytes following it.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUpdateRight) Opcode() network.PacketID {
	return network.MSG_SYS_UPDATE_RIGHT
}

// Parse parses the packet from binary
func (m *MsgSysUpdateRight) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysUpdateRight) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.ClientRespAckHandle)
	bf.WriteUint32(m.Unk1)
	bf.WriteUint16(uint16(len(m.Rights)))
	bf.WriteUint16(0)
	for _, v := range m.Rights {
		bf.WriteUint16(v.ID)
		bf.WriteUint16(v.Unk0)
		bf.WriteUint32(v.Timestamp)
	}

	bf.WriteUint16(m.UnkSize) // String of upto 0x800 bytes, update client login token / password in the game's launcherstate struct.
	//bf.WriteBytes(m.UpdatedClientLoginToken)

	return nil
}
