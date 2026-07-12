package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysRotateObject represents the MSG_SYS_ROTATE_OBJECT.
//
// Wire format recovered by decompiling the PC client's own dispatch handler
// (pkt_handler_MSG_SYS_ROTATE_OBJECT @ 0x11503040 in mhfo-hd.dll, see
// docs/network_protocol.md's PC Client Dispatch Table). The handler body
// only dereferences a single 4-byte value at offset 4 (passed straight
// through to an unresolved rotation-setter); offset 0 is inferred as ObjID
// by strict analogy with every sibling in this packet family (Delete/
// Position/Duplicate all lead with a 4-byte ObjID at offset 0), and because
// the server-side relay/broadcast has no other way to know which stage
// object rotated.
type MsgSysRotateObject struct {
	ObjID    uint32
	Rotation float32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRotateObject) Opcode() network.PacketID {
	return network.MSG_SYS_ROTATE_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysRotateObject) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.ObjID = bf.ReadUint32()
	m.Rotation = bf.ReadFloat32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysRotateObject) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.ObjID)
	bf.WriteFloat32(m.Rotation)
	return nil
}
