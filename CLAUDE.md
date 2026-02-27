# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Erupe is a Go server emulator for Monster Hunter Frontier, a shut-down MMORPG. It handles authentication, world selection, and gameplay in a single binary running four TCP/HTTP servers. Go 1.25+ required.

## Build & Test Commands

```bash
go build -o erupe-ce                    # Build server
go build -o protbot ./cmd/protbot/      # Build protocol bot
go test -race ./... -timeout=10m        # Run tests (race detection mandatory)
go test -v ./server/channelserver/...   # Test one package
go test -run TestHandleMsg ./server/channelserver/...  # Single test
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out  # Coverage (CI requires ≥50%)
gofmt -w .                              # Format
golangci-lint run ./...                 # Lint (v2 standard preset, must pass CI)
```

Docker (from `docker/`):
```bash
docker compose up db pgadmin            # PostgreSQL + pgAdmin (port 5050)
docker compose up server                # Erupe (after DB is healthy)
```

## Architecture

### Four-Server Model (single binary, orchestrated from `main.go`)

```
Client ←[Blowfish TCP]→ Sign Server (53312)      → Authentication, sessions
                       → Entrance Server (53310)  → Server list, character select
                       → Channel Servers (54001+) → Gameplay, quests, multiplayer
                       → API Server (8080)        → REST API (/health, /version, V2 sign)
```

Each server is in its own package under `server/`. The channel server is by far the largest (~200 files).

### Channel Server Packet Flow

1. `network/crypt_conn.go` decrypts TCP stream (Blowfish)
2. `network/mhfpacket/` deserializes binary packet into typed struct (~453 packet types, one file each)
3. `handlers_table.go` dispatches via `buildHandlerTable()` (~200+ `PacketID → handlerFunc` entries)
4. Handler in appropriate `handlers_*.go` processes it (organized by game system)

Handler signature: `func(s *Session, p mhfpacket.MHFPacket)`

### Layered Architecture

```
handlers_*.go  →  svc_*.go (service layer)  →  repo_*.go (data access)
                  (where needed)                     ↓
                                              repo_interfaces.go (21 interfaces)
                                                    ↓
                                              repo_mocks_test.go (test doubles)
```

- **Handlers**: Parse packets, call services or repos, build responses. Must always send ACK (see Error Handling below). Simple CRUD operations call repos directly; multi-step or cross-repo logic goes through services.
- **Services**: Encapsulate business logic that spans multiple repos or requires orchestration beyond simple CRUD. Not a mandatory pass-through — handlers call repos directly for straightforward data access.
- **Repositories**: All SQL lives in `repo_*.go` files behind interfaces in `repo_interfaces.go`. The `Server` struct holds interface types, not concrete implementations. Handler code must never contain inline SQL.
- **Sign server** has its own repo pattern: 3 interfaces in `server/signserver/repo_interfaces.go`.

#### Services

| Service | File | Methods | Purpose |
|---------|------|---------|---------|
| `GuildService` | `svc_guild.go` | 6 | Member operations, disband, resign, leave, scout — triggers cross-repo mail |
| `MailService` | `svc_mail.go` | 4 | Send/broadcast mail with message type routing |
| `GachaService` | `svc_gacha.go` | 6 | Gacha rolls (normal/stepup/box), point transactions, reward resolution |
| `AchievementService` | `svc_achievement.go` | 2 | Achievement fetch with score computation, increment |
| `TowerService` | `svc_tower.go` | 3 | Tower gem management, tenrourai progress capping, guild RP donation |
| `FestaService` | `svc_festa.go` | 2 | Event lifecycle (expiry/cleanup/creation), soul submission filtering |

Each service takes repo interfaces + `*zap.Logger` in its constructor, making it testable with mocks. Tests live in `svc_*_test.go` files alongside the service.

### Key Subsystems

| File(s) | Purpose |
|---------|---------|
| `sys_session.go` | Per-connection state: character, stage, semaphores, send queue |
| `sys_stage.go` | `StageMap` (`sync.Map`-backed), multiplayer rooms/lobbies |
| `sys_channel_server.go` | Server lifecycle, Raviente shared state, world management |
| `sys_semaphore.go` | Distributed locks for events (Raviente siege, guild ops) |
| `channel_registry.go` | Cross-channel operations (worldcast, session lookup, mail) |
| `handlers_cast_binary.go` | Binary state relay between clients (position, animation) |
| `handlers_helpers.go` | `loadCharacterData`/`saveCharacterData` shared helpers |
| `guild_model.go` | Guild data structures |

### Binary Serialization

`common/byteframe.ByteFrame` — sequential big-endian reads/writes with sticky error pattern (`bf.Err()`). Used for all packet parsing, response building, and save data manipulation. Use `encoding/binary` only for random-access reads at computed offsets on existing `[]byte` slices.

### Database

PostgreSQL with embedded auto-migrating schema in `server/migrations/`:
- `sql/0001_init.sql` — consolidated baseline
- `seed/*.sql` — demo data (applied via `migrations.ApplySeedData()` on fresh DB)
- New migrations: `sql/0002_description.sql`, etc. (each runs in its own transaction)

The server runs `migrations.Migrate()` automatically on startup.

### Configuration

Two reference files: `config.example.json` (minimal) and `config.reference.json` (all options). Loaded via Viper in `config/config.go`. All defaults registered in code. Supports 40 client versions (S1.0 → ZZ) via `ClientMode`. If `config.json` is missing, an interactive setup wizard launches at `http://localhost:8080`.

### Protocol Bot (`cmd/protbot/`)

Headless MHF client implementing the complete sign → entrance → channel flow. Shares `common/` and `network/crypto` but avoids `config` dependency via its own `conn/` package.

## Concurrency

Lock ordering: `Server.Mutex → Stage.RWMutex → semaphoreLock`. Stage map uses `sync.Map`; individual `Stage` structs have `sync.RWMutex`. Cross-channel operations go exclusively through `ChannelRegistry` — never access other servers' state directly.

## Error Handling in Handlers

The MHF client expects `MsgSysAck` for most requests. Missing ACKs cause client softlocks. On error paths, always send `doAckBufFail`/`doAckSimpleFail` before returning.

## Testing

- **Mock repos**: Handler tests use `repo_mocks_test.go` — no database needed
- **Table-driven tests**: Standard pattern (see `handlers_achievement_test.go`)
- **Race detection**: `go test -race` is mandatory in CI
- **Coverage floor**: CI enforces ≥50% total coverage

## Adding a New Packet

1. Define struct in `network/mhfpacket/msg_*.go` (implements `MHFPacket` interface: `Parse`, `Build`, `Opcode`)
2. Add packet ID constant in `network/packetid.go`
3. Register handler in `server/channelserver/handlers_table.go`
4. Implement handler in appropriate `handlers_*.go` file

## Adding a Database Query

1. Add method signature to the relevant interface in `repo_interfaces.go`
2. Implement in the corresponding `repo_*.go` file
3. Add mock implementation in `repo_mocks_test.go`

## Adding Business Logic

If the new logic involves multi-step orchestration, cross-repo coordination, or non-trivial data transformation:

1. Add or extend a service in the appropriate `svc_*.go` file
2. Wire it in `sys_channel_server.go` (constructor + field on `Server` struct)
3. Add tests in `svc_*_test.go` using mock repos
4. Call the service from the handler instead of the repo directly

Simple CRUD operations should stay as direct repo calls from handlers — not everything needs a service.

## Known Issues

See `docs/anti-patterns.md` for structural patterns and `docs/technical-debt.md` for specific fixable items with file paths and line numbers.

## Pre-Commit Checks

Before every commit, run `gofmt` and `golangci-lint` on changed Go files (excluding `vendor/`). Do not commit if either check fails.

```bash
gofmt -l .                    # Must produce no output
golangci-lint run ./...       # Must pass with zero errors
```

## Contributing

- Branch naming: `feature/`, `fix/`, `refactor/`, `docs/`
- Commit messages: conventional commits (`feat:`, `fix:`, `refactor:`, `docs:`)
- Update `CHANGELOG.md` under "Unreleased" for all changes
