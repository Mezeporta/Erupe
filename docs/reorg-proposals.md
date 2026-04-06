# Reorg Proposals

Architectural improvements identified from the `chore/reorg` branch (never merged — too much drift). Each item is self-contained and can be implemented independently.

---

## High Priority

### 1. Remove `clientctx` from the mhfpacket interface

`ClientContext` is an empty struct (`struct{}`) that was never used. It appears as a parameter in every `Parse` and `Build` signature across all ~453 mhfpacket files.

**Current:**
```go
type MHFPacket interface {
    Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error
    Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error
    Opcode() packetid.PacketID
}
```

**Proposed:**
```go
type MHFPacket interface {
    Parse(bf *byteframe.ByteFrame) error
    Build(bf *byteframe.ByteFrame) error
    Opcode() packetid.PacketID
}
```

Delete `network/mhfpacket/clientcontext.go` and remove the `ctx` argument from all 453 packet files and all call sites. Purely mechanical — no behaviour change.

---

### 2. Lazy packet two-buffer sendLoop

ACK responses (which the client blocks on) should always flush before broadcast state updates (position syncs, cast binaries from other players). Without priority ordering, a busy stage with many players can delay ACKs long enough that the client appears to softlock or experiences input lag.

**Proposed change to `sys_session.go`:**

Add a `Lazy bool` field to `queuedMHFPacket`. Add `QueueSendMHFLazy` for low-priority packets. In `sendLoop`, maintain two accumulators: flush `buffer` first, then `lazybuffer`.

```go
type queuedMHFPacket struct {
    pkt  mhfpacket.MHFPacket
    Lazy bool
}

func (s *Session) QueueSendMHFLazy(pkt mhfpacket.MHFPacket) {
    s.sendPackets <- queuedMHFPacket{pkt: pkt, Lazy: true}
}
```

`BroadcastMHF` and `Stage.BroadcastMHF` should use `QueueSendMHFLazy` since broadcast packets are lower priority than direct ACK responses.

The current `QueueSendNonBlocking` (which drops packets silently on full queue) should be merged into `QueueSendMHFLazy` with a warning log on drop.

---

### 3. `SessionStage` interface — decouple Stage from `*Session`

`Stage` currently stores `map[*Session]uint32` for its clients and `*Session` for its host. This creates a circular import: Stage imports Session, Session imports Stage. It makes Stage impossible to move to a shared package and impossible to test in isolation.

**Proposed — new interface in `sys_stage.go`:**
```go
type SessionStage interface {
    QueueSendMHFLazy(pkt mhfpacket.MHFPacket)
    GetCharID() uint32
    GetName() string
}
```

Change Stage fields:
```go
// Before
Clients map[*Session]uint32
Host    *Session

// After
Clients map[SessionStage]uint32
Host    SessionStage
```

`*Session` already satisfies this interface via existing methods. `Stage.BroadcastMHF` iterates `SessionStage` values — no concrete session reference needed.

This is a prerequisite for eventually moving Stage to a shared internal package and for writing Stage unit tests with mock sessions.

---

### 4. Fix `semaphoreIndex` data race

`semaphoreIndex uint32` is a shared incrementing counter on the `Server` struct, initialised to `7`. It is read and written from multiple goroutines without a lock — this is a data race.

**Current (`sys_channel_server.go`):**
```go
type Server struct {
    // ...
    semaphoreIndex uint32
}
```

**Proposed — remove from Server, derive per-session:**

Each session already tracks `semaphoreID []uint16` and `semaphoreMode bool`. Derive the semaphore ID from those:

```go
func (s *Session) GetSemaphoreID() uint32 {
    if s.semaphoreMode {
        return 0x000E0000 + uint32(s.semaphoreID[1])
    }
    return 0x000F0000 + uint32(s.semaphoreID[0])
}
```

No shared counter, no race. Verify with `go test -race` before and after.

---

## Medium Priority

### 5. Extract chat commands to `chat_commands.go`

`handlers_cast_binary.go` currently contains both packet handlers and all chat command implementations (`ban`, `timer`, `psn`, `reload`, `kqf`, `rights`, `course`, `ravi`, `teleport`, `discord`, `help`) — roughly 400 lines of command logic mixed into a handler file.

**Proposed:** Move `parseChatCommand` and all command implementations to a new `chat_commands.go`. The handler in `handlers_cast_binary.go` calls `parseChatCommand(s, message)` — that call site stays unchanged.

While doing this, change `Config.Commands` from `[]Command` (linear scan) to `map[string]Command` (O(1) lookup by name).

---

### 6. i18n rewrite

`sys_language.go` currently uses bare string literals concatenated per locale. The chore/reorg branch introduces a cleaner pattern worth adopting independently of the full reorg.

**Proposed `i18n.go`:**
```go
var translations = map[string]map[string]string{
    "en": {
        "commands.ban.success.permanent": "User {username} has been permanently banned.",
        "commands.ban.success.timed":     "User {username} has been banned for {duration}.",
        // ...
    },
    "jp": {
        "commands.ban.success.permanent": "ユーザー {username} は永久BANされました。",
        // ...
    },
}

type v = map[string]string

func t(key string, placeholders v) string {
    locale := currentLocale() // or derive from session
    tmpl, ok := translations[locale][key]
    if !ok {
        tmpl = translations["en"][key]
    }
    for k, val := range placeholders {
        tmpl = strings.ReplaceAll(tmpl, "{"+k+"}", val)
    }
    return tmpl
}
```

Usage: `t("commands.ban.success.permanent", v{"username": uname})`

---

### 7. Extract model structs to shared location

Model structs currently defined inside handler files (`Guild`, `Mail`, `GuildApplication`, `FestivalColor`, etc.) cannot be used by services or other packages without importing `channelserver`, which risks import cycles.

**Proposed:** Move `db:`-tagged model structs out of handler files and into a dedicated location (e.g. `server/channelserver/model/` or a future `internal/model/`). Local-only types (used by exactly one handler) can stay in place.

`guild_model.go` is already a partial example of this pattern — extend it.

---

### 8. `doAck*` as session methods

The four free functions `doAckBufSucceed`, `doAckBufFail`, `doAckSimpleSucceed`, `doAckSimpleFail` take `*Session` as their first argument. They are inherently session operations and should be methods.

**Current:**
```go
doAckBufSucceed(s, pkt.AckHandle, data)
doAckBufFail(s, pkt.AckHandle, data)
```

**Proposed:**
```go
s.DoAckBufSucceed(pkt.AckHandle, data)
s.DoAckBufFail(pkt.AckHandle, data)
```

Exporting them (`DoAck…` vs `doAck…`) makes them accessible from service packages without having to pass a raw response buffer around. Mechanical update across all handler files.

---

### 9. `Server` → `ChannelServer` rename

The channel server's `Server` struct is ambiguous when `signserver.Server` and `entranceserver.Server` all appear in `main.go`. Within handler files, the `s` receiver is used for both `*Session` and `*Server` methods — reading a handler requires tracking which `s` is which.

**Proposed:**
- Rename `type Server struct` → `type ChannelServer struct` in `sys_channel_server.go`
- Change `Server` method receivers from `s *Server` to `server *ChannelServer`
- Session method receivers remain `s *Session`
- `s.server` field (on Session) stays as-is but now types to `*ChannelServer`

This makes "am I reading session code or server code?" immediately obvious from the receiver name.

---

## Low Priority

### 10. Split `sys_channel_server.go` — extract Raviente to `sys_ravi.go`

`sys_channel_server.go` handles server lifecycle, world management, and all Raviente siege state. Raviente is conceptually a separate subsystem.

**Proposed:** Move `Raviente` struct, `resetRaviente()`, `GetRaviMultiplier()`, and `UpdateRavi()` to a new `sys_ravi.go`. No logic changes.

---

### 11. Split `handlers_cast_binary.go` — extract broadcast utils to `sys_broadcast.go`

`BroadcastMHF`, `WorldcastMHF`, `BroadcastChatMessage`, and `BroadcastRaviente` are server-level infrastructure, not packet handlers. They live in `handlers_cast_binary.go` for historical reasons.

**Proposed:** Move them to a new `sys_broadcast.go`. No logic changes.

---

## What Not to Adopt from chore/reorg

These patterns appeared in the branch but should not be brought over:

- **DB singleton (`database.GetDB()`)** — services calling a global DB directly lose the repo-interface/mock pattern we rely on for handler unit tests. The session-as-service-locator problem is better solved by passing the DB through constructors explicitly.
- **`db *sqlx.DB` added to every handler signature** — the intent (make DB dependency explicit) is right, but the implementation is inconsistent (many handlers still call `database.GetDB()` internally). The existing repo-mock pattern is a better testability mechanism.
- **Services in `internal/service/` with inline SQL** — the move out of `channelserver` is correct; dropping the repo-interface pattern is not. Any service extraction should preserve mockability.
- **Logger singleton** — same concern as DB singleton: `sync.Once` initialization cannot be reset between tests, complicating parallel or isolated test runs.
