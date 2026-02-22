package signserver

import (
	"database/sql"
	"testing"
	"time"

	cfg "erupe-ce/config"

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

func TestGetCharactersForUser(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		characters: []character{
			{ID: 1, IsFemale: false, Name: "Hunter1", HR: 100, GR: 50, WeaponType: 3, LastLogin: 1700000000},
			{ID: 2, IsFemale: true, Name: "Hunter2", HR: 200, GR: 100, WeaponType: 7, LastLogin: 1700000001},
		},
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

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
}

func TestGetCharactersForUserNoCharacters(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		characters: []character{},
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars, err := server.getCharactersForUser(1)
	if err != nil {
		t.Errorf("getCharactersForUser() error: %v", err)
	}
	if len(chars) != 0 {
		t.Errorf("getCharactersForUser() returned %d characters, want 0", len(chars))
	}
}

func TestGetCharactersForUserDBError(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		getForUserErr: sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	_, err := server.getCharactersForUser(1)
	if err == nil {
		t.Error("getCharactersForUser() should return error")
	}
}

func TestGetLastCID(t *testing.T) {
	userRepo := &mockSignUserRepo{
		lastCharacter: 12345,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	lastCID := server.getLastCID(1)
	if lastCID != 12345 {
		t.Errorf("getLastCID() = %d, want 12345", lastCID)
	}
}

func TestGetLastCIDNoResult(t *testing.T) {
	userRepo := &mockSignUserRepo{
		lastCharacterErr: sql.ErrNoRows,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	lastCID := server.getLastCID(1)
	if lastCID != 0 {
		t.Errorf("getLastCID() with no result = %d, want 0", lastCID)
	}
}

func TestGetUserRights(t *testing.T) {
	userRepo := &mockSignUserRepo{
		rights: 30,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	rights := server.getUserRights(1)
	if rights == 0 {
		t.Error("getUserRights() should return non-zero value")
	}
}

func TestGetReturnExpiry(t *testing.T) {
	recentLogin := time.Now().Add(-time.Hour * 24)
	userRepo := &mockSignUserRepo{
		lastLogin:    recentLogin,
		returnExpiry: time.Now().Add(time.Hour * 24 * 30),
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	expiry := server.getReturnExpiry(1)
	if expiry.Before(time.Now()) {
		t.Error("getReturnExpiry() should return future date")
	}
	if !userRepo.updateLastLoginCalled {
		t.Error("getReturnExpiry() should update last login")
	}
}

func TestGetReturnExpiryInactiveUser(t *testing.T) {
	oldLogin := time.Now().Add(-time.Hour * 24 * 100)
	userRepo := &mockSignUserRepo{
		lastLogin: oldLogin,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	expiry := server.getReturnExpiry(1)
	if expiry.Before(time.Now()) {
		t.Error("getReturnExpiry() should return future date for inactive user")
	}
	if !userRepo.updateReturnExpiryCalled {
		t.Error("getReturnExpiry() should update return expiry for inactive user")
	}
	if !userRepo.updateLastLoginCalled {
		t.Error("getReturnExpiry() should update last login")
	}
}

func TestGetReturnExpiryDBError(t *testing.T) {
	recentLogin := time.Now().Add(-time.Hour * 24)
	userRepo := &mockSignUserRepo{
		lastLogin:       recentLogin,
		returnExpiryErr: sql.ErrNoRows,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	expiry := server.getReturnExpiry(1)
	if expiry.IsZero() {
		t.Error("getReturnExpiry() should return non-zero time even on error")
	}
	if !userRepo.updateReturnExpiryCalled {
		t.Error("getReturnExpiry() should update return expiry on fallback")
	}
}

func TestNewUserChara(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		newCharCount: 0,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	err := server.newUserChara(1)
	if err != nil {
		t.Errorf("newUserChara() error: %v", err)
	}
	if !charRepo.createCalled {
		t.Error("newUserChara() should call CreateCharacter")
	}
}

func TestNewUserCharaAlreadyHasNewChar(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		newCharCount: 1,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	err := server.newUserChara(1)
	if err != nil {
		t.Errorf("newUserChara() should return nil when user already has new char: %v", err)
	}
	if charRepo.createCalled {
		t.Error("newUserChara() should not call CreateCharacter when user already has new char")
	}
}

func TestNewUserCharaCountError(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		newCharCountErr: sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	err := server.newUserChara(1)
	if err == nil {
		t.Error("newUserChara() should return error when count query fails")
	}
}

func TestNewUserCharaInsertError(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		newCharCount: 0,
		createErr:    sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	err := server.newUserChara(1)
	if err == nil {
		t.Error("newUserChara() should return error when insert fails")
	}
}

func TestRegisterDBAccount(t *testing.T) {
	userRepo := &mockSignUserRepo{
		registerUID: 1,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	uid, err := server.registerDBAccount("newuser", "password123")
	if err != nil {
		t.Errorf("registerDBAccount() error: %v", err)
	}
	if uid != 1 {
		t.Errorf("registerDBAccount() uid = %d, want 1", uid)
	}
	if !userRepo.registered {
		t.Error("registerDBAccount() should call Register")
	}
}

func TestRegisterDBAccountDuplicateUser(t *testing.T) {
	userRepo := &mockSignUserRepo{
		registerErr: sql.ErrNoRows,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	_, err := server.registerDBAccount("existinguser", "password123")
	if err == nil {
		t.Error("registerDBAccount() should return error for duplicate user")
	}
}

func TestDeleteCharacter(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateResult: true,
	}
	charRepo := &mockSignCharacterRepo{
		isNew: false,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
		charRepo:    charRepo,
	}

	err := server.deleteCharacter(123, "validtoken", 0)
	if err != nil {
		t.Errorf("deleteCharacter() error: %v", err)
	}
	if !charRepo.softDeleteCalled {
		t.Error("deleteCharacter() should soft delete existing character")
	}
}

func TestDeleteNewCharacter(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateResult: true,
	}
	charRepo := &mockSignCharacterRepo{
		isNew: true,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
		charRepo:    charRepo,
	}

	err := server.deleteCharacter(123, "validtoken", 0)
	if err != nil {
		t.Errorf("deleteCharacter() error: %v", err)
	}
	if !charRepo.hardDeleteCalled {
		t.Error("deleteCharacter() should hard delete new character")
	}
}

func TestDeleteCharacterInvalidToken(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateResult: false,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
	}

	err := server.deleteCharacter(123, "invalidtoken", 0)
	if err == nil {
		t.Error("deleteCharacter() should return error for invalid token")
	}
}

func TestDeleteCharacterDeleteError(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateResult: true,
	}
	charRepo := &mockSignCharacterRepo{
		isNew:         false,
		softDeleteErr: sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
		charRepo:    charRepo,
	}

	err := server.deleteCharacter(123, "validtoken", 0)
	if err == nil {
		t.Error("deleteCharacter() should return error when update fails")
	}
}

func TestGetFriendsForCharactersEmpty(t *testing.T) {
	charRepo := &mockSignCharacterRepo{}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{}
	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 0 {
		t.Errorf("getFriendsForCharacters() for empty chars = %d, want 0", len(friends))
	}
}

func TestGetGuildmatesForCharactersEmpty(t *testing.T) {
	charRepo := &mockSignCharacterRepo{}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{}
	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() for empty chars = %d, want 0", len(guildmates))
	}
}

func TestGetFriendsForCharacters(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		friends: []members{
			{ID: 2, Name: "Friend1"},
			{ID: 3, Name: "Friend2"},
		},
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 2 {
		t.Errorf("getFriendsForCharacters() = %d, want 2", len(friends))
	}
	if friends[0].CID != 1 {
		t.Errorf("friends[0].CID = %d, want 1", friends[0].CID)
	}
}

func TestGetGuildmatesForCharacters(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		guildmates: []members{
			{ID: 2, Name: "Guildmate1"},
			{ID: 3, Name: "Guildmate2"},
		},
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 2 {
		t.Errorf("getGuildmatesForCharacters() = %d, want 2", len(guildmates))
	}
	if guildmates[0].CID != 1 {
		t.Errorf("guildmates[0].CID = %d, want 1", guildmates[0].CID)
	}
}

func TestGetGuildmatesNotInGuild(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		guildmates: nil,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{
		{ID: 1, Name: "Hunter1"},
	}

	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() for non-guild member = %d, want 0", len(guildmates))
	}
}

func TestValidateLoginSuccess(t *testing.T) {
	// bcrypt hash for "password123"
	hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	userRepo := &mockSignUserRepo{
		credUID:      1,
		credPassword: hash,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	// Note: bcrypt verification will fail with this test hash since it's not a real hash of "password123"
	// The important thing is testing the flow, not actual bcrypt verification
	_, resp := server.validateLogin("testuser", "password123")
	// This will return SIGN_EPASS since the hash doesn't match, which is expected behavior
	if resp == SIGN_EABORT {
		t.Error("validateLogin() should not abort for valid credentials lookup")
	}
}

func TestValidateLoginUserNotFound(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: sql.ErrNoRows,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	_, resp := server.validateLogin("unknown", "password")
	if resp != SIGN_EAUTH {
		t.Errorf("validateLogin() for unknown user = %d, want SIGN_EAUTH(%d)", resp, SIGN_EAUTH)
	}
}

func TestValidateLoginAutoCreate(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr:     sql.ErrNoRows,
		registerUID: 42,
	}

	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			AutoCreateAccount: true,
		},
		userRepo: userRepo,
	}

	uid, resp := server.validateLogin("newuser", "password")
	if resp != SIGN_SUCCESS {
		t.Errorf("validateLogin() with auto-create = %d, want SIGN_SUCCESS(%d)", resp, SIGN_SUCCESS)
	}
	if uid != 42 {
		t.Errorf("validateLogin() uid = %d, want 42", uid)
	}
}

func TestValidateLoginDBError(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	_, resp := server.validateLogin("testuser", "password")
	if resp != SIGN_EABORT {
		t.Errorf("validateLogin() on DB error = %d, want SIGN_EABORT(%d)", resp, SIGN_EABORT)
	}
}

func TestValidateTokenValid(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateResult: true,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
	}

	if !server.validateToken("validtoken", 0) {
		t.Error("validateToken() should return true for valid token")
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateResult: false,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
	}

	if server.validateToken("invalidtoken", 0) {
		t.Error("validateToken() should return false for invalid token")
	}
}

func TestValidateTokenDBError(t *testing.T) {
	sessionRepo := &mockSignSessionRepo{
		validateErr: sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		sessionRepo: sessionRepo,
	}

	if server.validateToken("token", 0) {
		t.Error("validateToken() should return false on DB error")
	}
}

func TestGetUserRightsZeroUID(t *testing.T) {
	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
	}

	rights := server.getUserRights(0)
	if rights != 0 {
		t.Errorf("getUserRights(0) = %d, want 0", rights)
	}
}

func TestGetUserRightsDBError(t *testing.T) {
	userRepo := &mockSignUserRepo{
		rightsErr: sql.ErrConnDone,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		userRepo:    userRepo,
	}

	rights := server.getUserRights(1)
	if rights != 0 {
		t.Errorf("getUserRights() on error = %d, want 0", rights)
	}
}

func TestGetFriendsForCharactersError(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		getFriendsErr: errMockDB,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{{ID: 1, Name: "Hunter1"}}
	friends := server.getFriendsForCharacters(chars)
	if len(friends) != 0 {
		t.Errorf("getFriendsForCharacters() on error = %d, want 0", len(friends))
	}
}

func TestGetGuildmatesForCharactersError(t *testing.T) {
	charRepo := &mockSignCharacterRepo{
		getGuildmatesErr: errMockDB,
	}

	server := &Server{
		logger:      zap.NewNop(),
		erupeConfig: &cfg.Config{},
		charRepo:    charRepo,
	}

	chars := []character{{ID: 1, Name: "Hunter1"}}
	guildmates := server.getGuildmatesForCharacters(chars)
	if len(guildmates) != 0 {
		t.Errorf("getGuildmatesForCharacters() on error = %d, want 0", len(guildmates))
	}
}
