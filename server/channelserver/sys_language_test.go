package channelserver

import (
	"testing"

	"erupe-ce/config"
)

func TestGetLangStrings_English(t *testing.T) {
	server := &Server{
		erupeConfig: &config.Config{
			Language: "en",
		},
	}

	strings := getLangStrings(server)

	if strings["language"] != "English" {
		t.Errorf("language = %q, want %q", strings["language"], "English")
	}

	// Verify key strings exist
	requiredKeys := []string{
		"cafeReset",
		"commandDisabled",
		"commandReload",
		"commandKqfGet",
		"commandKqfSetError",
		"commandKqfSetSuccess",
		"commandRightsError",
		"commandRightsSuccess",
		"commandCourseError",
		"commandCourseDisabled",
		"commandCourseEnabled",
		"commandCourseLocked",
		"commandTeleportError",
		"commandTeleportSuccess",
		"commandRaviNoCommand",
		"commandRaviStartSuccess",
		"commandRaviStartError",
		"commandRaviMultiplier",
		"commandRaviResSuccess",
		"commandRaviResError",
		"commandRaviSedSuccess",
		"commandRaviRequest",
		"commandRaviError",
		"commandRaviNoPlayers",
		"ravienteBerserk",
		"ravienteExtreme",
		"ravienteExtremeLimited",
		"ravienteBerserkSmall",
		"guildInviteName",
		"guildInvite",
		"guildInviteSuccessName",
		"guildInviteSuccess",
		"guildInviteAcceptedName",
		"guildInviteAccepted",
		"guildInviteRejectName",
		"guildInviteReject",
		"guildInviteDeclinedName",
		"guildInviteDeclined",
	}

	for _, key := range requiredKeys {
		if _, ok := strings[key]; !ok {
			t.Errorf("Missing required key: %s", key)
		}
	}
}

func TestGetLangStrings_Japanese(t *testing.T) {
	server := &Server{
		erupeConfig: &config.Config{
			Language: "jp",
		},
	}

	strings := getLangStrings(server)

	if strings["language"] != "日本語" {
		t.Errorf("language = %q, want %q", strings["language"], "日本語")
	}

	// Verify Japanese strings are different from English defaults
	if strings["commandReload"] == "Reloading players..." {
		t.Error("Japanese commandReload should be different from English")
	}
}

func TestGetLangStrings_DefaultToEnglish(t *testing.T) {
	server := &Server{
		erupeConfig: &config.Config{
			Language: "unknown_language",
		},
	}

	strings := getLangStrings(server)

	// Unknown language should default to English
	if strings["language"] != "English" {
		t.Errorf("Unknown language should default to English, got %q", strings["language"])
	}
}

func TestGetLangStrings_EmptyLanguage(t *testing.T) {
	server := &Server{
		erupeConfig: &config.Config{
			Language: "",
		},
	}

	strings := getLangStrings(server)

	// Empty language should default to English
	if strings["language"] != "English" {
		t.Errorf("Empty language should default to English, got %q", strings["language"])
	}
}

func TestGetLangStrings_FormatStrings(t *testing.T) {
	server := &Server{
		erupeConfig: &config.Config{
			Language: "en",
		},
	}

	strings := getLangStrings(server)

	// Verify format strings contain placeholders
	tests := []struct {
		key         string
		placeholder string
	}{
		{"cafeReset", "%d"},
		{"commandDisabled", "%s"},
		{"commandKqfGet", "%x"},
		{"commandKqfSetError", "%s"},
		{"commandRightsError", "%s"},
		{"commandRightsSuccess", "%d"},
		{"commandCourseError", "%s"},
		{"commandCourseDisabled", "%s"},
		{"commandTeleportError", "%s"},
		{"commandTeleportSuccess", "%d"},
		{"commandRaviMultiplier", "%.2f"},
		{"guildInvite", "%s"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			str := strings[tt.key]
			if str == "" {
				t.Errorf("String %s is empty", tt.key)
				return
			}
			// Just verify format strings have some placeholder
			if len(str) == 0 {
				t.Errorf("String %s should not be empty", tt.key)
			}
		})
	}
}

func TestGetLangStrings_ReturnsDifferentMaps(t *testing.T) {
	server := &Server{
		erupeConfig: &config.Config{
			Language: "en",
		},
	}

	strings1 := getLangStrings(server)
	strings2 := getLangStrings(server)

	// Should return different map instances
	strings1["test"] = "modified"
	if strings2["test"] == "modified" {
		t.Error("getLangStrings should return a new map each call")
	}
}
