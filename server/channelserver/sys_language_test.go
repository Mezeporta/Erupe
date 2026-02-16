package channelserver

import (
	"testing"

	_config "erupe-ce/config"
)

func TestGetLangStrings_English(t *testing.T) {
	server := &Server{
		erupeConfig: &_config.Config{
			Language: "en",
		},
	}

	lang := getLangStrings(server)

	if lang.language != "English" {
		t.Errorf("language = %q, want %q", lang.language, "English")
	}

	// Verify key strings are not empty
	if lang.cafe.reset == "" {
		t.Error("cafe.reset should not be empty")
	}
	if lang.commands.disabled == "" {
		t.Error("commands.disabled should not be empty")
	}
	if lang.commands.reload == "" {
		t.Error("commands.reload should not be empty")
	}
	if lang.commands.ravi.noCommand == "" {
		t.Error("commands.ravi.noCommand should not be empty")
	}
	if lang.guild.invite.title == "" {
		t.Error("guild.invite.title should not be empty")
	}
}

func TestGetLangStrings_Japanese(t *testing.T) {
	server := &Server{
		erupeConfig: &_config.Config{
			Language: "jp",
		},
	}

	lang := getLangStrings(server)

	if lang.language != "日本語" {
		t.Errorf("language = %q, want %q", lang.language, "日本語")
	}

	// Verify Japanese strings are different from English
	enServer := &Server{
		erupeConfig: &_config.Config{
			Language: "en",
		},
	}
	enLang := getLangStrings(enServer)

	if lang.commands.reload == enLang.commands.reload {
		t.Error("Japanese commands.reload should be different from English")
	}
}

func TestGetLangStrings_DefaultToEnglish(t *testing.T) {
	server := &Server{
		erupeConfig: &_config.Config{
			Language: "unknown_language",
		},
	}

	lang := getLangStrings(server)

	// Unknown language should default to English
	if lang.language != "English" {
		t.Errorf("Unknown language should default to English, got %q", lang.language)
	}
}

func TestGetLangStrings_EmptyLanguage(t *testing.T) {
	server := &Server{
		erupeConfig: &_config.Config{
			Language: "",
		},
	}

	lang := getLangStrings(server)

	// Empty language should default to English
	if lang.language != "English" {
		t.Errorf("Empty language should default to English, got %q", lang.language)
	}
}
