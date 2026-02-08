package channelserver

import (
	"bytes"
	"encoding/binary"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfLoadLegendDispatch(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadLegendDispatch{
		AckHandle: 12345,
	}

	handleMsgMhfLoadLegendDispatch(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// --- NEW TESTS ---

// buildCatBytes constructs a binary cat data payload suitable for GetCatDetails.
func buildCatBytes(cats []CatDefinition) []byte {
	buf := new(bytes.Buffer)
	// catCount
	buf.WriteByte(byte(len(cats)))
	for _, cat := range cats {
		catBuf := new(bytes.Buffer)
		// catID (uint32)
		binary.Write(catBuf, binary.BigEndian, cat.CatID)
		// 1 byte skip (unknown bool)
		catBuf.WriteByte(0)
		// CatName (18 bytes)
		name := make([]byte, 18)
		copy(name, cat.CatName)
		catBuf.Write(name)
		// CurrentTask (uint8)
		catBuf.WriteByte(cat.CurrentTask)
		// 16 bytes skip (appearance data)
		catBuf.Write(make([]byte, 16))
		// Personality (uint8)
		catBuf.WriteByte(cat.Personality)
		// Class (uint8)
		catBuf.WriteByte(cat.Class)
		// 5 bytes skip (affection and colour sliders)
		catBuf.Write(make([]byte, 5))
		// Experience (uint32)
		binary.Write(catBuf, binary.BigEndian, cat.Experience)
		// 1 byte skip (bool for weapon equipped)
		catBuf.WriteByte(0)
		// WeaponType (uint8)
		catBuf.WriteByte(cat.WeaponType)
		// WeaponID (uint16)
		binary.Write(catBuf, binary.BigEndian, cat.WeaponID)

		catData := catBuf.Bytes()
		// catDefLen (uint32) - total length of the cat data after this field
		binary.Write(buf, binary.BigEndian, uint32(len(catData)))
		buf.Write(catData)
	}
	return buf.Bytes()
}

func TestGetCatDetails_Empty(t *testing.T) {
	// Zero cats
	data := []byte{0x00}
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetCatDetails(bf)

	if len(cats) != 0 {
		t.Errorf("Expected 0 cats, got %d", len(cats))
	}
}

func TestGetCatDetails_SingleCat(t *testing.T) {
	input := CatDefinition{
		CatID:       42,
		CatName:     []byte("TestCat"),
		CurrentTask: 4,
		Personality: 3,
		Class:       2,
		Experience:  1500,
		WeaponType:  6,
		WeaponID:    100,
	}

	data := buildCatBytes([]CatDefinition{input})
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetCatDetails(bf)

	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat, got %d", len(cats))
	}

	cat := cats[0]
	if cat.CatID != 42 {
		t.Errorf("CatID = %d, want 42", cat.CatID)
	}
	if cat.CurrentTask != 4 {
		t.Errorf("CurrentTask = %d, want 4", cat.CurrentTask)
	}
	if cat.Personality != 3 {
		t.Errorf("Personality = %d, want 3", cat.Personality)
	}
	if cat.Class != 2 {
		t.Errorf("Class = %d, want 2", cat.Class)
	}
	if cat.Experience != 1500 {
		t.Errorf("Experience = %d, want 1500", cat.Experience)
	}
	if cat.WeaponType != 6 {
		t.Errorf("WeaponType = %d, want 6", cat.WeaponType)
	}
	if cat.WeaponID != 100 {
		t.Errorf("WeaponID = %d, want 100", cat.WeaponID)
	}
	// Name should be 18 bytes (padded with nulls)
	if len(cat.CatName) != 18 {
		t.Errorf("CatName length = %d, want 18", len(cat.CatName))
	}
	// First bytes should match "TestCat"
	if !bytes.HasPrefix(cat.CatName, []byte("TestCat")) {
		t.Errorf("CatName does not start with 'TestCat', got %v", cat.CatName)
	}
}

func TestGetCatDetails_MultipleCats(t *testing.T) {
	inputs := []CatDefinition{
		{CatID: 1, CatName: []byte("Alpha"), CurrentTask: 1, Personality: 0, Class: 0, Experience: 100, WeaponType: 6, WeaponID: 10},
		{CatID: 2, CatName: []byte("Beta"), CurrentTask: 2, Personality: 1, Class: 1, Experience: 200, WeaponType: 6, WeaponID: 20},
		{CatID: 3, CatName: []byte("Gamma"), CurrentTask: 4, Personality: 2, Class: 2, Experience: 300, WeaponType: 6, WeaponID: 30},
	}

	data := buildCatBytes(inputs)
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetCatDetails(bf)

	if len(cats) != 3 {
		t.Fatalf("Expected 3 cats, got %d", len(cats))
	}

	for i, cat := range cats {
		if cat.CatID != inputs[i].CatID {
			t.Errorf("Cat %d: CatID = %d, want %d", i, cat.CatID, inputs[i].CatID)
		}
		if cat.CurrentTask != inputs[i].CurrentTask {
			t.Errorf("Cat %d: CurrentTask = %d, want %d", i, cat.CurrentTask, inputs[i].CurrentTask)
		}
		if cat.Experience != inputs[i].Experience {
			t.Errorf("Cat %d: Experience = %d, want %d", i, cat.Experience, inputs[i].Experience)
		}
		if cat.WeaponID != inputs[i].WeaponID {
			t.Errorf("Cat %d: WeaponID = %d, want %d", i, cat.WeaponID, inputs[i].WeaponID)
		}
	}
}

func TestGetCatDetails_ExtraTrailingBytes(t *testing.T) {
	// The GetCatDetails function handles extra bytes by seeking to catStart+catDefLen.
	// Simulate a cat definition with extra trailing bytes by increasing catDefLen.
	buf := new(bytes.Buffer)
	buf.WriteByte(1) // catCount = 1

	catBuf := new(bytes.Buffer)
	binary.Write(catBuf, binary.BigEndian, uint32(99))  // catID
	catBuf.WriteByte(0)                                 // skip
	catBuf.Write(make([]byte, 18))                      // name
	catBuf.WriteByte(3)                                 // currentTask
	catBuf.Write(make([]byte, 16))                      // appearance skip
	catBuf.WriteByte(1)                                 // personality
	catBuf.WriteByte(2)                                 // class
	catBuf.Write(make([]byte, 5))                       // affection skip
	binary.Write(catBuf, binary.BigEndian, uint32(500)) // experience
	catBuf.WriteByte(0)                                 // weapon equipped bool
	catBuf.WriteByte(6)                                 // weaponType
	binary.Write(catBuf, binary.BigEndian, uint16(50))  // weaponID

	catData := catBuf.Bytes()
	// Add 10 extra trailing bytes
	extra := make([]byte, 10)
	catDataWithExtra := append(catData, extra...)

	binary.Write(buf, binary.BigEndian, uint32(len(catDataWithExtra)))
	buf.Write(catDataWithExtra)

	bf := byteframe.NewByteFrameFromBytes(buf.Bytes())
	cats := GetCatDetails(bf)

	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat, got %d", len(cats))
	}
	if cats[0].CatID != 99 {
		t.Errorf("CatID = %d, want 99", cats[0].CatID)
	}
	if cats[0].Experience != 500 {
		t.Errorf("Experience = %d, want 500", cats[0].Experience)
	}
}

func TestGetCatDetails_CatNamePadding(t *testing.T) {
	// Verify that names shorter than 18 bytes are correctly padded with null bytes.
	input := CatDefinition{
		CatID:   1,
		CatName: []byte("Hi"),
	}

	data := buildCatBytes([]CatDefinition{input})
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetCatDetails(bf)

	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat, got %d", len(cats))
	}
	if len(cats[0].CatName) != 18 {
		t.Errorf("CatName length = %d, want 18", len(cats[0].CatName))
	}
	// "Hi" followed by null bytes
	if cats[0].CatName[0] != 'H' || cats[0].CatName[1] != 'i' {
		t.Errorf("CatName first bytes = %v, want 'Hi...'", cats[0].CatName[:2])
	}
}

// TestHandleMsgMhfMercenaryHuntdata_Unk0_1 tests with Unk0=1 (returns 1 byte)
func TestHandleMsgMhfMercenaryHuntdata_Unk0_1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle: 12345,
		Unk0:      1,
	}

	handleMsgMhfMercenaryHuntdata(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgMhfMercenaryHuntdata_Unk0_0 tests with Unk0=0 (returns 0 bytes payload)
func TestHandleMsgMhfMercenaryHuntdata_Unk0_0(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle: 12345,
		Unk0:      0,
	}

	handleMsgMhfMercenaryHuntdata(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgMhfEnumerateMercenaryLog tests the mercenary log enumeration handler
func TestHandleMsgMhfEnumerateMercenaryLog(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateMercenaryLog{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateMercenaryLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}
