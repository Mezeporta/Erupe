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

	adjustErr     error
	readErr       error
	saveErr       error
	loadColumnErr error

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

func (m *mockCharacterRepo) LoadColumn(_ uint32, column string) ([]byte, error) {
	if m.loadColumnErr != nil {
		return nil, m.loadColumnErr
	}
	return m.columns[column], nil
}
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

// --- mockGuildRepoOps (enhanced guild repo for ops/scout/board tests) ---

type mockGuildRepoOps struct {
	mockGuildRepoForMail

	// Configurable errors
	saveErr       error
	saveMemberErr error
	disbandErr    error
	getMembersErr error
	acceptErr     error
	rejectErr     error
	removeErr     error
	createAppErr  error
	getMemberErr  error
	hasAppResult  bool
	hasAppErr     error
	listPostsErr  error
	createPostErr error
	deletePostErr error

	// State tracking
	disbandedID    uint32
	removedCharID  uint32
	acceptedCharID uint32
	rejectedCharID uint32
	savedGuild     *Guild
	savedMembers   []*GuildMember
	createdAppArgs []interface{}
	createdPost    []interface{}
	deletedPostID  uint32

	// Alliance
	alliance          *GuildAlliance
	getAllianceErr     error
	createAllianceErr error
	deleteAllianceErr error
	removeAllyErr     error
	deletedAllianceID uint32
	removedAllyArgs   []uint32

	// Cooking
	meals         []*GuildMeal
	listMealsErr  error
	createdMealID uint32
	createMealErr error
	updateMealErr error

	// Adventure
	adventures       []*GuildAdventure
	listAdvErr       error
	createAdvErr     error
	collectAdvID     uint32
	chargeAdvID      uint32
	chargeAdvAmount  uint32

	// Treasure hunt
	pendingHunt      *TreasureHunt
	guildHunts       []*TreasureHunt
	listHuntsErr     error
	acquireHuntID    uint32
	reportHuntID     uint32
	collectHuntID    uint32
	claimHuntID      uint32
	createHuntErr    error

	// Hunt data
	guildKills       []*GuildKill
	listKillsErr     error
	countKills       int
	countKillsErr    error
	claimBoxCalled   bool

	// Data
	membership  *GuildMember
	application *GuildApplication
	posts       []*MessageBoardPost
}

func (m *mockGuildRepoOps) GetByID(guildID uint32) (*Guild, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.guild != nil && m.guild.ID == guildID {
		return m.guild, nil
	}
	return nil, errNotFound
}

func (m *mockGuildRepoOps) GetByCharID(charID uint32) (*Guild, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.guild, nil
}

func (m *mockGuildRepoOps) GetMembers(guildID uint32, applicants bool) ([]*GuildMember, error) {
	if m.getMembersErr != nil {
		return nil, m.getMembersErr
	}
	return m.members, nil
}

func (m *mockGuildRepoOps) GetCharacterMembership(_ uint32) (*GuildMember, error) {
	if m.getMemberErr != nil {
		return nil, m.getMemberErr
	}
	return m.membership, nil
}

func (m *mockGuildRepoOps) Save(guild *Guild) error {
	m.savedGuild = guild
	return m.saveErr
}

func (m *mockGuildRepoOps) SaveMember(member *GuildMember) error {
	m.savedMembers = append(m.savedMembers, member)
	return m.saveMemberErr
}

func (m *mockGuildRepoOps) Disband(guildID uint32) error {
	m.disbandedID = guildID
	return m.disbandErr
}

func (m *mockGuildRepoOps) RemoveCharacter(charID uint32) error {
	m.removedCharID = charID
	return m.removeErr
}

func (m *mockGuildRepoOps) AcceptApplication(guildID, charID uint32) error {
	m.acceptedCharID = charID
	return m.acceptErr
}

func (m *mockGuildRepoOps) RejectApplication(guildID, charID uint32) error {
	m.rejectedCharID = charID
	return m.rejectErr
}

func (m *mockGuildRepoOps) CreateApplication(guildID, charID, actorID uint32, appType GuildApplicationType) error {
	m.createdAppArgs = []interface{}{guildID, charID, actorID, appType}
	return m.createAppErr
}

func (m *mockGuildRepoOps) HasApplication(guildID, charID uint32) (bool, error) {
	return m.hasAppResult, m.hasAppErr
}

func (m *mockGuildRepoOps) GetApplication(guildID, charID uint32, appType GuildApplicationType) (*GuildApplication, error) {
	return m.application, nil
}

func (m *mockGuildRepoOps) ListPosts(guildID uint32, postType int) ([]*MessageBoardPost, error) {
	if m.listPostsErr != nil {
		return nil, m.listPostsErr
	}
	return m.posts, nil
}

func (m *mockGuildRepoOps) CreatePost(guildID, authorID, stampID uint32, postType int, title, body string, maxPosts int) error {
	m.createdPost = []interface{}{guildID, authorID, stampID, postType, title, body, maxPosts}
	return m.createPostErr
}

func (m *mockGuildRepoOps) DeletePost(postID uint32) error {
	m.deletedPostID = postID
	return m.deletePostErr
}

func (m *mockGuildRepoOps) GetAllianceByID(_ uint32) (*GuildAlliance, error) {
	return m.alliance, m.getAllianceErr
}

func (m *mockGuildRepoOps) CreateAlliance(_ string, _ uint32) error {
	return m.createAllianceErr
}

func (m *mockGuildRepoOps) DeleteAlliance(id uint32) error {
	m.deletedAllianceID = id
	return m.deleteAllianceErr
}

func (m *mockGuildRepoOps) RemoveGuildFromAlliance(allyID, guildID, sub1, sub2 uint32) error {
	m.removedAllyArgs = []uint32{allyID, guildID, sub1, sub2}
	return m.removeAllyErr
}

func (m *mockGuildRepoOps) ListMeals(_ uint32) ([]*GuildMeal, error) {
	return m.meals, m.listMealsErr
}

func (m *mockGuildRepoOps) CreateMeal(_, _, _ uint32, _ time.Time) (uint32, error) {
	return m.createdMealID, m.createMealErr
}

func (m *mockGuildRepoOps) UpdateMeal(_, _, _ uint32, _ time.Time) error {
	return m.updateMealErr
}

func (m *mockGuildRepoOps) ListAdventures(_ uint32) ([]*GuildAdventure, error) {
	return m.adventures, m.listAdvErr
}

func (m *mockGuildRepoOps) CreateAdventure(_, _ uint32, _, _ int64) error {
	return m.createAdvErr
}

func (m *mockGuildRepoOps) CreateAdventureWithCharge(_, _, _ uint32, _, _ int64) error {
	return m.createAdvErr
}

func (m *mockGuildRepoOps) CollectAdventure(id uint32, _ uint32) error {
	m.collectAdvID = id
	return nil
}

func (m *mockGuildRepoOps) ChargeAdventure(id uint32, amount uint32) error {
	m.chargeAdvID = id
	m.chargeAdvAmount = amount
	return nil
}

func (m *mockGuildRepoOps) GetPendingHunt(_ uint32) (*TreasureHunt, error) {
	return m.pendingHunt, nil
}

func (m *mockGuildRepoOps) ListGuildHunts(_, _ uint32) ([]*TreasureHunt, error) {
	return m.guildHunts, m.listHuntsErr
}

func (m *mockGuildRepoOps) CreateHunt(_, _, _, _ uint32, _ []byte, _ string) error {
	return m.createHuntErr
}

func (m *mockGuildRepoOps) AcquireHunt(id uint32) error {
	m.acquireHuntID = id
	return nil
}

func (m *mockGuildRepoOps) RegisterHuntReport(id, _ uint32) error {
	m.reportHuntID = id
	return nil
}

func (m *mockGuildRepoOps) CollectHunt(id uint32) error {
	m.collectHuntID = id
	return nil
}

func (m *mockGuildRepoOps) ClaimHuntReward(id, _ uint32) error {
	m.claimHuntID = id
	return nil
}

func (m *mockGuildRepoOps) ClaimHuntBox(_ uint32, _ time.Time) error {
	m.claimBoxCalled = true
	return nil
}

func (m *mockGuildRepoOps) ListGuildKills(_, _ uint32) ([]*GuildKill, error) {
	return m.guildKills, m.listKillsErr
}

func (m *mockGuildRepoOps) CountGuildKills(_, _ uint32) (int, error) {
	return m.countKills, m.countKillsErr
}

// --- mockUserRepoForItems ---

type mockUserRepoForItems struct {
	itemBoxData []byte
	itemBoxErr  error
	setData     []byte
}

func (m *mockUserRepoForItems) GetItemBox(_ uint32) ([]byte, error) {
	return m.itemBoxData, m.itemBoxErr
}

func (m *mockUserRepoForItems) SetItemBox(_ uint32, data []byte) error {
	m.setData = data
	return nil
}

// Stub all other UserRepo methods.
func (m *mockUserRepoForItems) GetGachaPoints(_ uint32) (uint32, uint32, uint32, error) { return 0, 0, 0, nil }
func (m *mockUserRepoForItems) GetTrialCoins(_ uint32) (uint16, error)                  { return 0, nil }
func (m *mockUserRepoForItems) DeductTrialCoins(_ uint32, _ uint32) error               { return nil }
func (m *mockUserRepoForItems) DeductPremiumCoins(_ uint32, _ uint32) error             { return nil }
func (m *mockUserRepoForItems) AddPremiumCoins(_ uint32, _ uint32) error                { return nil }
func (m *mockUserRepoForItems) AddTrialCoins(_ uint32, _ uint32) error                  { return nil }
func (m *mockUserRepoForItems) DeductFrontierPoints(_ uint32, _ uint32) error           { return nil }
func (m *mockUserRepoForItems) AddFrontierPoints(_ uint32, _ uint32) error              { return nil }
func (m *mockUserRepoForItems) AdjustFrontierPointsDeduct(_ uint32, _ int) (uint32, error) { return 0, nil }
func (m *mockUserRepoForItems) AdjustFrontierPointsCredit(_ uint32, _ int) (uint32, error) { return 0, nil }
func (m *mockUserRepoForItems) AddFrontierPointsFromGacha(_ uint32, _ uint32, _ uint8) error { return nil }
func (m *mockUserRepoForItems) GetRights(_ uint32) (uint32, error)                      { return 0, nil }
func (m *mockUserRepoForItems) SetRights(_ uint32, _ uint32) error                      { return nil }
func (m *mockUserRepoForItems) IsOp(_ uint32) (bool, error)                             { return false, nil }
func (m *mockUserRepoForItems) SetLastCharacter(_ uint32, _ uint32) error               { return nil }
func (m *mockUserRepoForItems) GetTimer(_ uint32) (bool, error)                         { return false, nil }
func (m *mockUserRepoForItems) SetTimer(_ uint32, _ bool) error                         { return nil }
func (m *mockUserRepoForItems) CountByPSNID(_ string) (int, error)                      { return 0, nil }
func (m *mockUserRepoForItems) SetPSNID(_ uint32, _ string) error                       { return nil }
func (m *mockUserRepoForItems) GetDiscordToken(_ uint32) (string, error)                { return "", nil }
func (m *mockUserRepoForItems) SetDiscordToken(_ uint32, _ string) error                { return nil }
func (m *mockUserRepoForItems) LinkDiscord(_ string, _ string) (string, error)          { return "", nil }
func (m *mockUserRepoForItems) SetPasswordByDiscordID(_ string, _ []byte) error         { return nil }
func (m *mockUserRepoForItems) GetByIDAndUsername(_ uint32) (uint32, string, error)      { return 0, "", nil }
func (m *mockUserRepoForItems) BanUser(_ uint32, _ *time.Time) error                    { return nil }

// --- mockStampRepoForItems ---

type mockStampRepoForItems struct {
	checkedTime  time.Time
	checkedErr   error
	totals       [2]uint16 // total, redeemed
	totalsErr    error
	initCalled   bool
	incrementCalled bool
	setCalled    bool
	exchangeResult [2]uint16
	exchangeErr  error
	yearlyResult [2]uint16
	yearlyErr    error
}

func (m *mockStampRepoForItems) GetChecked(_ uint32, _ string) (time.Time, error) {
	return m.checkedTime, m.checkedErr
}

func (m *mockStampRepoForItems) Init(_ uint32, _ time.Time) error {
	m.initCalled = true
	return nil
}

func (m *mockStampRepoForItems) SetChecked(_ uint32, _ string, _ time.Time) error {
	m.setCalled = true
	return nil
}

func (m *mockStampRepoForItems) IncrementTotal(_ uint32, _ string) error {
	m.incrementCalled = true
	return nil
}

func (m *mockStampRepoForItems) GetTotals(_ uint32, _ string) (uint16, uint16, error) {
	return m.totals[0], m.totals[1], m.totalsErr
}

func (m *mockStampRepoForItems) ExchangeYearly(_ uint32) (uint16, uint16, error) {
	return m.yearlyResult[0], m.yearlyResult[1], m.yearlyErr
}

func (m *mockStampRepoForItems) Exchange(_ uint32, _ string) (uint16, uint16, error) {
	return m.exchangeResult[0], m.exchangeResult[1], m.exchangeErr
}

// --- mockHouseRepoForItems ---

type mockHouseRepoForItems struct {
	warehouseItems map[uint8][]byte
	setData        map[uint8][]byte
	setErr         error
}

func newMockHouseRepoForItems() *mockHouseRepoForItems {
	return &mockHouseRepoForItems{
		warehouseItems: make(map[uint8][]byte),
		setData:        make(map[uint8][]byte),
	}
}

func (m *mockHouseRepoForItems) GetWarehouseItemData(_ uint32, index uint8) ([]byte, error) {
	return m.warehouseItems[index], nil
}

func (m *mockHouseRepoForItems) SetWarehouseItemData(_ uint32, index uint8, data []byte) error {
	m.setData[index] = data
	return m.setErr
}

func (m *mockHouseRepoForItems) InitializeWarehouse(_ uint32) error                           { return nil }

// Stub all other HouseRepo methods.
func (m *mockHouseRepoForItems) UpdateInterior(_ uint32, _ []byte) error                      { return nil }
func (m *mockHouseRepoForItems) GetHouseByCharID(_ uint32) (HouseData, error)                 { return HouseData{}, nil }
func (m *mockHouseRepoForItems) SearchHousesByName(_ string) ([]HouseData, error)              { return nil, nil }
func (m *mockHouseRepoForItems) UpdateHouseState(_ uint32, _ uint8, _ string) error            { return nil }
func (m *mockHouseRepoForItems) GetHouseAccess(_ uint32) (uint8, string, error)                { return 0, "", nil }
func (m *mockHouseRepoForItems) GetHouseContents(_ uint32) ([]byte, []byte, []byte, []byte, []byte, []byte, []byte, error) {
	return nil, nil, nil, nil, nil, nil, nil, nil
}
func (m *mockHouseRepoForItems) GetMission(_ uint32) ([]byte, error)                          { return nil, nil }
func (m *mockHouseRepoForItems) UpdateMission(_ uint32, _ []byte) error                       { return nil }
func (m *mockHouseRepoForItems) GetWarehouseNames(_ uint32) ([10]string, [10]string, error) {
	return [10]string{}, [10]string{}, nil
}
func (m *mockHouseRepoForItems) RenameWarehouseBox(_ uint32, _ uint8, _ uint8, _ string) error { return nil }
func (m *mockHouseRepoForItems) GetWarehouseEquipData(_ uint32, _ uint8) ([]byte, error)       { return nil, nil }
func (m *mockHouseRepoForItems) SetWarehouseEquipData(_ uint32, _ uint8, _ []byte) error       { return nil }
func (m *mockHouseRepoForItems) GetTitles(_ uint32) ([]Title, error)                           { return nil, nil }
func (m *mockHouseRepoForItems) AcquireTitle(_ uint16, _ uint32) error                         { return nil }

// --- mockSessionRepo ---

type mockSessionRepo struct {
	validateErr error
	bindErr     error
	clearErr    error
	updateErr   error

	boundToken  string
	clearedToken string
}

func (m *mockSessionRepo) ValidateLoginToken(_ string, _ uint32, _ uint32) error { return m.validateErr }
func (m *mockSessionRepo) BindSession(token string, _ uint16, _ uint32) error {
	m.boundToken = token
	return m.bindErr
}
func (m *mockSessionRepo) ClearSession(token string) error {
	m.clearedToken = token
	return m.clearErr
}
func (m *mockSessionRepo) UpdatePlayerCount(_ uint16, _ int) error { return m.updateErr }

// --- mockGachaRepo ---

type mockGachaRepo struct {
	// GetEntryForTransaction
	txItemType   uint8
	txItemNumber uint16
	txRolls      int
	txErr        error

	// GetRewardPool
	rewardPool    []GachaEntry
	rewardPoolErr error

	// GetItemsForEntry
	entryItems    map[uint32][]GachaItem
	entryItemsErr error

	// GetGuaranteedItems
	guaranteedItems []GachaItem

	// Stepup
	stepupStep    uint8
	stepupTime    time.Time
	stepupErr     error
	hasEntryType  bool
	deletedStepup bool
	insertedStep  uint8

	// Box
	boxEntryIDs    []uint32
	boxEntryIDsErr error
	insertedBoxIDs []uint32
	deletedBox     bool

	// Shop
	gachas       []Gacha
	listShopErr  error
	shopType     int
	allEntries   []GachaEntry
	allEntriesErr error
	weightDivisor float64

	// FrontierPoints from gacha
	addFPErr error
}

func (m *mockGachaRepo) GetEntryForTransaction(_ uint32, _ uint8) (uint8, uint16, int, error) {
	return m.txItemType, m.txItemNumber, m.txRolls, m.txErr
}
func (m *mockGachaRepo) GetRewardPool(_ uint32) ([]GachaEntry, error) {
	return m.rewardPool, m.rewardPoolErr
}
func (m *mockGachaRepo) GetItemsForEntry(entryID uint32) ([]GachaItem, error) {
	if m.entryItemsErr != nil {
		return nil, m.entryItemsErr
	}
	if m.entryItems != nil {
		return m.entryItems[entryID], nil
	}
	return nil, nil
}
func (m *mockGachaRepo) GetGuaranteedItems(_ uint8, _ uint32) ([]GachaItem, error) {
	return m.guaranteedItems, nil
}
func (m *mockGachaRepo) GetStepupStep(_ uint32, _ uint32) (uint8, error) {
	return m.stepupStep, m.stepupErr
}
func (m *mockGachaRepo) GetStepupWithTime(_ uint32, _ uint32) (uint8, time.Time, error) {
	return m.stepupStep, m.stepupTime, m.stepupErr
}
func (m *mockGachaRepo) HasEntryType(_ uint32, _ uint8) (bool, error) {
	return m.hasEntryType, nil
}
func (m *mockGachaRepo) DeleteStepup(_ uint32, _ uint32) error {
	m.deletedStepup = true
	return nil
}
func (m *mockGachaRepo) InsertStepup(_ uint32, step uint8, _ uint32) error {
	m.insertedStep = step
	return nil
}
func (m *mockGachaRepo) GetBoxEntryIDs(_ uint32, _ uint32) ([]uint32, error) {
	return m.boxEntryIDs, m.boxEntryIDsErr
}
func (m *mockGachaRepo) InsertBoxEntry(_ uint32, entryID uint32, _ uint32) error {
	m.insertedBoxIDs = append(m.insertedBoxIDs, entryID)
	return nil
}
func (m *mockGachaRepo) DeleteBoxEntries(_ uint32, _ uint32) error {
	m.deletedBox = true
	return nil
}
func (m *mockGachaRepo) ListShop() ([]Gacha, error)               { return m.gachas, m.listShopErr }
func (m *mockGachaRepo) GetShopType(_ uint32) (int, error)         { return m.shopType, nil }
func (m *mockGachaRepo) GetAllEntries(_ uint32) ([]GachaEntry, error) {
	return m.allEntries, m.allEntriesErr
}
func (m *mockGachaRepo) GetWeightDivisor(_ uint32) (float64, error) { return m.weightDivisor, nil }

// --- mockShopRepo ---

type mockShopRepo struct {
	shopItems       []ShopItem
	shopItemsErr    error
	purchases       []shopPurchaseRecord
	recordErr       error
	fpointQuantity  int
	fpointValue     int
	fpointItemErr   error
	fpointExchanges []FPointExchange
}

type shopPurchaseRecord struct {
	charID, itemHash, quantity uint32
}

func (m *mockShopRepo) GetShopItems(_ uint8, _ uint32, _ uint32) ([]ShopItem, error) {
	return m.shopItems, m.shopItemsErr
}
func (m *mockShopRepo) RecordPurchase(charID, itemHash, quantity uint32) error {
	m.purchases = append(m.purchases, shopPurchaseRecord{charID, itemHash, quantity})
	return m.recordErr
}
func (m *mockShopRepo) GetFpointItem(_ uint32) (int, int, error) {
	return m.fpointQuantity, m.fpointValue, m.fpointItemErr
}
func (m *mockShopRepo) GetFpointExchangeList() ([]FPointExchange, error) {
	return m.fpointExchanges, nil
}

// --- mockUserRepoGacha (UserRepo with configurable gacha fields) ---

type mockUserRepoGacha struct {
	mockUserRepoForItems

	gachaFP, gachaGP, gachaGT uint32
	trialCoins                 uint16
	deductTrialErr             error
	deductPremiumErr           error
	deductFPErr                error
	addFPFromGachaErr          error

	fpDeductBalance uint32
	fpDeductErr     error
	fpCreditBalance uint32
	fpCreditErr     error

	setLastCharErr error
	rights         uint32
	rightsErr      error
}

func (m *mockUserRepoGacha) GetGachaPoints(_ uint32) (uint32, uint32, uint32, error) {
	return m.gachaFP, m.gachaGP, m.gachaGT, nil
}
func (m *mockUserRepoGacha) GetTrialCoins(_ uint32) (uint16, error)    { return m.trialCoins, nil }
func (m *mockUserRepoGacha) DeductTrialCoins(_ uint32, _ uint32) error { return m.deductTrialErr }
func (m *mockUserRepoGacha) DeductPremiumCoins(_ uint32, _ uint32) error {
	return m.deductPremiumErr
}
func (m *mockUserRepoGacha) DeductFrontierPoints(_ uint32, _ uint32) error { return m.deductFPErr }
func (m *mockUserRepoGacha) AddFrontierPointsFromGacha(_ uint32, _ uint32, _ uint8) error {
	return m.addFPFromGachaErr
}
func (m *mockUserRepoGacha) AdjustFrontierPointsDeduct(_ uint32, _ int) (uint32, error) {
	return m.fpDeductBalance, m.fpDeductErr
}
func (m *mockUserRepoGacha) AdjustFrontierPointsCredit(_ uint32, _ int) (uint32, error) {
	return m.fpCreditBalance, m.fpCreditErr
}
func (m *mockUserRepoGacha) SetLastCharacter(_ uint32, _ uint32) error { return m.setLastCharErr }
func (m *mockUserRepoGacha) GetRights(_ uint32) (uint32, error)       { return m.rights, m.rightsErr }
