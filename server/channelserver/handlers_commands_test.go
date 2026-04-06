package channelserver

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"erupe-ce/common/mhfcourse"
	cfg "erupe-ce/config"
	"erupe-ce/network/clientctx"

	"go.uber.org/zap"
)

// syncOnceForTest returns a fresh sync.Once to reset the package-level commandsOnce.
func syncOnceForTest() sync.Once { return sync.Once{} }

// --- mockUserRepoCommands ---

type mockUserRepoCommands struct {
	mockUserRepoGacha

	opResult bool

	// Ban
	bannedUID uint32
	banExpiry *time.Time
	banErr    error
	foundUID  uint32
	foundName string
	findErr   error

	// Timer
	timerState     bool
	timerSetCalled bool
	timerNewState  bool

	// PSN
	psnCount int
	psnSetID string

	// Discord
	discordToken  string
	discordGetErr error
	discordSetTok string

	// Rights
	rightsVal    uint32
	setRightsVal uint32

	// Language
	langStored    string
	langSetCalled bool
	langSetErr    error
}

func (m *mockUserRepoCommands) IsOp(_ uint32) (bool, error) { return m.opResult, nil }
func (m *mockUserRepoCommands) GetByIDAndUsername(_ uint32) (uint32, string, error) {
	return m.foundUID, m.foundName, m.findErr
}
func (m *mockUserRepoCommands) BanUser(uid uint32, exp *time.Time) error {
	m.bannedUID = uid
	m.banExpiry = exp
	return m.banErr
}
func (m *mockUserRepoCommands) GetTimer(_ uint32) (bool, error) { return m.timerState, nil }
func (m *mockUserRepoCommands) SetTimer(_ uint32, v bool) error {
	m.timerSetCalled = true
	m.timerNewState = v
	return nil
}
func (m *mockUserRepoCommands) CountByPSNID(_ string) (int, error) { return m.psnCount, nil }
func (m *mockUserRepoCommands) SetPSNID(_ uint32, id string) error {
	m.psnSetID = id
	return nil
}
func (m *mockUserRepoCommands) GetDiscordToken(_ uint32) (string, error) {
	return m.discordToken, m.discordGetErr
}
func (m *mockUserRepoCommands) SetDiscordToken(_ uint32, tok string) error {
	m.discordSetTok = tok
	return nil
}
func (m *mockUserRepoCommands) GetRights(_ uint32) (uint32, error) { return m.rightsVal, nil }
func (m *mockUserRepoCommands) SetRights(_ uint32, v uint32) error {
	m.setRightsVal = v
	return nil
}
func (m *mockUserRepoCommands) GetLanguage(_ uint32) (string, error) { return m.langStored, nil }
func (m *mockUserRepoCommands) SetLanguage(_ uint32, lang string) error {
	m.langSetCalled = true
	m.langStored = lang
	return m.langSetErr
}

// --- helpers ---

func setupCommandsMap(allEnabled bool) {
	commands = map[string]cfg.Command{
		"Ban":      {Name: "Ban", Prefix: "ban", Enabled: allEnabled},
		"Timer":    {Name: "Timer", Prefix: "timer", Enabled: allEnabled},
		"PSN":      {Name: "PSN", Prefix: "psn", Enabled: allEnabled},
		"Reload":   {Name: "Reload", Prefix: "reload", Enabled: allEnabled},
		"KeyQuest": {Name: "KeyQuest", Prefix: "kqf", Enabled: allEnabled},
		"Rights":   {Name: "Rights", Prefix: "rights", Enabled: allEnabled},
		"Course":   {Name: "Course", Prefix: "course", Enabled: allEnabled},
		"Raviente": {Name: "Raviente", Prefix: "ravi", Enabled: allEnabled},
		"Teleport": {Name: "Teleport", Prefix: "tp", Enabled: allEnabled},
		"Discord":  {Name: "Discord", Prefix: "discord", Enabled: allEnabled},
		"Playtime": {Name: "Playtime", Prefix: "playtime", Enabled: allEnabled},
		"Help":     {Name: "Help", Prefix: "help", Enabled: allEnabled},
		"Language": {Name: "Language", Prefix: "lang", Enabled: allEnabled},
	}
}

func createCommandSession(repo *mockUserRepoCommands) *Session {
	server := createMockServer()
	server.erupeConfig.CommandPrefix = "!"
	server.userRepo = repo
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(1, server)
	session.userID = 1
	return session
}

func drainChatResponses(s *Session) int {
	count := 0
	for {
		select {
		case <-s.sendPackets:
			count++
		default:
			return count
		}
	}
}

// --- Timer ---

func TestParseChatCommand_Timer_TogglesOn(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{timerState: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!timer")

	if !repo.timerSetCalled {
		t.Fatal("SetTimer should be called")
	}
	if !repo.timerNewState {
		t.Error("timer should toggle from false to true")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Timer_TogglesOff(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{timerState: true}
	s := createCommandSession(repo)

	parseChatCommand(s, "!timer")

	if !repo.timerSetCalled {
		t.Fatal("SetTimer should be called")
	}
	if repo.timerNewState {
		t.Error("timer should toggle from true to false")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Timer_DisabledNonOp(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!timer")

	if repo.timerSetCalled {
		t.Error("SetTimer should not be called when disabled for non-op")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

func TestParseChatCommand_DisabledCommand_OpCanStillUse(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: true, timerState: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!timer")

	if !repo.timerSetCalled {
		t.Error("op should be able to use disabled commands")
	}
}

// --- PSN ---

func TestParseChatCommand_PSN_Success(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{psnCount: 0}
	s := createCommandSession(repo)

	parseChatCommand(s, "!psn MyPSNID")

	if repo.psnSetID != "MyPSNID" {
		t.Errorf("PSN ID = %q, want %q", repo.psnSetID, "MyPSNID")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_PSN_AlreadyExists(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{psnCount: 1}
	s := createCommandSession(repo)

	parseChatCommand(s, "!psn TakenID")

	if repo.psnSetID != "" {
		t.Error("PSN should not be set when ID already exists")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_PSN_MissingArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!psn")

	if repo.psnSetID != "" {
		t.Error("PSN should not be set with missing args")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- Rights ---

func TestParseChatCommand_Rights_Success(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!rights 30")

	if repo.setRightsVal != 30 {
		t.Errorf("rights = %d, want 30", repo.setRightsVal)
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Rights_MissingArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!rights")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- Discord ---

func TestParseChatCommand_Discord_ExistingToken(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{discordToken: "abc-123"}
	s := createCommandSession(repo)

	parseChatCommand(s, "!discord")

	if repo.discordSetTok != "" {
		t.Error("should not generate new token when existing one found")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Discord_NewToken(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{discordGetErr: errors.New("not found")}
	s := createCommandSession(repo)

	parseChatCommand(s, "!discord")

	if repo.discordSetTok == "" {
		t.Error("should generate and set a new token")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- Playtime ---

func TestParseChatCommand_Playtime(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.playtime = 3661 // 1h 1m 1s
	s.playtimeTime = time.Now()

	parseChatCommand(s, "!playtime")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- Help ---

func TestParseChatCommand_Help_ListsCommands(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!help")

	count := drainChatResponses(s)
	if count != len(commands) {
		t.Errorf("help messages = %d, want %d (one per enabled command)", count, len(commands))
	}
}

// --- Ban ---

func TestParseChatCommand_Ban_Success(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{
		opResult:  true,
		foundUID:  42,
		foundName: "TestUser",
	}
	s := createCommandSession(repo)

	// "211111" converts to CID 1 via ConvertCID (char '2' = value 1)
	parseChatCommand(s, "!ban 211111")

	if repo.bannedUID != 42 {
		t.Errorf("banned UID = %d, want 42", repo.bannedUID)
	}
	if repo.banExpiry != nil {
		t.Error("expiry should be nil for permanent ban")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Ban_WithDuration(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{
		opResult:  true,
		foundUID:  42,
		foundName: "TestUser",
	}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ban 211111 30d")

	if repo.bannedUID != 42 {
		t.Errorf("banned UID = %d, want 42", repo.bannedUID)
	}
	if repo.banExpiry == nil {
		t.Fatal("expiry should not be nil for timed ban")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Ban_NonOp(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ban 211111")

	if repo.bannedUID != 0 {
		t.Error("non-op should not be able to ban")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (noOp message)", n)
	}
}

func TestParseChatCommand_Ban_InvalidCID(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{opResult: true}
	s := createCommandSession(repo)

	// "abc" is not 6 chars, ConvertCID returns 0
	parseChatCommand(s, "!ban abc")

	if repo.bannedUID != 0 {
		t.Error("invalid CID should not result in a ban")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Ban_UserNotFound(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{
		opResult: true,
		findErr:  errors.New("not found"),
	}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ban 211111")

	if repo.bannedUID != 0 {
		t.Error("should not ban when user not found")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (noUser message)", n)
	}
}

func TestParseChatCommand_Ban_MissingArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{opResult: true}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ban")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Ban_DurationUnits(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"seconds", "10s"},
		{"minutes", "5m"},
		{"hours", "2h"},
		{"days", "30d"},
		{"months", "6mo"},
		{"years", "1y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupCommandsMap(true)
			repo := &mockUserRepoCommands{
				opResult:  true,
				foundUID:  1,
				foundName: "User",
			}
			s := createCommandSession(repo)

			parseChatCommand(s, "!ban 211111 "+tt.duration)

			if repo.banExpiry == nil {
				t.Errorf("expiry should not be nil for duration %s", tt.duration)
			}
		})
	}
}

// --- Teleport ---

func TestParseChatCommand_Teleport_Success(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!tp 100 200")

	// Teleport sends a CastedBinary + a chat message = 2 packets
	if n := drainChatResponses(s); n != 2 {
		t.Errorf("packets = %d, want 2 (teleport + message)", n)
	}
}

func TestParseChatCommand_Teleport_MissingArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!tp 100")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- KeyQuest ---

func TestParseChatCommand_KeyQuest_VersionCheck(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.RealClientMode = cfg.S6 // below G10

	parseChatCommand(s, "!kqf get")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (version error)", n)
	}
}

func TestParseChatCommand_KeyQuest_Get(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.RealClientMode = cfg.ZZ
	s.kqf = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	parseChatCommand(s, "!kqf get")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_KeyQuest_Set(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.RealClientMode = cfg.ZZ

	parseChatCommand(s, "!kqf set 0102030405060708")

	if !s.kqfOverride {
		t.Error("kqfOverride should be true after set")
	}
	if len(s.kqf) != 8 {
		t.Errorf("kqf length = %d, want 8", len(s.kqf))
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_KeyQuest_SetInvalid(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.RealClientMode = cfg.ZZ

	parseChatCommand(s, "!kqf set ABC") // not 16 hex chars

	if s.kqfOverride {
		t.Error("kqfOverride should not be set with invalid hex")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (error message)", n)
	}
}

// --- Raviente ---

func TestParseChatCommand_Raviente_NoSemaphore(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ravi start")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (noPlayers message)", n)
	}
}

func TestParseChatCommand_Raviente_MissingArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ravi")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (error message)", n)
	}
}

// --- Course ---

func TestParseChatCommand_Course_MissingArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!course")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (error message)", n)
	}
}

// --- Ban (additional) ---

func TestParseChatCommand_Ban_InvalidDurationFormat(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{opResult: true}
	s := createCommandSession(repo)

	// "30x" has an unparseable format — Sscanf fails
	parseChatCommand(s, "!ban 211111 badformat")

	if repo.bannedUID != 0 {
		t.Error("should not ban with invalid duration format")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (error message)", n)
	}
}

func TestParseChatCommand_Ban_BanUserError(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{
		opResult:  true,
		foundUID:  42,
		foundName: "TestUser",
		banErr:    errors.New("db error"),
	}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ban 211111")

	// Ban is attempted (bannedUID set by mock) but returns error.
	// The handler still sends a success message — it logs the error.
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Ban_WithExpiryBanError(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{
		opResult:  true,
		foundUID:  42,
		foundName: "TestUser",
		banErr:    errors.New("db error"),
	}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ban 211111 7d")

	// Even with error, handler sends success message (logs the error)
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Ban_DurationLongForm(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"seconds_long", "10seconds"},
		{"second_singular", "1second"},
		{"minutes_long", "5minutes"},
		{"minute_singular", "1minute"},
		{"hours_long", "2hours"},
		{"hour_singular", "1hour"},
		{"days_long", "30days"},
		{"day_singular", "1day"},
		{"months_long", "6months"},
		{"month_singular", "1month"},
		{"years_long", "2years"},
		{"year_singular", "1year"},
		{"mi_alias", "15mi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupCommandsMap(true)
			repo := &mockUserRepoCommands{
				opResult:  true,
				foundUID:  1,
				foundName: "User",
			}
			s := createCommandSession(repo)

			parseChatCommand(s, "!ban 211111 "+tt.duration)

			if repo.banExpiry == nil {
				t.Errorf("expiry should not be nil for duration %s", tt.duration)
			}
		})
	}
}

// --- Raviente (with semaphore) ---

// addRaviSemaphore sets up a Raviente semaphore on the server so getRaviSemaphore() returns non-nil.
func addRaviSemaphore(s *Server) {
	s.semaphore = map[string]*Semaphore{
		"hs_l0u3": {name: "hs_l0u3", clients: make(map[*Session]uint32)},
	}
}

func TestParseChatCommand_Raviente_StartSuccess(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	addRaviSemaphore(s.server)
	s.server.raviente.register[1] = 0
	s.server.raviente.register[3] = 100

	parseChatCommand(s, "!ravi start")

	if s.server.raviente.register[1] != 100 {
		t.Errorf("register[1] = %d, want 100", s.server.raviente.register[1])
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Raviente_StartAlreadyStarted(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	addRaviSemaphore(s.server)
	s.server.raviente.register[1] = 50 // already started

	parseChatCommand(s, "!ravi start")

	if s.server.raviente.register[1] != 50 {
		t.Errorf("register[1] should remain 50, got %d", s.server.raviente.register[1])
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (already started error)", n)
	}
}

func TestParseChatCommand_Raviente_CheckMultiplier(t *testing.T) {
	for _, alias := range []string{"cm", "check", "checkmultiplier", "multiplier"} {
		t.Run(alias, func(t *testing.T) {
			setupCommandsMap(true)
			repo := &mockUserRepoCommands{}
			s := createCommandSession(repo)
			addRaviSemaphore(s.server)
			// Add a client to the semaphore to avoid divide-by-zero in GetRaviMultiplier
			sema := s.server.getRaviSemaphore()
			sema.clients[s] = s.charID

			parseChatCommand(s, "!ravi "+alias)

			if n := drainChatResponses(s); n != 1 {
				t.Errorf("chat responses = %d, want 1", n)
			}
		})
	}
}

func TestParseChatCommand_Raviente_ZZCommands(t *testing.T) {
	tests := []struct {
		name    string
		aliases []string
	}{
		{"sendres", []string{"sr", "sendres", "resurrection"}},
		{"sendsed", []string{"ss", "sendsed"}},
		{"reqsed", []string{"rs", "reqsed"}},
	}

	for _, tt := range tests {
		for _, alias := range tt.aliases {
			t.Run(tt.name+"/"+alias, func(t *testing.T) {
				setupCommandsMap(true)
				repo := &mockUserRepoCommands{}
				s := createCommandSession(repo)
				addRaviSemaphore(s.server)
				s.server.erupeConfig.RealClientMode = cfg.ZZ
				// Set up HP for sendsed/reqsed
				s.server.raviente.state[0] = 100
				s.server.raviente.state[28] = 1 // res support available

				parseChatCommand(s, "!ravi "+alias)

				if n := drainChatResponses(s); n != 1 {
					t.Errorf("chat responses = %d, want 1", n)
				}
			})
		}
	}
}

func TestParseChatCommand_Raviente_ZZCommand_ResNoSupport(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	addRaviSemaphore(s.server)
	s.server.erupeConfig.RealClientMode = cfg.ZZ
	s.server.raviente.state[28] = 0 // no support available

	parseChatCommand(s, "!ravi sr")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (res error)", n)
	}
}

func TestParseChatCommand_Raviente_NonZZVersion(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	addRaviSemaphore(s.server)
	s.server.erupeConfig.RealClientMode = cfg.G10

	parseChatCommand(s, "!ravi sr")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (version error)", n)
	}
}

func TestParseChatCommand_Raviente_UnknownSubcommand(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	addRaviSemaphore(s.server)

	parseChatCommand(s, "!ravi unknown")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (error message)", n)
	}
}

func TestParseChatCommand_Raviente_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!ravi start")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Course (additional) ---

func TestParseChatCommand_Course_EnableCourse(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{rightsVal: 0}
	s := createCommandSession(repo)
	// "Trial" is alias for course ID 1; config must list it as enabled
	s.server.erupeConfig.Courses = []cfg.Course{{Name: "Trial", Enabled: true}}

	parseChatCommand(s, "!course Trial")

	if repo.setRightsVal == 0 {
		t.Error("rights should be updated when enabling a course")
	}
	// 1 chat message (enabled) + 1 updateRights packet = 2
	if n := drainChatResponses(s); n != 2 {
		t.Errorf("packets = %d, want 2 (course enabled message + rights update)", n)
	}
}

func TestParseChatCommand_Course_DisableCourse(t *testing.T) {
	setupCommandsMap(true)
	// Rights value = 2 means course ID 1 is active (2^1 = 2)
	repo := &mockUserRepoCommands{rightsVal: 2}
	s := createCommandSession(repo)
	s.server.erupeConfig.Courses = []cfg.Course{{Name: "Trial", Enabled: true}}
	// Pre-populate session courses so CourseExists returns true
	s.courses = []mhfcourse.Course{{ID: 1}}

	parseChatCommand(s, "!course Trial")

	// 1 chat message (disabled) + 1 updateRights packet = 2
	if n := drainChatResponses(s); n != 2 {
		t.Errorf("packets = %d, want 2 (course disabled message + rights update)", n)
	}
}

func TestParseChatCommand_Course_CaseInsensitive(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{rightsVal: 0}
	s := createCommandSession(repo)
	s.server.erupeConfig.Courses = []cfg.Course{{Name: "Trial", Enabled: true}}

	parseChatCommand(s, "!course trial")

	if repo.setRightsVal == 0 {
		t.Error("course lookup should be case-insensitive")
	}
}

func TestParseChatCommand_Course_AliasLookup(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{rightsVal: 0}
	s := createCommandSession(repo)
	s.server.erupeConfig.Courses = []cfg.Course{{Name: "Trial", Enabled: true}}

	// "TL" is an alias for Trial (course ID 1)
	parseChatCommand(s, "!course TL")

	if repo.setRightsVal == 0 {
		t.Error("course should be found by alias")
	}
}

func TestParseChatCommand_Course_Locked(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	// Course exists in game but NOT in config (or disabled in config)
	s.server.erupeConfig.Courses = []cfg.Course{}

	parseChatCommand(s, "!course Trial")

	// Should get "locked" message, no rights update
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (locked message)", n)
	}
}

func TestParseChatCommand_Course_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!course Trial")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Reload ---

func TestParseChatCommand_Reload_EmptyStage(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.stage = &Stage{
		id:      "test",
		objects: make(map[uint32]*Object),
		clients: make(map[*Session]uint32),
	}

	parseChatCommand(s, "!reload")

	// With no other sessions/objects: 1 chat message + 2 queue sends (delete + insert notifs)
	if n := drainChatResponses(s); n < 1 {
		t.Errorf("packets = %d, want >= 1", n)
	}
}

func TestParseChatCommand_Reload_WithOtherPlayersAndObjects(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	// Create another session in the server
	otherLogger, _ := zap.NewDevelopment()
	other := &Session{
		charID:        2,
		clientContext: &clientctx.ClientContext{},
		sendPackets:   make(chan packet, 20),
		server:        s.server,
		logger:        otherLogger,
	}
	s.server.sessions[&net.TCPConn{}] = other

	// Stage with an object owned by the other session
	s.stage = &Stage{
		id: "test",
		objects: map[uint32]*Object{
			1: {id: 1, ownerCharID: 2, x: 1.0, y: 2.0, z: 3.0},
			2: {id: 2, ownerCharID: s.charID}, // our own object — should be skipped
		},
		clients: map[*Session]uint32{s: s.charID, other: 2},
	}

	parseChatCommand(s, "!reload")

	// Should get: chat message + delete notif + reload notif (3 packets)
	if n := drainChatResponses(s); n != 3 {
		t.Errorf("packets = %d, want 3 (chat + delete + reload)", n)
	}
}

func TestParseChatCommand_Reload_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!reload")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Help (additional) ---

func TestParseChatCommand_Help_NonOpSeesOnlyEnabled(t *testing.T) {
	// Set up: only some commands enabled, user is not op
	commands = map[string]cfg.Command{
		"Ban":      {Name: "Ban", Prefix: "ban", Enabled: false},
		"Timer":    {Name: "Timer", Prefix: "timer", Enabled: true},
		"PSN":      {Name: "PSN", Prefix: "psn", Enabled: true},
		"Reload":   {Name: "Reload", Prefix: "reload", Enabled: false},
		"KeyQuest": {Name: "KeyQuest", Prefix: "kqf", Enabled: false},
		"Rights":   {Name: "Rights", Prefix: "rights", Enabled: false},
		"Course":   {Name: "Course", Prefix: "course", Enabled: true},
		"Raviente": {Name: "Raviente", Prefix: "ravi", Enabled: false},
		"Teleport": {Name: "Teleport", Prefix: "tp", Enabled: false},
		"Discord":  {Name: "Discord", Prefix: "discord", Enabled: true},
		"Playtime": {Name: "Playtime", Prefix: "playtime", Enabled: true},
		"Help":     {Name: "Help", Prefix: "help", Enabled: true},
	}
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!help")

	// Count enabled commands
	enabled := 0
	for _, cmd := range commands {
		if cmd.Enabled {
			enabled++
		}
	}

	count := drainChatResponses(s)
	if count != enabled {
		t.Errorf("help messages = %d, want %d (only enabled commands for non-op)", count, enabled)
	}
}

func TestParseChatCommand_Help_OpSeesAll(t *testing.T) {
	// Some disabled, but op sees all
	commands = map[string]cfg.Command{
		"Ban":      {Name: "Ban", Prefix: "ban", Enabled: false},
		"Timer":    {Name: "Timer", Prefix: "timer", Enabled: true},
		"PSN":      {Name: "PSN", Prefix: "psn", Enabled: false},
		"Reload":   {Name: "Reload", Prefix: "reload", Enabled: false},
		"KeyQuest": {Name: "KeyQuest", Prefix: "kqf", Enabled: false},
		"Rights":   {Name: "Rights", Prefix: "rights", Enabled: false},
		"Course":   {Name: "Course", Prefix: "course", Enabled: false},
		"Raviente": {Name: "Raviente", Prefix: "ravi", Enabled: false},
		"Teleport": {Name: "Teleport", Prefix: "tp", Enabled: false},
		"Discord":  {Name: "Discord", Prefix: "discord", Enabled: false},
		"Playtime": {Name: "Playtime", Prefix: "playtime", Enabled: false},
		"Help":     {Name: "Help", Prefix: "help", Enabled: true},
	}
	repo := &mockUserRepoCommands{opResult: true}
	s := createCommandSession(repo)

	parseChatCommand(s, "!help")

	count := drainChatResponses(s)
	if count != len(commands) {
		t.Errorf("help messages = %d, want %d (op sees all commands)", count, len(commands))
	}
}

func TestParseChatCommand_Help_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!help")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Rights (additional) ---

func TestParseChatCommand_Rights_SetRightsError(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	// Use a value that Atoi will parse but SetRights succeeds (no error mock needed here)
	// Instead test the "invalid" case: non-numeric argument
	parseChatCommand(s, "!rights notanumber")

	// Atoi("notanumber") returns 0 — SetRights(0) succeeds, sends success message
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

func TestParseChatCommand_Rights_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!rights 30")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Teleport (additional) ---

func TestParseChatCommand_Teleport_NoArgs(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	parseChatCommand(s, "!tp")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (error message)", n)
	}
}

func TestParseChatCommand_Teleport_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!tp 100 200")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- KeyQuest (additional) ---

func TestParseChatCommand_KeyQuest_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!kqf get")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- PSN (additional) ---

func TestParseChatCommand_PSN_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!psn MyPSNID")

	if repo.psnSetID != "" {
		t.Error("PSN should not be set when command is disabled")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Discord (additional) ---

func TestParseChatCommand_Discord_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)

	parseChatCommand(s, "!discord")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- Playtime (additional) ---

func TestParseChatCommand_Playtime_Disabled(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)
	s.playtimeTime = time.Now()

	parseChatCommand(s, "!playtime")

	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

// --- initCommands ---

func TestInitCommands(t *testing.T) {
	// Reset the sync.Once by replacing the package-level vars
	commandsOnce = syncOnceForTest()
	commands = nil

	logger, _ := zap.NewDevelopment()
	cmds := []cfg.Command{
		{Name: "TestCmd", Prefix: "test", Enabled: true},
		{Name: "Disabled", Prefix: "dis", Enabled: false},
	}

	initCommands(cmds, logger)

	if len(commands) != 2 {
		t.Fatalf("commands length = %d, want 2", len(commands))
	}
	if commands["TestCmd"].Prefix != "test" {
		t.Errorf("TestCmd prefix = %q, want %q", commands["TestCmd"].Prefix, "test")
	}
	if commands["Disabled"].Enabled {
		t.Error("Disabled command should not be enabled")
	}
}

// --- sendServerChatMessage ---

func TestSendServerChatMessage_CommandsContext(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	sendServerChatMessage(session, "Hello, World!")

	if n := drainChatResponses(session); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- sendDisabledCommandMessage ---

func TestSendDisabledCommandMessage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	sendDisabledCommandMessage(session, cfg.Command{Name: "TestCmd"})

	if n := drainChatResponses(session); n != 1 {
		t.Errorf("chat responses = %d, want 1", n)
	}
}

// --- Language ---

func TestParseChatCommand_Lang_ShowsUsageWhenNoArg(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.Language = "en"

	parseChatCommand(s, "!lang")

	if repo.langSetCalled {
		t.Error("SetLanguage should not be called when no argument provided")
	}
	// Two messages: current + usage.
	if n := drainChatResponses(s); n != 2 {
		t.Errorf("chat responses = %d, want 2 (current + usage)", n)
	}
}

func TestParseChatCommand_Lang_ValidCodePersistsAndSwitches(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.Language = "en"

	parseChatCommand(s, "!lang fr")

	if !repo.langSetCalled {
		t.Fatal("SetLanguage should be called")
	}
	if repo.langStored != "fr" {
		t.Errorf("langStored = %q, want %q", repo.langStored, "fr")
	}
	if got := s.Lang(); got != "fr" {
		t.Errorf("session Lang() = %q, want %q", got, "fr")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (success)", n)
	}
}

func TestParseChatCommand_Lang_InvalidCodeRejected(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.Language = "en"

	parseChatCommand(s, "!lang klingon")

	if repo.langSetCalled {
		t.Error("SetLanguage should not be called for unknown code")
	}
	if got := s.Lang(); got != "en" {
		t.Errorf("session Lang() = %q, want unchanged %q", got, "en")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (invalid message)", n)
	}
}

func TestParseChatCommand_Lang_DisabledNonOp(t *testing.T) {
	setupCommandsMap(false)
	repo := &mockUserRepoCommands{opResult: false}
	s := createCommandSession(repo)
	s.server.erupeConfig.Language = "en"

	parseChatCommand(s, "!lang fr")

	if repo.langSetCalled {
		t.Error("SetLanguage should not be called when command is disabled for non-op")
	}
	if got := s.Lang(); got != "en" {
		t.Errorf("session Lang() = %q, want unchanged %q", got, "en")
	}
	if n := drainChatResponses(s); n != 1 {
		t.Errorf("chat responses = %d, want 1 (disabled message)", n)
	}
}

func TestParseChatCommand_Lang_UppercaseNormalized(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)
	s.server.erupeConfig.Language = "en"

	parseChatCommand(s, "!lang FR")

	if repo.langStored != "fr" {
		t.Errorf("langStored = %q, want %q (should be lowercased)", repo.langStored, "fr")
	}
}

// --- Unknown command ---

func TestParseChatCommand_UnknownCommand(t *testing.T) {
	setupCommandsMap(true)
	repo := &mockUserRepoCommands{}
	s := createCommandSession(repo)

	// Command that doesn't match any registered prefix — should be a no-op
	parseChatCommand(s, "!nonexistent")

	if n := drainChatResponses(s); n != 0 {
		t.Errorf("chat responses = %d, want 0 (unknown command is silent)", n)
	}
}
