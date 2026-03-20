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

**Client-side size limits (confirmed from `FUN_11525c60` in `mhfo-hd.dll`):** each chunk is silently dropped (treated as size 0) if its size exceeds `0x8000` bytes (32 768). The client allocates three fixed 0x8000-byte buffers — one per chunk — so the server must not serve chunks larger than that limit.

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

**Metadata block** (partially decoded):

The metadata block is `MetadataSize` bytes long. Known sizes from real files:

- Chunk0 (flag 0x01): `MetadataSize = 0x14` (20 bytes = 10 × u16 LE)
- Chunk1 (flag 0x02): `MetadataSize = 0x2C` (44 bytes = 22 × u16 LE)

**Chunk0 metadata (20 bytes decoded from 145,000+ real scenario files):**

Client parser (`FUN_1080d310` in `mhfo-hd.dll`) extracts only m[0]–m[6]; fields m[7]–m[9] are not read.

| u16 index | Field | Notes |
|-----------|-------|-------|
| m[0] | CategoryID | Matches the first field of the filename (0=basic, 1=GR, 3=exchange, 6=pallone, 7=diva) |
| m[1] | MainID | Matches the `S` field of the filename |
| m[2] | 0x0000 | Always zero; used as offset to string 0 (i.e., strings section start = str0 start) |
| m[3–4] | 0x0000 | Reserved / always zero; not used by client |
| m[5] | str0_len | Byte length of string 0 in Shift-JIS including the null terminator; used as offset to string 1 |
| m[6] | SceneRef | `MainID` when CategoryID=0; `0xFFFF` (−1 as s16) when CategoryID≠0 — stored in client struct as signed short; purpose unclear |
| m[7] | 0x0000 | Not read by client parser |
| m[8] | 0x0005 | Not read by client parser; constant whose purpose is unknown |
| m[9] | varies | Not read by client parser |

**Chunk1 metadata (44 bytes decoded from multi-dialog scenario files):**

The 22 u16 fields encode string offsets and dialog script positions. Client parser (`FUN_1080d3b0`) interprets m[8]–m[17] as **signed** offsets: if the value is negative (as s16), the absolute position is `(~value) + dialog_base` where `dialog_base` is the start of the post-0xFF binary data; if non-negative, the position is `value + strings_base`.

| u16 index | Field | Notes |
|-----------|-------|-------|
| m[0] | ID byte 0 | Low byte only is read; typically 0 |
| m[1] | ID byte 1 | High byte only is read; varies |
| m[1] (u16) | TotalSize copy | Bytes 2–3 read as u16 LE; mirrors the sub-header TotalSize |
| m[2] | EntryCount (s16) | Read as signed short; number of strings or related count |
| m[3] | u16 at offset 6 | Read as u16 |
| m[4–5] | u32 at offset 8 | Read as single u32 |
| m[6] | u16 at offset 12 | Read as u16 |
| m[7] | u16 at offset 14 | Read as u16 |
| m[8] | signed offset | String/dialog pointer (see signed offset formula above) |
| m[9] | cumOff[2] | Byte offset to string 2 from strings section start (= str0_len + str1_len) |
| m[10] | cumOff[1] | Byte offset to string 1 = str0_len |
| m[11] | dialog offset | Offset into the post-0xFF dialog data section |
| m[12] | dialog offset | Offset into the post-0xFF dialog data section |
| m[13] | dialog offset | Offset into the post-0xFF dialog data section |
| m[14] | cumOff[3] | Byte offset to string 3 |
| m[15] | cumOff[4] | Total string bytes without the 0xFF sentinel |
| m[16] | dialog offset | Further offset into the post-0xFF dialog data section |
| m[17] | signed offset | Final offset; if negative, `(~m[17]) + dialog_base`; byte at m[18]×2 is also read |
| m[18–19] | byte fields | Individual bytes read (not as u16 pairs) |
| m[20] | 0x0005 | Constant (same as chunk0 m[8]); not confirmed whether client reads this |
| m[21] | DataSize − 4 | Approximately equal to `chunk1_size − 8 − MetadataSize + 4` |

The metadata is preserved verbatim in JSON as a base64 blob so that clients receive correct values for all fields including those not yet fully understood.

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
