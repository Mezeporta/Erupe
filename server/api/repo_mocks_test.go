package api

import (
	"context"
	"time"
)

// mockAPIUserRepo implements APIUserRepo for testing.
type mockAPIUserRepo struct {
	registerID     uint32
	registerRights uint32
	registerErr    error

	credentialsID       uint32
	credentialsPassword string
	credentialsRights   uint32
	credentialsErr      error

	lastLogin    time.Time
	lastLoginErr error

	returnExpiry    time.Time
	returnExpiryErr error

	updateReturnExpiryErr error
	updateLastLoginErr    error
}

func (m *mockAPIUserRepo) Register(_ context.Context, _, _ string, _ time.Time) (uint32, uint32, error) {
	return m.registerID, m.registerRights, m.registerErr
}

func (m *mockAPIUserRepo) GetCredentials(_ context.Context, _ string) (uint32, string, uint32, error) {
	return m.credentialsID, m.credentialsPassword, m.credentialsRights, m.credentialsErr
}

func (m *mockAPIUserRepo) GetLastLogin(_ uint32) (time.Time, error) {
	return m.lastLogin, m.lastLoginErr
}

func (m *mockAPIUserRepo) GetReturnExpiry(_ uint32) (time.Time, error) {
	return m.returnExpiry, m.returnExpiryErr
}

func (m *mockAPIUserRepo) UpdateReturnExpiry(_ uint32, _ time.Time) error {
	return m.updateReturnExpiryErr
}

func (m *mockAPIUserRepo) UpdateLastLogin(_ uint32, _ time.Time) error {
	return m.updateLastLoginErr
}

// mockAPICharacterRepo implements APICharacterRepo for testing.
type mockAPICharacterRepo struct {
	newCharacter    Character
	newCharacterErr error

	countForUser    int
	countForUserErr error

	createChar    Character
	createCharErr error

	isNewResult bool
	isNewErr    error

	hardDeleteErr error
	softDeleteErr error

	characters    []Character
	charactersErr error

	exportResult map[string]interface{}
	exportErr    error
}

func (m *mockAPICharacterRepo) GetNewCharacter(_ context.Context, _ uint32) (Character, error) {
	return m.newCharacter, m.newCharacterErr
}

func (m *mockAPICharacterRepo) CountForUser(_ context.Context, _ uint32) (int, error) {
	return m.countForUser, m.countForUserErr
}

func (m *mockAPICharacterRepo) Create(_ context.Context, _ uint32, _ uint32) (Character, error) {
	return m.createChar, m.createCharErr
}

func (m *mockAPICharacterRepo) IsNew(_ uint32) (bool, error) {
	return m.isNewResult, m.isNewErr
}

func (m *mockAPICharacterRepo) HardDelete(_ uint32) error {
	return m.hardDeleteErr
}

func (m *mockAPICharacterRepo) SoftDelete(_ uint32) error {
	return m.softDeleteErr
}

func (m *mockAPICharacterRepo) GetForUser(_ context.Context, _ uint32) ([]Character, error) {
	return m.characters, m.charactersErr
}

func (m *mockAPICharacterRepo) ExportSave(_ context.Context, _, _ uint32) (map[string]interface{}, error) {
	return m.exportResult, m.exportErr
}

// mockAPISessionRepo implements APISessionRepo for testing.
type mockAPISessionRepo struct {
	createTokenID  uint32
	createTokenErr error

	userID    uint32
	userIDErr error
}

func (m *mockAPISessionRepo) CreateToken(_ context.Context, _ uint32, _ string) (uint32, error) {
	return m.createTokenID, m.createTokenErr
}

func (m *mockAPISessionRepo) GetUserIDByToken(_ context.Context, _ string) (uint32, error) {
	return m.userID, m.userIDErr
}
