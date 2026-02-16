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
		HR:             999,
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
	if c.HR != 999 {
		t.Errorf("HR = %d, want 999", c.HR)
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
	if c.HR != 0 {
		t.Errorf("default HR = %d, want 0", c.HR)
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
	weaponTypes := []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

	for _, wt := range weaponTypes {
		c := character{WeaponType: wt}
		if c.WeaponType != wt {
			t.Errorf("WeaponType = %d, want %d", c.WeaponType, wt)
		}
	}
}

func TestCharacterHRRange(t *testing.T) {
	tests := []struct {
		name string
		hr   uint16
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
			c := character{HR: tt.hr}
			if c.HR != tt.hr {
				t.Errorf("HR = %d, want %d", c.HR, tt.hr)
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
	male := character{IsFemale: false}
	if male.IsFemale != false {
		t.Error("Male character should have IsFemale = false")
	}

	female := character{IsFemale: true}
	if female.IsFemale != true {
		t.Error("Female character should have IsFemale = true")
	}
}

func TestCharacterNewStatus(t *testing.T) {
	newChar := character{IsNewCharacter: true}
	if newChar.IsNewCharacter != true {
		t.Error("New character should have IsNewCharacter = true")
	}

	existingChar := character{IsNewCharacter: false}
	if existingChar.IsNewCharacter != false {
		t.Error("Existing character should have IsNewCharacter = false")
	}
}

func TestCharacterNameLength(t *testing.T) {
	names := []string{
		"",
		"A",
		"Hunter",
		"LongHunterName123",
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
	m := members{CID: 12345}
	if m.CID != 12345 {
		t.Errorf("CID = %d, want 12345", m.CID)
	}
}

func TestMultipleCharacters(t *testing.T) {
	chars := []character{
		{ID: 1, Name: "Char1", HR: 100},
		{ID: 2, Name: "Char2", HR: 200},
		{ID: 3, Name: "Char3", HR: 300},
	}

	for i, c := range chars {
		expectedID := uint32(i + 1)
		if c.ID != expectedID {
			t.Errorf("chars[%d].ID = %d, want %d", i, c.ID, expectedID)
		}
	}
}

func TestMultipleMembers(t *testing.T) {
	membersList := []members{
		{CID: 1, ID: 10, Name: "Friend1"},
		{CID: 1, ID: 20, Name: "Friend2"},
		{CID: 2, ID: 30, Name: "Friend3"},
	}

	if membersList[0].CID != membersList[1].CID {
		t.Error("First two members should share the same CID")
	}

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

	rows := sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hr", "gr", "weapon_type", "last_login"}).
		AddRow(1, false, false, "Hunter1", "desc1", 100, 50, 3, 1700000000).
		AddRow(2, true, false, "Hunter2", "desc2", 200, 100, 7, 1700000001)

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hr, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id").
		WithArgs(uint32(1)).
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

	rows := sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hr", "gr", "weapon_type", "last_login"})

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hr, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id").
		WithArgs(uint32(1)).
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

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hr, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id").
		WithArgs(uint32(1)).
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
		WithArgs(uint32(1)).
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
		WithArgs(uint32(1)).
		WillReturnError(sql.ErrNoRows)

	lastCID := server.getLastCID(1)
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
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"rights"}).AddRow(30))

	rights := server.getUserRights(1)
	if rights == 0 {
		t.Error("getUserRights() should return non-zero value")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetReturnExpiry(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	recentLogin := time.Now().Add(-time.Hour * 24)
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(recentLogin))

	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"return_expires"}).AddRow(time.Now().Add(time.Hour * 24 * 30)))

	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), uint32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	expiry := server.getReturnExpiry(1)

	if expiry.Before(time.Now()) {
		t.Error("getReturnExpiry() should return future date")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetReturnExpiryInactiveUser(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	oldLogin := time.Now().Add(-time.Hour * 24 * 100)
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(oldLogin))

	mock.ExpectExec("UPDATE users SET return_expires=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), uint32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), uint32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	expiry := server.getReturnExpiry(1)

	if expiry.Before(time.Now()) {
		t.Error("getReturnExpiry() should return future date for inactive user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetReturnExpiryDBError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	recentLogin := time.Now().Add(-time.Hour * 24)
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(recentLogin))

	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("UPDATE users SET return_expires=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), uint32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), uint32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	expiry := server.getReturnExpiry(1)

	if expiry.IsZero() {
		t.Error("getReturnExpiry() should return non-zero time even on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserChara(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec("INSERT INTO characters").
		WithArgs(uint32(1), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := server.newUserChara(1)
	if err != nil {
		t.Errorf("newUserChara() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserCharaAlreadyHasNewChar(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	err := server.newUserChara(1)
	if err != nil {
		t.Errorf("newUserChara() should return nil when user already has new char: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserCharaCountError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(uint32(1)).
		WillReturnError(sql.ErrConnDone)

	err := server.newUserChara(1)
	if err == nil {
		t.Error("newUserChara() should return error when count query fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestNewUserCharaInsertError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec("INSERT INTO characters").
		WithArgs(uint32(1), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err := server.newUserChara(1)
	if err == nil {
		t.Error("newUserChara() should return error when insert fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterDBAccount(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id").
		WithArgs("newuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	uid, err := server.registerDBAccount("newuser", "password123")
	if err != nil {
		t.Errorf("registerDBAccount() error: %v", err)
	}
	if uid != 1 {
		t.Errorf("registerDBAccount() uid = %d, want 1", uid)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterDBAccountDuplicateUser(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id").
		WithArgs("existinguser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	_, err := server.registerDBAccount("existinguser", "password123")
	if err == nil {
		t.Error("registerDBAccount() should return error for duplicate user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteCharacter(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	// validateToken: SELECT count(*) FROM sign_sessions WHERE token = $1
	// When tokenID=0, query has no AND clause but both args are still passed to QueryRow
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("validtoken", uint32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(false))

	mock.ExpectExec("UPDATE characters SET deleted = true WHERE id = \\$1").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := server.deleteCharacter(123, "validtoken", 0)
	if err != nil {
		t.Errorf("deleteCharacter() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteNewCharacter(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("validtoken", uint32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(true))

	mock.ExpectExec("DELETE FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := server.deleteCharacter(123, "validtoken", 0)
	if err != nil {
		t.Errorf("deleteCharacter() error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteCharacterInvalidToken(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("invalidtoken", uint32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	err := server.deleteCharacter(123, "invalidtoken", 0)
	if err == nil {
		t.Error("deleteCharacter() should return error for invalid token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteCharacterDeleteError(t *testing.T) {
	server, mock := newTestServerWithMock(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("validtoken", uint32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(false))

	mock.ExpectExec("UPDATE characters SET deleted = true WHERE id = \\$1").
		WithArgs(123).
		WillReturnError(sql.ErrConnDone)

	err := server.deleteCharacter(123, "validtoken", 0)
	if err == nil {
		t.Error("deleteCharacter() should return error when update fails")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetFriendsForCharactersEmpty(t *testing.T) {
	server, _ := newTestServerWithMock(t)

	chars := []character{}

	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 0 {
		t.Errorf("getFriendsForCharacters() for empty chars = %d, want 0", len(friends))
	}
}

func TestGetGuildmatesForCharactersEmpty(t *testing.T) {
	server, _ := newTestServerWithMock(t)

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

	mock.ExpectQuery("SELECT friends FROM characters WHERE id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"friends"}).AddRow("2,3"))

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

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT guild_id FROM guild_characters WHERE character_id=\\$1").
		WithArgs(uint32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"guild_id"}).AddRow(100))

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
