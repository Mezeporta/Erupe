# Scenario Binary Format

> Reference: `network/mhfpacket/msg_sys_get_file.go`, issue [#172](https://github.com/Mezeporta/Erupe/issues/172)

## Overview

Scenario files are binary blobs served by `MSG_SYS_GET_FILE` when `IsScenario` is true. They contain quest descriptions, NPC dialog, episode listings, and menu options for the game's scenario/story system.

## Request Format

When `IsScenario == true`, the client sends a `scenarioFileIdentifier`:

| Offset | Type   | Field       | Description |
|--------|--------|-------------|-------------|
| 0      | uint8  | CategoryID  | Scenario category (0=Basic, 1=Veteran, 3=Other, 6=Pallone, 7=Diva) |
| 1      | uint32 | MainID      | Main scenario identifier |
| 5      | uint8  | ChapterID   | Chapter within the scenario |
| 6      | uint8  | Flags       | Bit flags selecting chunk types (see below) |

The server constructs the filename as:

```
{CategoryID}_0_0_0_S{MainID}_T{Flags}_C{ChapterID}.bin   (or .json)
```

## Flags (Chunk Type Selection)

The `Flags` byte is a bitmask that selects which chunk types the client requests:

| Bit | Value | Format          | Content |
|-----|-------|-----------------|---------|
| 0   | 0x01  | Sub-header      | Quest name/description (chunk0) |
| 1   | 0x02  | Sub-header      | NPC dialog (chunk1) |
| 2   | 0x04  | —               | Unknown (no instances found) |
| 3   | 0x08  | Inline          | Episode listing (chunk0 inline) |
| 4   | 0x10  | JKR-compressed  | NPC dialog blob (chunk1) |
| 5   | 0x20  | JKR-compressed  | Menu options or quest titles (chunk2) |
| 6   | 0x40  | —               | Unknown (no instances found) |
| 7   | 0x80  | —               | Unknown (no instances found) |

The flags are part of the filename — each unique `(CategoryID, MainID, Flags, ChapterID)` tuple corresponds to its own file on disk.

## Container Format (big-endian)

```
Offset         Field
@0x00          u32 BE   chunk0_size
@0x04          u32 BE   chunk1_size
@0x08          bytes    chunk0_data  (chunk0_size bytes)
@0x08+c0       bytes    chunk1_data  (chunk1_size bytes)
@0x08+c0+c1    u32 BE   chunk2_size  (only present if file continues)
               bytes    chunk2_data  (chunk2_size bytes)
```

The 8-byte header is always present. Chunks with size 0 are absent. Chunk2 is only read if at least 4 bytes remain after chunk0+chunk1.

## Chunk Formats

### Sub-header Format (flags 0x01, 0x02)

Used for structured text chunks containing named strings with metadata.

**Sub-header (8 bytes, fields at byte offsets within the chunk):**

| Off | Type    | Field        | Notes |
|-----|---------|--------------|-------|
| 0   | u8      | Type         | Usually `0x01` |
| 1   | u8      | Pad          | Always `0x00`; used to detect this format vs inline |
| 2   | u16 LE  | TotalSize    | Total chunk size including this header |
| 4   | u8      | EntryCount   | Number of string entries |
| 5   | u8      | Unknown1     | Unknown; preserved in JSON for round-trip |
| 6   | u8      | MetadataSize | Total bytes of the metadata block that follows |
| 7   | u8      | Unknown2     | Unknown; preserved in JSON for round-trip |

**Layout after the 8-byte header:**

```
[MetadataSize bytes: opaque metadata block]
[null-terminated Shift-JIS string #1]
[null-terminated Shift-JIS string #2]
...
[0xFF end-of-strings sentinel]
```

**Metadata block** (partially understood):

The metadata block is `MetadataSize` bytes long and covers all entries collectively. Known sizes observed in real files:

- Chunk0 (flag 0x01): `MetadataSize = 0x14` (20 bytes)
- Chunk1 (flag 0x02): `MetadataSize = 0x2C` (44 bytes)

The internal structure of the metadata is not yet fully documented. It is preserved verbatim in the JSON format as a base64 blob so that clients receive correct values even for unknown fields.

**Format detection for chunk0:** if `chunk_data[1] == 0x00` → sub-header, else → inline.

### Inline Format (flag 0x08)

Used for episode listings. Each entry is:

```
{u8 index}{null-terminated Shift-JIS string}
```

Entries are sequential with no separator. Null bytes between entries are ignored during parsing.

### JKR-compressed Chunks (flags 0x10, 0x20)

Chunks with flags 0x10 (chunk1) and 0x20 (chunk2) are JKR-compressed blobs. The JKR header (magic `0x1A524B4A`) appears at the start of the chunk data.

The decompressed content contains metadata bytes interleaved with null-terminated Shift-JIS strings, but the detailed format is not yet fully documented. These chunks are stored as opaque base64 blobs in the JSON format and served to the client unchanged.

## JSON Format (for `.json` scenario files)

Erupe supports `.json` files in `bin/scenarios/` as an alternative to `.bin` files. The server compiles `.json` to wire format on demand. `.bin` takes priority if both exist.

Example `0_0_0_0_S102_T1_C0.json`:

```json
{
  "chunk0": {
    "subheader": {
      "type": 1,
      "unknown1": 0,
      "unknown2": 0,
      "metadata": "AAAAAAAAAAAAAAAAAAAAAAAAAAAA",
      "strings": ["Quest Name", "Quest description goes here."]
    }
  }
}
```

Example with inline chunk0 (flag 0x08):

```json
{
  "chunk0": {
    "inline": [
      {"index": 1, "text": "Chapter 1"},
      {"index": 2, "text": "Chapter 2"}
    ]
  }
}
```

Example with both chunk0 and chunk1:

```json
{
  "chunk0": {
    "subheader": {
      "type": 1, "unknown1": 0, "unknown2": 0,
      "metadata": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
      "strings": ["Quest Name"]
    }
  },
  "chunk1": {
    "subheader": {
      "type": 1, "unknown1": 0, "unknown2": 0,
      "metadata": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
      "strings": ["NPC: Welcome, hunter.", "NPC: Good luck!"]
    }
  }
}
```

**Key fields:**

- `metadata`: Base64-encoded opaque blob. Copy from `ParseScenarioBinary` output. For new scenarios with zero-filled metadata, use a base64 string of the right number of zero bytes.
- `strings`: UTF-8 text. The compiler converts to Shift-JIS on the wire.
- `chunk2.data`: Raw JKR-compressed bytes, base64-encoded. Copy from the original `.bin` file.

## JKR Compression

Chunks with flags 0x10 and 0x20 use JKR compression (magic `0x1A524B4A`, type 3 LZ77). The Go compressor is in `common/decryption.PackSimple` and the decompressor in `common/decryption.UnpackSimple`. These implement type-3 (LZ-only) compression, which is the format used throughout Erupe.

Type-4 (HFI = Huffman + LZ77) JKR blobs from real game files pass through as opaque base64 in `.json` — the server serves them as-is without re-compression.

## Implementation

- **Handler**: `server/channelserver/handlers_quest.go` → `handleMsgSysGetFile` → `loadScenarioBinary`
- **JSON schema + compiler**: `server/channelserver/scenario_json.go`
- **JKR compressor**: `common/decryption/jpk_compress.go` (`PackSimple`)
- **JKR decompressor**: `common/decryption/jpk.go` (`UnpackSimple`)
