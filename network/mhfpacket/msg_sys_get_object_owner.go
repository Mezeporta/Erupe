package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysGetObjectOwner represents the MSG_SYS_GET_OBJECT_OWNER.
//
// Wire format recovered by decompiling the PC client's own dispatch handler
// (pkt_handler_MSG_SYS_GET_OBJECT_OWNER @ 0x11503AC0 in mhfo-hd.dll, see
// docs/network_protocol.md's PC Client Dispatch Table): it reads a 4-byte
// AckHandle, then a 4-byte ObjID immediately after (no gap), resolves the
// owner locally, and replies via sync_man_put_ack(AckHandle, owner).
type MsgSysGetObjectOwner struct {
	AckHandle uint32
	ObjID     uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetObjectOwner) Opcode() network.PacketID {
	return network.MSG_SYS_GET_OBJECT_OWNER
}

// Parse parses the packet from binary
func (m *MsgSysGetObjectOwner) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.ObjID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetObjectOwner) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
