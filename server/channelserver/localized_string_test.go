package channelserver

import (
	"encoding/json"
	"testing"
)

func TestLocalizedString_UnmarshalPlainString(t *testing.T) {
	var l LocalizedString
	if err := json.Unmarshal([]byte(`"Rathalos"`), &l); err != nil {
		t.Fatalf("unmarshal plain: %v", err)
	}
	if l.IsLocalized() {
		t.Error("plain string should not be IsLocalized")
	}
	// Plain strings resolve to the same value regardless of language — this
	// is the backwards-compatibility contract that keeps existing single-
	// language quest JSONs working without a schema migration.
	for _, lang := range []string{"", "jp", "en", "fr", "es", "klingon"} {
		if got := l.Resolve(lang); got != "Rathalos" {
			t.Errorf("Resolve(%q) = %q, want %q", lang, got, "Rathalos")
		}
	}
}

func TestLocalizedString_UnmarshalMap(t *testing.T) {
	var l LocalizedString
	src := `{"jp": "リオレウス", "en": "Rathalos", "fr": "Rathalos"}`
	if err := json.Unmarshal([]byte(src), &l); err != nil {
		t.Fatalf("unmarshal map: %v", err)
	}
	if !l.IsLocalized() {
		t.Error("map form should be IsLocalized")
	}
	cases := map[string]string{
		"jp": "リオレウス",
		"en": "Rathalos",
		"fr": "Rathalos",
	}
	for lang, want := range cases {
		if got := l.Resolve(lang); got != want {
			t.Errorf("Resolve(%q) = %q, want %q", lang, got, want)
		}
	}
}

func TestLocalizedString_ResolveFallbackChain(t *testing.T) {
	// Spanish not provided — should fall back to jp first (canonical source
	// language for MHF), then en, then any non-empty.
	var l LocalizedString
	if err := json.Unmarshal([]byte(`{"jp": "リオレウス", "en": "Rathalos"}`), &l); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := l.Resolve("es"); got != "リオレウス" {
		t.Errorf("missing-lang fallback = %q, want jp value %q", got, "リオレウス")
	}

	// Only en provided → en must win even when jp is requested.
	var l2 LocalizedString
	if err := json.Unmarshal([]byte(`{"en": "Rathalos"}`), &l2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := l2.Resolve("jp"); got != "Rathalos" {
		t.Errorf("jp-missing fallback = %q, want %q", got, "Rathalos")
	}

	// Neither jp nor en → any non-empty value.
	var l3 LocalizedString
	if err := json.Unmarshal([]byte(`{"fr": "Rathalos FR"}`), &l3); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := l3.Resolve("jp"); got != "Rathalos FR" {
		t.Errorf("last-resort fallback = %q, want %q", got, "Rathalos FR")
	}

	// Empty string entries are skipped by the fallback chain.
	var l4 LocalizedString
	if err := json.Unmarshal([]byte(`{"jp": "", "en": "Rathalos"}`), &l4); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := l4.Resolve("fr"); got != "Rathalos" {
		t.Errorf("empty-jp fallback = %q, want %q (should skip empty jp)", got, "Rathalos")
	}
}

func TestLocalizedString_EmptyResolvesEmpty(t *testing.T) {
	var l LocalizedString
	if got := l.Resolve("jp"); got != "" {
		t.Errorf("zero value Resolve = %q, want empty", got)
	}
}

func TestLocalizedString_MarshalRoundTrip(t *testing.T) {
	// Plain string round-trip.
	var plain LocalizedString
	if err := json.Unmarshal([]byte(`"Rathalos"`), &plain); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(plain)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `"Rathalos"` {
		t.Errorf("plain round-trip = %s, want %q", out, "Rathalos")
	}

	// Map round-trip.
	var m LocalizedString
	if err := json.Unmarshal([]byte(`{"en":"Rathalos","jp":"リオレウス"}`), &m); err != nil {
		t.Fatal(err)
	}
	out, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	// Unmarshal again and compare — map key order is unstable so we don't
	// string-compare the marshaled form directly.
	var back LocalizedString
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if got := back.Resolve("jp"); got != "リオレウス" {
		t.Errorf("round-trip jp = %q", got)
	}
	if got := back.Resolve("en"); got != "Rathalos" {
		t.Errorf("round-trip en = %q", got)
	}
}

func TestLocalizedString_NullUnmarshal(t *testing.T) {
	// JSON null → zero value, no error.
	var l LocalizedString
	if err := json.Unmarshal([]byte(`null`), &l); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if l.IsLocalized() || l.Resolve("") != "" {
		t.Error("null should produce zero LocalizedString")
	}
}

func TestNewLocalizedPlain(t *testing.T) {
	l := NewLocalizedPlain("hello")
	if l.IsLocalized() {
		t.Error("NewLocalizedPlain should not be IsLocalized")
	}
	if got := l.Resolve("jp"); got != "hello" {
		t.Errorf("Resolve = %q, want %q", got, "hello")
	}
}
