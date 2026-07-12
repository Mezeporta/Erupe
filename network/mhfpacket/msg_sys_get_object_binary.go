package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgSysGetObjectBinary represents the MSG_SYS_GET_OBJECT_BINARY.
//
// Wire format recovered by decompiling the PC client's own dispatch handler
// (pkt_handler_MSG_SYS_GET_OBJECT_BINARY @ 0x115039D0 in mhfo-hd.dll, see
// docs/network_protocol.md's PC Client Dispatch Table): AckHandle at offset
// 0, ObjID at offset 8. The 4 bytes in between (offset 4) are never
// dereferenced by that handler -- consumed here as Unk0 rather than
// skipped, so the cursor lands correctly on ObjID.
type MsgSysGetObjectBinary struct {
	AckHandle uint32
	Unk0      uint32
	ObjID     uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_GET_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysGetObjectBinary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.ObjID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetObjectBinary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
