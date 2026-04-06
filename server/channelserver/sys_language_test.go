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

// TestGetLangStringsFor covers every supported language code and the
// unknown-code fallback, ensuring the new direct-dispatch primitive stays in
// sync with supportedLangs.
func TestGetLangStringsFor(t *testing.T) {
	cases := []struct {
		code       string
		wantLang   string
		wantNonJP  bool // ensure unknown falls back to English, not Japanese
		wantPrefix string
	}{
		{"en", "English", true, ""},
		{"jp", "日本語", false, ""},
		{"fr", "Français", true, ""},
		{"es", "Español", true, ""},
		{"", "English", true, ""},
		{"xx", "English", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			got := getLangStringsFor(tc.code)
			if got.language != tc.wantLang {
				t.Errorf("getLangStringsFor(%q).language = %q, want %q", tc.code, got.language, tc.wantLang)
			}
			if got.commands.lang.usage == "" {
				t.Errorf("%q: commands.lang.usage should not be empty", tc.code)
			}
		})
	}
}

func TestIsSupportedLang(t *testing.T) {
	for _, code := range []string{"en", "jp", "fr", "es"} {
		if !isSupportedLang(code) {
			t.Errorf("isSupportedLang(%q) = false, want true", code)
		}
	}
	for _, code := range []string{"", "de", "EN", "english"} {
		if isSupportedLang(code) {
			t.Errorf("isSupportedLang(%q) = true, want false", code)
		}
	}
}

// TestSessionLang_FallbackAndOverride verifies that Session.Lang() returns the
// server default when no per-session preference is set, and the preference
// once SetLang is called.
func TestSessionLang_FallbackAndOverride(t *testing.T) {
	server := &Server{erupeConfig: &cfg.Config{Language: "jp"}}
	s := &Session{server: server}

	if got := s.Lang(); got != "jp" {
		t.Errorf("Lang() with no override = %q, want %q (server default)", got, "jp")
	}

	s.SetLang("fr")
	if got := s.Lang(); got != "fr" {
		t.Errorf("Lang() after SetLang(fr) = %q, want %q", got, "fr")
	}

	// Empty SetLang clears the override — falls back to server default again.
	s.SetLang("")
	if got := s.Lang(); got != "jp" {
		t.Errorf("Lang() after SetLang(\"\") = %q, want %q (server default)", got, "jp")
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
