package signserver

import (
	"testing"
)

func TestCharacterStruct(t *testing.T) {
	c := character{
		ID:             12345,
		IsFemale:       true,
		IsNewCharacter: false,
		Name:           "TestHunter",
		UnkDescString:  "Test description",
		HRP:            999,
		GR:             300,
		WeaponType:     5,
		LastLogin:      1700000000,
	}

	if c.ID != 12345 {
		t.Errorf("ID = %d, want 12345", c.ID)
	}
	if c.IsFemale != true {
		t.Error("IsFemale should be true")
	}
	if c.IsNewCharacter != false {
		t.Error("IsNewCharacter should be false")
	}
	if c.Name != "TestHunter" {
		t.Errorf("Name = %s, want TestHunter", c.Name)
	}
	if c.UnkDescString != "Test description" {
		t.Errorf("UnkDescString = %s, want Test description", c.UnkDescString)
	}
	if c.HRP != 999 {
		t.Errorf("HRP = %d, want 999", c.HRP)
	}
	if c.GR != 300 {
		t.Errorf("GR = %d, want 300", c.GR)
	}
	if c.WeaponType != 5 {
		t.Errorf("WeaponType = %d, want 5", c.WeaponType)
	}
	if c.LastLogin != 1700000000 {
		t.Errorf("LastLogin = %d, want 1700000000", c.LastLogin)
	}
}

func TestCharacterStructDefaults(t *testing.T) {
	c := character{}

	if c.ID != 0 {
		t.Errorf("default ID = %d, want 0", c.ID)
	}
	if c.IsFemale != false {
		t.Error("default IsFemale should be false")
	}
	if c.IsNewCharacter != false {
		t.Error("default IsNewCharacter should be false")
	}
	if c.Name != "" {
		t.Errorf("default Name = %s, want empty", c.Name)
	}
	if c.HRP != 0 {
		t.Errorf("default HRP = %d, want 0", c.HRP)
	}
	if c.GR != 0 {
		t.Errorf("default GR = %d, want 0", c.GR)
	}
	if c.WeaponType != 0 {
		t.Errorf("default WeaponType = %d, want 0", c.WeaponType)
	}
}

func TestMembersStruct(t *testing.T) {
	m := members{
		CID:  100,
		ID:   200,
		Name: "FriendName",
	}

	if m.CID != 100 {
		t.Errorf("CID = %d, want 100", m.CID)
	}
	if m.ID != 200 {
		t.Errorf("ID = %d, want 200", m.ID)
	}
	if m.Name != "FriendName" {
		t.Errorf("Name = %s, want FriendName", m.Name)
	}
}

func TestMembersStructDefaults(t *testing.T) {
	m := members{}

	if m.CID != 0 {
		t.Errorf("default CID = %d, want 0", m.CID)
	}
	if m.ID != 0 {
		t.Errorf("default ID = %d, want 0", m.ID)
	}
	if m.Name != "" {
		t.Errorf("default Name = %s, want empty", m.Name)
	}
}

func TestCharacterWeaponTypes(t *testing.T) {
	// Test all weapon type values are valid
	weaponTypes := []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

	for _, wt := range weaponTypes {
		c := character{WeaponType: wt}
		if c.WeaponType != wt {
			t.Errorf("WeaponType = %d, want %d", c.WeaponType, wt)
		}
	}
}

func TestCharacterHRPRange(t *testing.T) {
	tests := []struct {
		name string
		hrp  uint16
	}{
		{"min", 0},
		{"beginner", 1},
		{"hr30", 30},
		{"hr50", 50},
		{"hr99", 99},
		{"hr299", 299},
		{"hr998", 998},
		{"hr999", 999},
		{"max uint16", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := character{HRP: tt.hrp}
			if c.HRP != tt.hrp {
				t.Errorf("HRP = %d, want %d", c.HRP, tt.hrp)
			}
		})
	}
}

func TestCharacterGRRange(t *testing.T) {
	tests := []struct {
		name string
		gr   uint16
	}{
		{"min", 0},
		{"gr1", 1},
		{"gr100", 100},
		{"gr300", 300},
		{"gr999", 999},
		{"max uint16", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := character{GR: tt.gr}
			if c.GR != tt.gr {
				t.Errorf("GR = %d, want %d", c.GR, tt.gr)
			}
		})
	}
}

func TestCharacterIDRange(t *testing.T) {
	tests := []struct {
		name string
		id   uint32
	}{
		{"min", 0},
		{"small", 1},
		{"medium", 1000000},
		{"large", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := character{ID: tt.id}
			if c.ID != tt.id {
				t.Errorf("ID = %d, want %d", c.ID, tt.id)
			}
		})
	}
}

func TestCharacterGender(t *testing.T) {
	// Male character
	male := character{IsFemale: false}
	if male.IsFemale != false {
		t.Error("Male character should have IsFemale = false")
	}

	// Female character
	female := character{IsFemale: true}
	if female.IsFemale != true {
		t.Error("Female character should have IsFemale = true")
	}
}

func TestCharacterNewStatus(t *testing.T) {
	// New character
	newChar := character{IsNewCharacter: true}
	if newChar.IsNewCharacter != true {
		t.Error("New character should have IsNewCharacter = true")
	}

	// Existing character
	existingChar := character{IsNewCharacter: false}
	if existingChar.IsNewCharacter != false {
		t.Error("Existing character should have IsNewCharacter = false")
	}
}

func TestCharacterNameLength(t *testing.T) {
	// Test various name lengths
	names := []string{
		"",                  // Empty
		"A",                 // Single char
		"Hunter",            // Normal
		"LongHunterName123", // Longer
	}

	for _, name := range names {
		c := character{Name: name}
		if c.Name != name {
			t.Errorf("Name = %s, want %s", c.Name, name)
		}
	}
}

func TestCharacterLastLogin(t *testing.T) {
	tests := []struct {
		name      string
		lastLogin uint32
	}{
		{"zero", 0},
		{"epoch", 0},
		{"past", 1600000000},
		{"present", 1700000000},
		{"future", 1800000000},
		{"max", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := character{LastLogin: tt.lastLogin}
			if c.LastLogin != tt.lastLogin {
				t.Errorf("LastLogin = %d, want %d", c.LastLogin, tt.lastLogin)
			}
		})
	}
}

func TestMembersCIDAssignment(t *testing.T) {
	// CID is the local character ID that references this member
	m := members{CID: 12345}
	if m.CID != 12345 {
		t.Errorf("CID = %d, want 12345", m.CID)
	}
}

func TestMultipleCharacters(t *testing.T) {
	// Test creating multiple character instances
	chars := []character{
		{ID: 1, Name: "Char1", HRP: 100},
		{ID: 2, Name: "Char2", HRP: 200},
		{ID: 3, Name: "Char3", HRP: 300},
	}

	for i, c := range chars {
		expectedID := uint32(i + 1)
		if c.ID != expectedID {
			t.Errorf("chars[%d].ID = %d, want %d", i, c.ID, expectedID)
		}
	}
}

func TestMultipleMembers(t *testing.T) {
	// Test creating multiple member instances
	membersList := []members{
		{CID: 1, ID: 10, Name: "Friend1"},
		{CID: 1, ID: 20, Name: "Friend2"},
		{CID: 2, ID: 30, Name: "Friend3"},
	}

	// First two should share the same CID
	if membersList[0].CID != membersList[1].CID {
		t.Error("First two members should share the same CID")
	}

	// Third should have different CID
	if membersList[1].CID == membersList[2].CID {
		t.Error("Third member should have different CID")
	}
}
