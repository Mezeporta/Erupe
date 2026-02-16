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

// buildCatBytes constructs a binary cat data payload suitable for GetAirouDetails.
func buildCatBytes(cats []Airou) []byte {
	buf := new(bytes.Buffer)
	// catCount
	buf.WriteByte(byte(len(cats)))
	for _, cat := range cats {
		catBuf := new(bytes.Buffer)
		// ID (uint32)
		binary.Write(catBuf, binary.BigEndian, cat.ID)
		// 1 byte skip (unknown bool)
		catBuf.WriteByte(0)
		// Name (18 bytes)
		name := make([]byte, 18)
		copy(name, cat.Name)
		catBuf.Write(name)
		// Task (uint8)
		catBuf.WriteByte(cat.Task)
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

func TestGetAirouDetails_Empty(t *testing.T) {
	// Zero cats
	data := []byte{0x00}
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetAirouDetails(bf)

	if len(cats) != 0 {
		t.Errorf("Expected 0 cats, got %d", len(cats))
	}
}

func TestGetAirouDetails_SingleCat(t *testing.T) {
	input := Airou{
		ID:       42,
		Name:     []byte("TestCat"),
		Task: 4,
		Personality: 3,
		Class:       2,
		Experience:  1500,
		WeaponType:  6,
		WeaponID:    100,
	}

	data := buildCatBytes([]Airou{input})
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetAirouDetails(bf)

	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat, got %d", len(cats))
	}

	cat := cats[0]
	if cat.ID != 42 {
		t.Errorf("ID = %d, want 42", cat.ID)
	}
	if cat.Task != 4 {
		t.Errorf("Task = %d, want 4", cat.Task)
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
	if len(cat.Name) != 18 {
		t.Errorf("Name length = %d, want 18", len(cat.Name))
	}
	// First bytes should match "TestCat"
	if !bytes.HasPrefix(cat.Name, []byte("TestCat")) {
		t.Errorf("Name does not start with 'TestCat', got %v", cat.Name)
	}
}

func TestGetAirouDetails_MultipleCats(t *testing.T) {
	inputs := []Airou{
		{ID: 1, Name: []byte("Alpha"), Task: 1, Personality: 0, Class: 0, Experience: 100, WeaponType: 6, WeaponID: 10},
		{ID: 2, Name: []byte("Beta"), Task: 2, Personality: 1, Class: 1, Experience: 200, WeaponType: 6, WeaponID: 20},
		{ID: 3, Name: []byte("Gamma"), Task: 4, Personality: 2, Class: 2, Experience: 300, WeaponType: 6, WeaponID: 30},
	}

	data := buildCatBytes(inputs)
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetAirouDetails(bf)

	if len(cats) != 3 {
		t.Fatalf("Expected 3 cats, got %d", len(cats))
	}

	for i, cat := range cats {
		if cat.ID != inputs[i].ID {
			t.Errorf("Cat %d: CatID = %d, want %d", i, cat.ID, inputs[i].ID)
		}
		if cat.Task != inputs[i].Task {
			t.Errorf("Cat %d: CurrentTask = %d, want %d", i, cat.Task, inputs[i].Task)
		}
		if cat.Experience != inputs[i].Experience {
			t.Errorf("Cat %d: Experience = %d, want %d", i, cat.Experience, inputs[i].Experience)
		}
		if cat.WeaponID != inputs[i].WeaponID {
			t.Errorf("Cat %d: WeaponID = %d, want %d", i, cat.WeaponID, inputs[i].WeaponID)
		}
	}
}

func TestGetAirouDetails_ExtraTrailingBytes(t *testing.T) {
	// The GetAirouDetails function handles extra bytes by seeking to catStart+catDefLen.
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
	cats := GetAirouDetails(bf)

	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat, got %d", len(cats))
	}
	if cats[0].ID != 99 {
		t.Errorf("ID = %d, want 99", cats[0].ID)
	}
	if cats[0].Experience != 500 {
		t.Errorf("Experience = %d, want 500", cats[0].Experience)
	}
}

func TestGetAirouDetails_CatNamePadding(t *testing.T) {
	// Verify that names shorter than 18 bytes are correctly padded with null bytes.
	input := Airou{
		ID:   1,
		Name: []byte("Hi"),
	}

	data := buildCatBytes([]Airou{input})
	bf := byteframe.NewByteFrameFromBytes(data)
	cats := GetAirouDetails(bf)

	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat, got %d", len(cats))
	}
	if len(cats[0].Name) != 18 {
		t.Errorf("Name length = %d, want 18", len(cats[0].Name))
	}
	// "Hi" followed by null bytes
	if cats[0].Name[0] != 'H' || cats[0].Name[1] != 'i' {
		t.Errorf("Name first bytes = %v, want 'Hi...'", cats[0].Name[:2])
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
