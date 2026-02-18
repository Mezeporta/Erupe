package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysUpdateRight represents the MSG_SYS_UPDATE_RIGHT
type MsgSysUpdateRight struct {
	ClientRespAckHandle uint32 // If non-0, requests the client to send back a MSG_SYS_ACK packet with this value.
	Bitfield            uint32
	Rights              []mhfcourse.Course
	TokenLength         uint16 // Length of the login token/password buffer (up to 0x800 bytes).
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
	bf.WriteUint32(m.Bitfield)
	bf.WriteUint16(uint16(len(m.Rights)))
	bf.WriteUint16(0)
	for _, v := range m.Rights {
		bf.WriteUint16(v.ID)
		bf.WriteUint16(0)
		if v.Expiry.IsZero() {
			bf.WriteUint32(0)
		} else {
			bf.WriteUint32(uint32(v.Expiry.Unix()))
		}
	}
	ps.Uint16(bf, "", false) // update client login token / password in the game's launcherstate struct
	return nil
}
