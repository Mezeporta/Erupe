package channelserver

import (
	"encoding/binary"
	"testing"
)

func TestCharacterSaveDataStruct(t *testing.T) {
	saveData := &CharacterSaveData{
		CharID:         12345,
		Name:           "TestHunter",
		IsNewCharacter: false,
		Gender:         true,
		RP:             1000,
		WeaponType:     5,
		WeaponID:       100,
		HRP:            500,
		GR:             50,
	}

	if saveData.CharID != 12345 {
		t.Errorf("CharID = %d, want 12345", saveData.CharID)
	}
	if saveData.Name != "TestHunter" {
		t.Errorf("Name = %s, want TestHunter", saveData.Name)
	}
	if saveData.Gender != true {
		t.Error("Gender should be true")
	}
	if saveData.RP != 1000 {
		t.Errorf("RP = %d, want 1000", saveData.RP)
	}
	if saveData.WeaponType != 5 {
		t.Errorf("WeaponType = %d, want 5", saveData.WeaponType)
	}
	if saveData.HRP != 500 {
		t.Errorf("HRP = %d, want 500", saveData.HRP)
	}
	if saveData.GR != 50 {
		t.Errorf("GR = %d, want 50", saveData.GR)
	}
}

func TestCharacterSaveData_InitialValues(t *testing.T) {
	saveData := &CharacterSaveData{}

	if saveData.CharID != 0 {
		t.Errorf("CharID should default to 0, got %d", saveData.CharID)
	}
	if saveData.IsNewCharacter != false {
		t.Error("IsNewCharacter should default to false")
	}
	if saveData.Gender != false {
		t.Error("Gender should default to false")
	}
}

func TestCharacterSaveData_BinarySlices(t *testing.T) {
	saveData := &CharacterSaveData{
		HouseTier:     make([]byte, 5),
		HouseData:     make([]byte, 195),
		BookshelfData: make([]byte, 5576),
		GalleryData:   make([]byte, 1748),
		ToreData:      make([]byte, 240),
		GardenData:    make([]byte, 68),
		KQF:           make([]byte, 8),
	}

	// Verify slice sizes match expected game data sizes
	if len(saveData.HouseTier) != 5 {
		t.Errorf("HouseTier len = %d, want 5", len(saveData.HouseTier))
	}
	if len(saveData.HouseData) != 195 {
		t.Errorf("HouseData len = %d, want 195", len(saveData.HouseData))
	}
	if len(saveData.BookshelfData) != 5576 {
		t.Errorf("BookshelfData len = %d, want 5576", len(saveData.BookshelfData))
	}
	if len(saveData.GalleryData) != 1748 {
		t.Errorf("GalleryData len = %d, want 1748", len(saveData.GalleryData))
	}
	if len(saveData.ToreData) != 240 {
		t.Errorf("ToreData len = %d, want 240", len(saveData.ToreData))
	}
	if len(saveData.GardenData) != 68 {
		t.Errorf("GardenData len = %d, want 68", len(saveData.GardenData))
	}
	if len(saveData.KQF) != 8 {
		t.Errorf("KQF len = %d, want 8", len(saveData.KQF))
	}
}

func TestPointerConstants(t *testing.T) {
	// Verify the pointer constants are set correctly based on the game's save format
	pointers := map[string]int{
		"pointerGender":        0x81,
		"pointerRP":            0x22D16,
		"pointerHouseTier":     0x1FB6C,
		"pointerHouseData":     0x1FE01,
		"pointerBookshelfData": 0x22298,
		"pointerGalleryData":   0x22320,
		"pointerToreData":      0x1FCB4,
		"pointerGardenData":    0x22C58,
		"pointerWeaponType":    0x1F715,
		"pointerWeaponID":      0x1F60A,
		"pointerHRP":           0x1FDF6,
		"pointerGRP":           0x1FDFC,
		"pointerKQF":           0x23D20,
	}

	// Verify constants are properly defined (non-zero and in expected ranges)
	if pointerGender != 0x81 {
		t.Errorf("pointerGender = 0x%X, want 0x81", pointerGender)
	}
	if pointerRP != 0x22D16 {
		t.Errorf("pointerRP = 0x%X, want 0x22D16", pointerRP)
	}
	if pointerKQF != 0x23D20 {
		t.Errorf("pointerKQF = 0x%X, want 0x23D20", pointerKQF)
	}

	// Verify pointers are all unique
	seen := make(map[int]string)
	for name, ptr := range pointers {
		if existingName, ok := seen[ptr]; ok {
			t.Errorf("Duplicate pointer value 0x%X: %s and %s", ptr, name, existingName)
		}
		seen[ptr] = name
	}
}

func TestCharacterSaveData_UpdateSaveDataWithStruct(t *testing.T) {
	// Create a save with enough data to hold all pointers
	// Maximum pointer is pointerKQF at 0x23D20 + 8 = 0x23D28
	saveSize := 0x23D30 // A bit more than needed
	saveData := &CharacterSaveData{
		decompSave: make([]byte, saveSize),
		RP:         1234,
		KQF:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	saveData.updateSaveDataWithStruct()

	// Check RP was written correctly (little endian)
	rpValue := binary.LittleEndian.Uint16(saveData.decompSave[pointerRP : pointerRP+2])
	if rpValue != 1234 {
		t.Errorf("RP in decompSave = %d, want 1234", rpValue)
	}

	// Check KQF was written correctly
	for i := 0; i < 8; i++ {
		if saveData.decompSave[pointerKQF+i] != byte(i+1) {
			t.Errorf("KQF[%d] = 0x%02X, want 0x%02X", i, saveData.decompSave[pointerKQF+i], i+1)
		}
	}
}

func TestCharacterSaveData_UpdateStructWithSaveData_Gender(t *testing.T) {
	// Create minimal save data for gender test
	saveSize := 0x23D30
	saveData := &CharacterSaveData{
		decompSave:     make([]byte, saveSize),
		IsNewCharacter: true, // New char doesn't read most fields
	}

	// Set gender to male (0)
	saveData.decompSave[pointerGender] = 0
	saveData.updateStructWithSaveData()

	if saveData.Gender != false {
		t.Error("Gender should be false (male) when byte is 0")
	}

	// Set gender to female (1)
	saveData.decompSave[pointerGender] = 1
	saveData.updateStructWithSaveData()

	if saveData.Gender != true {
		t.Error("Gender should be true (female) when byte is 1")
	}
}

func TestCharacterSaveData_NotNewCharacter(t *testing.T) {
	// Create save data for existing character
	saveSize := 0x23D30
	saveData := &CharacterSaveData{
		decompSave:     make([]byte, saveSize),
		IsNewCharacter: false,
	}

	// Set some values in the save data
	binary.LittleEndian.PutUint16(saveData.decompSave[pointerRP:], 5000)
	binary.LittleEndian.PutUint16(saveData.decompSave[pointerHRP:], 500)
	saveData.decompSave[pointerWeaponType] = 7

	saveData.updateStructWithSaveData()

	if saveData.RP != 5000 {
		t.Errorf("RP = %d, want 5000", saveData.RP)
	}
	if saveData.HRP != 500 {
		t.Errorf("HRP = %d, want 500", saveData.HRP)
	}
	if saveData.WeaponType != 7 {
		t.Errorf("WeaponType = %d, want 7", saveData.WeaponType)
	}
}

func TestCharacterSaveData_GR_MaxHRP(t *testing.T) {
	// When HRP is 999, GR is calculated from GRP
	saveSize := 0x23D30
	saveData := &CharacterSaveData{
		decompSave:     make([]byte, saveSize),
		IsNewCharacter: false,
	}

	// Set HRP to 999 (max HR)
	binary.LittleEndian.PutUint16(saveData.decompSave[pointerHRP:], 999)
	// Set GRP to 593400 (GR 100)
	binary.LittleEndian.PutUint32(saveData.decompSave[pointerGRP:], 593400)

	saveData.updateStructWithSaveData()

	if saveData.HRP != 999 {
		t.Errorf("HRP = %d, want 999", saveData.HRP)
	}
	// GR should be calculated via grpToGR
	expectedGR := grpToGR(593400)
	if saveData.GR != expectedGR {
		t.Errorf("GR = %d, want %d", saveData.GR, expectedGR)
	}
}

func TestCharacterSaveData_SliceExtraction(t *testing.T) {
	// Test that slices are extracted at correct offsets
	saveSize := 0x23D30
	saveData := &CharacterSaveData{
		decompSave:     make([]byte, saveSize),
		IsNewCharacter: false,
	}

	// Fill specific regions with identifiable patterns
	for i := 0; i < 5; i++ {
		saveData.decompSave[pointerHouseTier+i] = byte(0xAA)
	}
	for i := 0; i < 195; i++ {
		saveData.decompSave[pointerHouseData+i] = byte(0xBB)
	}
	for i := 0; i < 8; i++ {
		saveData.decompSave[pointerKQF+i] = byte(0xCC)
	}

	saveData.updateStructWithSaveData()

	// Verify HouseTier extraction
	if len(saveData.HouseTier) != 5 {
		t.Fatalf("HouseTier len = %d, want 5", len(saveData.HouseTier))
	}
	for i, b := range saveData.HouseTier {
		if b != 0xAA {
			t.Errorf("HouseTier[%d] = 0x%02X, want 0xAA", i, b)
		}
	}

	// Verify HouseData extraction
	if len(saveData.HouseData) != 195 {
		t.Fatalf("HouseData len = %d, want 195", len(saveData.HouseData))
	}
	for i, b := range saveData.HouseData {
		if b != 0xBB {
			t.Errorf("HouseData[%d] = 0x%02X, want 0xBB", i, b)
		}
	}

	// Verify KQF extraction
	if len(saveData.KQF) != 8 {
		t.Fatalf("KQF len = %d, want 8", len(saveData.KQF))
	}
	for i, b := range saveData.KQF {
		if b != 0xCC {
			t.Errorf("KQF[%d] = 0x%02X, want 0xCC", i, b)
		}
	}
}
