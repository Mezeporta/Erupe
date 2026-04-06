package channelserver

import (
	"bytes"
	"encoding/json"
)

// LocalizedString is a JSON field that unmarshals from either a plain string
// (backwards-compatible single-language behaviour — the value is returned for
// every language) or a map keyed by language code, e.g.
//
//	"title": "リオレウス"
//	"title": { "jp": "リオレウス", "en": "Rathalos", "fr": "Rathalos" }
//
// It is the core primitive used by phase B of #188 to localize server-sent
// content (quest text, scenario strings, mail templates, ...) per session
// without breaking any existing single-language JSON file.
//
// Encoding note: strings that end up on the wire as Shift-JIS (quest text,
// scenario strings) must only use characters representable in Shift-JIS —
// ASCII, kana, and CJK. Latin-extended characters commonly used in European
// languages (ê, ñ, ß, ...) will be rejected by the encoder at compile time.
// For those languages prefer ASCII-only romanizations ("Quete de test",
// "Espana") until the Frontier binary protocol is extended to a wider
// encoding.
type LocalizedString struct {
	// plain is set when the source was a bare JSON string. Treated as the
	// fallback for every language so legacy single-language files keep
	// working with no schema change.
	plain string
	// values is set when the source was a JSON object. Keys are language
	// codes (lowercase, e.g. "jp", "en", "fr", "es").
	values map[string]string
}

// NewLocalizedPlain wraps a single-language string — used internally by the
// binary-to-JSON reverse path (ParseQuestBinary etc.) where only one language
// is available from the source file.
func NewLocalizedPlain(s string) LocalizedString {
	return LocalizedString{plain: s}
}

// UnmarshalJSON accepts either a JSON string or a JSON object.
func (l *LocalizedString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	if trimmed[0] == '"' {
		return json.Unmarshal(trimmed, &l.plain)
	}
	return json.Unmarshal(trimmed, &l.values)
}

// MarshalJSON round-trips: plain strings stay plain; maps stay maps. This
// matters for the reverse ParseQuestBinary → JSON path, which should produce
// a plain string (backwards compatible), and for hypothetical editor tooling
// that reads a localized JSON, mutates it, and writes it back unchanged.
func (l LocalizedString) MarshalJSON() ([]byte, error) {
	if l.values != nil {
		return json.Marshal(l.values)
	}
	return json.Marshal(l.plain)
}

// Resolve returns the best available string for the requested language code.
// Fallback order when the requested language is missing:
//  1. The plain-string form (single-language source)
//  2. jp (the canonical source language for MH Frontier)
//  3. en (the common secondary)
//  4. Any non-empty value in the map
//
// Returns "" only when nothing is set — callers that need a non-empty value
// for binary serialization should treat that as an empty quest string, which
// the existing toShiftJIS encoder already accepts.
func (l LocalizedString) Resolve(lang string) string {
	if l.values == nil {
		return l.plain
	}
	if v, ok := l.values[lang]; ok && v != "" {
		return v
	}
	if l.plain != "" {
		return l.plain
	}
	for _, fb := range []string{"jp", "en"} {
		if v, ok := l.values[fb]; ok && v != "" {
			return v
		}
	}
	for _, v := range l.values {
		if v != "" {
			return v
		}
	}
	return ""
}

// IsLocalized reports whether the value was written as a language map (rather
// than a plain string). Mostly useful for tests and tooling.
func (l LocalizedString) IsLocalized() bool {
	return l.values != nil
}
