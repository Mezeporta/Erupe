package mhfpacket

import (
    "errors"

    "github.com/Andoryuuta/byteframe"
    "github.com/Solenataris/Erupe/network"
    "github.com/Solenataris/Erupe/network/clientctx"
)

// MsgMhfUpdateBeatLevel represents the MSG_MHF_UPDATE_BEAT_LEVEL
type MsgMhfUpdateBeatLevel struct {
    AckHandle uint32
    Unk1      uint32
    Unk2      uint32
    Unk3      uint16
    Unk4      uint16
    Unk5      uint32
    Unk6      uint32
    Unk7      uint32
    Unk8      uint16
    Unk9      uint16
    Unk10     uint32
    Unk11     uint32
    Unk12     uint32
    Unk13     uint32
    Unk14     uint32
    Unk15     uint32
    Unk16     uint32
    Unk17     uint32
    Unk18     uint32
    Unk19     uint32
    Unk20     uint32
    Unk21     uint16
    Unk22     uint16
    Unk23     uint16
    Unk24     uint32
    Unk25     uint32
    Unk26     uint16
    Unk27     uint32
    Unk28     uint32
    Unk29     uint32
    Unk30     uint32
    Unk31     uint32
    Unk32     uint32
    Unk33     uint32
    Unk34     uint32
    Unk35     uint32
    Unk36     uint32
    Unk37     uint32
    Unk38     uint16
    Unk39     uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateBeatLevel) Opcode() network.PacketID {
    return network.MSG_MHF_UPDATE_BEAT_LEVEL
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateBeatLevel) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
    m.AckHandle = bf.ReadUint32()
    m.Unk1 = bf.ReadUint32()
    m.Unk2 = bf.ReadUint32()
    m.Unk3 = bf.ReadUint16()
    m.Unk4 = bf.ReadUint16()
    m.Unk5 = bf.ReadUint32()
    m.Unk6 = bf.ReadUint32()
    m.Unk7 = bf.ReadUint32()
    m.Unk8 = bf.ReadUint16()
    m.Unk9 = bf.ReadUint16()
    m.Unk10 = bf.ReadUint32()
    m.Unk11 = bf.ReadUint32()
    m.Unk12 = bf.ReadUint32()
    m.Unk13 = bf.ReadUint32()
    m.Unk14 = bf.ReadUint32()
    m.Unk15 = bf.ReadUint32()
    m.Unk16 = bf.ReadUint32()
    m.Unk17 = bf.ReadUint32()
    m.Unk18 = bf.ReadUint32()
    m.Unk19 = bf.ReadUint32()
    m.Unk20 = bf.ReadUint32()
    m.Unk21 = bf.ReadUint16()
    m.Unk22 = bf.ReadUint16()
    m.Unk23 = bf.ReadUint16()
    m.Unk24 = bf.ReadUint32()
    m.Unk25 = bf.ReadUint32()
    m.Unk26 = bf.ReadUint16()
    m.Unk27 = bf.ReadUint32()
    m.Unk28 = bf.ReadUint32()
    m.Unk29 = bf.ReadUint32()
    m.Unk30 = bf.ReadUint32()
    m.Unk31 = bf.ReadUint32()
    m.Unk32 = bf.ReadUint32()
    m.Unk33 = bf.ReadUint32()
    m.Unk34 = bf.ReadUint32()
    m.Unk35 = bf.ReadUint32()
    m.Unk36 = bf.ReadUint32()
    m.Unk37 = bf.ReadUint32()
    m.Unk38 = bf.ReadUint16()
    m.Unk39 = bf.ReadUint8()
    return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateBeatLevel) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
    return errors.New("NOT IMPLEMENTED")
}
