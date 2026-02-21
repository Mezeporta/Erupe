package channelserver

import (
	"errors"
	"time"
)

// errNotFound is a sentinel for mock repos that simulate "not found".
var errNotFound = errors.New("not found")

// --- mockAchievementRepo ---

type mockAchievementRepo struct {
	scores        [33]int32
	ensureCalled  bool
	ensureErr     error
	getScoresErr  error
	incrementErr  error
	incrementedID uint8
}

func (m *mockAchievementRepo) EnsureExists(_ uint32) error {
	m.ensureCalled = true
	return m.ensureErr
}

func (m *mockAchievementRepo) GetAllScores(_ uint32) ([33]int32, error) {
	return m.scores, m.getScoresErr
}

func (m *mockAchievementRepo) IncrementScore(_ uint32, id uint8) error {
	m.incrementedID = id
	return m.incrementErr
}

// --- mockMailRepo ---

type mockMailRepo struct {
	mails          []Mail
	mailByID       map[int]*Mail
	listErr        error
	getByIDErr     error
	markReadCalled int
	markDeletedID  int
	lockID         int
	lockValue      bool
	itemReceivedID int
	sentMails      []sentMailRecord
	sendErr        error
}

type sentMailRecord struct {
	senderID, recipientID          uint32
	subject, body                  string
	itemID, itemAmount             uint16
	isGuildInvite, isSystemMessage bool
}

func (m *mockMailRepo) GetListForCharacter(_ uint32) ([]Mail, error) {
	return m.mails, m.listErr
}

func (m *mockMailRepo) GetByID(id int) (*Mail, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if mail, ok := m.mailByID[id]; ok {
		return mail, nil
	}
	return nil, errNotFound
}

func (m *mockMailRepo) MarkRead(id int) error {
	m.markReadCalled = id
	return nil
}

func (m *mockMailRepo) MarkDeleted(id int) error {
	m.markDeletedID = id
	return nil
}

func (m *mockMailRepo) SetLocked(id int, locked bool) error {
	m.lockID = id
	m.lockValue = locked
	return nil
}

func (m *mockMailRepo) MarkItemReceived(id int) error {
	m.itemReceivedID = id
	return nil
}

func (m *mockMailRepo) SendMail(senderID, recipientID uint32, subject, body string, itemID, itemAmount uint16, isGuildInvite, isSystemMessage bool) error {
	m.sentMails = append(m.sentMails, sentMailRecord{
		senderID: senderID, recipientID: recipientID,
		subject: subject, body: body,
		itemID: itemID, itemAmount: itemAmount,
		isGuildInvite: isGuildInvite, isSystemMessage: isSystemMessage,
	})
	return m.sendErr
}

// --- mockCharacterRepo ---

type mockCharacterRepo struct {
	ints    map[string]int
	times   map[string]time.Time
	columns map[string][]byte
	strings map[string]string
	bools   map[string]bool

	adjustErr error
	readErr   error
	saveErr   error

	// LoadSaveData mock fields
	loadSaveDataID   uint32
	loadSaveDataData []byte
	loadSaveDataNew  bool
	loadSaveDataName string
	loadSaveDataErr  error
}

func newMockCharacterRepo() *mockCharacterRepo {
	return &mockCharacterRepo{
		ints:    make(map[string]int),
		times:   make(map[string]time.Time),
		columns: make(map[string][]byte),
		strings: make(map[string]string),
		bools:   make(map[string]bool),
	}
}

func (m *mockCharacterRepo) ReadInt(_ uint32, column string) (int, error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	return m.ints[column], nil
}

func (m *mockCharacterRepo) AdjustInt(_ uint32, column string, delta int) (int, error) {
	if m.adjustErr != nil {
		return 0, m.adjustErr
	}
	m.ints[column] += delta
	return m.ints[column], nil
}

func (m *mockCharacterRepo) SaveInt(_ uint32, column string, value int) error {
	m.ints[column] = value
	return m.saveErr
}

func (m *mockCharacterRepo) ReadTime(_ uint32, column string, defaultVal time.Time) (time.Time, error) {
	if m.readErr != nil {
		return defaultVal, m.readErr
	}
	if t, ok := m.times[column]; ok {
		return t, nil
	}
	return defaultVal, errNotFound
}

func (m *mockCharacterRepo) SaveTime(_ uint32, column string, value time.Time) error {
	m.times[column] = value
	return m.saveErr
}

func (m *mockCharacterRepo) LoadColumn(_ uint32, column string) ([]byte, error)     { return m.columns[column], nil }
func (m *mockCharacterRepo) SaveColumn(_ uint32, column string, data []byte) error   { m.columns[column] = data; return m.saveErr }
func (m *mockCharacterRepo) GetName(_ uint32) (string, error)                        { return "TestChar", nil }
func (m *mockCharacterRepo) GetUserID(_ uint32) (uint32, error)                      { return 1, nil }
func (m *mockCharacterRepo) UpdateLastLogin(_ uint32, _ int64) error                 { return nil }
func (m *mockCharacterRepo) UpdateTimePlayed(_ uint32, _ int) error                  { return nil }
func (m *mockCharacterRepo) GetCharIDsByUserID(_ uint32) ([]uint32, error)           { return nil, nil }
func (m *mockCharacterRepo) SaveBool(_ uint32, col string, v bool) error             { m.bools[col] = v; return nil }
func (m *mockCharacterRepo) SaveString(_ uint32, col string, v string) error         { m.strings[col] = v; return nil }
func (m *mockCharacterRepo) ReadBool(_ uint32, col string) (bool, error)             { return m.bools[col], nil }
func (m *mockCharacterRepo) ReadString(_ uint32, col string) (string, error)         { return m.strings[col], nil }
func (m *mockCharacterRepo) LoadColumnWithDefault(_ uint32, col string, def []byte) ([]byte, error) {
	if d, ok := m.columns[col]; ok {
		return d, nil
	}
	return def, nil
}
func (m *mockCharacterRepo) SetDeleted(_ uint32) error                                           { return nil }
func (m *mockCharacterRepo) UpdateDailyCafe(_ uint32, _ time.Time, _, _ uint32) error            { return nil }
func (m *mockCharacterRepo) ResetDailyQuests(_ uint32) error                                     { return nil }
func (m *mockCharacterRepo) ReadEtcPoints(_ uint32) (uint32, uint32, uint32, error)              { return 0, 0, 0, nil }
func (m *mockCharacterRepo) ResetCafeTime(_ uint32, _ time.Time) error                           { return nil }
func (m *mockCharacterRepo) UpdateGuildPostChecked(_ uint32) error                               { return nil }
func (m *mockCharacterRepo) ReadGuildPostChecked(_ uint32) (time.Time, error)                    { return time.Time{}, nil }
func (m *mockCharacterRepo) SaveMercenary(_ uint32, _ []byte, _ uint32) error                    { return nil }
func (m *mockCharacterRepo) UpdateGCPAndPact(_ uint32, _ uint32, _ uint32) error                 { return nil }
func (m *mockCharacterRepo) FindByRastaID(_ int) (uint32, string, error)                         { return 0, "", nil }
func (m *mockCharacterRepo) SaveCharacterData(_ uint32, _ []byte, _, _ uint16, _ bool, _ uint8, _ uint16) error { return nil }
func (m *mockCharacterRepo) SaveHouseData(_ uint32, _ []byte, _, _, _, _, _ []byte) error        { return nil }
func (m *mockCharacterRepo) LoadSaveData(_ uint32) (uint32, []byte, bool, string, error) {
	return m.loadSaveDataID, m.loadSaveDataData, m.loadSaveDataNew, m.loadSaveDataName, m.loadSaveDataErr
}

// --- mockGoocooRepo ---

type mockGoocooRepo struct {
	slots         map[uint32][]byte
	ensureCalled  bool
	clearCalled   []uint32
	savedSlots    map[uint32][]byte
}

func newMockGoocooRepo() *mockGoocooRepo {
	return &mockGoocooRepo{
		slots:      make(map[uint32][]byte),
		savedSlots: make(map[uint32][]byte),
	}
}

func (m *mockGoocooRepo) EnsureExists(_ uint32) error {
	m.ensureCalled = true
	return nil
}

func (m *mockGoocooRepo) GetSlot(_ uint32, slot uint32) ([]byte, error) {
	if data, ok := m.slots[slot]; ok {
		return data, nil
	}
	return nil, nil
}

func (m *mockGoocooRepo) ClearSlot(_ uint32, slot uint32) error {
	m.clearCalled = append(m.clearCalled, slot)
	delete(m.slots, slot)
	return nil
}

func (m *mockGoocooRepo) SaveSlot(_ uint32, slot uint32, data []byte) error {
	m.savedSlots[slot] = data
	return nil
}

// --- mockGuildRepo (minimal, for SendMail guild path) ---

type mockGuildRepoForMail struct {
	guild   *Guild
	members []*GuildMember
	getErr  error
}

func (m *mockGuildRepoForMail) GetByCharID(_ uint32) (*Guild, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.guild, nil
}

func (m *mockGuildRepoForMail) GetMembers(_ uint32, _ bool) ([]*GuildMember, error) {
	return m.members, nil
}

// Stub out all other GuildRepo methods.
func (m *mockGuildRepoForMail) GetByID(_ uint32) (*Guild, error)       { return nil, errNotFound }
func (m *mockGuildRepoForMail) ListAll() ([]*Guild, error)             { return nil, nil }
func (m *mockGuildRepoForMail) Create(_ uint32, _ string) (int32, error) { return 0, nil }
func (m *mockGuildRepoForMail) Save(_ *Guild) error                    { return nil }
func (m *mockGuildRepoForMail) Disband(_ uint32) error                 { return nil }
func (m *mockGuildRepoForMail) RemoveCharacter(_ uint32) error         { return nil }
func (m *mockGuildRepoForMail) AcceptApplication(_, _ uint32) error    { return nil }
func (m *mockGuildRepoForMail) CreateApplication(_, _, _ uint32, _ GuildApplicationType) error {
	return nil
}
func (m *mockGuildRepoForMail) CreateApplicationWithMail(_, _, _ uint32, _ GuildApplicationType, _, _ uint32, _, _ string) error {
	return nil
}
func (m *mockGuildRepoForMail) CancelInvitation(_, _ uint32) error              { return nil }
func (m *mockGuildRepoForMail) RejectApplication(_, _ uint32) error             { return nil }
func (m *mockGuildRepoForMail) ArrangeCharacters(_ []uint32) error              { return nil }
func (m *mockGuildRepoForMail) GetApplication(_, _ uint32, _ GuildApplicationType) (*GuildApplication, error) {
	return nil, nil
}
func (m *mockGuildRepoForMail) HasApplication(_, _ uint32) (bool, error)        { return false, nil }
func (m *mockGuildRepoForMail) GetItemBox(_ uint32) ([]byte, error)             { return nil, nil }
func (m *mockGuildRepoForMail) SaveItemBox(_ uint32, _ []byte) error            { return nil }
func (m *mockGuildRepoForMail) GetCharacterMembership(_ uint32) (*GuildMember, error) { return nil, nil }
func (m *mockGuildRepoForMail) SaveMember(_ *GuildMember) error                 { return nil }
func (m *mockGuildRepoForMail) SetRecruiting(_ uint32, _ bool) error            { return nil }
func (m *mockGuildRepoForMail) SetPugiOutfits(_ uint32, _ uint32) error         { return nil }
func (m *mockGuildRepoForMail) SetRecruiter(_ uint32, _ bool) error             { return nil }
func (m *mockGuildRepoForMail) AddMemberDailyRP(_ uint32, _ uint16) error       { return nil }
func (m *mockGuildRepoForMail) ExchangeEventRP(_ uint32, _ uint16) (uint32, error) { return 0, nil }
func (m *mockGuildRepoForMail) AddRankRP(_ uint32, _ uint16) error              { return nil }
func (m *mockGuildRepoForMail) AddEventRP(_ uint32, _ uint16) error             { return nil }
func (m *mockGuildRepoForMail) GetRoomRP(_ uint32) (uint16, error)              { return 0, nil }
func (m *mockGuildRepoForMail) SetRoomRP(_ uint32, _ uint16) error              { return nil }
func (m *mockGuildRepoForMail) AddRoomRP(_ uint32, _ uint16) error              { return nil }
func (m *mockGuildRepoForMail) SetRoomExpiry(_ uint32, _ time.Time) error       { return nil }
func (m *mockGuildRepoForMail) ListPosts(_ uint32, _ int) ([]*MessageBoardPost, error) { return nil, nil }
func (m *mockGuildRepoForMail) CreatePost(_, _, _ uint32, _ int, _, _ string, _ int) error { return nil }
func (m *mockGuildRepoForMail) DeletePost(_ uint32) error                       { return nil }
func (m *mockGuildRepoForMail) UpdatePost(_ uint32, _, _ string) error          { return nil }
func (m *mockGuildRepoForMail) UpdatePostStamp(_, _ uint32) error               { return nil }
func (m *mockGuildRepoForMail) GetPostLikedBy(_ uint32) (string, error)         { return "", nil }
func (m *mockGuildRepoForMail) SetPostLikedBy(_ uint32, _ string) error         { return nil }
func (m *mockGuildRepoForMail) CountNewPosts(_ uint32, _ time.Time) (int, error) { return 0, nil }
func (m *mockGuildRepoForMail) GetAllianceByID(_ uint32) (*GuildAlliance, error) { return nil, nil }
func (m *mockGuildRepoForMail) ListAlliances() ([]*GuildAlliance, error)        { return nil, nil }
func (m *mockGuildRepoForMail) CreateAlliance(_ string, _ uint32) error         { return nil }
func (m *mockGuildRepoForMail) DeleteAlliance(_ uint32) error                   { return nil }
func (m *mockGuildRepoForMail) RemoveGuildFromAlliance(_, _, _, _ uint32) error  { return nil }
func (m *mockGuildRepoForMail) ListAdventures(_ uint32) ([]*GuildAdventure, error) { return nil, nil }
func (m *mockGuildRepoForMail) CreateAdventure(_, _ uint32, _, _ int64) error   { return nil }
func (m *mockGuildRepoForMail) CreateAdventureWithCharge(_, _, _ uint32, _, _ int64) error { return nil }
func (m *mockGuildRepoForMail) CollectAdventure(_ uint32, _ uint32) error       { return nil }
func (m *mockGuildRepoForMail) ChargeAdventure(_ uint32, _ uint32) error        { return nil }
func (m *mockGuildRepoForMail) GetPendingHunt(_ uint32) (*TreasureHunt, error)  { return nil, nil }
func (m *mockGuildRepoForMail) ListGuildHunts(_, _ uint32) ([]*TreasureHunt, error) { return nil, nil }
func (m *mockGuildRepoForMail) CreateHunt(_, _, _, _ uint32, _ []byte, _ string) error { return nil }
func (m *mockGuildRepoForMail) AcquireHunt(_ uint32) error                      { return nil }
func (m *mockGuildRepoForMail) RegisterHuntReport(_, _ uint32) error            { return nil }
func (m *mockGuildRepoForMail) CollectHunt(_ uint32) error                      { return nil }
func (m *mockGuildRepoForMail) ClaimHuntReward(_, _ uint32) error               { return nil }
func (m *mockGuildRepoForMail) ListMeals(_ uint32) ([]*GuildMeal, error)        { return nil, nil }
func (m *mockGuildRepoForMail) CreateMeal(_, _, _ uint32, _ time.Time) (uint32, error) { return 0, nil }
func (m *mockGuildRepoForMail) UpdateMeal(_, _, _ uint32, _ time.Time) error    { return nil }
func (m *mockGuildRepoForMail) ClaimHuntBox(_ uint32, _ time.Time) error        { return nil }
func (m *mockGuildRepoForMail) ListGuildKills(_, _ uint32) ([]*GuildKill, error) { return nil, nil }
func (m *mockGuildRepoForMail) CountGuildKills(_, _ uint32) (int, error)        { return 0, nil }
func (m *mockGuildRepoForMail) ClearTreasureHunt(_ uint32) error                { return nil }
func (m *mockGuildRepoForMail) InsertKillLog(_ uint32, _ int, _ uint8, _ time.Time) error { return nil }
func (m *mockGuildRepoForMail) ListInvitedCharacters(_ uint32) ([]*ScoutedCharacter, error) { return nil, nil }
func (m *mockGuildRepoForMail) RolloverDailyRP(_ uint32, _ time.Time) error     { return nil }
func (m *mockGuildRepoForMail) AddWeeklyBonusUsers(_ uint32, _ uint8) error     { return nil }
