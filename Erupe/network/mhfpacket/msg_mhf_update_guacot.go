package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// GuacotUpdateEntry represents an entry inside the MsgMhfUpdateGuacot packet.
type GuacotUpdateEntry struct {
	Unk0           uint32
	Unk1           uint16
	Unk2           uint16
	Unk3           uint16
	Unk4           uint16
	Unk5           uint16
	Unk6           uint16
	Unk7           uint16
	Unk8           uint16
	Unk9           uint16
	Unk10          uint16
	Unk11          uint16
	Unk12          uint16
	Unk13          uint16
	Unk14          uint16
	Unk15          uint16
	Unk16          uint16
	Unk17          uint16
	Unk18          uint16
	Unk19          uint16
	Unk20          uint16
	Unk21          uint16
	Unk22          uint16
	Unk23          uint32
	Unk24          uint32
	DataSize       uint8
	RawDataPayload []byte
}

// MsgMhfUpdateGuacot represents the MSG_MHF_UPDATE_GUACOT
type MsgMhfUpdateGuacot struct {
	AckHandle  uint32
	EntryCount uint16
	Unk0       uint16 // Hardcoded 0 in binary
	Entries    []*GuacotUpdateEntry
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuacot) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.EntryCount = bf.ReadUint16()
	m.Unk0 = bf.ReadUint16()
	for i := 0; i < int(m.EntryCount); i++ {
		// Yikes.
		e := &GuacotUpdateEntry{}

		e.Unk0 = bf.ReadUint32()
		e.Unk1 = bf.ReadUint16()
		e.Unk2 = bf.ReadUint16()
		e.Unk3 = bf.ReadUint16()
		e.Unk4 = bf.ReadUint16()
		e.Unk5 = bf.ReadUint16()
		e.Unk6 = bf.ReadUint16()
		e.Unk7 = bf.ReadUint16()
		e.Unk8 = bf.ReadUint16()
		e.Unk9 = bf.ReadUint16()
		e.Unk10 = bf.ReadUint16()
		e.Unk11 = bf.ReadUint16()
		e.Unk12 = bf.ReadUint16()
		e.Unk13 = bf.ReadUint16()
		e.Unk14 = bf.ReadUint16()
		e.Unk15 = bf.ReadUint16()
		e.Unk16 = bf.ReadUint16()
		e.Unk17 = bf.ReadUint16()
		e.Unk18 = bf.ReadUint16()
		e.Unk19 = bf.ReadUint16()
		e.Unk20 = bf.ReadUint16()
		e.Unk21 = bf.ReadUint16()
		e.Unk22 = bf.ReadUint16()
		e.Unk23 = bf.ReadUint32()
		e.Unk24 = bf.ReadUint32()
		e.DataSize = bf.ReadUint8()
		e.RawDataPayload = bf.ReadBytes(uint(e.DataSize))

		m.Entries = append(m.Entries, e)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuacot) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

func (m *MsgMhfUpdateGuacot) GuacotUpdateEntryToBytes(Entry *GuacotUpdateEntry) []byte {
	resp:= byteframe.NewByteFrame()
	resp.WriteUint32(Entry.Unk0)
	resp.WriteUint16(Entry.Unk1)
	resp.WriteUint16(Entry.Unk1)
	resp.WriteUint16(Entry.Unk2)
	resp.WriteUint16(Entry.Unk3)
	resp.WriteUint16(Entry.Unk4)
	resp.WriteUint16(Entry.Unk5)
	resp.WriteUint16(Entry.Unk6)
	resp.WriteUint16(Entry.Unk7)
	resp.WriteUint16(Entry.Unk8)
	resp.WriteUint16(Entry.Unk9)
	resp.WriteUint16(Entry.Unk10)
	resp.WriteUint16(Entry.Unk11)
	resp.WriteUint16(Entry.Unk12)
	resp.WriteUint16(Entry.Unk13)
	resp.WriteUint16(Entry.Unk14)
	resp.WriteUint16(Entry.Unk15)
	resp.WriteUint16(Entry.Unk16)
	resp.WriteUint16(Entry.Unk17)
	resp.WriteUint16(Entry.Unk18)
	resp.WriteUint16(Entry.Unk19)
	resp.WriteUint16(Entry.Unk20)
	resp.WriteUint16(Entry.Unk21)
	resp.WriteUint16(Entry.Unk22)
	resp.WriteUint32(Entry.Unk23)
	resp.WriteUint32(Entry.Unk24)
	resp.WriteUint8(Entry.DataSize)
	resp.WriteBytes(Entry.RawDataPayload)
	return resp.Data()
}