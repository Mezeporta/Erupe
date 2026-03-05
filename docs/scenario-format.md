# Scenario Binary Format

> Reference: `network/mhfpacket/msg_sys_get_file.go`, issue [#172](https://github.com/Mezeporta/Erupe/issues/172)

## Overview

Scenario files are binary blobs served by `MSG_SYS_GET_FILE` when `IsScenario` is true. They contain quest descriptions, NPC dialog, episode listings, and menu options for the game's scenario/story system.

## Request Format

When `IsScenario == true`, the client sends a `scenarioFileIdentifier`:

| Offset | Type   | Field       | Description |
|--------|--------|-------------|-------------|
| 0      | uint8  | CategoryID  | Scenario category |
| 1      | uint32 | MainID      | Main scenario identifier |
| 5      | uint8  | ChapterID   | Chapter within the scenario |
| 6      | uint8  | Flags       | Bit flags selecting chunk types (see below) |

## Flags (Chunk Type Selection)

The `Flags` byte is a bitmask that selects which chunk types the client requests:

| Bit  | Value | Type    | Recursive | Content |
|------|-------|---------|-----------|---------|
| 0    | 0x01  | Chunk0  | Yes       | Quest name/description + 0x14 byte info block |
| 1    | 0x02  | Chunk1  | Yes       | NPC dialog(?) + 0x2C byte info block |
| 2    | 0x04  | —       | —         | Unknown (no instances found; possibly Chunk2) |
| 3    | 0x08  | Chunk0  | No        | Episode listing (0x1 prefixed?) |
| 4    | 0x10  | Chunk1  | No        | JKR-compressed blob, NPC dialog(?) |
| 5    | 0x20  | Chunk2  | No        | JKR-compressed blob, menu options or quest titles(?) |
| 6    | 0x40  | —       | —         | Unknown (no instances found) |
| 7    | 0x80  | —       | —         | Unknown (no instances found) |

### Chunk Types

- **Chunk0**: Contains text data (quest names, descriptions, episode titles) with an accompanying fixed-size info block.
- **Chunk1**: Contains dialog or narrative text with a larger info block (0x2C bytes).
- **Chunk2**: Contains menu/selection text.

### Recursive vs Non-Recursive

- **Recursive chunks** (flags 0x01, 0x02): The chunk data itself contains nested sub-chunks that must be parsed recursively.
- **Non-recursive chunks** (flags 0x08, 0x10, 0x20): The chunk is a flat binary blob. Flags 0x10 and 0x20 are JKR-compressed and must be decompressed before reading.

## Response Format

The server responds with the scenario file data via `doAckBufSucceed`. The response is the raw binary blob matching the requested chunk types. If the scenario file is not found, the server sends `doAckBufFail` to prevent a client crash.

## Current Implementation

Scenario files are loaded from `quests/scenarios/` on disk. The server currently serves them as opaque binary blobs with no parsing. Issue #172 proposes adding JSON/CSV support for easier editing, which would require implementing a parser/serializer for this format.

## JKR Compression

Chunks with flags 0x10 and 0x20 use JPK compression (magic bytes `0x1A524B4A`). See the ReFrontier tool for decompression utilities.
