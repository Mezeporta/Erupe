package mhfpacket

import (
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysAck represents the MSG_SYS_ACK
type MsgSysAck struct {
	AckHandle        uint32
	IsBufferResponse bool
	ErrorCode        uint8
	AckData          []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAck) Opcode() network.PacketID {
	return network.MSG_SYS_ACK
}

// Parse parses the packet from binary
func (m *MsgSysAck) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.IsBufferResponse = bf.ReadBool()
	m.ErrorCode = bf.ReadUint8()

	payloadSize := uint(bf.ReadUint16())
	// Extended data size field
	if payloadSize == 0xFFFF {
		payloadSize = uint(bf.ReadUint32())
	}

	if m.IsBufferResponse {
		m.AckData = bf.ReadBytes(payloadSize)
	} else {
		// endian-swapped 4 bytes, could be any type. Unknown purpose.
		// Probably a fixed type like (int32 or uint32), but unknown.
		m.AckData = bf.ReadBytes(4)
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysAck) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteBool(m.IsBufferResponse)
	bf.WriteUint8(m.ErrorCode)

	if m.IsBufferResponse {
		if len(m.AckData) < 0xFFFF {
			bf.WriteUint16(uint16(len(m.AckData)))
		} else {
			bf.WriteUint16(0xFFFF)
			bf.WriteUint32(uint32(len(m.AckData)))
		}
	} else {
		bf.WriteUint16(0x00)
	}

	if m.IsBufferResponse {
		bf.WriteBytes(m.AckData)
	} else if len(m.AckData) >= 4 {
		bf.WriteBytes(m.AckData[:4])
	} else {
		bf.WriteBytes([]byte{0x00, 0x00, 0x00, 0x00})
	}

	return nil
}
