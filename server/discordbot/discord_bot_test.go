package discordbot

import (
	"errors"
	cfg "erupe-ce/config"
	"regexp"
	"testing"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// mockSession implements the Session interface for testing.
type mockSession struct {
	openErr               error
	channelResult         *discordgo.Channel
	channelErr            error
	userResults           map[string]*discordgo.User
	userErr               error
	messageSentTo         string
	messageSentContent    string
	messageErr            error
	addHandlerCalls       int
	bulkOverwriteAppID    string
	bulkOverwriteCommands []*discordgo.ApplicationCommand
	bulkOverwriteErr      error
}

func (m *mockSession) Open() error {
	return m.openErr
}

func (m *mockSession) Channel(_ string, _ ...discordgo.RequestOption) (*discordgo.Channel, error) {
	return m.channelResult, m.channelErr
}

func (m *mockSession) User(userID string, _ ...discordgo.RequestOption) (*discordgo.User, error) {
	if m.userResults != nil {
		if u, ok := m.userResults[userID]; ok {
			return u, nil
		}
	}
	return nil, m.userErr
}

func (m *mockSession) ChannelMessageSend(channelID string, content string, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.messageSentTo = channelID
	m.messageSentContent = content
	return &discordgo.Message{}, m.messageErr
}

func (m *mockSession) AddHandler(_ interface{}) func() {
	m.addHandlerCalls++
	return func() {}
}

func (m *mockSession) ApplicationCommandBulkOverwrite(appID string, _ string, commands []*discordgo.ApplicationCommand, _ ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	m.bulkOverwriteAppID = appID
	m.bulkOverwriteCommands = commands
	return commands, m.bulkOverwriteErr
}

func newTestBot(session *mockSession) *DiscordBot {
	return &DiscordBot{
		Session: session,
		config:  &cfg.Config{},
		logger:  zap.NewNop(),
	}
}

func TestStart_Success(t *testing.T) {
	ms := &mockSession{}
	bot := newTestBot(ms)

	if err := bot.Start(); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
}

func TestStart_OpenError(t *testing.T) {
	ms := &mockSession{openErr: errors.New("connection refused")}
	bot := newTestBot(ms)

	err := bot.Start()
	if err == nil {
		t.Fatal("Start() expected error, got nil")
	}
	if err.Error() != "connection refused" {
		t.Errorf("Start() error = %q, want %q", err.Error(), "connection refused")
	}
}

func TestNormalizeDiscordMessage(t *testing.T) {
	tests := []struct {
		name     string
		users    map[string]*discordgo.User
		userErr  error
		message  string
		expected string
	}{
		{
			name: "replace user mention with username",
			users: map[string]*discordgo.User{
				"123456789012345678": {Username: "TestUser"},
			},
			message:  "Hello <@123456789012345678>!",
			expected: "Hello @TestUser!",
		},
		{
			name: "replace nickname mention",
			users: map[string]*discordgo.User{
				"123456789012345678": {Username: "NickUser"},
			},
			message:  "Hello <@!123456789012345678>!",
			expected: "Hello @NickUser!",
		},
		{
			name:     "unknown user fallback",
			userErr:  errors.New("not found"),
			message:  "Hello <@123456789012345678>!",
			expected: "Hello @unknown!",
		},
		{
			name:     "simple emoji preserved",
			message:  "Hello :smile:!",
			expected: "Hello :smile:!",
		},
		{
			name:     "custom emoji normalized",
			message:  "Nice <:custom:123456789012345678>",
			expected: "Nice :custom:",
		},
		{
			name:     "animated emoji normalized",
			message:  "Fun <a:dance:123456789012345678>",
			expected: "Fun :dance:",
		},
		{
			name: "mixed mentions and emoji",
			users: map[string]*discordgo.User{
				"111111111111111111": {Username: "Alice"},
			},
			message:  "<@111111111111111111> says :wave:",
			expected: "@Alice says :wave:",
		},
		{
			name:     "plain text unchanged",
			message:  "Hello World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &mockSession{
				userResults: tt.users,
				userErr:     tt.userErr,
			}
			bot := newTestBot(ms)
			result := bot.NormalizeDiscordMessage(tt.message)
			if result != tt.expected {
				t.Errorf("NormalizeDiscordMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRealtimeChannelSend_NilRelayChannel(t *testing.T) {
	ms := &mockSession{}
	bot := newTestBot(ms)
	bot.RelayChannel = nil

	if err := bot.RealtimeChannelSend("test"); err != nil {
		t.Fatalf("RealtimeChannelSend() unexpected error: %v", err)
	}
	if ms.messageSentTo != "" {
		t.Error("RealtimeChannelSend() should not send when RelayChannel is nil")
	}
}

func TestRealtimeChannelSend_Success(t *testing.T) {
	ms := &mockSession{}
	bot := newTestBot(ms)
	bot.RelayChannel = &discordgo.Channel{ID: "chan123"}

	if err := bot.RealtimeChannelSend("hello"); err != nil {
		t.Fatalf("RealtimeChannelSend() unexpected error: %v", err)
	}
	if ms.messageSentTo != "chan123" {
		t.Errorf("sent to channel %q, want %q", ms.messageSentTo, "chan123")
	}
	if ms.messageSentContent != "hello" {
		t.Errorf("sent content %q, want %q", ms.messageSentContent, "hello")
	}
}

func TestRealtimeChannelSend_Error(t *testing.T) {
	ms := &mockSession{messageErr: errors.New("send failed")}
	bot := newTestBot(ms)
	bot.RelayChannel = &discordgo.Channel{ID: "chan123"}

	err := bot.RealtimeChannelSend("hello")
	if err == nil {
		t.Fatal("RealtimeChannelSend() expected error, got nil")
	}
	if err.Error() != "send failed" {
		t.Errorf("error = %q, want %q", err.Error(), "send failed")
	}
}

func TestRegisterCommands_Success(t *testing.T) {
	ms := &mockSession{}
	bot := newTestBot(ms)
	bot.userID = "bot123"

	if err := bot.RegisterCommands(); err != nil {
		t.Fatalf("RegisterCommands() unexpected error: %v", err)
	}
	if ms.bulkOverwriteAppID != "bot123" {
		t.Errorf("appID = %q, want %q", ms.bulkOverwriteAppID, "bot123")
	}
	if len(ms.bulkOverwriteCommands) != len(Commands) {
		t.Errorf("commands count = %d, want %d", len(ms.bulkOverwriteCommands), len(Commands))
	}
}

func TestRegisterCommands_Error(t *testing.T) {
	ms := &mockSession{bulkOverwriteErr: errors.New("forbidden")}
	bot := newTestBot(ms)

	err := bot.RegisterCommands()
	if err == nil {
		t.Fatal("RegisterCommands() expected error, got nil")
	}
}

func TestAddHandler(t *testing.T) {
	ms := &mockSession{}
	bot := newTestBot(ms)

	cleanup := bot.AddHandler(func() {})
	if cleanup == nil {
		t.Fatal("AddHandler() returned nil cleanup func")
	}
	if ms.addHandlerCalls != 1 {
		t.Errorf("addHandlerCalls = %d, want 1", ms.addHandlerCalls)
	}
}

func TestUserID(t *testing.T) {
	bot := &DiscordBot{userID: "abc123"}
	if bot.UserID() != "abc123" {
		t.Errorf("UserID() = %q, want %q", bot.UserID(), "abc123")
	}
}

func TestCommands_Structure(t *testing.T) {
	if len(Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(Commands))
	}

	expectedNames := []string{"link", "password"}
	for i, name := range expectedNames {
		if Commands[i].Name != name {
			t.Errorf("Commands[%d].Name = %q, want %q", i, Commands[i].Name, name)
		}
		if Commands[i].Description == "" {
			t.Errorf("Commands[%d] (%s) has empty description", i, name)
		}
		if len(Commands[i].Options) == 0 {
			t.Errorf("Commands[%d] (%s) has no options", i, name)
		}
		for _, opt := range Commands[i].Options {
			if !opt.Required {
				t.Errorf("Commands[%d] (%s) option %q should be required", i, name, opt.Name)
			}
		}
	}
}

func TestReplaceTextAll(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		regex    *regexp.Regexp
		handler  func(string) string
		expected string
	}{
		{
			name:  "replace single match",
			text:  "Hello @123456789012345678",
			regex: regexp.MustCompile(`@(\d+)`),
			handler: func(id string) string {
				return "@user_" + id
			},
			expected: "Hello @user_123456789012345678",
		},
		{
			name:  "replace multiple matches",
			text:  "Users @111 and @222",
			regex: regexp.MustCompile(`@(\d+)`),
			handler: func(id string) string {
				return "@user_" + id
			},
			expected: "Users @user_111 and @user_222",
		},
		{
			name:  "no matches",
			text:  "Hello World",
			regex: regexp.MustCompile(`@(\d+)`),
			handler: func(id string) string {
				return "@user_" + id
			},
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceTextAll(tt.text, tt.regex, tt.handler)
			if result != tt.expected {
				t.Errorf("ReplaceTextAll() = %q, want %q", result, tt.expected)
			}
		})
	}
}
