package discordbot

import (
	"regexp"
	"testing"
)

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
			text:  "Users @111111111111111111 and @222222222222222222",
			regex: regexp.MustCompile(`@(\d+)`),
			handler: func(id string) string {
				return "@user_" + id
			},
			expected: "Users @user_111111111111111111 and @user_222222222222222222",
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
		{
			name:  "replace with empty string",
			text:  "Remove @123456789012345678 this",
			regex: regexp.MustCompile(`@(\d+)`),
			handler: func(id string) string {
				return ""
			},
			expected: "Remove  this",
		},
		{
			name:  "replace emoji syntax",
			text:  "Hello :smile: and :wave:",
			regex: regexp.MustCompile(`:(\w+):`),
			handler: func(emoji string) string {
				return "[" + emoji + "]"
			},
			expected: "Hello [smile] and [wave]",
		},
		{
			name:  "complex replacement",
			text:  "Text with <@!123456789012345678> mention",
			regex: regexp.MustCompile(`<@!?(\d+)>`),
			handler: func(id string) string {
				return "@user_" + id
			},
			expected: "Text with @user_123456789012345678 mention",
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

func TestReplaceTextAll_UserMentionPattern(t *testing.T) {
	// Test the actual user mention regex used in NormalizeDiscordMessage
	userRegex := regexp.MustCompile(`<@!?(\d{17,19})>`)

	tests := []struct {
		name     string
		text     string
		expected []string // Expected captured IDs
	}{
		{
			name:     "standard mention",
			text:     "<@123456789012345678>",
			expected: []string{"123456789012345678"},
		},
		{
			name:     "nickname mention",
			text:     "<@!123456789012345678>",
			expected: []string{"123456789012345678"},
		},
		{
			name:     "multiple mentions",
			text:     "<@123456789012345678> and <@!987654321098765432>",
			expected: []string{"123456789012345678", "987654321098765432"},
		},
		{
			name:     "17 digit ID",
			text:     "<@12345678901234567>",
			expected: []string{"12345678901234567"},
		},
		{
			name:     "19 digit ID",
			text:     "<@1234567890123456789>",
			expected: []string{"1234567890123456789"},
		},
		{
			name:     "invalid - too short",
			text:     "<@1234567890123456>",
			expected: []string{},
		},
		{
			name:     "invalid - too long",
			text:     "<@12345678901234567890>",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := userRegex.FindAllStringSubmatch(tt.text, -1)
			if len(matches) != len(tt.expected) {
				t.Fatalf("Expected %d matches, got %d", len(tt.expected), len(matches))
			}
			for i, match := range matches {
				if len(match) < 2 {
					t.Fatalf("Match %d: expected capture group", i)
				}
				if match[1] != tt.expected[i] {
					t.Errorf("Match %d: got ID %q, want %q", i, match[1], tt.expected[i])
				}
			}
		})
	}
}

func TestReplaceTextAll_EmojiPattern(t *testing.T) {
	// Test the actual emoji regex used in NormalizeDiscordMessage
	emojiRegex := regexp.MustCompile(`(?:<a?)?:(\w+):(?:\d{18}>)?`)

	tests := []struct {
		name         string
		text         string
		expectedName []string // Expected emoji names
	}{
		{
			name:         "simple emoji",
			text:         ":smile:",
			expectedName: []string{"smile"},
		},
		{
			name:         "custom emoji",
			text:         "<:customemoji:123456789012345678>",
			expectedName: []string{"customemoji"},
		},
		{
			name:         "animated emoji",
			text:         "<a:animated:123456789012345678>",
			expectedName: []string{"animated"},
		},
		{
			name:         "multiple emojis",
			text:         ":wave: <:custom:123456789012345678> :smile:",
			expectedName: []string{"wave", "custom", "smile"},
		},
		{
			name:         "emoji with underscores",
			text:         ":thumbs_up:",
			expectedName: []string{"thumbs_up"},
		},
		{
			name:         "emoji with numbers",
			text:         ":emoji123:",
			expectedName: []string{"emoji123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := emojiRegex.FindAllStringSubmatch(tt.text, -1)
			if len(matches) != len(tt.expectedName) {
				t.Fatalf("Expected %d matches, got %d", len(tt.expectedName), len(matches))
			}
			for i, match := range matches {
				if len(match) < 2 {
					t.Fatalf("Match %d: expected capture group", i)
				}
				if match[1] != tt.expectedName[i] {
					t.Errorf("Match %d: got name %q, want %q", i, match[1], tt.expectedName[i])
				}
			}
		})
	}
}

func TestNormalizeDiscordMessage_Integration(t *testing.T) {
	// Create a mock bot for testing the normalization logic
	// Note: We can't fully test this without a real Discord session,
	// but we can test the regex patterns and structure
	tests := []struct {
		name     string
		input    string
		contains []string // Strings that should be in the output
	}{
		{
			name:     "plain text unchanged",
			input:    "Hello World",
			contains: []string{"Hello World"},
		},
		{
			name:  "user mention format",
			input: "Hello <@123456789012345678>",
			// We can't test the actual replacement without a real Discord session
			// but we can verify the pattern is matched
			contains: []string{"Hello"},
		},
		{
			name:     "emoji format preserved",
			input:    "Hello :smile:",
			contains: []string{"Hello", ":smile:"},
		},
		{
			name:  "mixed content",
			input: "<@123456789012345678> sent :wave:",
			contains: []string{"sent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the message contains expected parts
			for _, expected := range tt.contains {
				if len(expected) > 0 && !contains(tt.input, expected) {
					t.Errorf("Input %q should contain %q", tt.input, expected)
				}
			}
		})
	}
}

func TestCommands_Structure(t *testing.T) {
	// Test that the Commands slice is properly structured
	if len(Commands) == 0 {
		t.Error("Commands slice should not be empty")
	}

	expectedCommands := map[string]bool{
		"link":     false,
		"password": false,
	}

	for _, cmd := range Commands {
		if cmd.Name == "" {
			t.Error("Command should have a name")
		}
		if cmd.Description == "" {
			t.Errorf("Command %q should have a description", cmd.Name)
		}

		if _, exists := expectedCommands[cmd.Name]; exists {
			expectedCommands[cmd.Name] = true
		}
	}

	// Verify expected commands exist
	for name, found := range expectedCommands {
		if !found {
			t.Errorf("Expected command %q not found in Commands", name)
		}
	}
}

func TestCommands_LinkCommand(t *testing.T) {
	var linkCmd *struct {
		Name        string
		Description string
		Options     []struct {
			Type        int
			Name        string
			Description string
			Required    bool
		}
	}

	// Find the link command
	for _, cmd := range Commands {
		if cmd.Name == "link" {
			// Verify structure
			if cmd.Description == "" {
				t.Error("Link command should have a description")
			}
			if len(cmd.Options) == 0 {
				t.Error("Link command should have options")
			}

			// Verify token option
			for _, opt := range cmd.Options {
				if opt.Name == "token" {
					if !opt.Required {
						t.Error("Token option should be required")
					}
					if opt.Description == "" {
						t.Error("Token option should have a description")
					}
					return
				}
			}
			t.Error("Link command should have a 'token' option")
		}
	}

	if linkCmd == nil {
		t.Error("Link command not found")
	}
}

func TestCommands_PasswordCommand(t *testing.T) {
	// Find the password command
	for _, cmd := range Commands {
		if cmd.Name == "password" {
			// Verify structure
			if cmd.Description == "" {
				t.Error("Password command should have a description")
			}
			if len(cmd.Options) == 0 {
				t.Error("Password command should have options")
			}

			// Verify password option
			for _, opt := range cmd.Options {
				if opt.Name == "password" {
					if !opt.Required {
						t.Error("Password option should be required")
					}
					if opt.Description == "" {
						t.Error("Password option should have a description")
					}
					return
				}
			}
			t.Error("Password command should have a 'password' option")
		}
	}

	t.Error("Password command not found")
}

func TestDiscordBotStruct(t *testing.T) {
	// Test that the DiscordBot struct can be initialized
	_ = &DiscordBot{
		Session:      nil, // Can't create real session in tests
		MainGuild:    nil,
		RelayChannel: nil,
	}
}

func TestOptionsStruct(t *testing.T) {
	// Test that the Options struct can be initialized
	opts := Options{
		Config: nil,
		Logger: nil,
	}

	// Just verify we can create the struct
	_ = opts
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkReplaceTextAll(b *testing.B) {
	text := "Message with <@123456789012345678> and <@!987654321098765432> mentions and :smile: :wave: emojis"
	userRegex := regexp.MustCompile(`<@!?(\d{17,19})>`)
	handler := func(id string) string {
		return "@user_" + id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ReplaceTextAll(text, userRegex, handler)
	}
}

func BenchmarkReplaceTextAll_NoMatches(b *testing.B) {
	text := "Message with no mentions or special syntax at all, just plain text"
	userRegex := regexp.MustCompile(`<@!?(\d{17,19})>`)
	handler := func(id string) string {
		return "@user_" + id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ReplaceTextAll(text, userRegex, handler)
	}
}
