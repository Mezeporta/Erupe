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
		{"zh", "中文", true, ""},
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
	for _, code := range []string{"en", "jp", "fr", "es", "zh"} {
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
		"zh": langChinese(),
	}
	for code, lang := range languages {
		t.Run(code, func(t *testing.T) {
			checkNoEmptyStrings(t, reflect.ValueOf(lang), code)
		})
	}
}

// TestSessionI18n_Cached verifies that Session.I18n() returns an i18n table
// resolved against the session's language and caches the pointer until
// SetLang invalidates it (phase C of #188).
func TestSessionI18n_Cached(t *testing.T) {
	server := &Server{erupeConfig: &cfg.Config{Language: "en"}}
	s := &Session{server: server}

	en1 := s.I18n()
	en2 := s.I18n()
	if en1 != en2 {
		t.Error("Session.I18n() should return the same pointer until SetLang")
	}
	if en1.language != "English" {
		t.Errorf("server-default I18n = %q, want English", en1.language)
	}

	s.SetLang("jp")
	jp := s.I18n()
	if jp == en1 {
		t.Error("SetLang should invalidate the cached I18n pointer")
	}
	if jp.language != "日本語" {
		t.Errorf("after SetLang(jp) I18n = %q, want 日本語", jp.language)
	}

	// And another call returns the same (now-cached) jp pointer.
	if s.I18n() != jp {
		t.Error("Session.I18n() should be cached after SetLang rebuild")
	}
}

// TestParseChatCommand_RepliesInSessionLanguage confirms the mechanical
// s.server.i18n → s.I18n() refactor routes chat responses through the
// session's language.
func TestParseChatCommand_RepliesInSessionLanguage_Placeholder(t *testing.T) {
	// Sanity: for a French session, the i18n table returned by I18n() must
	// be the French one, and its commands.timer.enabled must not equal the
	// English string.
	server := &Server{erupeConfig: &cfg.Config{Language: "en"}}
	s := &Session{server: server}
	s.SetLang("fr")

	frTable := s.I18n()
	enTable := getLangStringsFor("en")
	if frTable.commands.timer.enabled == enTable.commands.timer.enabled {
		t.Error("fr and en timer.enabled strings should differ — refactor may have reverted")
	}
	if frTable.language != "Français" {
		t.Errorf("session I18n language = %q, want Français", frTable.language)
	}
}
