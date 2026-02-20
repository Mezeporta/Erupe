package mhfpacket

import (
	"errors"
	"erupe-ce/common/bfutil"
	"erupe-ce/common/stringsupport"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfApplyBbsArticle represents the MSG_MHF_APPLY_BBS_ARTICLE
type MsgMhfApplyBbsArticle struct {
	AckHandle   uint32
	Unk0        uint32
	Unk1        []byte
	Name        string
	Title       string
	Description string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyBbsArticle) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_BBS_ARTICLE
}

// Parse parses the packet from binary
func (m *MsgMhfApplyBbsArticle) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadBytes(16)
	m.Name, _ = stringsupport.SJISToUTF8(bfutil.UpToNull(bf.ReadBytes(32)))
	m.Title, _ = stringsupport.SJISToUTF8(bfutil.UpToNull(bf.ReadBytes(128)))
	m.Description, _ = stringsupport.SJISToUTF8(bfutil.UpToNull(bf.ReadBytes(256)))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyBbsArticle) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
