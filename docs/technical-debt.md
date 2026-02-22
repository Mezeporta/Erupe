# Erupe Technical Debt & Suggested Next Steps

> Analysis date: 2026-02-22

This document tracks actionable technical debt items discovered during a codebase audit. It complements `anti-patterns.md` (which covers structural patterns) by focusing on specific, fixable items with file paths and line numbers.

## Table of Contents

- [High Priority](#high-priority)
  - [1. Broken game features (gameplay-impacting TODOs)](#1-broken-game-features-gameplay-impacting-todos)
  - [2. Test gaps on critical paths](#2-test-gaps-on-critical-paths)
  - [3. Sign server has no repository layer](#3-sign-server-has-no-repository-layer)
- [Medium Priority](#medium-priority)
  - [4. Split repo_guild.go](#4-split-repo_guildgo)
  - [5. Logging anti-patterns](#5-logging-anti-patterns)
  - [6. Inconsistent transaction API](#6-inconsistent-transaction-api)
  - [7. LoopDelay config has no Viper default](#7-loopdelay-config-has-no-viper-default)
- [Low Priority](#low-priority)
  - [8. Typos and stale comments](#8-typos-and-stale-comments)
  - [9. CI updates](#9-ci-updates)
- [Suggested Execution Order](#suggested-execution-order)

---

## High Priority

### 1. Broken game features (gameplay-impacting TODOs)

These TODOs represent features that are visibly broken for players.

| Location | Issue | Impact |
|----------|-------|--------|
| `model_character.go:88,101,113` | `TODO: fix bookshelf data pointer` for G10-ZZ, F4-F5, and S6 versions | Wrong pointer corrupts character save reads for three game versions |
| ~~`handlers_guild.go:389`~~ | ~~`TODO: Implement month-by-month tracker` — always returns `0x01` (claimed)~~ | ~~Players can never claim monthly guild items~~ **Fixed.** Now tracks per-character per-type monthly claims via `stamps` table. |
| `handlers_guild_ops.go:148` | `TODO: Move this value onto rp_yesterday and reset to 0... daily?` | Guild daily RP rollover logic is missing entirely |
| `handlers_achievement.go:125` | `TODO: Notify on rank increase` — always returns `false` | Achievement rank-up notifications are silently suppressed |
| `handlers_guild_info.go:443` | `TODO: Enable GuildAlliance applications` — hardcoded `true` | Guild alliance applications are always open regardless of setting |
| `handlers_session.go:397` | `TODO(Andoryuuta): log key index off-by-one` | Known off-by-one in log key indexing is unresolved |
| `handlers_session.go:577` | `TODO: This case might be <=G2` | Uncertain version detection in switch case |
| `handlers_session.go:777` | `TODO: Retail returned the number of clients in quests` | Player count reported to clients does not match retail behavior |

### 2. Test gaps on critical paths

**Handler files with no test file:**

| File | Lines | Priority | Reason |
|------|-------|----------|--------|
| `handlers_session.go` | 833 | HIGH | Login/logout, log key, character enumeration |
| `handlers_gacha.go` | 411 | HIGH | Economy system with DB writes |
| `handlers_commands.go` | 421 | HIGH | Admin command system |
| `handlers_data_paper.go` | 621 | MEDIUM | Daily paper data |
| `handlers_plate.go` | 294 | MEDIUM | Armor plate system |
| `handlers_shop.go` | 291 | MEDIUM | Shopping system |
| `handlers_seibattle.go` | 259 | MEDIUM | Sei battle system |
| `handlers_scenario.go` | ~100 | LOW | Mostly complete, uses repo |
| `handlers_distitem.go` | small | LOW | Distribution items |
| `handlers_guild_mission.go` | small | LOW | Guild missions |
| `handlers_kouryou.go` | small | LOW | Kouryou system |

**Repository files with no store-level test file (17 total):**

`repo_achievement.go`, `repo_cafe.go`, `repo_distribution.go`, `repo_diva.go`, `repo_festa.go`, `repo_gacha.go`, `repo_goocoo.go`, `repo_house.go`, `repo_mail.go`, `repo_mercenary.go`, `repo_misc.go`, `repo_rengoku.go`, `repo_scenario.go`, `repo_session.go`, `repo_shop.go`, `repo_stamp.go`, `repo_tower.go`

These are validated indirectly through mock-based handler tests but have no SQL-level integration tests.

### 3. Sign server has no repository layer

The channelserver was refactored to use repository interfaces (commits `a9cca84`, `6fbd294`, `1d5026c`), but `server/signserver/` was not included. It still does raw `db.QueryRow`/`db.Exec` with **8 silently discarded errors** on write paths:

```
server/signserver/dbutils.go:86,91,94,100,107,123,149
server/signserver/session.go
server/signserver/dsgn_resp.go
```

Examples of discarded errors (login timestamps, return-to-player expiry, rights queries):
```go
_, _ = s.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
_ = s.db.QueryRow("SELECT last_character FROM users WHERE id=$1", uid).Scan(&lastPlayed)
```

A database connectivity issue during login would be invisible.

---

## Medium Priority

### 4. Split `repo_guild.go`

At 1004 lines with 71 functions, `repo_guild.go` mixes 6 distinct concerns:

- Guild CRUD and metadata
- Member management
- Applications/recruitment
- RP tracking
- Item box operations
- Message board posts

Suggested split: `repo_guild.go` (core CRUD), `repo_guild_members.go`, `repo_guild_items.go`, `repo_guild_board.go`.

### 5. Logging anti-patterns

~~**a) `fmt.Sprintf` inside structured logger calls (6 sites):**~~ **Fixed.** All 6 sites now use `zap.Uint32`/`zap.Uint8`/`zap.String` structured fields instead of `fmt.Sprintf`.

**b) 20 silently discarded SJIS encoding errors in packet parsing:**

The pattern `m.Field, _ = stringsupport.SJISToUTF8(...)` appears across:
- `network/binpacket/msg_bin_chat.go:43-44`
- `network/mhfpacket/msg_mhf_apply_bbs_article.go:33-35`
- `network/mhfpacket/msg_mhf_send_mail.go:38-39`
- `network/mhfpacket/msg_mhf_update_guild_message_board.go:41-42,50-51`
- `server/channelserver/model_character.go:175`
- And 7+ more packet files

A malformed SJIS string from a client yields an empty string with no log output, making garbled text impossible to debug. The error return should at least be logged at debug level.

### 6. Inconsistent transaction API

`repo_guild.go` mixes two transaction styles in the same file:

```go
// Line 175 — no context, old style
tx, err := r.db.Begin()

// Line 518 — sqlx-idiomatic, with context
tx, err := r.db.BeginTxx(context.Background(), nil)
```

Should standardize on `BeginTxx` throughout. The `Begin()` calls cannot carry a context for cancellation or timeout.

### ~~7. `LoopDelay` config has no Viper default~~ (Fixed)

**Status:** Fixed. `viper.SetDefault("LoopDelay", 50)` added in `config/config.go`, matching the `config.example.json` value.

---

## Low Priority

### 8. Typos and stale comments

| Location | Issue |
|----------|-------|
| `sys_session.go:73` | Comment says "For Debuging" — typo, and the field is used in production logging, not just debugging |
| `handlers_session.go:397` | "offical" should be "official" |
| `handlers_session.go:324` | `if s.server.db != nil` guard wraps repo calls that are already nil-safe — refactoring artifact |

### 9. CI updates

- `codecov-action@v3` could be updated to `v4` (current stable)
- No coverage threshold is enforced — coverage is uploaded but regressions aren't caught

---

## Suggested Execution Order

Based on impact and the momentum from recent repo-interface refactoring:

1. **Add tests for `handlers_session.go` and `handlers_gacha.go`** — highest-risk untested code on the critical login and economy paths
2. **Refactor signserver to use repository interfaces** — completes the pattern established in channelserver and surfaces 8 hidden error paths
3. ~~**Fix monthly guild item claim**~~ (`handlers_guild.go:389`) — **Done**
4. **Split `repo_guild.go`** — last oversized file after the recent refactoring push
5. ~~**Fix `fmt.Sprintf` in logger calls**~~ — **Done**
6. ~~**Add `LoopDelay` Viper default**~~ — **Done**
7. **Log SJIS decoding errors** — improves debuggability for text issues
8. **Standardize on `BeginTxx`** — consistency fix in `repo_guild.go`
