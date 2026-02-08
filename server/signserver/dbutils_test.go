package signserver

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
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

// Helper to create a test server with mocked database
func newTestServerWithMock(t *testing.T) (*Server, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger: zap.NewNop(),
		db:     sqlxDB,
	}

	return server, mock
}

func TestGetCharactersForUser(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	rows := sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hrp", "gr", "weapon_type", "last_login"}).
		AddRow(1, false, false, "Hunter1", "desc1", 100, 50, 3, 1700000000).
		AddRow(2, true, false, "Hunter2", "desc2", 200, 100, 7, 1700000001)

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id ASC").
		WithArgs(1).
		WillReturnRows(rows)

	chars, err := server.getCharactersForUser(1)
	if err != nil {
		t.Errorf("getCharactersForUser() error: %v", err)
	}

	if len(chars) != 2 {
		t.Errorf("getCharactersForUser() returned %d characters, want 2", len(chars))
	}

	if chars[0].Name != "Hunter1" {
		t.Errorf("First character name = %s, want Hunter1", chars[0].Name)
	}

	if chars[1].IsFemale != true {
		t.Error("Second character should be female")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetCharactersForUserNoCharacters(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	rows := sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hrp", "gr", "weapon_type", "last_login"})

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id ASC").
		WithArgs(1).
		WillReturnRows(rows)

	chars, err := server.getCharactersForUser(1)
	if err != nil {
		t.Errorf("getCharactersForUser() error: %v", err)
	}

	if len(chars) != 0 {
		t.Errorf("getCharactersForUser() returned %d characters, want 0", len(chars))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetCharactersForUserDBError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id ASC").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	_, err := server.getCharactersForUser(1)
	if err == nil {
		t.Error("getCharactersForUser() should return error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetLastCID(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT last_character FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_character"}).AddRow(12345))

	lastCID := server.getLastCID(1)
	if lastCID != 12345 {
		t.Errorf("getLastCID() = %d, want 12345", lastCID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetLastCIDNoResult(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT last_character FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	lastCID := server.getLastCID(1)
	// Should return 0 on error
	if lastCID != 0 {
		t.Errorf("getLastCID() with no result = %d, want 0", lastCID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetUserRights(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT rights FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"rights"}).AddRow(30))

	rights := server.getUserRights(1)
	// Rights value is transformed by mhfcourse.GetCourseStruct
	// The function should return a non-zero value when rights is set
	if rights == 0 {
		t.Error("getUserRights() should return non-zero value")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetUserRightsDefault(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT rights FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	rights := server.getUserRights(1)
	// Default rights is 2, which is transformed by mhfcourse.GetCourseStruct
	if rights == 0 {
		t.Error("getUserRights() should return default rights on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCheckToken(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE user_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := server.checkToken(1)
	if err != nil {
		t.Errorf("checkToken() error: %v", err)
	}
	if !exists {
		t.Error("checkToken() should return true when token exists")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCheckTokenNotExists(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE user_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := server.checkToken(1)
	if err != nil {
		t.Errorf("checkToken() error: %v", err)
	}
	if exists {
		t.Error("checkToken() should return false when token doesn't exist")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCheckTokenError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE user_id = \\$1").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	_, err := server.checkToken(1)
	if err == nil {
		t.Error("checkToken() should return error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterToken(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectExec("INSERT INTO sign_sessions \\(user_id, token\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(1, "testtoken123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := server.registerToken(1, "testtoken123")
	if err != nil {
		t.Errorf("registerToken() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterTokenError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectExec("INSERT INTO sign_sessions \\(user_id, token\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(1, "testtoken123").
		WillReturnError(sql.ErrConnDone)

	err := server.registerToken(1, "testtoken123")
	if err == nil {
		t.Error("registerToken() should return error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteCharacter(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Token verification
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("validtoken").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Check if new character
	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(false))

	// Soft delete (update deleted flag)
	mock.ExpectExec("UPDATE characters SET deleted = true WHERE id = \\$1").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := server.deleteCharacter(123, "validtoken")
	if err != nil {
		t.Errorf("deleteCharacter() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteNewCharacter(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Token verification
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("validtoken").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Check if new character (is_new_character = true)
	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(true))

	// Hard delete for new characters
	mock.ExpectExec("DELETE FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := server.deleteCharacter(123, "validtoken")
	if err != nil {
		t.Errorf("deleteCharacter() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteCharacterInvalidToken(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Token verification fails
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("invalidtoken").
		WillReturnError(sql.ErrNoRows)

	err := server.deleteCharacter(123, "invalidtoken")
	if err == nil {
		t.Error("deleteCharacter() should return error for invalid token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserChara(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Check for existing new characters
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Insert new character
	mock.ExpectExec("INSERT INTO characters").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := server.newUserChara("testuser")
	if err != nil {
		t.Errorf("newUserChara() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserCharaAlreadyHasNewChar(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Check for existing new characters - already has one
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Should not insert since user already has a new character
	err := server.newUserChara("testuser")
	// Error is nil but no insert happens
	if err != nil {
		t.Errorf("newUserChara() should return nil when user already has new char: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserCharaUserNotFound(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get user ID - not found
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("unknownuser").
		WillReturnError(sql.ErrNoRows)

	err := server.newUserChara("unknownuser")
	if err == nil {
		t.Error("newUserChara() should return error when user not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterDBAccount(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Insert user
	mock.ExpectExec("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs("newuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Insert initial character
	mock.ExpectExec("INSERT INTO characters").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := server.registerDBAccount("newuser", "password123")
	if err != nil {
		t.Errorf("registerDBAccount() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterDBAccountDuplicateUser(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Insert user fails (duplicate)
	mock.ExpectExec("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs("existinguser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	err := server.registerDBAccount("existinguser", "password123")
	if err == nil {
		t.Error("registerDBAccount() should return error for duplicate user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetReturnExpiry(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get last login (recent)
	recentLogin := time.Now().Add(-time.Hour * 24) // 1 day ago
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(recentLogin))

	// Get return expiry
	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"return_expires"}).AddRow(time.Now().Add(time.Hour * 24 * 30)))

	// Update last login
	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	expiry := server.getReturnExpiry(1)

	// Should return a future date
	if expiry.Before(time.Now()) {
		t.Error("getReturnExpiry() should return future date")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetReturnExpiryInactiveUser(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get last login (inactive - over 90 days ago)
	oldLogin := time.Now().Add(-time.Hour * 24 * 100) // 100 days ago
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(oldLogin))

	// Update return expiry for returning user
	mock.ExpectExec("UPDATE users SET return_expires=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Update last login
	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	expiry := server.getReturnExpiry(1)

	// Should return a future date (30 days from now for returning user)
	if expiry.Before(time.Now()) {
		t.Error("getReturnExpiry() should return future date for inactive user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetFriendsForCharactersEmpty(t *testing.T) {
	server, _ := newTestServerWithMock(t)

	// Empty character list
	chars := []character{}

	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 0 {
		t.Errorf("getFriendsForCharacters() for empty chars = %d, want 0", len(friends))
	}
}

func TestGetGuildmatesForCharactersEmpty(t *testing.T) {
	server, _ := newTestServerWithMock(t)

	// Empty character list
	chars := []character{}

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() for empty chars = %d, want 0", len(guildmates))
	}
}

func TestGetFriendsForCharacters(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	// Get friends CSV for character
	mock.ExpectQuery("SELECT friends FROM characters WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"friends"}).AddRow("2,3"))

	// Query friends
	mock.ExpectQuery("SELECT id, name FROM characters WHERE id=2 OR id=3").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "Friend1").
			AddRow(3, "Friend2"))

	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 2 {
		t.Errorf("getFriendsForCharacters() = %d, want 2", len(friends))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetGuildmatesForCharacters(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	// Check if in guild
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Get guild ID
	mock.ExpectQuery("SELECT guild_id FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"guild_id"}).AddRow(100))

	// Get guildmates
	mock.ExpectQuery("SELECT character_id AS id, c.name FROM guild_characters gc JOIN characters c ON c.id = gc.character_id WHERE guild_id=\\$1 AND character_id!=\\$2").
		WithArgs(100, uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "Guildmate1").
			AddRow(3, "Guildmate2"))

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 2 {
		t.Errorf("getGuildmatesForCharacters() = %d, want 2", len(guildmates))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetGuildmatesNotInGuild(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	// Check if in guild - not in guild
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() for non-guild member = %d, want 0", len(guildmates))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestGetFriendsForCharactersDBError tests getFriendsForCharacters when DB query fails
func TestGetFriendsForCharactersDBError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	// Get friends CSV for character - DB error
	mock.ExpectQuery("SELECT friends FROM characters WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnError(sql.ErrNoRows)

	// Even on error, still produces the friend query (with empty/error friendsCSV)
	// The function calls Scan which fails, then continues to build a query
	// with the empty string. The query then fails as well.
	mock.ExpectQuery("SELECT id, name FROM characters").
		WillReturnError(sql.ErrConnDone)

	friends := server.getFriendsForCharacters(chars)
	// Should return 0 friends on error
	if len(friends) != 0 {
		t.Errorf("getFriendsForCharacters() with DB error = %d, want 0", len(friends))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestGetGuildmatesForCharactersGuildQueryError tests guild ID query failure
func TestGetGuildmatesForCharactersGuildQueryError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	// Check if in guild - yes
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Get guild ID - error
	mock.ExpectQuery("SELECT guild_id FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnError(sql.ErrConnDone)

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() with guild query error = %d, want 0", len(guildmates))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestGetGuildmatesForCharactersGuildmatesQueryError tests guildmates query failure
func TestGetGuildmatesForCharactersGuildmatesQueryError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	// Check if in guild - yes
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Get guild ID
	mock.ExpectQuery("SELECT guild_id FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"guild_id"}).AddRow(100))

	// Get guildmates - error
	mock.ExpectQuery("SELECT character_id AS id, c.name FROM guild_characters gc JOIN characters c ON c.id = gc.character_id WHERE guild_id=\\$1 AND character_id!=\\$2").
		WithArgs(100, uint32(1)).
		WillReturnError(sql.ErrConnDone)

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() with guildmates query error = %d, want 0", len(guildmates))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestDeleteCharacterDeleteError tests deleteCharacter when the delete/update query fails
func TestDeleteCharacterDeleteError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Token verification
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("validtoken").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Check if new character
	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(false))

	// Soft delete fails
	mock.ExpectExec("UPDATE characters SET deleted = true WHERE id = \\$1").
		WithArgs(123).
		WillReturnError(sql.ErrConnDone)

	err := server.deleteCharacter(123, "validtoken")
	if err == nil {
		t.Error("deleteCharacter() should return error when update fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNewUserCharaInsertError tests newUserChara when the INSERT fails
func TestNewUserCharaInsertError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Check for existing new characters
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Insert new character - error
	mock.ExpectExec("INSERT INTO characters").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err := server.newUserChara("testuser")
	if err == nil {
		t.Error("newUserChara() should return error when insert fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNewUserCharaCountError tests newUserChara when the COUNT query fails
func TestNewUserCharaCountError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Check for existing new characters - error
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	err := server.newUserChara("testuser")
	if err == nil {
		t.Error("newUserChara() should return error when count query fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestRegisterDBAccountGetIDError tests registerDBAccount when getting the new user ID fails
func TestRegisterDBAccountGetIDError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Insert user succeeds
	mock.ExpectExec("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs("newuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get user ID - error
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("newuser").
		WillReturnError(sql.ErrConnDone)

	err := server.registerDBAccount("newuser", "password123")
	if err == nil {
		t.Error("registerDBAccount() should return error when getting ID fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestRegisterDBAccountCharacterInsertError tests registerDBAccount when character insert fails
func TestRegisterDBAccountCharacterInsertError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Insert user
	mock.ExpectExec("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs("newuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Insert character - error
	mock.ExpectExec("INSERT INTO characters").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err := server.registerDBAccount("newuser", "password123")
	if err == nil {
		t.Error("registerDBAccount() should return error when character insert fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestGetReturnExpiryDBError tests getReturnExpiry when the return_expires query fails
func TestGetReturnExpiryDBError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// Get last login - recent
	recentLogin := time.Now().Add(-time.Hour * 24)
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(recentLogin))

	// Get return expiry - error
	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	// Should set return_expires to now
	mock.ExpectExec("UPDATE users SET return_expires=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Update last login
	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	expiry := server.getReturnExpiry(1)

	// Should still return a valid time (approximately now)
	if expiry.IsZero() {
		t.Error("getReturnExpiry() should return non-zero time even on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestGetFriendsForCharactersMultipleChars tests with multiple characters
func TestGetFriendsForCharactersMultipleChars(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
		{ID: 2, Name: "Hunter2"},
	}

	// First character friends
	mock.ExpectQuery("SELECT friends FROM characters WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"friends"}).AddRow("10"))

	mock.ExpectQuery("SELECT id, name FROM characters WHERE id=10").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(10, "Friend1"))

	// Second character friends
	mock.ExpectQuery("SELECT friends FROM characters WHERE id=\\$1").
		WithArgs(uint32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"friends"}).AddRow("20"))

	mock.ExpectQuery("SELECT id, name FROM characters WHERE id=20").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(20, "Friend2"))

	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 2 {
		t.Errorf("getFriendsForCharacters() = %d, want 2", len(friends))
	}

	// Verify CID assignment
	if len(friends) >= 2 {
		if friends[0].CID != 1 {
			t.Errorf("friends[0].CID = %d, want 1", friends[0].CID)
		}
		if friends[1].CID != 2 {
			t.Errorf("friends[1].CID = %d, want 2", friends[1].CID)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestGetGuildmatesForCharactersMultipleChars tests with multiple characters in guilds
func TestGetGuildmatesForCharactersMultipleChars(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	chars := []character{
		{ID: 1, Name: "Hunter1"},
		{ID: 2, Name: "Hunter2"},
	}

	// First character in guild
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT guild_id FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"guild_id"}).AddRow(100))

	mock.ExpectQuery("SELECT character_id AS id, c.name FROM guild_characters gc JOIN characters c ON c.id = gc.character_id WHERE guild_id=\\$1 AND character_id!=\\$2").
		WithArgs(100, uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(10, "Guildmate1"))

	// Second character not in guild
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 1 {
		t.Errorf("getGuildmatesForCharacters() = %d, want 1", len(guildmates))
	}

	if len(guildmates) >= 1 && guildmates[0].CID != 1 {
		t.Errorf("guildmates[0].CID = %d, want 1", guildmates[0].CID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
