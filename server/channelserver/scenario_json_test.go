package channelserver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"
)

// ── test helpers ─────────────────────────────────────────────────────────────

// buildTestSubheaderChunk constructs a minimal sub-header format chunk.
// metadata is zero-filled to metaSize bytes.
func buildTestSubheaderChunk(t *testing.T, strings []string, metaSize int) []byte {
	t.Helper()
	var strBuf bytes.Buffer
	for _, s := range strings {
		sjis, err := scenarioEncodeShiftJIS(s)
		if err != nil {
			t.Fatalf("encode %q: %v", s, err)
		}
		strBuf.Write(sjis)
	}
	strBuf.WriteByte(0xFF) // end sentinel

	totalSize := 8 + metaSize + strBuf.Len()
	meta := make([]byte, metaSize) // zero metadata

	var buf bytes.Buffer
	buf.WriteByte(0x01)                 // type
	buf.WriteByte(0x00)                 // pad
	buf.WriteByte(byte(totalSize))      // size lo
	buf.WriteByte(byte(totalSize >> 8)) // size hi
	buf.WriteByte(byte(len(strings)))   // entry count
	buf.WriteByte(0x00)                 // unknown1
	buf.WriteByte(byte(metaSize))       // metadata total
	buf.WriteByte(0x00)                 // unknown2
	buf.Write(meta)
	buf.Write(strBuf.Bytes())
	return buf.Bytes()
}

// buildTestInlineChunk constructs an inline-format chunk0.
func buildTestInlineChunk(t *testing.T, strings []string) []byte {
	t.Helper()
	var buf bytes.Buffer
	for i, s := range strings {
		buf.WriteByte(byte(i + 1)) // 1-based index
		sjis, err := scenarioEncodeShiftJIS(s)
		if err != nil {
			t.Fatalf("encode %q: %v", s, err)
		}
		buf.Write(sjis)
	}
	return buf.Bytes()
}

// buildTestScenarioBinary assembles a complete scenario container for testing.
func buildTestScenarioBinary(t *testing.T, c0, c1 []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(c0))); err != nil {
		t.Fatal(err)
	}
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(c1))); err != nil {
		t.Fatal(err)
	}
	buf.Write(c0)
	buf.Write(c1)
	// c2 size = 0
	if err := binary.Write(&buf, binary.BigEndian, uint32(0)); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// extractStringsFromScenario parses a binary and returns all strings it contains.
func extractStringsFromScenario(t *testing.T, data []byte) []string {
	t.Helper()
	s, err := ParseScenarioBinary(data)
	if err != nil {
		t.Fatalf("ParseScenarioBinary: %v", err)
	}
	var result []string
	if s.Chunk0 != nil {
		if s.Chunk0.Subheader != nil {
			result = append(result, s.Chunk0.Subheader.Strings...)
		}
		for _, e := range s.Chunk0.Inline {
			result = append(result, e.Text)
		}
	}
	if s.Chunk1 != nil && s.Chunk1.Subheader != nil {
		result = append(result, s.Chunk1.Subheader.Strings...)
	}
	return result
}

// ── parse tests ──────────────────────────────────────────────────────────────

func TestParseScenarioBinary_TooShort(t *testing.T) {
	_, err := ParseScenarioBinary([]byte{0x00, 0x01})
	if err == nil {
		t.Error("expected error for short input")
	}
}

func TestParseScenarioBinary_EmptyChunks(t *testing.T) {
	data := buildTestScenarioBinary(t, nil, nil)
	s, err := ParseScenarioBinary(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Chunk0 != nil || s.Chunk1 != nil || s.Chunk2 != nil {
		t.Error("expected all chunks nil for empty scenario")
	}
}

func TestParseScenarioBinary_SubheaderChunk0(t *testing.T) {
	c0 := buildTestSubheaderChunk(t, []string{"Quest A", "Quest B"}, 4)
	data := buildTestScenarioBinary(t, c0, nil)

	s, err := ParseScenarioBinary(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Chunk0 == nil || s.Chunk0.Subheader == nil {
		t.Fatal("expected chunk0 subheader")
	}
	got := s.Chunk0.Subheader.Strings
	want := []string{"Quest A", "Quest B"}
	if len(got) != len(want) {
		t.Fatalf("string count: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseScenarioBinary_InlineChunk0(t *testing.T) {
	c0 := buildTestInlineChunk(t, []string{"Item1", "Item2"})
	data := buildTestScenarioBinary(t, c0, nil)

	s, err := ParseScenarioBinary(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Chunk0 == nil || len(s.Chunk0.Inline) == 0 {
		t.Fatal("expected chunk0 inline entries")
	}
	want := []string{"Item1", "Item2"}
	for i, e := range s.Chunk0.Inline {
		if e.Text != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, e.Text, want[i])
		}
	}
}

func TestParseScenarioBinary_BothChunks(t *testing.T) {
	c0 := buildTestSubheaderChunk(t, []string{"Quest"}, 4)
	c1 := buildTestSubheaderChunk(t, []string{"NPC1", "NPC2"}, 8)
	data := buildTestScenarioBinary(t, c0, c1)

	strings := extractStringsFromScenario(t, data)
	want := []string{"Quest", "NPC1", "NPC2"}
	if len(strings) != len(want) {
		t.Fatalf("string count: got %d, want %d", len(strings), len(want))
	}
	for i := range want {
		if strings[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, strings[i], want[i])
		}
	}
}

func TestParseScenarioBinary_Japanese(t *testing.T) {
	c0 := buildTestSubheaderChunk(t, []string{"テスト", "日本語"}, 4)
	data := buildTestScenarioBinary(t, c0, nil)

	s, err := ParseScenarioBinary(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"テスト", "日本語"}
	got := s.Chunk0.Subheader.Strings
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

// ── compile tests ─────────────────────────────────────────────────────────────

func TestCompileScenarioJSON_Subheader(t *testing.T) {
	input := &ScenarioJSON{
		Chunk0: &ScenarioChunk0JSON{
			Subheader: &ScenarioSubheaderJSON{
				Type:     0x01,
				Unknown1: 0x00,
				Unknown2: 0x00,
				Metadata: "AAAABBBB", // base64 of 6 zero bytes
				Strings:  []string{"Hello", "World"},
			},
		},
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	compiled, err := CompileScenarioJSON(jsonData)
	if err != nil {
		t.Fatalf("CompileScenarioJSON: %v", err)
	}

	// Parse the compiled output and verify strings survive
	result, err := ParseScenarioBinary(compiled)
	if err != nil {
		t.Fatalf("ParseScenarioBinary on compiled output: %v", err)
	}
	if result.Chunk0 == nil || result.Chunk0.Subheader == nil {
		t.Fatal("expected chunk0 subheader in compiled output")
	}
	want := []string{"Hello", "World"}
	got := result.Chunk0.Subheader.Strings
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompileScenarioJSON_Inline(t *testing.T) {
	input := &ScenarioJSON{
		Chunk0: &ScenarioChunk0JSON{
			Inline: []ScenarioInlineEntry{
				{Index: 1, Text: "Sword"},
				{Index: 2, Text: "Shield"},
			},
		},
	}
	jsonData, _ := json.Marshal(input)
	compiled, err := CompileScenarioJSON(jsonData)
	if err != nil {
		t.Fatalf("CompileScenarioJSON: %v", err)
	}

	result, err := ParseScenarioBinary(compiled)
	if err != nil {
		t.Fatalf("ParseScenarioBinary: %v", err)
	}
	if result.Chunk0 == nil || len(result.Chunk0.Inline) != 2 {
		t.Fatal("expected 2 inline entries")
	}
	if result.Chunk0.Inline[0].Text != "Sword" {
		t.Errorf("got %q, want Sword", result.Chunk0.Inline[0].Text)
	}
	if result.Chunk0.Inline[1].Text != "Shield" {
		t.Errorf("got %q, want Shield", result.Chunk0.Inline[1].Text)
	}
}

// ── round-trip tests ─────────────────────────────────────────────────────────

func TestScenarioRoundTrip_Subheader(t *testing.T) {
	original := buildTestScenarioBinary(t,
		buildTestSubheaderChunk(t, []string{"QuestName", "Description"}, 0x14),
		buildTestSubheaderChunk(t, []string{"Dialog1", "Dialog2", "Dialog3"}, 0x2C),
	)

	s, err := ParseScenarioBinary(original)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	jsonData, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	compiled, err := CompileScenarioJSON(jsonData)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	// Re-parse compiled and compare strings
	wantStrings := []string{"QuestName", "Description", "Dialog1", "Dialog2", "Dialog3"}
	gotStrings := extractStringsFromScenario(t, compiled)
	if len(gotStrings) != len(wantStrings) {
		t.Fatalf("string count: got %d, want %d", len(gotStrings), len(wantStrings))
	}
	for i := range wantStrings {
		if gotStrings[i] != wantStrings[i] {
			t.Errorf("[%d]: got %q, want %q", i, gotStrings[i], wantStrings[i])
		}
	}
}

func TestScenarioRoundTrip_Inline(t *testing.T) {
	original := buildTestScenarioBinary(t,
		buildTestInlineChunk(t, []string{"EpisodeA", "EpisodeB"}),
		nil,
	)

	s, _ := ParseScenarioBinary(original)
	jsonData, _ := json.Marshal(s)
	compiled, err := CompileScenarioJSON(jsonData)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	got := extractStringsFromScenario(t, compiled)
	want := []string{"EpisodeA", "EpisodeB"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScenarioRoundTrip_MetadataPreserved(t *testing.T) {
	// The metadata block must survive parse → JSON → compile unchanged.
	metaBytes := []byte{0x01, 0x02, 0x03, 0x04, 0xFF, 0xFE, 0xFD, 0xFC}
	// Build a chunk with custom metadata and unknown field values by hand.
	var buf bytes.Buffer
	str := []byte("A\x00\xFF")
	totalSize := 8 + len(metaBytes) + len(str)
	buf.WriteByte(0x01)
	buf.WriteByte(0x00)
	buf.WriteByte(byte(totalSize))
	buf.WriteByte(byte(totalSize >> 8))
	buf.WriteByte(0x01) // entry count
	buf.WriteByte(0xAA) // unknown1
	buf.WriteByte(byte(len(metaBytes)))
	buf.WriteByte(0xBB) // unknown2
	buf.Write(metaBytes)
	buf.Write(str)
	c0 := buf.Bytes()

	data := buildTestScenarioBinary(t, c0, nil)
	s, err := ParseScenarioBinary(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	sh := s.Chunk0.Subheader
	if sh.Type != 0x01 || sh.Unknown1 != 0xAA || sh.Unknown2 != 0xBB {
		t.Errorf("header fields: type=%02X unk1=%02X unk2=%02X", sh.Type, sh.Unknown1, sh.Unknown2)
	}

	// Compile and parse again — metadata must survive
	jsonData, _ := json.Marshal(s)
	compiled, err := CompileScenarioJSON(jsonData)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	s2, err := ParseScenarioBinary(compiled)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	sh2 := s2.Chunk0.Subheader
	if sh2.Metadata != sh.Metadata {
		t.Errorf("metadata changed:\n  before: %s\n   after: %s", sh.Metadata, sh2.Metadata)
	}
	if sh2.Unknown1 != sh.Unknown1 || sh2.Unknown2 != sh.Unknown2 {
		t.Errorf("unknown fields changed: unk1 %02X→%02X  unk2 %02X→%02X",
			sh.Unknown1, sh2.Unknown1, sh.Unknown2, sh2.Unknown2)
	}
}

// ── real-file round-trip tests ────────────────────────────────────────────────

// scenarioBinPath is the relative path from the package to the scenario files.
// These tests are skipped if the directory does not exist (CI without game data).
const scenarioBinPath = "../../bin/scenarios"

func TestScenarioRoundTrip_RealFiles(t *testing.T) {
	samples := []struct {
		name   string
		wantC0 bool // expect chunk0 subheader
		wantC1 bool // expect chunk1 (subheader or JKR)
	}{
		// cat=0 basic quest scenarios (chunk0 subheader, no chunk1)
		{"0_0_0_0_S0_T101_C0", true, false},
		{"0_0_0_0_S1_T101_C0", true, false},
		{"0_0_0_0_S5_T101_C0", true, false},
		// cat=1 GR scenarios (chunk0 subheader, T101 has no chunk1)
		{"1_0_0_0_S0_T101_C0", true, false},
		{"1_0_0_0_S1_T101_C0", true, false},
		// cat=3 item exchange (chunk0 subheader, chunk1 subheader with extra data)
		{"3_0_0_0_S0_T103_C0", true, true},
		// multi-chapter file with chunk1 subheader
		{"0_0_0_0_S0_T103_C0", true, true},
	}

	for _, tc := range samples {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			path := scenarioBinPath + "/" + tc.name + ".bin"
			original, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("scenario file not found (game data not present): %v", err)
			}

			// Parse binary → JSON schema
			parsed, err := ParseScenarioBinary(original)
			if err != nil {
				t.Fatalf("ParseScenarioBinary: %v", err)
			}

			// Verify expected chunk presence
			if tc.wantC0 && (parsed.Chunk0 == nil || parsed.Chunk0.Subheader == nil) {
				t.Error("expected chunk0 subheader")
			}
			if tc.wantC1 && parsed.Chunk1 == nil {
				t.Error("expected chunk1")
			}

			// Marshal to JSON
			jsonData, err := json.Marshal(parsed)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			// Compile JSON → binary
			compiled, err := CompileScenarioJSON(jsonData)
			if err != nil {
				t.Fatalf("CompileScenarioJSON: %v", err)
			}

			// Re-parse compiled output
			result, err := ParseScenarioBinary(compiled)
			if err != nil {
				t.Fatalf("ParseScenarioBinary on compiled output: %v", err)
			}

			// Verify strings survive round-trip unchanged
			origStrings := extractStringsFromScenario(t, original)
			gotStrings := extractStringsFromScenario(t, compiled)
			if len(gotStrings) != len(origStrings) {
				t.Fatalf("string count changed: %d → %d", len(origStrings), len(gotStrings))
			}
			for i := range origStrings {
				if gotStrings[i] != origStrings[i] {
					t.Errorf("[%d]: %q → %q", i, origStrings[i], gotStrings[i])
				}
			}

			// Verify metadata is preserved byte-for-byte
			if parsed.Chunk0 != nil && parsed.Chunk0.Subheader != nil {
				if result.Chunk0 == nil || result.Chunk0.Subheader == nil {
					t.Fatal("chunk0 subheader lost in round-trip")
				}
				if result.Chunk0.Subheader.Metadata != parsed.Chunk0.Subheader.Metadata {
					t.Errorf("chunk0 metadata changed after round-trip")
				}
			}
			if parsed.Chunk1 != nil && parsed.Chunk1.Subheader != nil {
				if result.Chunk1 == nil || result.Chunk1.Subheader == nil {
					t.Fatal("chunk1 subheader lost in round-trip")
				}
				if result.Chunk1.Subheader.Metadata != parsed.Chunk1.Subheader.Metadata {
					t.Errorf("chunk1 metadata changed after round-trip")
				}
			}
		})
	}
}
