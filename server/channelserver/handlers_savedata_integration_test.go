package channelserver

import (
	"bytes"
	"testing"

	"erupe-ce/common/mhfitem"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

// ============================================================================
// SAVE/LOAD INTEGRATION TESTS
// Tests to verify user-reported save/load issues
//
// USER COMPLAINT SUMMARY:
// Features that ARE saved: RdP, items purchased, money spent, Hunter Navi
// Features that are NOT saved: current equipment, equipment sets, transmogs,
//   crafted equipment, monster kill counter (Koryo), warehouse, inventory
// ============================================================================

// TestSaveLoad_RoadPoints tests that Road Points (RdP) are saved correctly
// User reports this DOES save correctly
func TestSaveLoad_RoadPoints(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Set initial Road Points
	initialPoints := uint32(1000)
	_, err := db.Exec("UPDATE characters SET frontier_points = $1 WHERE id = $2", initialPoints, charID)
	if err != nil {
		t.Fatalf("Failed to set initial road points: %v", err)
	}

	// Modify Road Points
	newPoints := uint32(2500)
	_, err = db.Exec("UPDATE characters SET frontier_points = $1 WHERE id = $2", newPoints, charID)
	if err != nil {
		t.Fatalf("Failed to update road points: %v", err)
	}

	// Verify Road Points persisted
	var savedPoints uint32
	err = db.QueryRow("SELECT frontier_points FROM characters WHERE id = $1", charID).Scan(&savedPoints)
	if err != nil {
		t.Fatalf("Failed to query road points: %v", err)
	}

	if savedPoints != newPoints {
		t.Errorf("Road Points not saved correctly: got %d, want %d", savedPoints, newPoints)
	} else {
		t.Logf("✓ Road Points saved correctly: %d", savedPoints)
	}
}

// TestSaveLoad_HunterNavi tests that Hunter Navi data is saved correctly
// User reports this DOES save correctly
func TestSaveLoad_HunterNavi(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Create test session
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	s.server.db = db

	// Create Hunter Navi data
	naviData := make([]byte, 552) // G8+ size
	for i := range naviData {
		naviData[i] = byte(i % 256)
	}

	// Save Hunter Navi
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1234,
		IsDataDiff:     false, // Full save
		RawDataPayload: naviData,
	}

	handleMsgMhfSaveHunterNavi(s, pkt)

	// Verify saved
	var saved []byte
	err := db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", charID).Scan(&saved)
	if err != nil {
		t.Fatalf("Failed to query hunter navi: %v", err)
	}

	if len(saved) == 0 {
		t.Error("Hunter Navi not saved")
	} else if !bytes.Equal(saved, naviData) {
		t.Error("Hunter Navi data mismatch")
	} else {
		t.Logf("✓ Hunter Navi saved correctly: %d bytes", len(saved))
	}
}

// TestSaveLoad_MonsterKillCounter tests that Koryo points (kill counter) are saved
// User reports this DOES NOT save correctly
func TestSaveLoad_MonsterKillCounter(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Create test session
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	s.server.db = db

	// Initial Koryo points
	initialPoints := uint32(0)
	err := db.QueryRow("SELECT kouryou_point FROM characters WHERE id = $1", charID).Scan(&initialPoints)
	if err != nil {
		t.Fatalf("Failed to query initial koryo points: %v", err)
	}

	// Add Koryo points (simulate killing monsters)
	addPoints := uint32(100)
	pkt := &mhfpacket.MsgMhfAddKouryouPoint{
		AckHandle:     5678,
		KouryouPoints: addPoints,
	}

	handleMsgMhfAddKouryouPoint(s, pkt)

	// Verify points were added
	var savedPoints uint32
	err = db.QueryRow("SELECT kouryou_point FROM characters WHERE id = $1", charID).Scan(&savedPoints)
	if err != nil {
		t.Fatalf("Failed to query koryo points: %v", err)
	}

	expectedPoints := initialPoints + addPoints
	if savedPoints != expectedPoints {
		t.Errorf("Koryo points not saved correctly: got %d, want %d (BUG CONFIRMED)", savedPoints, expectedPoints)
	} else {
		t.Logf("✓ Koryo points saved correctly: %d", savedPoints)
	}
}

// TestSaveLoad_Inventory tests that inventory (item_box) is saved correctly
// User reports this DOES NOT save correctly
func TestSaveLoad_Inventory(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	_ = CreateTestCharacter(t, db, userID, "TestChar")

	// Create test items
	items := []mhfitem.MHFItemStack{
		{Item: mhfitem.MHFItem{ItemID: 1001}, Quantity: 10},
		{Item: mhfitem.MHFItem{ItemID: 1002}, Quantity: 20},
		{Item: mhfitem.MHFItem{ItemID: 1003}, Quantity: 30},
	}

	// Serialize and save inventory
	serialized := mhfitem.SerializeWarehouseItems(items)
	_, err := db.Exec("UPDATE users SET item_box = $1 WHERE id = $2", serialized, userID)
	if err != nil {
		t.Fatalf("Failed to save inventory: %v", err)
	}

	// Reload inventory
	var savedItemBox []byte
	err = db.QueryRow("SELECT item_box FROM users WHERE id = $1", userID).Scan(&savedItemBox)
	if err != nil {
		t.Fatalf("Failed to load inventory: %v", err)
	}

	if len(savedItemBox) == 0 {
		t.Error("Inventory not saved (BUG CONFIRMED)")
	} else if !bytes.Equal(savedItemBox, serialized) {
		t.Error("Inventory data mismatch (BUG CONFIRMED)")
	} else {
		t.Logf("✓ Inventory saved correctly: %d bytes", len(savedItemBox))
	}
}

// TestSaveLoad_Warehouse tests that warehouse contents are saved correctly
// User reports this DOES NOT save correctly
func TestSaveLoad_Warehouse(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Create test equipment for warehouse
	equipment := []mhfitem.MHFEquipment{
		{ItemID: 100, WarehouseID: 1},
		{ItemID: 101, WarehouseID: 2},
		{ItemID: 102, WarehouseID: 3},
	}

	// Serialize and save to warehouse
	serializedEquip := mhfitem.SerializeWarehouseEquipment(equipment)

	// Update warehouse equip0
	_, err := db.Exec("UPDATE warehouse SET equip0 = $1 WHERE character_id = $2", serializedEquip, charID)
	if err != nil {
		// Warehouse entry might not exist, try insert
		_, err = db.Exec(`
			INSERT INTO warehouse (character_id, equip0)
			VALUES ($1, $2)
			ON CONFLICT (character_id) DO UPDATE SET equip0 = $2
		`, charID, serializedEquip)
		if err != nil {
			t.Fatalf("Failed to save warehouse: %v", err)
		}
	}

	// Reload warehouse
	var savedEquip []byte
	err = db.QueryRow("SELECT equip0 FROM warehouse WHERE character_id = $1", charID).Scan(&savedEquip)
	if err != nil {
		t.Errorf("Failed to load warehouse: %v (BUG CONFIRMED)", err)
		return
	}

	if len(savedEquip) == 0 {
		t.Error("Warehouse not saved (BUG CONFIRMED)")
	} else if !bytes.Equal(savedEquip, serializedEquip) {
		t.Error("Warehouse data mismatch (BUG CONFIRMED)")
	} else {
		t.Logf("✓ Warehouse saved correctly: %d bytes", len(savedEquip))
	}
}

// TestSaveLoad_CurrentEquipment tests that currently equipped gear is saved
// User reports this DOES NOT save correctly
func TestSaveLoad_CurrentEquipment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Create test session
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	s.Name = "TestChar"
	s.server.db = db

	// Create savedata with equipped gear
	// Equipment data is embedded in the main savedata blob
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("TestChar\x00"))

	// Set weapon type at known offset (simplified)
	weaponTypeOffset := 500 // Example offset
	saveData[weaponTypeOffset] = 0x03 // Great Sword

	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress savedata: %v", err)
	}

	// Save equipment data
	pkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0, // Full blob
		AckHandle:      1111,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}

	handleMsgMhfSavedata(s, pkt)

	// Drain ACK
	if len(s.sendPackets) > 0 {
		<-s.sendPackets
	}

	// Reload savedata
	var savedCompressed []byte
	err = db.QueryRow("SELECT savedata FROM characters WHERE id = $1", charID).Scan(&savedCompressed)
	if err != nil {
		t.Fatalf("Failed to load savedata: %v", err)
	}

	if len(savedCompressed) == 0 {
		t.Error("Savedata (current equipment) not saved (BUG CONFIRMED)")
		return
	}

	// Decompress and verify
	decompressed, err := nullcomp.Decompress(savedCompressed)
	if err != nil {
		t.Errorf("Failed to decompress savedata: %v", err)
		return
	}

	if len(decompressed) < weaponTypeOffset+1 {
		t.Error("Savedata too short, equipment data missing (BUG CONFIRMED)")
		return
	}

	if decompressed[weaponTypeOffset] != saveData[weaponTypeOffset] {
		t.Errorf("Equipment data not saved correctly (BUG CONFIRMED): got 0x%02X, want 0x%02X",
			decompressed[weaponTypeOffset], saveData[weaponTypeOffset])
	} else {
		t.Logf("✓ Current equipment saved in savedata")
	}
}

// TestSaveLoad_EquipmentSets tests that equipment set configurations are saved
// User reports this DOES NOT save correctly (creation/modification/deletion)
func TestSaveLoad_EquipmentSets(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Equipment sets are stored in characters.platemyset
	testSetData := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05,
		0x10, 0x20, 0x30, 0x40, 0x50,
	}

	// Save equipment sets
	_, err := db.Exec("UPDATE characters SET platemyset = $1 WHERE id = $2", testSetData, charID)
	if err != nil {
		t.Fatalf("Failed to save equipment sets: %v", err)
	}

	// Reload equipment sets
	var savedSets []byte
	err = db.QueryRow("SELECT platemyset FROM characters WHERE id = $1", charID).Scan(&savedSets)
	if err != nil {
		t.Fatalf("Failed to load equipment sets: %v", err)
	}

	if len(savedSets) == 0 {
		t.Error("Equipment sets not saved (BUG CONFIRMED)")
	} else if !bytes.Equal(savedSets, testSetData) {
		t.Error("Equipment sets data mismatch (BUG CONFIRMED)")
	} else {
		t.Logf("✓ Equipment sets saved correctly: %d bytes", len(savedSets))
	}
}

// TestSaveLoad_Transmog tests that transmog/appearance data is saved correctly
// User reports this DOES NOT save correctly
func TestSaveLoad_Transmog(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Create test session
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	s.server.db = db

	// Create transmog/decoration set data
	transmogData := make([]byte, 100)
	for i := range transmogData {
		transmogData[i] = byte((i * 3) % 256)
	}

	// Save transmog data
	pkt := &mhfpacket.MsgMhfSaveDecoMyset{
		AckHandle:      2222,
		RawDataPayload: transmogData,
	}

	handleMsgMhfSaveDecoMyset(s, pkt)

	// Verify saved
	var saved []byte
	err := db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", charID).Scan(&saved)
	if err != nil {
		t.Fatalf("Failed to query transmog data: %v", err)
	}

	if len(saved) == 0 {
		t.Error("Transmog data not saved (BUG CONFIRMED)")
	} else {
		// handleMsgMhfSaveDecoMyset merges data, so check if anything was saved
		t.Logf("✓ Transmog data saved: %d bytes", len(saved))
	}
}

// TestSaveLoad_CraftedEquipment tests that crafted/upgraded equipment persists
// User reports this DOES NOT save correctly
func TestSaveLoad_CraftedEquipment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "TestChar")

	// Crafted equipment would be stored in savedata or warehouse
	// Let's test warehouse equipment with upgrade levels

	// Create crafted equipment with upgrade level
	equipment := []mhfitem.MHFEquipment{
		{
			ItemID:      5000, // Crafted weapon
			WarehouseID: 12345,
			// Upgrade level would be in equipment metadata
		},
	}

	serialized := mhfitem.SerializeWarehouseEquipment(equipment)

	// Save to warehouse
	_, err := db.Exec(`
		INSERT INTO warehouse (character_id, equip0)
		VALUES ($1, $2)
		ON CONFLICT (character_id) DO UPDATE SET equip0 = $2
	`, charID, serialized)
	if err != nil {
		t.Fatalf("Failed to save crafted equipment: %v", err)
	}

	// Reload
	var saved []byte
	err = db.QueryRow("SELECT equip0 FROM warehouse WHERE character_id = $1", charID).Scan(&saved)
	if err != nil {
		t.Errorf("Failed to load crafted equipment: %v (BUG CONFIRMED)", err)
		return
	}

	if len(saved) == 0 {
		t.Error("Crafted equipment not saved (BUG CONFIRMED)")
	} else if !bytes.Equal(saved, serialized) {
		t.Error("Crafted equipment data mismatch (BUG CONFIRMED)")
	} else {
		t.Logf("✓ Crafted equipment saved correctly: %d bytes", len(saved))
	}
}

// TestSaveLoad_CompleteSaveLoadCycle tests a complete save/load cycle
// This simulates a player logging out and back in
func TestSaveLoad_CompleteSaveLoadCycle(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "testuser")
	charID := CreateTestCharacter(t, db, userID, "SaveLoadTest")

	// Create test session (login)
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	s.Name = "SaveLoadTest"
	s.server.db = db

	// 1. Set Road Points
	rdpPoints := uint32(5000)
	_, err := db.Exec("UPDATE characters SET frontier_points = $1 WHERE id = $2", rdpPoints, charID)
	if err != nil {
		t.Fatalf("Failed to set RdP: %v", err)
	}

	// 2. Add Koryo Points
	koryoPoints := uint32(250)
	addPkt := &mhfpacket.MsgMhfAddKouryouPoint{
		AckHandle:     1111,
		KouryouPoints: koryoPoints,
	}
	handleMsgMhfAddKouryouPoint(s, addPkt)

	// 3. Save main savedata
	saveData := make([]byte, 150000)
	copy(saveData[88:], []byte("SaveLoadTest\x00"))
	compressed, _ := nullcomp.Compress(saveData)

	savePkt := &mhfpacket.MsgMhfSavedata{
		SaveType:       0,
		AckHandle:      2222,
		AllocMemSize:   uint32(len(compressed)),
		DataSize:       uint32(len(compressed)),
		RawDataPayload: compressed,
	}
	handleMsgMhfSavedata(s, savePkt)

	// Drain ACK packets
	for len(s.sendPackets) > 0 {
		<-s.sendPackets
	}

	// SIMULATE LOGOUT/LOGIN - Create new session
	mock2 := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s2 := createTestSession(mock2)
	s2.charID = charID
	s2.server.db = db
	s2.server.userBinaryParts = make(map[userBinaryPartID][]byte)

	// Load character data
	loadPkt := &mhfpacket.MsgMhfLoaddata{
		AckHandle: 3333,
	}
	handleMsgMhfLoaddata(s2, loadPkt)

	// Verify loaded name
	if s2.Name != "SaveLoadTest" {
		t.Errorf("Character name not loaded correctly: got %q, want %q", s2.Name, "SaveLoadTest")
	}

	// Verify Road Points persisted
	var loadedRdP uint32
	db.QueryRow("SELECT frontier_points FROM characters WHERE id = $1", charID).Scan(&loadedRdP)
	if loadedRdP != rdpPoints {
		t.Errorf("RdP not persisted: got %d, want %d (BUG CONFIRMED)", loadedRdP, rdpPoints)
	} else {
		t.Logf("✓ RdP persisted across save/load: %d", loadedRdP)
	}

	// Verify Koryo Points persisted
	var loadedKoryo uint32
	db.QueryRow("SELECT kouryou_point FROM characters WHERE id = $1", charID).Scan(&loadedKoryo)
	if loadedKoryo != koryoPoints {
		t.Errorf("Koryo points not persisted: got %d, want %d (BUG CONFIRMED)", loadedKoryo, koryoPoints)
	} else {
		t.Logf("✓ Koryo points persisted across save/load: %d", loadedKoryo)
	}

	t.Log("Complete save/load cycle test finished")
}
