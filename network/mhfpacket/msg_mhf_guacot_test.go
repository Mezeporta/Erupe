package mhfpacket

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
)

func TestMsgMhfUpdateGuacotOpcode_Guacot(t *testing.T) {
	pkt := &MsgMhfUpdateGuacot{}
	if pkt.Opcode() != network.MSG_MHF_UPDATE_GUACOT {
		t.Errorf("Opcode() = %s, want MSG_MHF_UPDATE_GUACOT", pkt.Opcode())
	}
}

func TestMsgMhfEnumerateGuacotOpcode_Guacot(t *testing.T) {
	pkt := &MsgMhfEnumerateGuacot{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_GUACOT {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_GUACOT", pkt.Opcode())
	}
}

func TestMsgMhfUpdateGuacotParse_SingleEntry(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD) // AckHandle
	bf.WriteUint16(1)          // EntryCount
	bf.WriteUint16(0)          // Zeroed

	// Goocoo entry
	bf.WriteUint32(2) // Index
	for i := 0; i < 22; i++ {
		bf.WriteInt16(int16(i + 1)) // Data1
	}
	bf.WriteUint32(100) // Data2[0]
	bf.WriteUint32(200) // Data2[1]
	bf.WriteUint8(5)    // Name length
	bf.WriteBytes([]byte("Porky"))

	pkt := &MsgMhfUpdateGuacot{}
	_, _ = bf.Seek(0, 0)
	err := pkt.Parse(bf, nil)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if pkt.AckHandle != 0xAABBCCDD {
		t.Errorf("AckHandle = 0x%X, want 0xAABBCCDD", pkt.AckHandle)
	}
	if pkt.EntryCount != 1 {
		t.Errorf("EntryCount = %d, want 1", pkt.EntryCount)
	}
	if len(pkt.Goocoos) != 1 {
		t.Fatalf("len(Goocoos) = %d, want 1", len(pkt.Goocoos))
	}

	g := pkt.Goocoos[0]
	if g.Index != 2 {
		t.Errorf("Index = %d, want 2", g.Index)
	}
	if len(g.Data1) != 22 {
		t.Fatalf("len(Data1) = %d, want 22", len(g.Data1))
	}
	for i := 0; i < 22; i++ {
		if g.Data1[i] != int16(i+1) {
			t.Errorf("Data1[%d] = %d, want %d", i, g.Data1[i], i+1)
		}
	}
	if len(g.Data2) != 2 {
		t.Fatalf("len(Data2) = %d, want 2", len(g.Data2))
	}
	if g.Data2[0] != 100 {
		t.Errorf("Data2[0] = %d, want 100", g.Data2[0])
	}
	if g.Data2[1] != 200 {
		t.Errorf("Data2[1] = %d, want 200", g.Data2[1])
	}
	if string(g.Name) != "Porky" {
		t.Errorf("Name = %q, want %q", string(g.Name), "Porky")
	}
}

func TestMsgMhfUpdateGuacotParse_MultipleEntries(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)  // AckHandle
	bf.WriteUint16(3)  // EntryCount
	bf.WriteUint16(0)  // Zeroed

	for idx := uint32(0); idx < 3; idx++ {
		bf.WriteUint32(idx) // Index
		for i := 0; i < 22; i++ {
			bf.WriteInt16(int16(idx*100 + uint32(i)))
		}
		bf.WriteUint32(idx * 10) // Data2[0]
		bf.WriteUint32(idx * 20) // Data2[1]
		name := []byte("Pog")
		bf.WriteUint8(uint8(len(name)))
		bf.WriteBytes(name)
	}

	pkt := &MsgMhfUpdateGuacot{}
	_, _ = bf.Seek(0, 0)
	err := pkt.Parse(bf, nil)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(pkt.Goocoos) != 3 {
		t.Fatalf("len(Goocoos) = %d, want 3", len(pkt.Goocoos))
	}
	for idx := uint32(0); idx < 3; idx++ {
		g := pkt.Goocoos[idx]
		if g.Index != idx {
			t.Errorf("Goocoos[%d].Index = %d, want %d", idx, g.Index, idx)
		}
		if g.Data1[0] != int16(idx*100) {
			t.Errorf("Goocoos[%d].Data1[0] = %d, want %d", idx, g.Data1[0], idx*100)
		}
		if g.Data2[0] != idx*10 {
			t.Errorf("Goocoos[%d].Data2[0] = %d, want %d", idx, g.Data2[0], idx*10)
		}
	}
}

func TestMsgMhfUpdateGuacotParse_ZeroEntries(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(42) // AckHandle
	bf.WriteUint16(0)  // EntryCount
	bf.WriteUint16(0)  // Zeroed

	pkt := &MsgMhfUpdateGuacot{}
	_, _ = bf.Seek(0, 0)
	err := pkt.Parse(bf, nil)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if pkt.EntryCount != 0 {
		t.Errorf("EntryCount = %d, want 0", pkt.EntryCount)
	}
	if len(pkt.Goocoos) != 0 {
		t.Errorf("len(Goocoos) = %d, want 0", len(pkt.Goocoos))
	}
}

func TestMsgMhfUpdateGuacotParse_DeletionEntry(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint16(1) // EntryCount
	bf.WriteUint16(0) // Zeroed

	bf.WriteUint32(0) // Index
	// Data1[0] = 0 signals deletion
	bf.WriteInt16(0)
	for i := 1; i < 22; i++ {
		bf.WriteInt16(0)
	}
	bf.WriteUint32(0) // Data2[0]
	bf.WriteUint32(0) // Data2[1]
	bf.WriteUint8(0)  // Empty name

	pkt := &MsgMhfUpdateGuacot{}
	_, _ = bf.Seek(0, 0)
	err := pkt.Parse(bf, nil)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	g := pkt.Goocoos[0]
	if g.Data1[0] != 0 {
		t.Errorf("Data1[0] = %d, want 0 (deletion marker)", g.Data1[0])
	}
}

func TestMsgMhfUpdateGuacotParse_EmptyName(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint16(1) // EntryCount
	bf.WriteUint16(0) // Zeroed

	bf.WriteUint32(0) // Index
	for i := 0; i < 22; i++ {
		bf.WriteInt16(1)
	}
	bf.WriteUint32(0) // Data2[0]
	bf.WriteUint32(0) // Data2[1]
	bf.WriteUint8(0)  // Empty name

	pkt := &MsgMhfUpdateGuacot{}
	_, _ = bf.Seek(0, 0)
	err := pkt.Parse(bf, nil)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(pkt.Goocoos[0].Name) != 0 {
		t.Errorf("Name length = %d, want 0", len(pkt.Goocoos[0].Name))
	}
}

func TestMsgMhfEnumerateGuacotParse(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint32(0)          // Unk0
	bf.WriteUint16(0)          // Zeroed

	pkt := &MsgMhfEnumerateGuacot{}
	_, _ = bf.Seek(0, 0)
	err := pkt.Parse(bf, nil)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.Unk0 != 0 {
		t.Errorf("Unk0 = %d, want 0", pkt.Unk0)
	}
}

func TestMsgMhfUpdateGuacotBuild_NotImplemented(t *testing.T) {
	pkt := &MsgMhfUpdateGuacot{}
	err := pkt.Build(byteframe.NewByteFrame(), nil)
	if err == nil {
		t.Error("Build() should return error (not implemented)")
	}
}

func TestMsgMhfEnumerateGuacotBuild_NotImplemented(t *testing.T) {
	pkt := &MsgMhfEnumerateGuacot{}
	err := pkt.Build(byteframe.NewByteFrame(), nil)
	if err == nil {
		t.Error("Build() should return error (not implemented)")
	}
}

func TestGoocooStruct_Data1Size(t *testing.T) {
	// Verify 22 int16 entries = 44 bytes of outfit/appearance data
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint16(1) // EntryCount
	bf.WriteUint16(0) // Zeroed

	bf.WriteUint32(0) // Index
	for i := 0; i < 22; i++ {
		bf.WriteInt16(int16(i * 3))
	}
	bf.WriteUint32(0xDEAD) // Data2[0]
	bf.WriteUint32(0xBEEF) // Data2[1]
	bf.WriteUint8(0)       // No name

	pkt := &MsgMhfUpdateGuacot{}
	_, _ = bf.Seek(0, 0)
	_ = pkt.Parse(bf, nil)

	g := pkt.Goocoos[0]

	// Verify all 22 data slots are correctly read
	for i := 0; i < 22; i++ {
		expected := int16(i * 3)
		if g.Data1[i] != expected {
			t.Errorf("Data1[%d] = %d, want %d", i, g.Data1[i], expected)
		}
	}

	if g.Data2[0] != 0xDEAD {
		t.Errorf("Data2[0] = 0x%X, want 0xDEAD", g.Data2[0])
	}
	if g.Data2[1] != 0xBEEF {
		t.Errorf("Data2[1] = 0x%X, want 0xBEEF", g.Data2[1])
	}
}

func TestGoocooSerialization_Roundtrip(t *testing.T) {
	// Simulate what handleMsgMhfUpdateGuacot does when saving to DB
	goocoo := Goocoo{
		Index: 1,
		Data1: make([]int16, 22),
		Data2: []uint32{0x1234, 0x5678},
		Name:  []byte("MyPoogie"),
	}
	goocoo.Data1[0] = 5    // outfit type (non-zero = exists)
	goocoo.Data1[1] = 100  // some appearance data
	goocoo.Data1[21] = -50 // test negative int16

	// Serialize (matches handler logic)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(goocoo.Index)
	for i := range goocoo.Data1 {
		bf.WriteInt16(goocoo.Data1[i])
	}
	for i := range goocoo.Data2 {
		bf.WriteUint32(goocoo.Data2[i])
	}
	bf.WriteUint8(uint8(len(goocoo.Name)))
	bf.WriteBytes(goocoo.Name)

	// Deserialize and verify
	data := bf.Data()
	rbf := byteframe.NewByteFrameFromBytes(data)

	index := rbf.ReadUint32()
	if index != 1 {
		t.Errorf("index = %d, want 1", index)
	}

	data1_0 := rbf.ReadInt16()
	if data1_0 != 5 {
		t.Errorf("data1[0] = %d, want 5", data1_0)
	}
	data1_1 := rbf.ReadInt16()
	if data1_1 != 100 {
		t.Errorf("data1[1] = %d, want 100", data1_1)
	}
	// Skip to data1[21]
	for i := 2; i < 21; i++ {
		rbf.ReadInt16()
	}
	data1_21 := rbf.ReadInt16()
	if data1_21 != -50 {
		t.Errorf("data1[21] = %d, want -50", data1_21)
	}

	d2_0 := rbf.ReadUint32()
	if d2_0 != 0x1234 {
		t.Errorf("data2[0] = 0x%X, want 0x1234", d2_0)
	}
	d2_1 := rbf.ReadUint32()
	if d2_1 != 0x5678 {
		t.Errorf("data2[1] = 0x%X, want 0x5678", d2_1)
	}

	nameLen := rbf.ReadUint8()
	if nameLen != 8 {
		t.Errorf("nameLen = %d, want 8", nameLen)
	}
	name := rbf.ReadBytes(uint(nameLen))
	if string(name) != "MyPoogie" {
		t.Errorf("name = %q, want %q", string(name), "MyPoogie")
	}
}

func TestGoocooEntrySize(t *testing.T) {
	// Each goocoo entry in the packet should be:
	// 4 (index) + 22*2 (data1) + 2*4 (data2) + 1 (name len) + N (name)
	// = 4 + 44 + 8 + 1 + N = 57 + N bytes
	name := []byte("Test")
	expectedSize := 4 + 44 + 8 + 1 + len(name)

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // index
	for i := 0; i < 22; i++ {
		bf.WriteInt16(0)
	}
	bf.WriteUint32(0)               // data2[0]
	bf.WriteUint32(0)               // data2[1]
	bf.WriteUint8(uint8(len(name))) // name len
	bf.WriteBytes(name)

	if len(bf.Data()) != expectedSize {
		t.Errorf("entry size = %d bytes, want %d bytes (57 + %d name)", len(bf.Data()), expectedSize, len(name))
	}
}
