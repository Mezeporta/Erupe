package signserver

import (
	"errors"
	"time"
)

// errMockDB is a sentinel for mock repo error injection.
var errMockDB = errors.New("mock database error")

// --- mockSignUserRepo ---

type mockSignUserRepo struct {
	// GetCredentials
	credUID      uint32
	credPassword string
	credErr      error

	// Register
	registerUID uint32
	registerErr error
	registered  bool

	// GetRights
	rights    uint32
	rightsErr error

	// GetLastCharacter
	lastCharacter    uint32
	lastCharacterErr error

	// GetLastLogin
	lastLogin    time.Time
	lastLoginErr error

	// GetReturnExpiry
	returnExpiry    time.Time
	returnExpiryErr error

	// UpdateReturnExpiry
	updateReturnExpiryErr    error
	updateReturnExpiryCalled bool

	// UpdateLastLogin
	updateLastLoginErr    error
	updateLastLoginCalled bool

	// CountPermanentBans
	permanentBans    int
	permanentBansErr error

	// CountActiveBans
	activeBans    int
	activeBansErr error

	// GetByWiiUKey
	wiiuUID uint32
	wiiuErr error

	// GetByPSNID
	psnUID uint32
	psnErr error

	// CountByPSNID
	psnCount    int
	psnCountErr error

	// GetPSNIDForUsername
	psnIDForUsername    string
	psnIDForUsernameErr error

	// SetPSNID
	setPSNIDErr    error
	setPSNIDCalled bool

	// GetPSNIDForUser
	psnIDForUser    string
	psnIDForUserErr error
}

func (m *mockSignUserRepo) GetCredentials(username string) (uint32, string, error) {
	return m.credUID, m.credPassword, m.credErr
}

func (m *mockSignUserRepo) Register(username, passwordHash string, returnExpires time.Time) (uint32, error) {
	m.registered = true
	return m.registerUID, m.registerErr
}

func (m *mockSignUserRepo) GetRights(uid uint32) (uint32, error) {
	return m.rights, m.rightsErr
}

func (m *mockSignUserRepo) GetLastCharacter(uid uint32) (uint32, error) {
	return m.lastCharacter, m.lastCharacterErr
}

func (m *mockSignUserRepo) GetLastLogin(uid uint32) (time.Time, error) {
	return m.lastLogin, m.lastLoginErr
}

func (m *mockSignUserRepo) GetReturnExpiry(uid uint32) (time.Time, error) {
	return m.returnExpiry, m.returnExpiryErr
}

func (m *mockSignUserRepo) UpdateReturnExpiry(uid uint32, expiry time.Time) error {
	m.updateReturnExpiryCalled = true
	return m.updateReturnExpiryErr
}

func (m *mockSignUserRepo) UpdateLastLogin(uid uint32, loginTime time.Time) error {
	m.updateLastLoginCalled = true
	return m.updateLastLoginErr
}

func (m *mockSignUserRepo) CountPermanentBans(uid uint32) (int, error) {
	return m.permanentBans, m.permanentBansErr
}

func (m *mockSignUserRepo) CountActiveBans(uid uint32) (int, error) {
	return m.activeBans, m.activeBansErr
}

func (m *mockSignUserRepo) GetByWiiUKey(wiiuKey string) (uint32, error) {
	return m.wiiuUID, m.wiiuErr
}

func (m *mockSignUserRepo) GetByPSNID(psnID string) (uint32, error) {
	return m.psnUID, m.psnErr
}

func (m *mockSignUserRepo) CountByPSNID(psnID string) (int, error) {
	return m.psnCount, m.psnCountErr
}

func (m *mockSignUserRepo) GetPSNIDForUsername(username string) (string, error) {
	return m.psnIDForUsername, m.psnIDForUsernameErr
}

func (m *mockSignUserRepo) SetPSNID(username, psnID string) error {
	m.setPSNIDCalled = true
	return m.setPSNIDErr
}

func (m *mockSignUserRepo) GetPSNIDForUser(uid uint32) (string, error) {
	return m.psnIDForUser, m.psnIDForUserErr
}

// --- mockSignCharacterRepo ---

type mockSignCharacterRepo struct {
	// CountNewCharacters
	newCharCount    int
	newCharCountErr error

	// CreateCharacter
	createErr    error
	createCalled bool

	// GetForUser
	characters    []character
	getForUserErr error

	// IsNewCharacter
	isNew    bool
	isNewErr error

	// HardDelete
	hardDeleteErr    error
	hardDeleteCalled bool

	// SoftDelete
	softDeleteErr    error
	softDeleteCalled bool

	// GetFriends
	friends       []members
	getFriendsErr error

	// GetGuildmates
	guildmates       []members
	getGuildmatesErr error
}

func (m *mockSignCharacterRepo) CountNewCharacters(uid uint32) (int, error) {
	return m.newCharCount, m.newCharCountErr
}

func (m *mockSignCharacterRepo) CreateCharacter(uid uint32, lastLogin uint32) error {
	m.createCalled = true
	return m.createErr
}

func (m *mockSignCharacterRepo) GetForUser(uid uint32) ([]character, error) {
	return m.characters, m.getForUserErr
}

func (m *mockSignCharacterRepo) IsNewCharacter(cid int) (bool, error) {
	return m.isNew, m.isNewErr
}

func (m *mockSignCharacterRepo) HardDelete(cid int) error {
	m.hardDeleteCalled = true
	return m.hardDeleteErr
}

func (m *mockSignCharacterRepo) SoftDelete(cid int) error {
	m.softDeleteCalled = true
	return m.softDeleteErr
}

func (m *mockSignCharacterRepo) GetFriends(charID uint32) ([]members, error) {
	return m.friends, m.getFriendsErr
}

func (m *mockSignCharacterRepo) GetGuildmates(charID uint32) ([]members, error) {
	return m.guildmates, m.getGuildmatesErr
}

// --- mockSignSessionRepo ---

type mockSignSessionRepo struct {
	// RegisterUID
	registerUIDTokenID uint32
	registerUIDErr     error

	// RegisterPSN
	registerPSNTokenID uint32
	registerPSNErr     error

	// Validate
	validateResult bool
	validateErr    error

	// GetPSNIDByToken
	psnIDByToken    string
	psnIDByTokenErr error
}

func (m *mockSignSessionRepo) RegisterUID(uid uint32, token string) (uint32, error) {
	return m.registerUIDTokenID, m.registerUIDErr
}

func (m *mockSignSessionRepo) RegisterPSN(psnID, token string) (uint32, error) {
	return m.registerPSNTokenID, m.registerPSNErr
}

func (m *mockSignSessionRepo) Validate(token string, tokenID uint32) (bool, error) {
	return m.validateResult, m.validateErr
}

func (m *mockSignSessionRepo) GetPSNIDByToken(token string) (string, error) {
	return m.psnIDByToken, m.psnIDByTokenErr
}

// newTestServer creates a Server with mock repos for testing.
func newTestServer(userRepo SignUserRepo, charRepo SignCharacterRepo, sessionRepo SignSessionRepo) *Server {
	return &Server{
		userRepo:    userRepo,
		charRepo:    charRepo,
		sessionRepo: sessionRepo,
	}
}
