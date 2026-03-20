package channelserver

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ── JSON schema types ────────────────────────────────────────────────────────

// ScenarioJSON is the open, human-editable representation of a scenario .bin file.
// Strings are stored as UTF-8; the compiler converts to/from Shift-JIS.
//
// Container layout (big-endian sizes):
//
//	@0x00: u32 BE  chunk0_size
//	@0x04: u32 BE  chunk1_size
//	       [chunk0_data]
//	       [chunk1_data]
//	       u32 BE  chunk2_size  (only present when non-zero)
//	       [chunk2_data]
type ScenarioJSON struct {
	// Chunk0 holds quest name/description data (sub-header or inline format).
	Chunk0 *ScenarioChunk0JSON `json:"chunk0,omitempty"`
	// Chunk1 holds NPC dialog data (sub-header format or raw JKR blob).
	Chunk1 *ScenarioChunk1JSON `json:"chunk1,omitempty"`
	// Chunk2 holds JKR-compressed menu/title data.
	Chunk2 *ScenarioRawChunkJSON `json:"chunk2,omitempty"`
}

// ScenarioChunk0JSON represents chunk0, which is either sub-header or inline format.
// Exactly one of Subheader/Inline is non-nil.
type ScenarioChunk0JSON struct {
	Subheader *ScenarioSubheaderJSON `json:"subheader,omitempty"`
	Inline    []ScenarioInlineEntry  `json:"inline,omitempty"`
}

// ScenarioChunk1JSON represents chunk1, which is either sub-header or raw JKR.
// Exactly one of Subheader/JKR is non-nil.
type ScenarioChunk1JSON struct {
	Subheader *ScenarioSubheaderJSON `json:"subheader,omitempty"`
	JKR       *ScenarioRawChunkJSON  `json:"jkr,omitempty"`
}

// ScenarioSubheaderJSON represents a chunk in sub-header format.
//
// Sub-header binary layout (8 bytes, little-endian where applicable):
//
//	@0: u8   Type     (usually 0x01)
//	@1: u8   0x00     (pad; distinguishes this format from inline)
//	@2: u16  Size     (total chunk size including this header)
//	@4: u8   Count    (number of string entries)
//	@5: u8   Unknown1
//	@6: u8   MetaSize (total bytes of metadata block)
//	@7: u8   Unknown2
//	[MetaSize bytes: opaque metadata (string IDs, offsets, flags — partially unknown)]
//	[null-terminated Shift-JIS strings, one per entry]
//	[0xFF end-of-strings sentinel]
type ScenarioSubheaderJSON struct {
	// Type is the chunk type byte (almost always 0x01).
	Type     uint8 `json:"type"`
	Unknown1 uint8 `json:"unknown1"`
	Unknown2 uint8 `json:"unknown2"`
	// Metadata is the opaque metadata block, base64-encoded.
	// Preserving it unchanged ensures correct client behavior for fields
	// whose meaning is not yet fully understood.
	Metadata string `json:"metadata"`
	// Strings contains the human-editable text (UTF-8).
	Strings []string `json:"strings"`
}

// ScenarioInlineEntry is one entry in an inline-format chunk0.
// Format on wire: {u8 index}{Shift-JIS string}{0x00}.
type ScenarioInlineEntry struct {
	Index uint8  `json:"index"`
	Text  string `json:"text"`
}

// ScenarioRawChunkJSON stores a JKR-compressed chunk as its raw compressed bytes.
// The data is served to the client as-is; the format of the decompressed content
// is not yet fully documented.
type ScenarioRawChunkJSON struct {
	// Data is the raw JKR-compressed bytes, base64-encoded.
	Data string `json:"data"`
}

// ── Parse: binary → JSON ─────────────────────────────────────────────────────

// ParseScenarioBinary reads a scenario .bin file and returns a ScenarioJSON
// suitable for editing and re-compilation with CompileScenarioJSON.
func ParseScenarioBinary(data []byte) (*ScenarioJSON, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("scenario data too short: %d bytes", len(data))
	}

	c0Size := int(binary.BigEndian.Uint32(data[0:4]))
	c1Size := int(binary.BigEndian.Uint32(data[4:8]))

	result := &ScenarioJSON{}

	// Chunk0
	c0Off := 8
	if c0Size > 0 {
		if c0Off+c0Size > len(data) {
			return nil, fmt.Errorf("chunk0 size %d overruns data at offset %d", c0Size, c0Off)
		}
		chunk0, err := parseScenarioChunk0(data[c0Off : c0Off+c0Size])
		if err != nil {
			return nil, fmt.Errorf("chunk0: %w", err)
		}
		result.Chunk0 = chunk0
	}

	// Chunk1
	c1Off := c0Off + c0Size
	if c1Size > 0 {
		if c1Off+c1Size > len(data) {
			return nil, fmt.Errorf("chunk1 size %d overruns data at offset %d", c1Size, c1Off)
		}
		chunk1, err := parseScenarioChunk1(data[c1Off : c1Off+c1Size])
		if err != nil {
			return nil, fmt.Errorf("chunk1: %w", err)
		}
		result.Chunk1 = chunk1
	}

	// Chunk2 (preceded by its own 4-byte size field)
	c2HdrOff := c1Off + c1Size
	if c2HdrOff+4 <= len(data) {
		c2Size := int(binary.BigEndian.Uint32(data[c2HdrOff : c2HdrOff+4]))
		if c2Size > 0 {
			c2DataOff := c2HdrOff + 4
			if c2DataOff+c2Size > len(data) {
				return nil, fmt.Errorf("chunk2 size %d overruns data at offset %d", c2Size, c2DataOff)
			}
			result.Chunk2 = &ScenarioRawChunkJSON{
				Data: base64.StdEncoding.EncodeToString(data[c2DataOff : c2DataOff+c2Size]),
			}
		}
	}

	return result, nil
}

// parseScenarioChunk0 auto-detects sub-header vs inline format.
// The second byte being 0x00 is the pad byte in sub-headers; non-zero means inline.
func parseScenarioChunk0(data []byte) (*ScenarioChunk0JSON, error) {
	if len(data) < 2 {
		return &ScenarioChunk0JSON{}, nil
	}
	if data[1] == 0x00 {
		sh, err := parseScenarioSubheader(data)
		if err != nil {
			return nil, err
		}
		return &ScenarioChunk0JSON{Subheader: sh}, nil
	}
	entries, err := parseScenarioInline(data)
	if err != nil {
		return nil, err
	}
	return &ScenarioChunk0JSON{Inline: entries}, nil
}

// parseScenarioChunk1 parses chunk1 as JKR or sub-header depending on magic bytes.
func parseScenarioChunk1(data []byte) (*ScenarioChunk1JSON, error) {
	if len(data) >= 4 && binary.LittleEndian.Uint32(data[0:4]) == 0x1A524B4A {
		return &ScenarioChunk1JSON{
			JKR: &ScenarioRawChunkJSON{
				Data: base64.StdEncoding.EncodeToString(data),
			},
		}, nil
	}
	sh, err := parseScenarioSubheader(data)
	if err != nil {
		return nil, err
	}
	return &ScenarioChunk1JSON{Subheader: sh}, nil
}

// parseScenarioSubheader parses the 8-byte sub-header + metadata + strings.
func parseScenarioSubheader(data []byte) (*ScenarioSubheaderJSON, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("sub-header chunk too short: %d bytes", len(data))
	}

	// Sub-header fields
	chunkType := data[0]
	// data[1] is the 0x00 pad (not stored; implicit)
	// data[2:4] is the u16 LE total size (recomputed on compile)
	entryCount := int(data[4])
	unknown1 := data[5]
	metaSize := int(data[6])
	unknown2 := data[7]

	metaEnd := 8 + metaSize
	if metaEnd > len(data) {
		return nil, fmt.Errorf("metadata block (size %d) overruns chunk (len %d)", metaSize, len(data))
	}

	metadata := base64.StdEncoding.EncodeToString(data[8:metaEnd])

	strings, err := scenarioReadStrings(data, metaEnd, entryCount)
	if err != nil {
		return nil, err
	}

	return &ScenarioSubheaderJSON{
		Type:     chunkType,
		Unknown1: unknown1,
		Unknown2: unknown2,
		Metadata: metadata,
		Strings:  strings,
	}, nil
}

// parseScenarioInline parses chunk0 inline format: {u8 index}{Shift-JIS string}{0x00}.
func parseScenarioInline(data []byte) ([]ScenarioInlineEntry, error) {
	var result []ScenarioInlineEntry
	pos := 0
	for pos < len(data) {
		if data[pos] == 0x00 {
			pos++
			continue
		}
		idx := data[pos]
		pos++
		if pos >= len(data) {
			break
		}
		end := pos
		for end < len(data) && data[end] != 0x00 {
			end++
		}
		if end > pos {
			text, err := scenarioDecodeShiftJIS(data[pos:end])
			if err != nil {
				return nil, fmt.Errorf("inline entry at 0x%x: %w", pos, err)
			}
			result = append(result, ScenarioInlineEntry{Index: idx, Text: text})
		}
		pos = end + 1 // skip null terminator
	}
	return result, nil
}

// scenarioReadStrings scans for null-terminated Shift-JIS strings starting at
// offset start, reading at most maxCount strings (0 = unlimited). Stops on 0xFF.
func scenarioReadStrings(data []byte, start, maxCount int) ([]string, error) {
	var result []string
	pos := start
	for pos < len(data) {
		if maxCount > 0 && len(result) >= maxCount {
			break
		}
		if data[pos] == 0x00 {
			pos++
			continue
		}
		if data[pos] == 0xFF {
			break
		}
		end := pos
		for end < len(data) && data[end] != 0x00 {
			end++
		}
		if end > pos {
			text, err := scenarioDecodeShiftJIS(data[pos:end])
			if err != nil {
				return nil, fmt.Errorf("string at 0x%x: %w", pos, err)
			}
			result = append(result, text)
		}
		pos = end + 1
	}
	return result, nil
}

// ── Compile: JSON → binary ───────────────────────────────────────────────────

// CompileScenarioJSON parses jsonData and compiles it to MHF scenario binary format.
func CompileScenarioJSON(jsonData []byte) ([]byte, error) {
	var s ScenarioJSON
	if err := json.Unmarshal(jsonData, &s); err != nil {
		return nil, fmt.Errorf("unmarshal scenario JSON: %w", err)
	}
	return compileScenario(&s)
}

func compileScenario(s *ScenarioJSON) ([]byte, error) {
	var chunk0, chunk1, chunk2 []byte
	var err error

	if s.Chunk0 != nil {
		chunk0, err = compileScenarioChunk0(s.Chunk0)
		if err != nil {
			return nil, fmt.Errorf("chunk0: %w", err)
		}
	}
	if s.Chunk1 != nil {
		chunk1, err = compileScenarioChunk1(s.Chunk1)
		if err != nil {
			return nil, fmt.Errorf("chunk1: %w", err)
		}
	}
	if s.Chunk2 != nil {
		chunk2, err = compileScenarioRawChunk(s.Chunk2)
		if err != nil {
			return nil, fmt.Errorf("chunk2: %w", err)
		}
	}

	var buf bytes.Buffer
	// Container header: c0_size, c1_size (big-endian u32)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(chunk0)))
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(chunk1)))
	buf.Write(chunk0)
	buf.Write(chunk1)
	// Chunk2 preceded by its own size field
	if len(chunk2) > 0 {
		_ = binary.Write(&buf, binary.BigEndian, uint32(len(chunk2)))
		buf.Write(chunk2)
	}

	return buf.Bytes(), nil
}

func compileScenarioChunk0(c *ScenarioChunk0JSON) ([]byte, error) {
	if c.Subheader != nil {
		return compileScenarioSubheader(c.Subheader)
	}
	return compileScenarioInline(c.Inline)
}

func compileScenarioChunk1(c *ScenarioChunk1JSON) ([]byte, error) {
	if c.JKR != nil {
		return compileScenarioRawChunk(c.JKR)
	}
	if c.Subheader != nil {
		return compileScenarioSubheader(c.Subheader)
	}
	return nil, nil
}

// compileScenarioSubheader builds the binary sub-header chunk:
// [8-byte header][metadata][null-terminated Shift-JIS strings][0xFF]
func compileScenarioSubheader(sh *ScenarioSubheaderJSON) ([]byte, error) {
	meta, err := base64.StdEncoding.DecodeString(sh.Metadata)
	if err != nil {
		return nil, fmt.Errorf("decode metadata base64: %w", err)
	}

	var strBuf bytes.Buffer
	for _, s := range sh.Strings {
		sjis, err := scenarioEncodeShiftJIS(s)
		if err != nil {
			return nil, err
		}
		strBuf.Write(sjis) // sjis already has null terminator from helper
	}
	strBuf.WriteByte(0xFF) // end-of-strings sentinel

	// Total size = 8-byte header + metadata + strings
	totalSize := 8 + len(meta) + strBuf.Len()

	var buf bytes.Buffer
	buf.WriteByte(sh.Type)
	buf.WriteByte(0x00) // pad (format detector)
	// u16 LE total size
	buf.WriteByte(byte(totalSize))
	buf.WriteByte(byte(totalSize >> 8))
	buf.WriteByte(byte(len(sh.Strings))) // entry count
	buf.WriteByte(sh.Unknown1)
	buf.WriteByte(byte(len(meta))) // metadata total size
	buf.WriteByte(sh.Unknown2)
	buf.Write(meta)
	buf.Write(strBuf.Bytes())

	return buf.Bytes(), nil
}

// compileScenarioInline builds the inline-format chunk0 bytes.
func compileScenarioInline(entries []ScenarioInlineEntry) ([]byte, error) {
	var buf bytes.Buffer
	for _, e := range entries {
		buf.WriteByte(e.Index)
		sjis, err := scenarioEncodeShiftJIS(e.Text)
		if err != nil {
			return nil, err
		}
		buf.Write(sjis) // includes null terminator
	}
	return buf.Bytes(), nil
}

// compileScenarioRawChunk decodes the base64 raw chunk bytes.
// These are served to the client as-is (no re-compression).
func compileScenarioRawChunk(rc *ScenarioRawChunkJSON) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(rc.Data)
	if err != nil {
		return nil, fmt.Errorf("decode raw chunk base64: %w", err)
	}
	return data, nil
}

// ── String helpers ───────────────────────────────────────────────────────────

// scenarioDecodeShiftJIS converts a raw Shift-JIS byte slice to UTF-8 string.
func scenarioDecodeShiftJIS(b []byte) (string, error) {
	dec := japanese.ShiftJIS.NewDecoder()
	out, _, err := transform.Bytes(dec, b)
	if err != nil {
		return "", fmt.Errorf("shift-jis decode: %w", err)
	}
	return string(out), nil
}

// scenarioEncodeShiftJIS converts a UTF-8 string to a null-terminated Shift-JIS byte slice.
func scenarioEncodeShiftJIS(s string) ([]byte, error) {
	enc := japanese.ShiftJIS.NewEncoder()
	out, _, err := transform.Bytes(enc, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("shift-jis encode %q: %w", s, err)
	}
	return append(out, 0x00), nil
}

