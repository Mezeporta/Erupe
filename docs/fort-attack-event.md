# Fort Attack Event (迎撃拠点 / Interceptor's Base)

Tracks what is known about the Interceptor's Base fort attack event system and what remains to be
reverse-engineered before it can be implemented in Erupe.

The `feature/enum-event` branch (origin) attempted a partial implementation but is not mergeable in
its current state. Its useful findings are incorporated below.

---

## Game Context

The **Interceptor's Base** (迎撃拠点) is a persistent field introduced in Forward.1 (April 2011).
Guilds defend a fortress adjacent to Mezeporta against invading Elder Dragons. The fort has a
**durability meter** — if monster attacks reduce it to 0% the quest fails regardless of time or
lives remaining. Monsters known to attack include Rukodiora, Rebidiora, Teoleskatle, Yamatukami,
Shengaroren, Harudomerugu, Rusted Kushala Daora, Belkyruros, Abiologu, and Keoaruboru.

**Keoaruboru** (古龍の侵攻 culmination, added MHF-Z Z1.1) is the hardest variant. Its limbs
accumulate heat as the fight progresses; if any limb reaches maximum heat it fires a beam at the
fort dealing 20% durability damage and resetting all heat. The fort starts at 80% integrity, meaning
four unchecked beams cause quest failure. Managing heat across limbs is the central mechanic.

The event was scheduled by Capcom's live servers on a cycle. The exact trigger frequency is not
publicly documented in either English or Japanese sources.

---

## Packet Overview

Five packets are involved. All live in `network/mhfpacket/`.

### `MsgMhfEnumerateEvent` (0x72) — Client → Server → Client

The client polls this on login to learn what fort attack events are currently scheduled.

**Request** (`msg_mhf_enumerate_event.go`): `AckHandle uint32` + two zeroed `uint16`.

**Response** built in `handleMsgMhfEnumerateEvent` (`handlers_event.go`):

```
[uint8]  event count
for each event:
  [uint16] EventType   — 0 = nothing; 1 or 2 = "Ancient Dragon has attacked the fort"
  [uint16] Unk1        — unknown; always 0 in known captures
  [uint16] Unk2        — unknown; always 0
  [uint16] Unk3        — unknown; always 0
  [uint16] Unk4        — unknown; always 0
  [uint32] StartTime   — Unix timestamp (seconds) when event begins
  [uint32] EndTime     — Unix timestamp when event ends
  if EventType == 2:
    [uint8]  quest file count
    [uint16] quest file ID × N
```

What `EventType == 1` means vs `EventType == 2` is not known. The quest file ID list only appears
when `EventType == 2`. The semantics of Unk1–Unk4 are entirely unknown.

**Current state**: Handler returns an empty event list (0 events). The `feature/enum-event` branch
adds DB-backed scheduling with a configurable `Duration` / `RestartAfter` cycle and a hardcoded
list of 19 quest IDs, but has a logic bug (inverted `rows.Next()` check) and uses raw DB calls
instead of the repo pattern.

---

### `MsgMhfRegisterEvent` — Client → Server → Client

Sent when a player attempts to join a fort attack session.

**Request** (`msg_mhf_register_event.go`):
```
AckHandle uint32
Unk0      uint16   — unknown
WorldID   uint16
LandID    uint16
CheckOnly bool     — if true, only check whether an event is active (don't join)
[uint8 zeroed padding]
```

**Response** (4 bytes): `WorldID uint8 | LandID uint8 | RaviID uint16`

**Current state**: Implemented in `handlers_register.go`. On `CheckOnly=true` with no active
Raviente semaphore it returns a zeroed 4-byte success. Otherwise it echoes back the world/land IDs
and `s.server.raviente.id`. This is the Raviente siege plumbing reused — whether it is correct for
fort attack (as opposed to the Raviente siege proper) is unknown.

---

### `MsgMhfReleaseEvent` — Client → Server

Sent when a player leaves a fort attack session. Carries `RaviID uint32` (the session ID returned
by RegisterEvent) plus a zeroed `uint32`.

**Current state**: Always returns `_ACK_EFAILED` (0x41). The correct success response format is
unknown — packet `Build()` is also unimplemented.

---

### `MsgMhfGetRestrictionEvent` — Client → Server → Client

Purpose unknown. Likely fetches per-player or per-world restrictions for event participation
(e.g. quest rank gate, prior completion check).

**Current state**: Packet `Parse()` and `Build()` both return `NOT IMPLEMENTED`. Handler is an
empty no-op (`handleMsgMhfGetRestrictionEvent`). No captures of this packet are known.

---

### `MsgMhfSetRestrictionEvent` — Client → Server → Client

Purpose unknown. Likely sets restriction state after an event completes or a player qualifies.

**Request** (`msg_mhf_set_restriction_event.go`):
```
AckHandle uint32
Unk0      uint32   — unknown
Unk1      uint32   — unknown
Unk2      uint32   — unknown
Unk3      uint8    — unknown
```

**Current state**: Handler returns a zeroed 4-byte success. `Build()` is unimplemented. Packet
semantics are entirely unknown.

---

## Shared State (Registers)

The Raviente siege uses three named register banks (`raviRegisterState`, `raviRegisterSupport`,
`raviRegisterGeneral`) served via `MsgSysLoadRegister` and mutated via `MsgSysOperateRegister`.
The fort attack event likely uses the same register mechanism for shared state (fort durability,
Keoaruboru heat accumulation, etc.), but which register IDs and slot indices map to which fort
variables has not been reverse-engineered.

`handleMsgSysNotifyRegister` is a stub (`// stub: unimplemented`) — this handler broadcasts
register updates to other players in the session. It must be implemented for multi-player fort
state synchronisation to work.

---

## Database

The `events` table (`server/migrations/sql/0001_init.sql`) already supports timestamped events:

```sql
CREATE TABLE public.events (
    id         SERIAL PRIMARY KEY,
    event_type event_type NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL
);
```

The `event_type` enum currently contains `festa`, `diva`, `vs`, `mezfes`. Adding `ancientdragon`
requires a migration:

```sql
ALTER TYPE event_type ADD VALUE 'ancientdragon';
```

The `feature/enum-event` branch placed this in `schemas/patch-schema/event-ancientdragon.sql`,
which is outside the numbered migration sequence and will not be auto-applied. It needs to be
added as `server/migrations/sql/0002_ancientdragon_event_type.sql` (or folded into the next
migration).

---

## What Needs RE Before Implementation

| Unknown | Where to look | Priority |
|---------|---------------|---------|
| Semantics of `EventType` values (1 vs 2, others?) | Packet captures during event window | High |
| Meaning of Unk1–Unk4 in the EnumerateEvent response | Packet captures + client disassembly | Medium |
| Correct `MsgMhfReleaseEvent` success response format | Packet captures | High |
| `MsgMhfGetRestrictionEvent` full structure (parse + response) | Packet captures | High |
| `MsgMhfSetRestrictionEvent` field semantics (Unk0–Unk3) | Packet captures | Medium |
| Which register IDs / slots carry fort durability | Packet captures during fort quest | High |
| Keoaruboru heat accumulation register mapping | Packet captures during Keoaruboru quest | High |
| Whether `MsgMhfRegisterEvent` reuses Raviente state correctly for fort | Packet captures + comparison with Raviente behaviour | Medium |
| Original event scheduling cadence (cycle length, trigger time) | Live server logs / JP wiki sources | Low |

---

## What Is Already Understood

- `MsgMhfEnumerateEvent` response wire format (field order, types, conditional quest ID list)
- `StartTime` / `EndTime` are Unix timestamps (confirmed by the feature branch)
- `MsgMhfRegisterEvent` request structure and plausible response format (echoes world/land + ravi ID)
- `MsgMhfReleaseEvent` request structure (carries the ravi session ID)
- `MsgMhfSetRestrictionEvent` request structure (5 fields, semantics unknown)
- The fort event cycles via the `events` table and can share the existing Raviente semaphore infrastructure
- Quest file IDs for fort quests: `20001, 20004–20006, 20011–20013, 20018–20029` (from feature branch config; unvalidated against captures)

---

## Relation to Raviente

The Raviente siege (`sys_channel_server.go`, `handlers_register.go`) is the closest implemented
analogue. It uses the same `MsgMhfRegisterEvent` / `MsgSysOperateRegister` / `MsgSysLoadRegister`
pipeline. Fort attack implementation can likely reuse or extend this infrastructure rather than
building a separate system. The key difference is that Raviente is always available (with its own
scheduling), while fort attacks are event-gated via `MsgMhfEnumerateEvent`.
