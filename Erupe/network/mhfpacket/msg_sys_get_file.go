package mhfpacket

import (
	"errors"

	"github.com/Solenataris/Erupe/common/bfutil"
	"github.com/Solenataris/Erupe/network"
	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Andoryuuta/byteframe"
)

type scenarioFileIdentifer struct {
	CategoryID uint8
	MainID     uint32
	ChapterID  uint8
	/*
		Flags represent the following bit flags:
		11111111 -> Least significant bit on the right.
		||||||||
		|||||||0x1: Chunk0-type, recursive chunks, quest name/description + 0x14 byte unk info
		||||||0x2:  Chunk1-type, recursive chunks, npc dialog? + 0x2C byte unk info
		|||||0x4:   UNK NONE FOUND. (Guessing from the following that this might be a chunk2-type)
		||||0x8:    Chunk0-type, NO RECURSIVE CHUNKS ([0x1] prefixed?), Episode listing
		|||0x10:    Chunk1-type, NO RECURSIVE CHUNKS, JKR blob, npc dialog?
		||0x20:     Chunk2-type, NO RECURSIVE CHUNKS, JKR blob, Menu options or quest titles?
		|0x40:      UNK NONE FOUND
		0x80:       UNK NONE FOUND
	*/
	Flags uint8
}

// MsgSysGetFile represents the MSG_SYS_GET_FILE
type MsgSysGetFile struct {
	AckHandle         uint32
	IsScenario        bool
	Filename          string
	ScenarioIdentifer scenarioFileIdentifer
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetFile) Opcode() network.PacketID {
	return network.MSG_SYS_GET_FILE
}

// Parse parses the packet from binary
func (m *MsgSysGetFile) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.IsScenario = bf.ReadBool()
	filenameLength := bf.ReadUint8()
	if filenameLength > 0 {
		m.Filename = string(bfutil.UpToNull(bf.ReadBytes(uint(filenameLength))))
	}

	if m.IsScenario {
		m.ScenarioIdentifer = scenarioFileIdentifer{
			bf.ReadUint8(),
			bf.ReadUint32(),
			bf.ReadUint8(),
			bf.ReadUint8(),
		}
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetFile) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
