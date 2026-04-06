# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Fixed quest tune-value filter silently dropping user-configured multipliers set to `0.0`: previously setting e.g. `ZennyMultiplier: 0.0` would strip the entry from the table and fall back to the client's default (100%), producing the opposite of the intended "no zenny" configuration. The `Value > 0` filter in `handleMsgMhfEnumerateQuest` has been removed so zero values are now sent verbatim. Affects HRP/SRP/GRP/GSRP/Zenny/GZenny/Material/GMaterial/GCP/GUrgent multipliers and their NC variants.
- Fixed float32 truncation in quest multiplier conversion: `uint16(0.20 * 100)` yielded `19` instead of `20` because `float32(0.20) ≈ 0.19999998`. Replaced with a `multiplierToTuneValue` helper that rounds via `math.Round`. Applied to all 18 multiplier call sites.
- Fixed `DisableLoginBoost` and `DisableBoostTime` config flags not fully honored ([#187](https://github.com/Mezeporta/Erupe/issues/187)): `GetBoostTimeLimit`/`GetBoostRight` now respect `DisableBoostTime` and `UseKeepLoginBoost` now respects `DisableLoginBoost`. Also fixes a zero-`time.Time` wraparound in `GetBoostTimeLimit` that made the "Boost Time" overlay appear on fresh characters.
- Fixed playtime regression across sessions: `updateSaveDataWithStruct` now writes the accumulated playtime back into the binary save blob, preventing each reconnect from loading a stale in-game counter and rolling back progress.
- Fixed player softlock when buying items at the forge: `MSG_CA_EXCHANGE_ITEM` `Parse()` was returning `NOT IMPLEMENTED`, causing the dispatch loop to drop the packet without sending an ACK. Now parses the `AckHandle` and responds with `doAckBufFail` so the client's error branch exits cleanly.
- Fixed player softlock on N-points (Hunting Road) interactions: same root cause for `MSG_MHF_USE_UD_SHOP_COIN` — `Parse()` now reads the `AckHandle` and responds with `doAckBufFail`.

## [9.3.1] - 2026-03-23

### Added

- `DisableSaveIntegrityCheck` config flag: when `true`, the SHA-256 savedata integrity check is skipped on load. 
Intended for cross-server save transfers where the stored hash in the database does not match the imported save blob. 
Defaults to `false`.
Affected characters can alternatively be unblocked per-character with `UPDATE characters SET savedata_hash = NULL WHERE id = <id>`.

## [9.3.0] - 2026-03-19

### Fixed

- Fixed G-rank Workshop and Master Felyne (Cog) softlock: `MSG_MHF_GET_EXTRA_INFO` and `MSG_MHF_GET_COG_INFO` now parse correctly and return a fail ACK instead of dropping the packet silently ([#180](https://github.com/Mezeporta/Erupe/issues/180))
- A second SIGINT/Ctrl+C during the shutdown countdown now force-stops the server immediately
- Fixed `ecdMagic` constant byte order causing encryption failures on some platforms ([#174](https://github.com/Mezeporta/Erupe/issues/174))
- Fixed guild nil panics: variable shadowing causing nil panic in scout list ([#171](https://github.com/Mezeporta/Erupe/issues/171))
- Fixed guild nil panics: added nil guards in cancel and answer scout handlers ([#171](https://github.com/Mezeporta/Erupe/issues/171))
- Fixed guild nil panics: added nil guards for alliance guild lookups ([#171](https://github.com/Mezeporta/Erupe/issues/171))
- Fixed `rasta_id=0` overwriting NULL in mercenary save, preventing game state saving ([#163](https://github.com/Mezeporta/Erupe/issues/163))
- Fixed false race condition in `PacketDuringLogout` test
- Fixed bookshelf save data pointers for non-ZZ client versions (G1–G10, F4–F5, S6.0)
- Fixed Forward.5 festa crashes: skip trials referencing monsters added after em106 (Odibatorasu) and filter out item 7011 which does not exist before G1 ([#156](https://github.com/Mezeporta/Erupe/pull/156))

### Changed

- Cached `rengoku_data.bin` at startup for improved channel server performance

### Added

- Achievement rank-up notifications: the client now shows rank-up popups when achievements level up, using per-character tracking of last-displayed levels ([#165](https://github.com/Mezeporta/Erupe/issues/165))
- Database migration `0008_achievement_displayed_levels` (tracks last-displayed achievement levels)
- Diva Defense point accumulation: `MsgMhfAddUdPoint` now stores per-character quest and bonus points in a dedicated `diva_points` table, RE'd from the ZZ client DLL ([#168](https://github.com/Mezeporta/Erupe/issues/168))
- Database migration `0009_diva_points` (per-character per-event point tracking)
- Savedata corruption defense (tier 1): bounded decompression in nullcomp prevents OOM from crafted payloads, bounds-checked delta patching prevents buffer overflows, compressed payload size limits (512KB) and decompressed size limits (1MB) reject oversized saves, rotating savedata backups (3 slots, 30-minute interval) provide recovery points
- Savedata corruption defense (tier 2): SHA-256 checksum on decompressed savedata verified on every load, atomic DB transactions wrapping character data + house data + hash + backup in a single commit, per-character save mutex preventing concurrent save races
- Database migration `0007_savedata_integrity` (rotating backup table + integrity checksum column)
- Tests for `logoutPlayer`, `saveAllCharacterData`, and transit message handlers
- Alliance `scanAllianceWithGuilds` test for missing guild (nil return from GetByID)
- Handler dispatch table test verifying all expected packet IDs are mapped
- Scenario binary format documentation (`docs/scenario-format.md`)

### Infrastructure

- Updated `go.mod` dependencies
- Added `IF NOT EXISTS` guard to alliance recruiting column migration

## [9.3.0-rc1] - 2026-02-28

900 commits, 860 files changed, ~100,000 lines of new code. The largest Erupe release ever.

### Added

#### Architecture
- Repository pattern: 21 interfaces in `repo_interfaces.go` replace all inline SQL in handlers (`CharacterRepo`, `GuildRepo`, `UserRepo`, `SessionRepo`, `AchievementRepo`, `CafeRepo`, `DistributionRepo`, `DivaRepo`, `EventRepo`, `FestaRepo`, `GachaRepo`, `GoocooRepo`, `HouseRepo`, `MailRepo`, `MercenaryRepo`, `MiscRepo`, `RengokuRepo`, `ScenarioRepo`, `ShopRepo`, `StampRepo`, `TowerRepo`)
- Service layer: 6 services encapsulating multi-step business logic (`GuildService`, `MailService`, `GachaService`, `AchievementService`, `TowerService`, `FestaService`)
- `ChannelRegistry` interface for cross-channel operations (worldcast, session lookup, mail, disconnect) — channels decoupled for independent operation
- Sign server converted to repository pattern with 3 interfaces (`SignUserRepo`, `SignCharacterRepo`, `SignSessionRepo`)

#### Database & Schema
- Embedded auto-migrating database schema system (`server/migrations/`): the server binary now contains all SQL and runs migrations automatically on startup — no more `pg_restore`, manual patch ordering, or external `schemas/` directory
- Catch-up migration (`0002_catch_up_patches.sql`) for databases with partially-applied patch schemas — idempotent no-op on fresh or fully-patched databases, fills gaps for partial installations
- Setup wizard: web-based first-run configuration at `http://localhost:8080` when `config.json` is missing — guides through database connection, schema initialization, and server settings
- Seed data embedded and applied automatically on fresh installs (shops, events, gacha, scenarios, etc.)
- Database connection pool configuration

#### Game Systems
- Quest enumeration completely rewritten with quest caching system
- Event quest cycling with database-driven rotation and season/time override
- Quest stamp card system with retro stamp rewards
- Raviente v3 rework: ID system, semaphore handling, party broadcasting, customizable latency and max players
- Warehouse v2 rewrite with proper serialization across game versions
- Distribution system completely rewritten with proper typing and version support
- Conquest/Earth status: multiple war targets, rewritten handlers, status override options
- Festa bonus categories, trial voting, version-gated info (S6.0, Z2)
- Monthly guild item claim tracking per character per type (standard/HLC/EXC)
- Scenario counter implementation with database-driven defaults
- Campaign structs ported with backwards compatibility
- Trend weapons implementation
- Clan Changing Room support
- Operator accounts and ban system
- NG word filter (ASCII and SJIS)
- Custom command prefixes with help command

#### Client Version Support
- Season 6.0: savedata compatibility, encryption fix, semaphore fix, terminal log fix
- Forward.4–F.5: ClientMode support added
- G1–G2: save pointers enabled, gacha shop fix, compatibility fixes
- G3–G9.1: save pointers verified, DecoMyset response fixes
- < G10: InfoGuild response fix, semaphore backwards compatibility
- PS3: PS3SGN support, PSN account linking, trophy course
- PS Vita: VITASGN support, PSN linking
- Wii U: WIIUSGN support
- 40 client versions supported (S1.0 through ZZ) via `ClientMode` config option

#### API
- `/v2/` route prefix with HTTP method enforcement alongside legacy routes
- `GET /v2/server/status` endpoint returning MezFes schedule, featured weapon, and festa/diva event status
- `DELETE /v2/characters/{id}` route
- `GET /version` endpoint returning server name and client mode
- `GET /health` endpoint with Docker healthchecks
- Auth middleware extracting `Authorization: Bearer <token>` header for v2 routes; legacy body-token auth preserved
- Standardized JSON error responses (`{"error":"...","message":"..."}`) across all endpoints
- `returning` field on characters (true if last login > 90 days ago) and `courses` field on auth data
- `APIEventRepo` interface and read-only implementation for feature weapons and events
- OpenAPI spec at `docs/openapi.yaml`

#### Developer Tooling
- Protocol bot (`cmd/protbot/`): headless MHF client implementing the complete sign → entrance → channel flow for automated testing and protocol debugging
- Packet capture & replay system (`network/pcap/`): transparent recording, filtering, metadata, and standalone replay tool
- Mock repository implementations for all 21 interfaces — handler unit tests without PostgreSQL
- 120+ new test files, coverage pushed from ~7% to 65%+
- CI: GitHub Actions with race detector, coverage threshold (≥50%), `golangci-lint` v2, automated release builds (Linux/Windows)
- CI: Docker CD workflow pushing images to GHCR

#### Logging & Observability
- Standardized `zap` structured logging across all packages (replaced `fmt.Printf`, `log.*`)
- Comprehensive production logging for save operations (warehouse, Koryo points, savedata, Hunter Navi, plate equipment)
- Disconnect type tracking (graceful, connection_lost, error) with detailed logging
- Session lifecycle logging with duration and metrics tracking
- Plate data (transmog) safety net in logout flow

### Changed

- Minimum Go version: 1.21 → 1.25
- Monolithic `handlers.go` split into ~30 domain-specific files; guild handlers split from 1 file to 10
- Handler registration: replaced `init()` with explicit `buildHandlerTable()` construction
- Eliminated `ErupeConfig` global variable — config passed explicitly throughout
- Schema management consolidated: replaced 4 independent code paths (Docker shell script, setup wizard, test helpers, manual psql) with single embedded migration runner
- Docker simplified: removed schema volume mounts and init script — the server binary handles everything
- `config.json` removed from repo; `config.example.json` minimized, `config.reference.json` added with all options
- Stage map replaced with `sync.Map`-backed `StageMap` implementation
- Refactored logout flow to save all data before cleanup (prevents data loss race conditions)
- Unified save operation into single `saveAllCharacterData()` function with proper error handling
- SignV2 server removed — merged into unified API server
- `ByteFrame` read-overflow panic replaced with sticky error pattern (`bf.Err()`)
- `panic()` calls replaced with structured error handling throughout
- 15+ `Unk*` packet fields renamed to meaningful names across the protocol
- `errcheck` lint compliance across entire codebase

### Fixed

#### Gameplay
- Fixed lobby search returning all reserved players instead of only quest-bound players — `QuestReserved` now counts only clients in "Qs" stages, matching retail ([#167](https://github.com/Mezeporta/Erupe/issues/167))
- Fixed bookshelf save data pointer being off by 14810 bytes for G1–Z2, F4–F5, and S6 game versions ([#164](https://github.com/Mezeporta/Erupe/issues/164))
- Fixed guild alliance application toggle being hardcoded to always-open ([#166](https://github.com/Mezeporta/Erupe/issues/166))
- Fixed gacha shop not working on G1–GG clients due to protocol differences — thanks @Sin365 (#150)
- Fixed save data corruption check rejecting valid saves due to name encoding mismatches (SJIS/UTF-8)
- Fixed incomplete saves during logout — character savedata now persisted even during ungraceful disconnects
- Fixed stale transmog/armor appearance shown to other players — user binary cache invalidated on save
- Fixed login boost creating hanging connections
- Fixed MezFes tickets not resetting weekly
- Fixed event quests not selectable
- Fixed inflated festa rewards
- Fixed RP inconsistent between clients
- Fixed limited friends & clanmates display
- Fixed house theme corruption on save (#92)
- Fixed Sky Corridor race condition preventing skill data wipe (#85)
- Fixed `CafeDuration` and `AcquireCafeItem` cost for G1–G5.2 clients
- Fixed quest mark labelling
- Fixed HunterNavi savedata clipping (last 2 bytes)

#### Stability
- Fixed deadlock in zone change causing 60-second timeout
- Fixed 3 critical race conditions in `handlers_stage.go`
- Fixed data race in `token.RNG` global used concurrently across goroutines
- Fixed JPK decompression data race
- Fixed session lifecycle races in shutdown path
- Fixed concurrent quest cache map write
- Fixed guild RP rollover race
- Fixed crash on clans with 0 members
- Fixed crash when sending empty packets in `QueueSend`/`QueueSendNonBlocking`
- Fixed `LoadDecoMyset` crash with 40+ decoration presets on older versions
- Fixed `WaitStageBinary` handler hanging indefinitely
- Fixed double-save bug in logout flow
- Fixed save operation ordering — data saved before session cleanup instead of after

#### Protocol
- Fixed missing ACK responses across handlers — prevents client softlocks
- Fixed missing stage transfer packet for empty zones
- Fixed client crash when quest or scenario files are missing — sends failure ACK instead of nil data
- Fixed server crash when Discord relay receives unsupported Shift-JIS characters (emoji, Lenny faces, cuneiform, etc.)

### Removed

- Compatibility with Go 1.21
- Old `schemas/` and `bundled-schema/` directories (replaced by embedded migrations)
- `distribution.data` column (unused, prevented seed data from matching Go code expectations) (#169)
- SignV2 server (merged into unified API server)
- Unused `timeserver` module

### Security

- Bumped `golang.org/x/net` from 0.18.0 to 0.38.0
- Bumped `golang.org/x/crypto` from 0.15.0 to 0.35.0
- Path traversal fix in screenshot API endpoint
- CodeQL scanning added to CI
- Binary blob size guards on save handlers
- Database connection arguments escaped

## [9.2.0] - 2023-04-01

### Added in 9.2.0

- Gacha system with box gacha and stepup gacha support
- Multiple login notices support
- Daily quest allowance configuration
- Gameplay options system
- Support for stepping stone gacha rewards
- Guild semaphore locking mechanism
- Feature weapon schema and generation system
- Gacha reward tracking and fulfillment
- Koban my mission exchange for gacha

### Changed in 9.2.0

- Reworked logging code and syntax
- Reworked broadcast functions
- Reworked netcafe course activation
- Reworked command responses for JP chat
- Refactored guild message board code
- Separated out gacha function code
- Rearranged gacha functions
- Updated golang dependencies
- Made various handlers non-fatal errors
- Moved various packet handlers
- Moved caravan event handlers
- Enhanced feature weapon RNG

### Fixed in 9.2.0

- Mail item workaround removed (replaced with proper implementation)
- Possible infinite loop in gacha rolls
- Feature weapon RNG and generation
- Feature weapon times and return expiry
- Netcafe timestamp handling
- Guild meal enumeration and timer
- Guild message board enumerating too many posts
- Gacha koban my mission exchange
- Gacha rolling and reward handling
- Gacha enumeration recommendation tag
- Login boost creating hanging connections
- Shop-db schema issues
- Scout enumeration data
- Missing primary key in schema
- Time fixes and initialization
- Concurrent stage map write issue
- Nil savedata errors on logout
- Patch schema inconsistencies
- Edge cases in rights integer handling
- Missing period in broadcast strings

### Removed in 9.2.0

- Unused database tables
- Obsolete LauncherServer code
- Unused code from gacha functionality
- Mail item workaround (replaced with proper implementation)

### Security in 9.2.0

- Escaped database connection arguments

## [9.1.1] - 2022-11-10

### Changed in 9.1.1

- Temporarily reverted versioning system
- Fixed netcafe time reset behavior

## [9.1.0] - 2022-11-04

### Added in 9.1.0

- Multi-language support system
- Support for JP strings in broadcasts
- Guild scout language support
- Screenshot sharing support
- New sign server implementation
- Multi-language string mappings
- Language-based chat command responses

### Changed in 9.1.0

- Rearranged configuration options
- Converted token to library
- Renamed sign server
- Mapped language to server instead of session

### Fixed in 9.1.0

- Various packet responses

## [9.1.0-rc3] - 2022-11-02

### Fixed in 9.1.0-rc3

- Prevented invalid bitfield issues

## [9.1.0-rc2] - 2022-10-28

### Changed in 9.1.0-rc2

- Set default featured weapons to 1

## [9.1.0-rc1] - 2022-10-24

### Removed in 9.1.0-rc1

- Migrations directory

## [9.0.1] - 2022-08-04

### Changed in 9.0.1

- Updated login notice

## [9.0.0] - 2022-08-03

### Fixed in 9.0.0

- Fixed readlocked channels issue
- Prevent rp logs being nil
- Prevent applicants from receiving message board notifications

### Added in 9.0.0

- Implement guild semaphore locking
- Support for more courses
- Option to flag corruption attempted saves as deleted
- Point limitations for currency

---

## Pre-9.0.0 Development (2022-02-25 to 2022-08-03)

The period before version 9.0.0 represents the early community development phase, starting with the Community Edition reupload and continuing through multiple feature additions leading up to the first semantic versioning release.

### [Pre-release] - 2022-06-01 to 2022-08-03

Major feature implementations leading to 9.0.0:

#### Added (June-August 2022)

- **Friend System**: Friend list functionality with cross-character enumeration
- **Blacklist System**: Player blocking functionality
- **My Series System**: Basic My Series functionality with shared data and bookshelf support
- **Guild Treasure Hunts**: Complete guild treasure hunting system with cooldowns
- **House System**:
  - House interior updates and furniture loading
  - House entry handling improvements
  - Visit other players' houses with correct furniture display
- **Festa System**:
  - Initial Festa build and decoding
  - Canned Festa prizes implementation
  - Festa finale acquisition handling
  - Festa info and packet handling improvements
- **Achievement System**: Hunting career achievements concept implementation
- **Object System**:
  - Object indexing (v3, v3.1)
  - Semaphore indexes
  - Object index limits and reuse prevention
- **Transit Message**: Correct parsing of transit messages for minigames
- **World Chat**: Enabled world chat functionality
- **Rights System**: Rights command and permission updates on login
- **Customizable Login Notice**: Support for custom login notices

#### Changed (June-August 2022)

- **Stage System**: Major stage rework and improvements
- **Raviente System**: Cleanup, fixes, and announcement improvements
- **Discord Integration**: Mediated Discord handling improvements
- **Server Logging**: Improved server logging throughout
- **Configuration**: Edited default configs
- **Repository**: Extensive repository cleanup
- **Build System**: Implemented build actions and artifact generation

#### Fixed (June-August 2022)

- Critical semaphore bug fixes
- Raviente-related fixes and cleanup
- Read-locked channels issue
- Stubbed title enumeration
- Object index reuse prevention
- Crash when not in guild on logout
- Invalid schema issues
- Stage enumeration crash prevention
- Gook (book) enumeration and cleanup
- Guild SQL fixes
- Various packet parsing improvements
- Semaphore checking changes
- User insertion not broadcasting

### [Pre-release] - 2022-05-01 to 2022-06-01

Guild system enhancements and social features:

#### Added (May-June 2022)

- **Guild Features**:
  - Guild alliance support with complete implementation
  - Guild member (Pugi) management and renaming
  - Guild post SJIS (Japanese) character encoding support
  - Guild message board functionality
  - Guild meal system
  - Diva Hall adventure cat support
  - Guild adventure cat implementation
  - Alliance members included in guild member enumeration
- **Character System**:
  - Mail locking mechanism
  - Favorite quest save/load functionality
  - Title/achievement enumeration parsing
  - Character data handler rewrite
- **Game Features**:
  - Item distribution handling system
  - Road Shop weekly rotation
  - Scenario counter implementation
  - Diva adventure dispatch parsing
  - House interior query support
  - Entrance and sign server response improvements
- **Launcher**:
  - Discord bot integration with configurable channels and dev roles
  - Launcher error handling improvements
  - Launcher finalization with modal, news, menu, safety links
  - Auto character addition
  - Variable centered text support
  - Last login timestamp updates

#### Changed (May-June 2022)

- Stage and semaphore overhaul with improved casting handling
- Simplified guild handler code
- String support improvements with PascalString helpers
- Byte frame converted to local package
- Local package conversions (byteframe, pascalstring)

#### Fixed (May-June 2022)

- SJIS guild post support
- Nil guild failsafes
- SQL queries with missing counter functionality
- Enumerate airoulist parsing
- Mail item description crashes
- Ambiguous mail query
- Last character updates
- Compatibility issues
- Various packet files

### [Pre-release] - 2022-02-25 to 2022-05-01

Initial Community Edition and foundational work:

#### Added (February-May 2022)

- **Core Systems**:
  - Japanese Shift-JIS character name support
  - Character creation with automatic addition
  - Raviente system patches
  - Diva reward handling
  - Conquest quest support
  - Quest clear timer
  - Garden cat/shared account box implementation
- **Guild Features**:
  - Guild hall available on creation
  - Unlocked all street titles
  - Guild schema corrections
- **Launcher**:
  - Complete launcher implementation
  - Modal dialogs
  - News system
  - Menu and safety links
  - Button functionality
  - Caching system

#### Changed (February-May 2022)

- Save compression updates
- Migration folder moved to root
- Improved launcher code structure

#### Fixed (February-May 2022)

- Mercenary/cat handler fixes
- Error code 10054 (savedata directory creation)
- Conflicts resolution
- Various syntax corrections

---

## Historical Context

This changelog documents all known changes from the Community Edition reupload (February 25, 2022) onwards. The period before this (Einherjar Team era, ~2020-2022) has no public git history.

Earlier development by Cappuccino/Ellie42 (March 2020) focused on basic server infrastructure, multiplayer systems, and core functionality. See [AUTHORS.md](AUTHORS.md) for detailed development history.

The project began following semantic versioning with v9.0.0 (August 3, 2022) and maintains tagged releases for stable versions. Development continues on the main branch with features merged from feature branches.
