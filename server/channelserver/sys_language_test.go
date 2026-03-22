package channelserver

import (
	"fmt"
	"reflect"
	"testing"

	cfg "erupe-ce/config"
)

func TestGetLangStrings_English(t *testing.T) {
	server := &Server{
		erupeConfig: &cfg.Config{
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
		erupeConfig: &cfg.Config{
			Language: "jp",
		},
	}

	lang := getLangStrings(server)

	if lang.language != "日本語" {
		t.Errorf("language = %q, want %q", lang.language, "日本語")
	}

	// Verify Japanese strings are different from English
	enServer := &Server{
		erupeConfig: &cfg.Config{
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
		erupeConfig: &cfg.Config{
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
		erupeConfig: &cfg.Config{
			Language: "",
		},
	}

	lang := getLangStrings(server)

	// Empty language should default to English
	if lang.language != "English" {
		t.Errorf("Empty language should default to English, got %q", lang.language)
	}
}

// checkNoEmptyStrings recursively walks v and fails the test for any empty string field.
func checkNoEmptyStrings(t *testing.T, v reflect.Value, path string) {
	t.Helper()
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			t.Errorf("missing translation: %s is empty", path)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			checkNoEmptyStrings(t, v.Field(i), path+"."+v.Type().Field(i).Name)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			checkNoEmptyStrings(t, v.Index(i), fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

func TestLangCompleteness(t *testing.T) {
	languages := map[string]i18n{
		"en": langEnglish(),
		"jp": langJapanese(),
		"fr": langFrench(),
		"es": langSpanish(),
	}
	for code, lang := range languages {
		t.Run(code, func(t *testing.T) {
			checkNoEmptyStrings(t, reflect.ValueOf(lang), code)
		})
	}
}
