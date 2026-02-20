# Erupe Codebase Anti-Patterns Analysis

> Analysis date: 2026-02-20

## Table of Contents

- [1. God Files — Massive Handler Files](#1-god-files--massive-handler-files)
- [2. Silently Swallowed Errors](#2-silently-swallowed-errors)
- [3. No Architectural Layering](#3-no-architectural-layering--handlers-do-everything)
- [4. Magic Numbers Everywhere](#4-magic-numbers-everywhere)
- [5. Inconsistent Binary I/O Patterns](#5-inconsistent-binary-io-patterns)
- [6. Session God Object](#6-session-struct-is-a-god-object)
- [7. Mutex Granularity Issues](#7-mutex-granularity-issues)
- [8. Copy-Paste Handler Patterns](#8-copy-paste-handler-patterns)
- [9. Raw SQL Scattered in Handlers](#9-raw-sql-strings-scattered-in-handlers)
- [10. init() Handler Registration](#10-init-function-for-handler-registration)
- [11. Panic-Based Flow](#11-panic-based-flow-in-some-paths)
- [12. Inconsistent Logging](#12-inconsistent-logging)
- [13. Tight Coupling to PostgreSQL](#13-tight-coupling-to-postgresql)
- [Summary](#summary-by-severity)

---

## 1. God Files — Massive Handler Files

The channel server has large handler files, each mixing DB queries, business logic, binary serialization, and response writing with no layering. Actual line counts (non-test files):

| File | Lines | Purpose |
|------|-------|---------|
| `server/channelserver/handlers_session.go` | 794 | Session setup/teardown |
| `server/channelserver/handlers_data_paper_tables.go` | 765 | Paper table data |
| `server/channelserver/handlers_quest.go` | 722 | Quest lifecycle |
| `server/channelserver/handlers_house.go` | 638 | Housing system |
| `server/channelserver/handlers_festa.go` | 637 | Festival events |
| `server/channelserver/handlers_data_paper.go` | 621 | Paper/data system |
| `server/channelserver/handlers_tower.go` | 529 | Tower gameplay |
| `server/channelserver/handlers_mercenary.go` | 495 | Mercenary system |
| `server/channelserver/handlers_stage.go` | 492 | Stage/lobby management |
| `server/channelserver/handlers_guild_info.go` | 473 | Guild info queries |

These sizes (~500-800 lines) are not extreme by Go standards, but the files mix all architectural concerns. The bigger problem is the lack of layering within each file (see [#3](#3-no-architectural-layering--handlers-do-everything)), not the file sizes themselves.

**Impact:** Each handler function is a monolith mixing data access, business logic, and protocol serialization. Testing or reusing any single concern is impossible.

---

## 2. Missing ACK Responses on Error Paths (Client Softlocks)

Some handler error paths log the error and return without sending any ACK response to the client. The MHF client uses `MsgSysAck` with an `ErrorCode` field (0 = success, 1 = failure) to complete request/response cycles. When no ACK is sent at all, the client softlocks waiting for a response that never arrives.

### The three error handling patterns in the codebase

**Pattern A — Silent return (the bug):** Error logged, no ACK sent, client hangs.

```go
if err != nil {
    s.logger.Error("Failed to get ...", zap.Error(err))
    return  // BUG: client gets no response, softlocks
}
```

**Pattern B — Log and continue (acceptable):** Error logged, handler continues and sends a success ACK with default/empty data. The client proceeds with fallback behavior.

```go
if err != nil {
    s.logger.Error("Failed to load mezfes data", zap.Error(err))
}
// Falls through to doAckBufSucceed with empty data
```

**Pattern C — Fail ACK (correct):** Error logged, explicit fail ACK sent. The client shows an appropriate error dialog and stays connected.

```go
if err != nil {
    s.logger.Error("Failed to read rengoku_data.bin", zap.Error(err))
    doAckBufFail(s, pkt.AckHandle, nil)
    return
}
```

### Evidence that fail ACKs are safe

The codebase already sends ~70 `doAckSimpleFail`/`doAckBufFail` calls in production handler code across 15 files. The client handles them gracefully in all observed cases:

| File | Fail ACKs | Client behavior |
|------|-----------|-----------------|
| `handlers_guild_scout.go` | 17 | Guild recruitment error dialogs |
| `handlers_guild_ops.go` | 10 | Permission denied, guild not found dialogs |
| `handlers_stage.go` | 8 | "Room is full", "wrong password", "stage locked" |
| `handlers_house.go` | 6 | Wrong password, invalid box index |
| `handlers_guild.go` | 9 | Guild icon update errors, unimplemented features |
| `handlers_guild_alliance.go` | 4 | Alliance permission errors |
| `handlers_data.go` | 4 | Decompression failures, oversized payloads |
| `handlers_festa.go` | 4 | Festival entry errors |
| `handlers_quest.go` | 3 | Missing quest/scenario files |

A comment in `handlers_quest.go:188` explicitly documents the mechanism:

> sends doAckBufFail, which triggers the client's error dialog (snj_questd_matching_fail → SetDialogData) instead of a softlock

The original `mhfo-hd.dll` client reads the `ErrorCode` byte from `MsgSysAck` and dispatches to per-message error UI. A fail ACK causes the client to show an error dialog and remain functional. A missing ACK causes a softlock.

### Scope

A preliminary grep for `logger.Error` followed by bare `return` (no doAck call) found instances across ~25 handler files. However, a thorough manual audit (2026-02-20) revealed that the vast majority are Pattern B (log-and-continue to a success ACK with empty data) or Pattern C (explicit fail ACK). Only one true Pattern A instance was found, in `handleMsgSysOperateRegister` (`handlers_register.go`), which has been fixed.

**Status:** ~~Players experience softlocks on error paths.~~ **Fixed.** The last Pattern A instance (`handlers_register.go:62`) now sends `doAckBufSucceed` with nil data before returning. The ~87 existing `doAckSimpleFail`/`doAckBufFail` calls and the helper functions (`loadCharacterData`, `saveCharacterData`, `stubEnumerateNoResults`) provide comprehensive ACK coverage across all handler error paths.

---

## 3. No Architectural Layering — Handlers Do Everything

Handler functions directly embed raw SQL, binary parsing, business logic, and response building in a single function body. For example, a typical guild handler will:

1. Parse the incoming packet
2. Run 3-5 inline SQL queries
3. Apply business logic (permission checks, state transitions)
4. Manually serialize a binary response

```go
func handleMsgMhfCreateGuild(s *Session, p mhfpacket.MHFPacket) {
    pkt := p.(*mhfpacket.MsgMhfCreateGuild)

    // Direct SQL in the handler
    var guildCount int
    err := s.Server.DB.QueryRow("SELECT count(*) FROM guilds WHERE leader_id=$1", s.CharID).Scan(&guildCount)
    if err != nil {
        s.logger.Error(...)
        return
    }

    // Business logic inline
    if guildCount > 0 { ... }

    // More SQL
    _, err = s.Server.DB.Exec("INSERT INTO guilds ...")

    // Binary response building
    bf := byteframe.NewByteFrame()
    bf.WriteUint32(...)
    doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}
```

There is no repository layer, no service layer — just handlers.

**Impact:** Testing individual concerns is impossible without a real database and a full session. Business logic can't be reused. Schema changes require updating dozens of handler files.

**Recommendation:** Introduce at minimum a repository layer for data access and a service layer for business logic. Handlers should only deal with packet parsing and response serialization.

---

## 4. ~~Magic Numbers Everywhere~~ (Substantially Fixed)

**Status:** Two rounds of extraction have replaced the highest-impact magic numbers with named constants:

- **Round 1** (commit `7c444b0`): `constants_quest.go`, `handlers_guild_info.go`, `handlers_quest.go`, `handlers_rengoku.go`, `handlers_session.go`, `model_character.go`
- **Round 2**: `constants_time.go` (shared `secsPerDay`, `secsPerWeek`), `constants_raviente.go` (register IDs, semaphore constants), plus constants in `handlers_register.go`, `handlers_semaphore.go`, `handlers_session.go`, `handlers_festa.go`, `handlers_diva.go`, `handlers_event.go`, `handlers_mercenary.go`, `handlers_misc.go`, `handlers_plate.go`, `handlers_cast_binary.go`, `handlers_commands.go`, `handlers_reward.go`, `handlers_guild_mission.go`, `sys_channel_server.go`

**Remaining:** Unknown protocol fields (e.g., `handlers_diva.go:112-115` `0x19, 0x2D, 0x02, 0x02`) are intentionally left as literals until their meaning is understood. Data tables (monster point tables, item IDs) are data, not protocol constants. Standard empty ACK payloads (`make([]byte, 4)`) are idiomatic Go.

**Impact:** ~~New contributors can't understand what these values mean.~~ Most protocol-meaningful constants now have names and comments.

---

## 5. ~~Inconsistent Binary I/O Patterns~~ (Resolved)

**Status:** Non-issue on closer inspection. The codebase has already standardized on `byteframe` for all sequential packet building and parsing.

The 12 remaining `encoding/binary` call sites (across `sys_session.go`, `handlers_session.go`, `model_character.go`, `handlers_quest.go`, `handlers_rengoku.go`) are all cases where `byteframe` is structurally wrong:

- **Zero-allocation spot-reads on existing `[]byte`** — reading an opcode or ack handle from an already-serialized packet for logging, or sentinel guard checks on raw blobs. Allocating a byteframe for a 2-byte read in a log path would be wasteful.
- **Random-access reads/writes at computed offsets** — patching fields in the decompressed game save blob (`model_character.go`) or copying fields within quest binaries during version backport (`handlers_quest.go`). Byteframe is a sequential cursor and cannot do `buf[offset:offset+4]` style access.

Pattern C (raw `data[i] = byte(...)` serialization) does not exist in production code — only in test fixtures as loop fills for dummy payloads.

---

## 6. Session Struct is a God Object

`sys_session.go` defines a `Session` struct that carries everything a handler could possibly need:

- Database connection (`*sql.DB`)
- Logger
- Server reference (which itself contains more shared state)
- Character state (ID, name, stats)
- Stage/lobby state
- Semaphore state
- Send channels
- Various flags and locks

Every handler receives this god object, coupling all handlers to the entire server's internal state.

**Impact:** Any handler can modify any part of the session or server state. There's no encapsulation. Testing requires constructing a fully populated Session with all dependencies. It's unclear which fields a given handler actually needs.

**Recommendation:** Pass narrower interfaces to handlers (e.g., a `DBQuerier` interface instead of the full server, a `ResponseWriter` instead of the raw send channel).

---

## 7. Mutex Granularity Issues

`sys_stage.go` and `sys_channel_server.go` use coarse-grained `sync.RWMutex` locks on entire maps:

```go
// A single lock for ALL stages
s.stageMapLock.Lock()
defer s.stageMapLock.Unlock()
// Any operation on any stage blocks all other stage operations
```

The Raviente shared state uses a single mutex for all Raviente data fields.

**Impact:** Contention scales with player count. Operations on unrelated stages block each other unnecessarily. Under load, this becomes a bottleneck.

**Recommendation:** Use per-stage locks (e.g., `sync.Map` or a map of per-key mutexes) so operations on different stages don't contend. For Raviente, consider splitting the mutex by data group.

---

## 8. Copy-Paste Handler Patterns

~~Many handlers follow an identical template with minor variations but no shared abstraction.~~ **Substantially fixed.** `loadCharacterData` and `saveCharacterData` helpers in `handlers_helpers.go` now cover all standard character blob load/save patterns (11 load handlers, 6 save handlers including `handleMsgMhfSaveScenarioData`). The `saveCharacterData` helper sends `doAckSimpleFail` on oversized payloads and DB errors, matching the correct error-handling pattern.

Remaining inline DB patterns were audited and are genuinely different (non-blob types, wrong tables, diff compression, read-modify-write with bit ops, multi-column updates, or queries against other characters).

---

## 9. Raw SQL Strings Scattered in Handlers

SQL queries are string literals directly embedded in handler functions with no constants, no query builder, and no repository abstraction:

```go
err := s.Server.DB.QueryRow(
    "SELECT id, name, leader_id, ... FROM guilds WHERE id=$1", guildID,
).Scan(&id, &name, &leaderID, ...)
```

The same table is queried in different handlers with slightly different column sets and joins.

**Impact:** Schema changes (renaming a column, adding a field) require finding and updating every handler that touches that table. There's no way to ensure all queries stay in sync. SQL injection risk is low (parameterized queries are used), but query correctness is hard to verify.

**Recommendation:** At minimum, define query constants. Ideally, introduce a repository layer that encapsulates all queries for a given entity.

**Status (partial):** A `CharacterRepository` layer has been introduced in `repo_character.go`, centralizing all `characters` table access behind a concrete struct. The 4 existing helpers (`loadCharacterData`, `saveCharacterData`, `readCharacterInt`, `adjustCharacterInt`) now delegate to the repository, covering ~70% of character queries. Direct queries in `handlers_session.go` (login/logout), `sys_channel_server.go` (`DisconnectUser`), and `handlers_mail.go` (name lookup) have also been migrated. Remaining work: guild repository (second-highest duplication), per-handler migration of remaining inline character queries (plate, mercenary, rengoku, cafe, clients), and column allowlist for SQL injection hardening.

---

## 10. init() Function for Handler Registration

`handlers_table.go` uses a massive `init()` function to register ~200+ handlers in a global map:

```go
func init() {
    handlers[network.MsgMhfSaveFoo] = handleMsgMhfSaveFoo
    handlers[network.MsgMhfLoadFoo] = handleMsgMhfLoadFoo
    // ... 200+ more entries
}
```

**Impact:** Registration is implicit and happens at package load time. It's impossible to selectively register handlers (e.g., for testing). The handler map can't be mocked. The `init()` function is ~200+ lines of boilerplate.

**Recommendation:** Use explicit registration (a function called from `main` or server setup) that builds and returns the handler map.

---

## 11. Panic-Based Flow in Some Paths

~~Some error paths use `panic()` or `log.Fatal()` (which calls `os.Exit`) instead of returning errors.~~ **Substantially fixed.** The 5 production `panic()` calls (4 in mhfpacket `Build()` stubs, 1 in binpacket `Parse()`) have been replaced with `fmt.Errorf` returns. The `byteframe.go` read-overflow panic has been replaced with a sticky error pattern (`ByteFrame.Err()`), and the packet dispatch loop in `sys_session.go` now checks `bf.Err()` after parsing to reject malformed packets cleanly.

**Remaining:** The `recover()` in `handlePacketGroup` is retained as a safety net for any future unexpected panics.

---

## 12. Inconsistent Logging

The codebase mixes logging approaches:

- `zap.Logger` (structured logging) — primary approach
- Remnants of `fmt.Printf` / `log.Printf` in some packages
- Some packages accept a logger parameter, others create their own

**Impact:** Log output format is inconsistent. Some logs lack structure (no fields, no levels). Filtering and aggregation in production is harder.

**Recommendation:** Standardize on `zap.Logger` everywhere. Pass the logger via dependency injection. Remove all `fmt.Printf` / `log.Printf` usage from non-CLI code.

---

## 13. Tight Coupling to PostgreSQL

Database operations use raw `database/sql` with PostgreSQL-specific syntax throughout:

- `$1` parameter placeholders (PostgreSQL-specific)
- PostgreSQL-specific types and functions in queries
- `*sql.DB` passed directly through the server struct to every handler
- No interface abstraction over data access

**Impact:** Unit tests require a real PostgreSQL instance. Storage can't be swapped (e.g., SQLite for development). Mocking data access for handler tests is impossible.

**Recommendation:** While PostgreSQL is the correct production choice, introducing a repository interface would enable in-memory or mock implementations for testing.

---

## Summary by Severity

| Severity | Anti-patterns |
|----------|--------------|
| **High** | ~~Missing ACK responses / softlocks (#2)~~ **Fixed**, no architectural layering (#3), tight DB coupling (#13) |
| **Medium** | ~~Magic numbers (#4)~~ **Fixed**, ~~inconsistent binary I/O (#5)~~ **Resolved**, Session god object (#6), ~~copy-paste handlers (#8)~~ **Fixed**, raw SQL duplication (#9) |
| **Low** | God files (#1), ~~`init()` registration (#10)~~ **Fixed**, ~~inconsistent logging (#12)~~ **Fixed**, mutex granularity (#7), ~~panic-based flow (#11)~~ **Fixed** |

### Root Cause

Most of these anti-patterns stem from a single root cause: **the codebase grew organically from a protocol reverse-engineering effort without introducing architectural boundaries**. When the primary goal is "make this packet work," it's natural to put the SQL, logic, and response all in one function. Over time, this produces the pattern seen here — hundreds of handler functions that each independently implement the full stack.

### Recommended Refactoring Priority

1. **Add fail ACKs to silent error paths** — prevents player softlocks, ~70 existing doAckFail calls prove safety, low risk, can be done handler-by-handler
2. **Extract a character repository layer** — 152 queries across 26 files touch the `characters` table, highest SQL duplication
3. **Extract load/save helpers** — 38 handlers repeat the same ~10-15 line template, mechanical extraction
4. **Extract a guild repository layer** — 32 queries across 8-15 files, second-highest SQL duplication
5. **Define protocol constants** — 1,052 hex literals with 174 unique values, improves documentation
6. ~~**Standardize binary I/O**~~ — already standardized on `byteframe`; remaining `encoding/binary` uses are correct (see #5)
