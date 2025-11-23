# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Erupe is a community-maintained server emulator for Monster Hunter Frontier written in Go. It reverse-engineers the game server protocol to enable self-hosted Monster Hunter Frontier servers, supporting multiple game versions (Season 6.0 through ZZ) and platforms (PC, PS3, PS Vita, Wii U).

**Module name:** `erupe-ce`
**Go version:** 1.25+
**Current branch purpose:** Clean transition from 9.2.0 to future 9.3.0 release

## Build and Development Commands

### Building and Running

```bash
# Build the server
go build

# Build and produce binary
go build -o erupe-ce

# Run without building
go run .

# Run with hot reload during development
go run main.go
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./server/channelserver/...

# Run specific test
go test -v ./server/channelserver/... -run TestName

# Check for race conditions (important!)
go test -v -race ./...

# Generate test coverage
go test -v -cover ./...
go test ./... -coverprofile=/tmp/coverage.out
go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
```

### Code Quality

```bash
# Format code (ALWAYS run before committing)
gofmt -w .

# Lint code
golangci-lint run ./...

# Fix linting issues automatically
golangci-lint run ./... --fix
```

### Database Operations

```bash
# Connect to PostgreSQL
psql -U postgres -d erupe

# Apply schema patches (must be done in order)
psql -U postgres -d erupe -f patch-schema/01_patch.sql
psql -U postgres -d erupe -f patch-schema/02_patch.sql
# ... continue in order

# With password from environment
PGPASSWORD='password' psql -U postgres -d erupe -f schema.sql
```

### Docker Development

```bash
# Start database and pgadmin
docker compose up db pgadmin

# Start server (after configuring database)
docker compose up server

# Full stack
docker compose up
```

### Log Analysis

The project includes a comprehensive log analyzer tool in `tools/loganalyzer/`:

```bash
# Build log analyzer
cd tools/loganalyzer
go build -o loganalyzer

# Filter logs by level
./loganalyzer filter -f ../../logs/erupe.log -level error

# Analyze errors with stack traces
./loganalyzer errors -f ../../logs/erupe.log -stack -detailed

# Track player connections
./loganalyzer connections -f ../../logs/erupe.log -sessions

# Real-time log monitoring
./loganalyzer tail -f ../../logs/erupe.log -level error

# Generate statistics
./loganalyzer stats -f ../../logs/erupe.log -detailed
```

## Architecture

### Three-Server Architecture

Erupe uses a multi-server architecture that mirrors the original Monster Hunter Frontier server design:

1. **Sign Server** (Port 53312)
   - Handles authentication and account management
   - Located in `server/signserver/`
   - Two versions: legacy (signserver) and modern HTTP-based (signv2server)
   - Creates sign sessions with tokens for channel server authentication

2. **Entrance Server** (Port 53310)
   - Manages world/server selection and character list
   - Located in `server/entranceserver/`
   - Routes players to available channel servers
   - Maintains server availability information

3. **Channel Servers** (Ports 54001+)
   - Handle actual gameplay sessions, quests, and player interactions
   - Located in `server/channelserver/`
   - Multiple instances can run simultaneously
   - Organized by world types (Newbie, Normal, Cities, Tavern, Return, MezFes)

### Channel Server Internal Architecture

The channel server is the most complex component:

**Session Management** ([sys_session.go](server/channelserver/sys_session.go))

- Each player connection creates a `Session` struct
- Sessions handle packet queuing, encryption, and client state
- Uses goroutines for send/receive loops
- Thread-safe with mutex locks

**Handler System** ([handlers_table.go](server/channelserver/handlers_table.go))

- Packet handlers registered in `handlerTable` map
- Maps `network.PacketID` to `handlerFunc`
- Handlers organized by feature in separate files:
  - `handlers_quest.go` - Quest system
  - `handlers_guild.go` - Guild operations
  - `handlers_stage.go` - Stage/room management
  - `handlers_character.go` - Character data
  - `handlers_*.go` - Feature-specific handlers

**Stage System** ([sys_stage.go](server/channelserver/sys_stage.go))

- Stages are game rooms/areas where players interact
- Thread-safe stage creation, entry, movement, and destruction
- Stage locking/unlocking for privacy
- Stage binary data for synchronizing state between players

**Semaphore System** ([sys_semaphore.go](server/channelserver/sys_semaphore.go))

- Resource locking mechanism for shared game resources
- Used for quests, events, and multiplayer coordination
- Global and local semaphores

**Special Event Systems**

- Raviente: Large-scale raid event (in `Server.raviente`)
- Diva Defense, Hunter's Festa, VS Tournament (handlers in `handlers_event.go`, `handlers_festa.go`, `handlers_tournament.go`)

### Network Layer

**Packet Structure** ([network/mhfpacket/](network/mhfpacket/))

- Packet definitions in `msg_*.go` files
- Each packet type implements `MHFPacket` interface
- Packets prefixed by type: `MSG_SYS_*`, `MSG_MHF_*`, `MSG_CA_*`
- Binary packet handling in `network/binpacket/`

**Encryption** ([network/crypt_packet.go](network/crypt_packet.go))

- Custom encryption layer wrapping connections
- Different crypto for different server types

**Compression** ([server/channelserver/compression/](server/channelserver/compression/))

- Delta compression (`deltacomp`) for bandwidth optimization
- Null compression (`nullcomp`) for debugging

### Common Utilities

**ByteFrame** ([common/byteframe/](common/byteframe/))

- Buffer for reading/writing binary data
- Provides methods for reading/writing various data types
- Critical for packet construction and parsing

**PascalString** ([common/pascalstring/](common/pascalstring/))

- Length-prefixed string format used by game protocol
- Different variants for different string types

**Client Context** ([network/clientctx/](network/clientctx/))

- Stores client version and capabilities
- Used for multi-version support

## Database Schema Management

**Schema Types:**

- **Initialization Schema:** Bootstraps database to version 9.1.0 (in `schemas/`)
- **Patch Schemas:** Development updates in `patch-schema/` (numbered, apply in order)
- **Update Schemas:** Production-ready consolidated updates (for releases)
- **Bundled Schemas:** Demo data templates in `schemas/bundled-schema/` (shops, events, gacha)

**Important:** Patch schemas are subject to change during development. They get consolidated into update schemas on release. Apply patches in numerical order.

## Configuration System

Configuration uses Viper and is defined in `config/config.go`. The `ErupeConfig` global variable holds runtime configuration loaded from `config.json`.

**Key configuration sections:**

- `DevMode` and `DevModeOptions` - Development flags and debugging
- `GameplayOptions` - Gameplay modifiers (NP/RP caps, boost times, quest allowances)
- `Database` - PostgreSQL connection settings
- `Sign`, `SignV2`, `Entrance`, `Channel` - Server enable/disable and ports
- `Discord` - Discord bot integration
- `Logging` - File logging with rotation (uses lumberjack)

## Concurrency Patterns

**Critical:** This codebase heavily uses goroutines and shared state. Always:

1. Use mutexes when accessing shared state:
   - Server-level: `s.server.Lock()` / `s.server.Unlock()`
   - Stage-level: `s.server.stagesLock.Lock()` / `s.server.stagesLock.Unlock()`
   - Session-level: `s.Lock()` / `s.Unlock()`

2. Use RWMutex for read-heavy operations:
   - `s.server.stagesLock.RLock()` for reads
   - `s.server.stagesLock.Lock()` for writes

3. Test with race detector:

   ```bash
   go test -race ./...
   ```

4. Common concurrency scenarios:
   - Stage access: Always lock `s.server.stagesLock` when reading/writing stage map
   - Session broadcasts: Iterate sessions under lock
   - Database operations: Use transactions for multi-step operations

## Packet Handler Development

When adding new packet handlers:

1. Define packet structure in `network/mhfpacket/msg_*.go`
2. Implement `Build()` and `Parse()` methods
3. Add handler function in appropriate `handlers_*.go` file
4. Register in `handlerTable` map in `handlers_table.go`
5. Use helper functions:
   - `doAckBufSucceed(s, ackHandle, data)` - Success response with data
   - `doAckBufFail(s, ackHandle, data)` - Failure response
   - `stubEnumerateNoResults(s, ackHandle)` - Empty enumerate response
   - `stubGetNoResults(s, ackHandle)` - Empty get response

Example handler pattern:

```go
func handleMsgMhfYourPacket(s *Session, p mhfpacket.MHFPacket) {
    pkt := p.(*mhfpacket.MsgMhfYourPacket)

    // Process packet
    resp := byteframe.NewByteFrame()
    resp.WriteUint32(someValue)

    doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}
```

## Testing Practices

- Use table-driven tests for multiple scenarios
- Mock database operations where appropriate
- Test concurrent access patterns with goroutines
- Test both success and error paths
- Add tests in `*_test.go` files next to source files

## Common Pitfalls

1. **Thread Safety:** Always consider concurrent access. If unsure, add locks.
2. **Database Queries:** Use parameterized queries (`$1`, `$2`) to prevent SQL injection
3. **Error Handling:** Never ignore errors - log them or handle appropriately
4. **Session State:** Be careful with session state during disconnects
5. **Packet Ordering:** Some packets have ordering requirements - check client expectations
6. **Binary Data:** Always use `byteframe` for binary reads/writes to ensure correct endianness

## Commit and PR Guidelines

**Commit message format:**

- `feat: description` - New features
- `fix: description` - Bug fixes
- `refactor: description` - Code refactoring
- `docs: description` - Documentation
- `chore: description` - Maintenance tasks

**Before committing:**

1. Run `gofmt -w .`
2. Run tests: `go test -v ./...`
3. Check for race conditions: `go test -race ./...`
4. Update CHANGELOG.md under "Unreleased" section

**PR Requirements:**

- Clear description of changes
- Test coverage for new features
- No race conditions
- Passes all existing tests
- Updated documentation if needed

## Remote Server Operations

Environment variables for remote operations are stored in `.env` (gitignored). Copy from `.env.example` if needed.

### Fetch Logs

```bash
# Load environment and fetch logs
source .env
scp -r $SERVER:$REMOTE_LOGS/* $LOCAL_LOGS/
```

### Deploy

```bash
# Build for Linux, upload, and restart
source .env
GOOS=linux GOARCH=amd64 go build -o erupe-ce
scp erupe-ce $SERVER:$REMOTE_BIN
ssh $SERVER "cd $REMOTE_DIR && sudo systemctl restart erupe"
```

### Quick Commands

```bash
# Check server status
source .env && ssh $SERVER "systemctl status erupe"

# Tail remote logs
source .env && ssh $SERVER "tail -f $REMOTE_LOGS/erupe.log"

# View recent errors
source .env && ssh $SERVER "grep -E 'level.*(error|warn)' $REMOTE_LOGS/erupe.log | tail -50"
```

## Discord Integration

Optional Discord bot in `server/discordbot/` provides:

- Real-time server activity notifications
- Player connection/disconnection events
- Quest completions and event notifications

Configured via `Discord` section in config.json.

## Multi-Version Support

The codebase supports multiple game client versions through:

- `ClientMode` configuration setting
- Client context detection
- Version-specific packet handling
- Binary file compatibility (quests/scenarios in `bin/`)

Primary focus: G10-ZZ (ClientMode), with varying support for older versions.
