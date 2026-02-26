package channelserver

import (
	"bytes"
	"encoding/binary"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	cfg "erupe-ce/config"
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
		_ = binary.Write(catBuf, binary.BigEndian, cat.ID)
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
		_ = binary.Write(catBuf, binary.BigEndian, cat.Experience)
		// 1 byte skip (bool for weapon equipped)
		catBuf.WriteByte(0)
		// WeaponType (uint8)
		catBuf.WriteByte(cat.WeaponType)
		// WeaponID (uint16)
		_ = binary.Write(catBuf, binary.BigEndian, cat.WeaponID)

		catData := catBuf.Bytes()
		// catDefLen (uint32) - total length of the cat data after this field
		_ = binary.Write(buf, binary.BigEndian, uint32(len(catData)))
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
		ID:          42,
		Name:        []byte("TestCat"),
		Task:        4,
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
	_ = binary.Write(catBuf, binary.BigEndian, uint32(99))  // catID
	catBuf.WriteByte(0)                                     // skip
	catBuf.Write(make([]byte, 18))                          // name
	catBuf.WriteByte(3)                                     // currentTask
	catBuf.Write(make([]byte, 16))                          // appearance skip
	catBuf.WriteByte(1)                                     // personality
	catBuf.WriteByte(2)                                     // class
	catBuf.Write(make([]byte, 5))                           // affection skip
	_ = binary.Write(catBuf, binary.BigEndian, uint32(500)) // experience
	catBuf.WriteByte(0)                                     // weapon equipped bool
	catBuf.WriteByte(6)                                     // weaponType
	_ = binary.Write(catBuf, binary.BigEndian, uint16(50))  // weaponID

	catData := catBuf.Bytes()
	// Add 10 extra trailing bytes
	extra := make([]byte, 10)
	catDataWithExtra := append(catData, extra...)

	_ = binary.Write(buf, binary.BigEndian, uint32(len(catDataWithExtra)))
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

func TestHandleMsgMhfMercenaryHuntdata_RequestTypeIs1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle:   12345,
		RequestType: 1,
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

func TestHandleMsgMhfMercenaryHuntdata_RequestTypeIs0(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle:   12345,
		RequestType: 0,
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

func TestHandleMsgMhfMercenaryHuntdata_RequestTypeIs2(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle:   12345,
		RequestType: 2,
	}

	handleMsgMhfMercenaryHuntdata(session, pkt)

	// RequestType=2 takes the else branch (same as 0)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfLoadLegendDispatch_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadLegendDispatch{
		AckHandle: 22222,
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

func TestAirouStruct(t *testing.T) {
	cat := Airou{
		ID:          42,
		Name:        []byte("TestCat"),
		Task:        4,
		Personality: 2,
		Class:       1,
		Experience:  1500,
		WeaponType:  6,
		WeaponID:    100,
	}

	if cat.ID != 42 {
		t.Errorf("ID = %d, want 42", cat.ID)
	}
	if cat.Task != 4 {
		t.Errorf("Task = %d, want 4", cat.Task)
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
}

// TestHandleMsgMhfMercenaryHuntdata_Unk0_1 tests with Unk0=1 (returns 1 byte)
func TestHandleMsgMhfMercenaryHuntdata_Unk0_1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfMercenaryHuntdata{
		AckHandle:   12345,
		RequestType: 1,
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
		AckHandle:   12345,
		RequestType: 0,
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

func TestHandleMsgMhfLoadPartner(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadPartner{AckHandle: 1}
	handleMsgMhfLoadPartner(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSavePartner(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfSavePartner{AckHandle: 1, RawDataPayload: []byte{1, 2, 3, 4}}
	handleMsgMhfSavePartner(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfLoadHunterNavi_G8(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	server.erupeConfig.RealClientMode = cfg.G10
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadHunterNavi{AckHandle: 1}
	handleMsgMhfLoadHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfLoadHunterNavi_G7(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	server.erupeConfig.RealClientMode = cfg.G7
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadHunterNavi{AckHandle: 1}
	handleMsgMhfLoadHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveHunterNavi_NoDiff(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	data := make([]byte, 100)
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1,
		IsDataDiff:     false,
		RawDataPayload: data,
	}
	handleMsgMhfSaveHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveHunterNavi_Diff(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["hunternavi"] = make([]byte, 552)
	server.charRepo = charRepo
	server.erupeConfig.RealClientMode = cfg.G10
	session := createMockSession(100, server)

	// Create a valid diff payload (deltacomp format: pairs of offset+data)
	// A simple diff: zero length means no changes
	diffData := make([]byte, 4) // minimal diff
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1,
		IsDataDiff:     true,
		RawDataPayload: diffData,
	}
	handleMsgMhfSaveHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveHunterNavi_OversizedPayload(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	data := make([]byte, 5000) // > 4096
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1,
		IsDataDiff:     false,
		RawDataPayload: data,
	}
	handleMsgMhfSaveHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfCreateMercenary_Success(t *testing.T) {
	server := createMockServer()
	server.mercenaryRepo = &mockMercenaryRepo{nextRastaID: 42}
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfCreateMercenary{AckHandle: 1}
	handleMsgMhfCreateMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfCreateMercenary_Error(t *testing.T) {
	server := createMockServer()
	server.mercenaryRepo = &mockMercenaryRepo{rastaIDErr: errNotFound}
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfCreateMercenary{AckHandle: 1}
	handleMsgMhfCreateMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveMercenary_Normal(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	mercData := make([]byte, 100)
	// Write a uint32 index at the start
	mercData[0] = 0
	mercData[1] = 0
	mercData[2] = 0
	mercData[3] = 1
	pkt := &mhfpacket.MsgMhfSaveMercenary{
		AckHandle:  1,
		GCP:        500,
		PactMercID: 10,
		MercData:   mercData,
	}
	handleMsgMhfSaveMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveMercenary_Oversized(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfSaveMercenary{
		AckHandle: 1,
		MercData:  make([]byte, 70000),
	}
	handleMsgMhfSaveMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryM_EmptyData(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryM{AckHandle: 1, CharID: 200}
	handleMsgMhfReadMercenaryM(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryM_WithData(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["savemercenary"] = []byte{0x01, 0x02, 0x03, 0x04}
	server.charRepo = charRepo
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryM{AckHandle: 1, CharID: 100}
	handleMsgMhfReadMercenaryM(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfContractMercenary_Op0(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfContractMercenary{AckHandle: 1, Op: 0, CID: 200, PactMercID: 42}
	handleMsgMhfContractMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfContractMercenary_Op1(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfContractMercenary{AckHandle: 1, Op: 1}
	handleMsgMhfContractMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfContractMercenary_Op2(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfContractMercenary{AckHandle: 1, Op: 2, CID: 200}
	handleMsgMhfContractMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryW_NoPact(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	server.mercenaryRepo = &mockMercenaryRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryW{AckHandle: 1, Op: 0}
	handleMsgMhfReadMercenaryW(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryW_WithPact(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["pact_id"] = 42
	server.charRepo = charRepo
	server.mercenaryRepo = &mockMercenaryRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryW{AckHandle: 1, Op: 0}
	handleMsgMhfReadMercenaryW(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryW_Op2(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	server.mercenaryRepo = &mockMercenaryRepo{}
	session := createMockSession(100, server)

	// Op 2 skips loan enumeration
	pkt := &mhfpacket.MsgMhfReadMercenaryW{AckHandle: 1, Op: 2}
	handleMsgMhfReadMercenaryW(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfLoadOtomoAirou(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadOtomoAirou{AckHandle: 1}
	handleMsgMhfLoadOtomoAirou(session, pkt)
	<-session.sendPackets
}
