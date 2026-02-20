package channelserver

import (
	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/common/mhfitem"
	"erupe-ce/common/token"
	"erupe-ce/network/mhfpacket"
	"testing"

	"github.com/jmoiron/sqlx"
)

// ackResponse holds parsed fields from a queued MsgSysAck packet.
type ackResponse struct {
	AckHandle        uint32
	IsBufferResponse bool
	ErrorCode        uint8
	PayloadSize      uint
	Payload          []byte
}

// readAck drains one packet from the session's sendPackets channel and
// parses the MsgSysAck wire format that QueueSendMHF produces.
func readAck(t *testing.T, session *Session) ackResponse {
	t.Helper()
	select {
	case p := <-session.sendPackets:
		bf := byteframe.NewByteFrameFromBytes(p.data)
		_ = bf.ReadUint16() // opcode
		ack := ackResponse{}
		ack.AckHandle = bf.ReadUint32()
		ack.IsBufferResponse = bf.ReadBool()
		ack.ErrorCode = bf.ReadUint8()
		size := uint(bf.ReadUint16())
		if size == 0xFFFF {
			size = uint(bf.ReadUint32())
		}
		ack.PayloadSize = size
		if ack.IsBufferResponse {
			ack.Payload = bf.ReadBytes(size)
		} else {
			ack.Payload = bf.ReadBytes(4)
		}
		return ack
	default:
		t.Fatal("No response packet queued")
		return ackResponse{}
	}
}

// setupHouseTest creates DB, server, session, and a character with user_binary row.
func setupHouseTest(t *testing.T) (*sqlx.DB, *Server, *Session, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	server := createMockServer()
	server.erupeConfig.RealClientMode = _config.ZZ
	SetTestDB(server, db)

	userID := CreateTestUser(t, db, "house_test_user")
	charID := CreateTestCharacter(t, db, userID, "HousePlayer")

	_, err := db.Exec(`INSERT INTO user_binary (id) VALUES ($1) ON CONFLICT DO NOTHING`, charID)
	if err != nil {
		t.Fatalf("Failed to create user_binary row: %v", err)
	}

	session := createMockSession(charID, server)
	return db, server, session, charID
}

// createTestEquipment creates properly initialized test equipment
func createTestEquipment(itemIDs []uint16, warehouseIDs []uint32) []mhfitem.MHFEquipment {
	var equip []mhfitem.MHFEquipment
	for i, itemID := range itemIDs {
		e := mhfitem.MHFEquipment{
			ItemID:      itemID,
			WarehouseID: warehouseIDs[i],
			Decorations: make([]mhfitem.MHFItem, 3),
			Sigils:      make([]mhfitem.MHFSigil, 3),
		}
		// Initialize Sigils Effects arrays
		for j := 0; j < 3; j++ {
			e.Sigils[j].Effects = make([]mhfitem.MHFSigilEffect, 3)
		}
		equip = append(equip, e)
	}
	return equip
}

// =============================================================================
// Unit Tests — guard paths, no database
// =============================================================================

func TestUpdateInterior_PayloadTooLarge(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateInterior{
		AckHandle:    1,
		InteriorData: make([]byte, 65), // > 64 triggers guard
	}
	handleMsgMhfUpdateInterior(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Errorf("expected success ACK (guard returns succeed), got error code %d", ack.ErrorCode)
	}
}

func TestUpdateMyhouseInfo_PayloadTooLarge(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateMyhouseInfo{
		AckHandle: 2,
		Data:      make([]byte, 513), // > 512 triggers guard
	}
	handleMsgMhfUpdateMyhouseInfo(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Errorf("expected success ACK on oversized payload, got error code %d", ack.ErrorCode)
	}
}

func TestSaveDecoMyset_PayloadTooShort(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSaveDecoMyset{
		AckHandle:      3,
		RawDataPayload: []byte{0x00, 0x01}, // < 3 bytes
	}
	handleMsgMhfSaveDecoMyset(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Errorf("expected success ACK on short payload, got error code %d", ack.ErrorCode)
	}
}

func TestUpdateWarehouse_BoxIndexTooHigh(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateWarehouse{
		AckHandle: 4,
		BoxIndex:  11, // > 10 triggers fail
	}
	handleMsgMhfUpdateWarehouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 1 {
		t.Errorf("expected fail ACK for out-of-bounds box index, got error code %d", ack.ErrorCode)
	}
}

func TestEnumerateHouse_Method5_EmptyResult(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = _config.ZZ
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{
		AckHandle: 5,
		Method:    5, // Recent visitors — always returns empty
	}
	handleMsgMhfEnumerateHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	if !ack.IsBufferResponse {
		t.Fatal("expected buffer response")
	}
	// First 2 bytes = count, should be 0
	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	count := bf.ReadUint16()
	if count != 0 {
		t.Errorf("expected 0 houses for method 5, got %d", count)
	}
}

func TestResetTitle_NoOp(t *testing.T) {
	// handleMsgMhfResetTitle is an empty function — just verify no panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfResetTitle panicked: %v", r)
		}
	}()
	handleMsgMhfResetTitle(nil, nil)
}

func TestOperateWarehouse_RenameBoxIndexTooHigh(t *testing.T) {
	// Operation 2 = Rename. BoxIndex > 9 should skip the rename.
	// This needs a DB for initializeWarehouse, so the full test is the
	// integration test TestOperateWarehouse_Op2_RenameBoxIndexTooHigh below.
}

// =============================================================================
// Integration Tests — real PostgreSQL via SetupTestDB
// =============================================================================

func TestUpdateInterior_SavesData(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	interiorData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}
	pkt := &mhfpacket.MsgMhfUpdateInterior{
		AckHandle:    10,
		InteriorData: interiorData,
	}
	handleMsgMhfUpdateInterior(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}

	// Verify data was persisted
	_, _, furniture, _, _, _, _, err := session.server.houseRepo.GetHouseContents(charID)
	if err != nil {
		t.Fatalf("GetHouseContents failed: %v", err)
	}
	if len(furniture) < len(interiorData) {
		t.Fatalf("furniture data too short: got %d bytes", len(furniture))
	}
	for i, b := range interiorData {
		if furniture[i] != b {
			t.Errorf("furniture[%d] = %#x, want %#x", i, furniture[i], b)
		}
	}
}

func TestUpdateHouse_SetsStateAndPassword(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfUpdateHouse{
		AckHandle: 11,
		State:     3,
		Password:  "secret",
	}
	handleMsgMhfUpdateHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}

	state, password, err := session.server.houseRepo.GetHouseAccess(charID)
	if err != nil {
		t.Fatalf("GetHouseAccess failed: %v", err)
	}
	if state != 3 {
		t.Errorf("state = %d, want 3", state)
	}
	if password != "secret" {
		t.Errorf("password = %q, want %q", password, "secret")
	}
}

func TestEnumerateHouse_Method4_ByCharID(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{
		AckHandle: 12,
		Method:    4,
		CharID:    charID,
	}
	handleMsgMhfEnumerateHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	count := bf.ReadUint16()
	if count != 1 {
		t.Errorf("expected 1 house for charID lookup, got %d", count)
	}
}

func TestEnumerateHouse_Method3_ByName(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{
		AckHandle: 13,
		Method:    3,
		Name:      "HousePlayer",
	}
	handleMsgMhfEnumerateHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	count := bf.ReadUint16()
	if count < 1 {
		t.Errorf("expected at least 1 house for name search, got %d", count)
	}
}

func TestLoadHouse_OwnHouse_Destination9(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	// Set some interior data first
	interior := make([]byte, 20)
	interior[0] = 0xAB
	_ = session.server.houseRepo.UpdateInterior(charID, interior)

	pkt := &mhfpacket.MsgMhfLoadHouse{
		AckHandle:   14,
		CharID:      charID,
		Destination: 9, // Own house — bypasses access control
	}
	handleMsgMhfLoadHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success loading own house, got error code %d", ack.ErrorCode)
	}
	if !ack.IsBufferResponse {
		t.Fatal("expected buffer response")
	}
	if len(ack.Payload) == 0 {
		t.Error("expected non-empty house data")
	}
}

func TestLoadHouse_WrongPassword_Fails(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	// Set a password on the house
	_ = session.server.houseRepo.UpdateHouseState(charID, 2, "correct")

	pkt := &mhfpacket.MsgMhfLoadHouse{
		AckHandle:   15,
		CharID:      charID,
		Destination: 3, // Others house
		CheckPass:   true,
		Password:    "wrong",
	}
	handleMsgMhfLoadHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 1 {
		t.Errorf("expected fail ACK for wrong password, got error code %d", ack.ErrorCode)
	}
}

func TestLoadHouse_CorrectPassword_Succeeds(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	_ = session.server.houseRepo.UpdateHouseState(charID, 2, "correct")

	pkt := &mhfpacket.MsgMhfLoadHouse{
		AckHandle:   16,
		CharID:      charID,
		Destination: 3,
		CheckPass:   true,
		Password:    "correct",
	}
	handleMsgMhfLoadHouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Errorf("expected success for correct password, got error code %d", ack.ErrorCode)
	}
	if !ack.IsBufferResponse {
		t.Fatal("expected buffer response for house data")
	}
}

func TestGetMyhouseInfo_NoData(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfGetMyhouseInfo{AckHandle: 17}
	handleMsgMhfGetMyhouseInfo(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	// When no mission data exists, handler returns 9-byte default
	if len(ack.Payload) != 9 {
		t.Errorf("expected 9-byte default payload, got %d bytes", len(ack.Payload))
	}
}

func TestGetMyhouseInfo_WithData(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	missionData := make([]byte, 50)
	missionData[0] = 0xDE
	missionData[1] = 0xAD
	_ = session.server.houseRepo.UpdateMission(charID, missionData)

	pkt := &mhfpacket.MsgMhfGetMyhouseInfo{AckHandle: 18}
	handleMsgMhfGetMyhouseInfo(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	if len(ack.Payload) != 50 {
		t.Fatalf("expected 50-byte payload, got %d bytes", len(ack.Payload))
	}
	if ack.Payload[0] != 0xDE || ack.Payload[1] != 0xAD {
		t.Errorf("payload mismatch: got %#x %#x, want 0xDE 0xAD", ack.Payload[0], ack.Payload[1])
	}
}

func TestUpdateMyhouseInfo_SavesData(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	missionData := make([]byte, 100)
	missionData[0] = 0xCA
	missionData[1] = 0xFE

	pkt := &mhfpacket.MsgMhfUpdateMyhouseInfo{
		AckHandle: 19,
		Data:      missionData,
	}
	handleMsgMhfUpdateMyhouseInfo(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}

	// Verify via repository
	data, err := session.server.houseRepo.GetMission(charID)
	if err != nil {
		t.Fatalf("GetMission failed: %v", err)
	}
	if len(data) != 100 {
		t.Fatalf("mission data length = %d, want 100", len(data))
	}
	if data[0] != 0xCA || data[1] != 0xFE {
		t.Errorf("mission data mismatch: got %#x %#x, want 0xCA 0xFE", data[0], data[1])
	}
}

func TestEnumerateTitle_Empty(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfEnumerateTitle{AckHandle: 20}
	handleMsgMhfEnumerateTitle(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	count := bf.ReadUint16()
	if count != 0 {
		t.Errorf("expected 0 titles, got %d", count)
	}
}

func TestAcquireTitle_AndEnumerate(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	// Acquire two titles
	acquirePkt := &mhfpacket.MsgMhfAcquireTitle{
		AckHandle: 21,
		TitleIDs:  []uint16{100, 200},
	}
	handleMsgMhfAcquireTitle(session, acquirePkt)
	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("acquire failed: error code %d", ack.ErrorCode)
	}

	// Enumerate
	enumPkt := &mhfpacket.MsgMhfEnumerateTitle{AckHandle: 22}
	handleMsgMhfEnumerateTitle(session, enumPkt)
	ack = readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("enumerate failed: error code %d", ack.ErrorCode)
	}

	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	count := bf.ReadUint16()
	if count != 2 {
		t.Errorf("expected 2 titles, got %d", count)
	}

	// Read title IDs
	_ = bf.ReadUint16() // unk
	ids := make(map[uint16]bool)
	for i := 0; i < int(count); i++ {
		id := bf.ReadUint16()
		ids[id] = true
		_ = bf.ReadUint16() // unk
		_ = bf.ReadUint32() // acquired timestamp
		_ = bf.ReadUint32() // updated timestamp
	}
	if !ids[100] || !ids[200] {
		t.Errorf("expected title IDs 100 and 200, got %v", ids)
	}
}

func TestAcquireTitle_Duplicate(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	// Acquire title 300
	pkt1 := &mhfpacket.MsgMhfAcquireTitle{AckHandle: 23, TitleIDs: []uint16{300}}
	handleMsgMhfAcquireTitle(session, pkt1)
	_ = readAck(t, session)

	// Acquire same title again
	pkt2 := &mhfpacket.MsgMhfAcquireTitle{AckHandle: 24, TitleIDs: []uint16{300}}
	handleMsgMhfAcquireTitle(session, pkt2)
	_ = readAck(t, session)

	// Should still have exactly 1 title (upsert)
	titles, err := session.server.houseRepo.GetTitles(charID)
	if err != nil {
		t.Fatalf("GetTitles failed: %v", err)
	}
	if len(titles) != 1 {
		t.Errorf("expected 1 title after duplicate acquire, got %d", len(titles))
	}
}

func TestOperateWarehouse_Op0_GetBoxNames(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	// Initialize warehouse and rename a box
	session.server.houseRepo.InitializeWarehouse(charID)
	_ = session.server.houseRepo.RenameWarehouseBox(charID, 0, 0, "MyItems")

	pkt := &mhfpacket.MsgMhfOperateWarehouse{
		AckHandle: 25,
		Operation: 0,
	}
	handleMsgMhfOperateWarehouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	if !ack.IsBufferResponse {
		t.Fatal("expected buffer response")
	}
	// Response format: op(1) + renewal(4) + usages(2) + count(1) + entries
	if len(ack.Payload) < 8 {
		t.Fatalf("payload too short: %d bytes", len(ack.Payload))
	}
	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	op := bf.ReadUint8()
	if op != 0 {
		t.Errorf("op = %d, want 0", op)
	}
}

func TestOperateWarehouse_Op3_GetUsageLimit(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{
		AckHandle: 26,
		Operation: 3,
	}
	handleMsgMhfOperateWarehouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	// Response: op(1) + renewal_time(4) + usages(2) = 7 bytes
	bf := byteframe.NewByteFrameFromBytes(ack.Payload)
	op := bf.ReadUint8()
	if op != 3 {
		t.Errorf("op = %d, want 3", op)
	}
	renewalTime := bf.ReadUint32()
	usages := bf.ReadUint16()
	if renewalTime != 0 {
		t.Errorf("renewal time = %d, want 0", renewalTime)
	}
	if usages != 10000 {
		t.Errorf("usages = %d, want 10000", usages)
	}
}

func TestOperateWarehouse_Op2_RenameBoxIndexTooHigh(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{
		AckHandle: 27,
		Operation: 2,
		BoxIndex:  10, // > 9, rename should be skipped
		Name:      "ShouldNotRename",
	}
	handleMsgMhfOperateWarehouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success ACK even with skipped rename, got error code %d", ack.ErrorCode)
	}
}

func TestEnumerateWarehouse_EmptyBox(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfEnumerateWarehouse{
		AckHandle: 28,
		BoxType:   0, // Items
		BoxIndex:  0,
	}
	handleMsgMhfEnumerateWarehouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	if !ack.IsBufferResponse {
		t.Fatal("expected buffer response")
	}
	// Empty box returns serialized empty list: count(2) + unk(2) = 4 bytes minimum
	if len(ack.Payload) < 4 {
		t.Errorf("expected at least 4-byte payload for empty box, got %d", len(ack.Payload))
	}
}

func TestUpdateWarehouse_Items(t *testing.T) {
	_, _, session, charID := setupHouseTest(t)

	items := []mhfitem.MHFItemStack{
		{Item: mhfitem.MHFItem{ItemID: 42}, Quantity: 10, WarehouseID: token.RNG.Uint32()},
		{Item: mhfitem.MHFItem{ItemID: 99}, Quantity: 5, WarehouseID: token.RNG.Uint32()},
	}
	pkt := &mhfpacket.MsgMhfUpdateWarehouse{
		AckHandle:    29,
		BoxType:      0,
		BoxIndex:     0,
		UpdatedItems: items,
	}
	handleMsgMhfUpdateWarehouse(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}

	// Read back via enumerate
	session2 := createMockSession(charID, session.server)
	enumPkt := &mhfpacket.MsgMhfEnumerateWarehouse{
		AckHandle: 30,
		BoxType:   0,
		BoxIndex:  0,
	}
	handleMsgMhfEnumerateWarehouse(session2, enumPkt)

	ack2 := readAck(t, session2)
	if ack2.ErrorCode != 0 {
		t.Fatalf("enumerate failed: error code %d", ack2.ErrorCode)
	}
	// Parse the serialized items
	bf := byteframe.NewByteFrameFromBytes(ack2.Payload)
	count := bf.ReadUint16()
	if count != 2 {
		t.Errorf("expected 2 items in warehouse, got %d", count)
	}
}

func TestLoadDecoMyset_Default(t *testing.T) {
	_, _, session, _ := setupHouseTest(t)

	pkt := &mhfpacket.MsgMhfLoadDecoMyset{AckHandle: 31}
	handleMsgMhfLoadDecoMyset(session, pkt)

	ack := readAck(t, session)
	if ack.ErrorCode != 0 {
		t.Fatalf("expected success, got error code %d", ack.ErrorCode)
	}
	if !ack.IsBufferResponse {
		t.Fatal("expected buffer response")
	}
	// G10+ mode returns {0x01, 0x00}
	if len(ack.Payload) < 2 {
		t.Fatalf("expected at least 2-byte payload, got %d", len(ack.Payload))
	}
	if ack.Payload[0] != 0x01 || ack.Payload[1] != 0x00 {
		t.Errorf("expected default {0x01, 0x00}, got {%#x, %#x}", ack.Payload[0], ack.Payload[1])
	}
}

// =============================================================================
// Existing pure-logic tests and benchmarks (unchanged)
// =============================================================================

// TestWarehouseItemSerialization verifies warehouse item serialization
func TestWarehouseItemSerialization(t *testing.T) {
	tests := []struct {
		name  string
		items []mhfitem.MHFItemStack
	}{
		{
			name: "empty_warehouse",
			items: []mhfitem.MHFItemStack{},
		},
		{
			name: "single_item",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
			},
		},
		{
			name: "multiple_items",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 20},
				{Item: mhfitem.MHFItem{ItemID: 3}, Quantity: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized := mhfitem.SerializeWarehouseItems(tt.items)

			// Basic validation
			if serialized == nil {
				t.Error("serialization returned nil")
			}

			// Verify we can work with the serialized data
			if serialized == nil {
				t.Error("invalid serialized length")
			}
		})
	}
}

// TestWarehouseEquipmentSerialization verifies warehouse equipment serialization
func TestWarehouseEquipmentSerialization(t *testing.T) {
	tests := []struct {
		name      string
		equipment []mhfitem.MHFEquipment
	}{
		{
			name:      "empty_equipment",
			equipment: []mhfitem.MHFEquipment{},
		},
		{
			name: "single_equipment",
			equipment: createTestEquipment([]uint16{100}, []uint32{1}),
		},
		{
			name: "multiple_equipment",
			equipment: createTestEquipment([]uint16{100, 101, 102}, []uint32{1, 2, 3}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized := mhfitem.SerializeWarehouseEquipment(tt.equipment, _config.ZZ)

			// Basic validation
			if serialized == nil {
				t.Error("serialization returned nil")
			}

			// Verify we can work with the serialized data
			if serialized == nil {
				t.Error("invalid serialized length")
			}
		})
	}
}

// TestWarehouseItemDiff verifies the item diff calculation
func TestWarehouseItemDiff(t *testing.T) {
	tests := []struct {
		name     string
		oldItems []mhfitem.MHFItemStack
		newItems []mhfitem.MHFItemStack
		wantDiff bool
	}{
		{
			name:     "no_changes",
			oldItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			newItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			wantDiff: false,
		},
		{
			name:     "quantity_changed",
			oldItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			newItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 15}},
			wantDiff: true,
		},
		{
			name:     "item_added",
			oldItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			newItems: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 5},
			},
			wantDiff: true,
		},
		{
			name: "item_removed",
			oldItems: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 5},
			},
			newItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := mhfitem.DiffItemStacks(tt.oldItems, tt.newItems)

			// Verify that diff returns a valid result (not nil)
			if diff == nil {
				t.Error("diff should not be nil")
			}

			// The diff function returns items where Quantity > 0
			// So with no changes (all same quantity), diff should have same items
			if tt.name == "no_changes" {
				if len(diff) == 0 {
					t.Error("no_changes should return items")
				}
			}
		})
	}
}

// TestWarehouseEquipmentMerge verifies equipment merging logic
func TestWarehouseEquipmentMerge(t *testing.T) {
	tests := []struct {
		name        string
		oldEquip    []mhfitem.MHFEquipment
		newEquip    []mhfitem.MHFEquipment
		wantMerged  int
	}{
		{
			name:        "merge_empty",
			oldEquip:    []mhfitem.MHFEquipment{},
			newEquip:    []mhfitem.MHFEquipment{},
			wantMerged:  0,
		},
		{
			name: "add_new_equipment",
			oldEquip: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
			},
			newEquip: []mhfitem.MHFEquipment{
				{ItemID: 101, WarehouseID: 0}, // New item, no warehouse ID yet
			},
			wantMerged: 2, // Old + new
		},
		{
			name: "update_existing_equipment",
			oldEquip: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
			},
			newEquip: []mhfitem.MHFEquipment{
				{ItemID: 101, WarehouseID: 1}, // Update existing
			},
			wantMerged: 1, // Updated in place
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the merge logic from handleMsgMhfUpdateWarehouse
			var finalEquip []mhfitem.MHFEquipment
			oEquips := tt.oldEquip

			for _, uEquip := range tt.newEquip {
				exists := false
				for i := range oEquips {
					if oEquips[i].WarehouseID == uEquip.WarehouseID && uEquip.WarehouseID != 0 {
						exists = true
						oEquips[i].ItemID = uEquip.ItemID
						break
					}
				}
				if !exists {
					// Generate new warehouse ID
					uEquip.WarehouseID = token.RNG.Uint32()
					finalEquip = append(finalEquip, uEquip)
				}
			}

			for _, oEquip := range oEquips {
				if oEquip.ItemID > 0 {
					finalEquip = append(finalEquip, oEquip)
				}
			}

			// Verify merge result count
			if len(finalEquip) != tt.wantMerged {
				t.Errorf("expected %d merged equipment, got %d", tt.wantMerged, len(finalEquip))
			}
		})
	}
}

// TestWarehouseIDGeneration verifies warehouse ID uniqueness
func TestWarehouseIDGeneration(t *testing.T) {
	// Generate multiple warehouse IDs and verify they're unique
	idCount := 100
	ids := make(map[uint32]bool)

	for i := 0; i < idCount; i++ {
		id := token.RNG.Uint32()
		if id == 0 {
			t.Error("generated warehouse ID is 0 (invalid)")
		}
		if ids[id] {
			// While collisions are possible with random IDs,
			// they should be extremely rare
			t.Logf("Warning: duplicate warehouse ID generated: %d", id)
		}
		ids[id] = true
	}

	if len(ids) < idCount*90/100 {
		t.Errorf("too many duplicate IDs: got %d unique out of %d", len(ids), idCount)
	}
}

// TestWarehouseItemRemoval verifies item removal logic
func TestWarehouseItemRemoval(t *testing.T) {
	tests := []struct {
		name       string
		items      []mhfitem.MHFItemStack
		removeID   uint16
		wantRemain int
	}{
		{
			name: "remove_existing",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 20},
			},
			removeID:   1,
			wantRemain: 1,
		},
		{
			name: "remove_non_existing",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
			},
			removeID:   999,
			wantRemain: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var remaining []mhfitem.MHFItemStack
			for _, item := range tt.items {
				if item.Item.ItemID != tt.removeID {
					remaining = append(remaining, item)
				}
			}

			if len(remaining) != tt.wantRemain {
				t.Errorf("expected %d remaining items, got %d", tt.wantRemain, len(remaining))
			}
		})
	}
}

// TestWarehouseEquipmentRemoval verifies equipment removal logic
func TestWarehouseEquipmentRemoval(t *testing.T) {
	tests := []struct {
		name       string
		equipment  []mhfitem.MHFEquipment
		setZeroID  uint32
		wantActive int
	}{
		{
			name: "remove_by_setting_zero",
			equipment: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
				{ItemID: 101, WarehouseID: 2},
			},
			setZeroID:  1,
			wantActive: 1,
		},
		{
			name: "all_active",
			equipment: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
				{ItemID: 101, WarehouseID: 2},
			},
			setZeroID:  999,
			wantActive: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate removal by setting ItemID to 0
			equipment := make([]mhfitem.MHFEquipment, len(tt.equipment))
			copy(equipment, tt.equipment)

			for i := range equipment {
				if equipment[i].WarehouseID == tt.setZeroID {
					equipment[i].ItemID = 0
				}
			}

			// Count active equipment (ItemID > 0)
			activeCount := 0
			for _, eq := range equipment {
				if eq.ItemID > 0 {
					activeCount++
				}
			}

			if activeCount != tt.wantActive {
				t.Errorf("expected %d active equipment, got %d", tt.wantActive, activeCount)
			}
		})
	}
}

// TestWarehouseBoxIndexValidation verifies box index bounds
func TestWarehouseBoxIndexValidation(t *testing.T) {
	tests := []struct {
		name     string
		boxIndex uint8
		isValid  bool
	}{
		{
			name:     "box_0",
			boxIndex: 0,
			isValid:  true,
		},
		{
			name:     "box_1",
			boxIndex: 1,
			isValid:  true,
		},
		{
			name:     "box_9",
			boxIndex: 9,
			isValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify box index is within reasonable bounds
			if tt.isValid && tt.boxIndex > 100 {
				t.Error("box index unreasonably high")
			}
		})
	}
}

// TestWarehouseErrorRecovery verifies error handling doesn't corrupt state
func TestWarehouseErrorRecovery(t *testing.T) {
	t.Run("database_error_handling", func(t *testing.T) {
		// After our fix, database errors should:
		// 1. Be logged with s.logger.Error()
		// 2. Send doAckSimpleFail()
		// 3. Return immediately
		// 4. NOT send doAckSimpleSucceed() (the bug we fixed)

		// This test documents the expected behavior
	})

	t.Run("serialization_error_handling", func(t *testing.T) {
		// Test that serialization errors are handled gracefully
		emptyItems := []mhfitem.MHFItemStack{}
		serialized := mhfitem.SerializeWarehouseItems(emptyItems)

		// Should handle empty gracefully
		if serialized == nil {
			t.Error("serialization of empty items should not return nil")
		}
	})
}

// BenchmarkWarehouseSerialization benchmarks warehouse serialization performance
func BenchmarkWarehouseSerialization(b *testing.B) {
	items := []mhfitem.MHFItemStack{
		{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
		{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 20},
		{Item: mhfitem.MHFItem{ItemID: 3}, Quantity: 30},
		{Item: mhfitem.MHFItem{ItemID: 4}, Quantity: 40},
		{Item: mhfitem.MHFItem{ItemID: 5}, Quantity: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mhfitem.SerializeWarehouseItems(items)
	}
}

// BenchmarkWarehouseEquipmentMerge benchmarks equipment merge performance
func BenchmarkWarehouseEquipmentMerge(b *testing.B) {
	oldEquip := make([]mhfitem.MHFEquipment, 50)
	for i := range oldEquip {
		oldEquip[i] = mhfitem.MHFEquipment{
			ItemID:      uint16(100 + i),
			WarehouseID: uint32(i + 1),
		}
	}

	newEquip := make([]mhfitem.MHFEquipment, 10)
	for i := range newEquip {
		newEquip[i] = mhfitem.MHFEquipment{
			ItemID:      uint16(200 + i),
			WarehouseID: uint32(i + 1),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var finalEquip []mhfitem.MHFEquipment
		oEquips := oldEquip

		for _, uEquip := range newEquip {
			exists := false
			for j := range oEquips {
				if oEquips[j].WarehouseID == uEquip.WarehouseID {
					exists = true
					oEquips[j].ItemID = uEquip.ItemID
					break
				}
			}
			if !exists {
				finalEquip = append(finalEquip, uEquip)
			}
		}

		for _, oEquip := range oEquips {
			if oEquip.ItemID > 0 {
				finalEquip = append(finalEquip, oEquip)
			}
		}
		_ = finalEquip // Use finalEquip to avoid unused variable warning
	}
}
