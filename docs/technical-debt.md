# Erupe Technical Debt & Suggested Next Steps

> Last updated: 2026-03-05

This document tracks actionable technical debt items discovered during a codebase audit. It complements `anti-patterns.md` (which covers structural patterns) by focusing on specific, fixable items with file paths and line numbers.

## Table of Contents

- [High Priority](#high-priority)
  - [1. Broken game features (gameplay-impacting TODOs)](#1-broken-game-features-gameplay-impacting-todos)
  - [2. Test gaps on critical paths](#2-test-gaps-on-critical-paths)
- [Medium Priority](#medium-priority)
  - [3. Logging anti-patterns](#3-logging-anti-patterns)
- [Low Priority](#low-priority)
  - [4. CI updates](#4-ci-updates)
- [Completed Items](#completed-items)
- [Suggested Execution Order](#suggested-execution-order)

---

## High Priority

### 1. Broken game features (gameplay-impacting TODOs)

These TODOs represent features that are visibly broken for players.

| Location | Issue | Impact | Tracker |
|----------|-------|--------|---------|
| ~~`model_character.go:88,101,113`~~ | ~~`TODO: fix bookshelf data pointer` for G10-ZZ, F4-F5, and S6 versions~~ | ~~Wrong pointer corrupts character save reads for three game versions.~~ **Fixed.** Corrected offsets to 103928 (G1–Z2), 71928 (F4–F5), 23928 (S6) — validated via inter-version delta analysis and Ghidra decompilation of `snj_db_get_housedata` in the ZZ DLL. | [#164](https://github.com/Mezeporta/Erupe/issues/164) |
| ~~`handlers_achievement.go:117`~~ | ~~`TODO: Notify on rank increase` — always returns `false`~~ | ~~Achievement rank-up notifications are silently suppressed.~~ **Fixed.** RE'd `putDisplayed_achievement` in ZZ DLL: sends opcode + 1 zero byte (no achievement ID). Added `displayed_levels` column to track per-character last-displayed levels; `GetAchievement` compares current vs displayed; `DisplayedAchievement` snapshots current levels. | [#165](https://github.com/Mezeporta/Erupe/issues/165) |
| ~~`handlers_guild_info.go:443`~~ | ~~`TODO: Enable GuildAlliance applications` — hardcoded `true`~~ | ~~Guild alliance applications are always open regardless of setting.~~ **Fixed.** Added `recruiting` column to `guild_alliances`, wired `OperateJoint` actions `0x06`/`0x07`, reads from DB. | [#166](https://github.com/Mezeporta/Erupe/issues/166) |
| ~~`handlers_session.go:410`~~ | ~~`TODO(Andoryuuta): log key index off-by-one`~~ | ~~Known off-by-one in log key indexing is unresolved~~ **Documented.** RE'd from ZZ DLL: `putRecord_log`/`putTerminal_log` don't embed the key (size 0), so the off-by-one only matters in pre-ZZ clients and is benign server-side. | [#167](https://github.com/Mezeporta/Erupe/issues/167) |
| ~~`handlers_session.go:551`~~ | ~~`TODO: This case might be <=G2`~~ | ~~Uncertain version detection in switch case~~ **Documented.** RE'd ZZ per-entry parser (FUN_115868a0) confirms 40-byte padding. G2 DLL analysis inconclusive (stripped, no shared struct sizes). Kept <=G1 boundary with RE documentation. | [#167](https://github.com/Mezeporta/Erupe/issues/167) |
| ~~`handlers_session.go:714`~~ | ~~`TODO: Retail returned the number of clients in quests`~~ | ~~Player count reported to clients does not match retail behavior~~ **Fixed.** Added `QuestReserved` field to `StageSnapshot` that counts only clients in "Qs" stages, pre-collected under server lock to respect lock ordering. | [#167](https://github.com/Mezeporta/Erupe/issues/167) |
| `msg_mhf_add_ud_point.go:28` | `TODO: Parse is a stub` — field meanings unknown | UD point packet fields unnamed, `Build` not implemented | [#168](https://github.com/Mezeporta/Erupe/issues/168) |

### 2. Test gaps on critical paths

**All handler files now have test coverage.**

~~**Repository files with no store-level test file (17 total):**~~ **Fixed.** All 20 repo source files now have corresponding `_test.go` files. The split guild files (`repo_guild_adventure.go`, `repo_guild_alliance.go`, etc.) are covered by `repo_guild_test.go`. These are mock-based unit tests; SQL-level integration tests against a live database remain a future goal.

---

## Medium Priority

### 3. Logging anti-patterns

~~**a) `fmt.Sprintf` inside structured logger calls (6 sites):**~~ **Fixed.** All 6 sites now use `zap.Uint32`/`zap.Uint8`/`zap.String` structured fields instead of `fmt.Sprintf`.

~~**b) 20+ silently discarded SJIS encoding errors in packet parsing:**~~ **Fixed.** All call sites now use `SJISToUTF8Lossy()` which logs decode errors at `slog.Debug` level.

---

## Low Priority

### 4. CI updates

- ~~`codecov-action@v4` could be updated to `v5` (current stable)~~ **Removed.** Replaced with local `go tool cover` threshold check (no Codecov account needed).
- ~~No coverage threshold is enforced — coverage is uploaded but regressions aren't caught~~ **Fixed.** CI now fails if total coverage drops below 50% (current: ~58%).

---

## Completed Items

Items resolved since the original audit:

| # | Item | Resolution |
|---|------|------------|
| ~~3~~ | **Sign server has no repository layer** | Fully refactored with `repo_interfaces.go`, `repo_user.go`, `repo_session.go`, `repo_character.go`, and mock tests. All 8 previously-discarded error paths are now handled. |
| ~~4~~ | **Split `repo_guild.go`** | Split from 1004 lines into domain-focused files: `repo_guild.go` (466 lines, core CRUD), `repo_guild_posts.go`, `repo_guild_alliance.go`, `repo_guild_adventure.go`, `repo_guild_hunt.go`, `repo_guild_cooking.go`, `repo_guild_rp.go`. |
| ~~6~~ | **Inconsistent transaction API** | All call sites now use `BeginTxx(context.Background(), nil)` with deferred rollback. |
| ~~7~~ | **`LoopDelay` config has no Viper default** | `viper.SetDefault("LoopDelay", 50)` added in `config/config.go`. |
| — | **Monthly guild item claim** (`handlers_guild.go:389`) | Now tracks per-character per-type monthly claims via `stamps` table. |
| — | **Handler test coverage (4 files)** | Tests added for `handlers_session.go`, `handlers_gacha.go`, `handlers_plate.go`, `handlers_shop.go`. |
| — | **Handler test coverage (`handlers_commands.go`)** | 62 tests covering all 12 commands, disabled-command gating, op overrides, error paths, raviente with semaphore, course enable/disable/locked, reload with players/objects. |
| — | **Handler test coverage (`handlers_data_paper.go`)** | 20 tests covering all DataType branches (0/5/6/gift/>1000/unknown), ACK payload structure, earth succeed entry counts, timetable content, serialization round-trips, and paperGiftData table integrity. |
| — | **Handler test coverage (5 files)** | Tests added for `handlers_seibattle.go` (9 tests), `handlers_kouryou.go` (7 tests), `handlers_scenario.go` (6 tests), `handlers_distitem.go` (8 tests), `handlers_guild_mission.go` (5 tests in coverage5). |
| — | **Entrance server raw SQL** | Refactored to repository interfaces (`repo_interfaces.go`, `repo_session.go`, `repo_server.go`). |
| — | **Guild daily RP rollover** (`handlers_guild_ops.go:148`) | Implemented via lazy rollover in `handlers_guild.go:110-119` using `RolloverDailyRP()`. Stale TODO removed. |
| — | **Typos** (`sys_session.go`, `handlers_session.go`) | "For Debuging" and "offical" typos already fixed in previous commits. |
| — | **`db != nil` guard** (`handlers_session.go:322`) | Investigated — this guard is intentional. Test servers run without repos; the guard protects the entire logout path from nil repo dereferences. Not a leaky abstraction. |
| ~~2~~ | **Repo test coverage (17 files)** | All 20 repo source files now have `_test.go` files with mock-based unit tests. |
| — | **Bookshelf data pointer** ([#164](https://github.com/Mezeporta/Erupe/issues/164)) | Corrected `pBookshelfData` offsets for G1–Z2 (103928), F4–F5 (71928), S6 (23928). Validated via inter-version delta analysis and Ghidra decompilation of ZZ `snj_db_get_housedata`. |
| — | **Guild nil panics** ([#171](https://github.com/Mezeporta/Erupe/issues/171)) | Three fixes merged post-RC1: nil guards for alliance guild lookups (aee5353), variable shadowing fix in scout list (8e79fe6), nil guards in cancel/answer scout handlers (8717fb9). Clan hall softlock resolved. |
| — | **ecdMagic byte order** ([#174](https://github.com/Mezeporta/Erupe/issues/174)) | Corrected constant byte order in `crypt_conn.go` (10ac803). |
| — | **Rengoku caching** | Cached `rengoku_data.bin` at startup to avoid repeated disk reads (5b631d1). |
| — | **rasta_id=0 save issue** ([#163](https://github.com/Mezeporta/Erupe/issues/163)) | `SaveMercenary` now skips rasta_id update when value is 0, preserving NULL for characters without a mercenary. |
| — | **fmt.Printf in setup wizard** | Replaced `fmt.Printf` in `server/setup/setup.go` with structured `zap` logging. |

---

## Suggested Execution Order

Based on remaining impact:

1. ~~**Add tests for `handlers_commands.go`**~~ — **Done.** 62 tests covering all 12 commands (ban, timer, PSN, reload, key quest, rights, course, raviente, teleport, discord, playtime, help), disabled-command gating, op overrides, error paths, and `initCommands`.
2. ~~**Fix bookshelf data pointer** ([#164](https://github.com/Mezeporta/Erupe/issues/164))~~ — **Done.** Corrected offsets for G1–Z2, F4–F5, S6 via delta analysis + Ghidra RE
3. ~~**Fix achievement rank-up notifications** ([#165](https://github.com/Mezeporta/Erupe/issues/165))~~ — **Done.** RE'd `putDisplayed_achievement` from ZZ DLL; added `displayed_levels` column + service methods + migration 0008
4. ~~**Add coverage threshold** to CI~~ — **Done.** 50% floor enforced via `go tool cover` in CI; Codecov removed.
5. ~~**Fix guild alliance toggle** ([#166](https://github.com/Mezeporta/Erupe/issues/166))~~ — **Done.** `recruiting` column + `OperateJoint` allow/deny actions + DB toggle
6. ~~**Fix session handler retail mismatches** ([#167](https://github.com/Mezeporta/Erupe/issues/167))~~ — **Documented.** RE'd from ZZ DLL; log key off-by-one is benign server-side, player count fixed via `QuestReserved`.
7. **Reverse-engineer MhfAddUdPoint fields** ([#168](https://github.com/Mezeporta/Erupe/issues/168)) — needs packet captures
