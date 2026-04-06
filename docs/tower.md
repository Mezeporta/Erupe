# Tower / Sky Corridor (天廊 / Tenrou)

Tracks what is known about the Tower system and what remains to be reverse-engineered
or implemented in Erupe.

The core of this system **is already implemented on `develop`** via the repository and
service pattern (`repo_tower.go`, `svc_tower.go`, `handlers_tower.go`). Two branches carry
earlier work: `wip/tower` predates the refactor and uses direct SQL; `feature/tower` merges
`wip/tower` with `feature/earth`. Their useful findings are incorporated below.

The `feature/tower` branch is **not mergeable in its current state** — it diverged too far
from `develop` and was superseded by the direct integration of tower code into `develop`.
Its main remaining value is the `PresentBox` handler logic and the `TimeTaken`/`CID`
field naming for `MsgMhfPostTowerInfo`.

---

## Game Context

The **Sky Corridor** (天廊, *Tenrou*) is a permanent dungeon introduced in MHFG6. Players
explore a multi-floor structure built by an ancient civilization on a desolate island. Unlike
quest zones, the Sky Corridor has its own progression systems independent of normal quests.

### Individual Tower Progression

- Each clear adds **10 minutes** to the timer; quests consist of 2–4 randomly selected floors.
- Each floor contains 2–3 rooms separated by gates, with Felynes that drop **purple medals**,
  fixed-position treasure chests, and shiny crystals.
- Players accumulate **Tower Rank (TR)**, **Tower Rank Points (TRP)**, and
  **Tower Skill Points (TSP)**.
- TSP is spent to level up one of 64 tower-specific skills stored server-side.
- Two floor-count columns exist in the DB (`block1`, `block2`), corresponding to the
  pre-G7 and G7+ floor tiers respectively.

### Duremudira (The Guardian)

**Duremudira** (天廊の番人, Tower Guardian) is the Emperor Ice Dragon — an Elder Dragon
introduced in MHFG6. It appears as an optional boss in the second district of the Sky
Corridor. Slaying it yields **Red and Grey Liquids** used to craft Sky Corridor Gems and
Sigils.

**Arrogant Duremudira** is a harder variant; less information about it is available in
English sources.

### Gems (Sky Corridor Decorations)

Gems are collectibles (30 slots, organized as 6 tiers × 5 per tier) obtained from the Tower.
They slot into equipment to activate skills without consuming armor skill slots. Stored
server-side as a 30-element CSV in `tower.gems`.

### Tenrouirai (天廊威来) — Guild Mission System

A guild-parallel challenge system layered over the Sky Corridor. Each guild has a
**mission page** (1-based) containing 3 active missions. Guild members contribute scores
by completing tower runs; when cumulative scores meet all three mission goals, the page
advances (after a guild RP donation threshold is also met). There are 33 hardcoded
missions across 3 blocks.

**Mission types** (the `Mission` field in `TenrouiraiData`):
| Value | Objective |
|-------|-----------|
| 1 | Floors climbed |
| 2 | Antiques collected |
| 3 | Chests opened |
| 4 | Felynes (cats) saved |
| 5 | TRP acquisition |
| 6 | Monster slays |

### Relation to the Earth Event Cycle

The Tower phase is **week 3** of the three-week Earth event rotation (see `docs/conquest-war.md`).
The `GetWeeklySeibatuRankingReward` handler (Operation=5, IDs 260001 and 260003) already
handles Tower dure kill rewards and the 155-entry floor reward table. These do not need to be
reimplemented here.

---

## Database Schema

All tower tables are defined in `server/migrations/sql/0001_init.sql`.

### `tower` — per-character progression

```sql
CREATE TABLE public.tower (
    char_id  integer,
    tr       integer,    -- Tower Rank
    trp      integer,    -- Tower Rank Points
    tsp      integer,    -- Tower Skill Points
    block1   integer,    -- Floor count, era 1 (pre-G7)
    block2   integer,    -- Floor count, era 2 (G7+)
    skills   text,       -- CSV of 64 skill levels
    gems     text        -- CSV of 30 gem quantities (6 tiers × 5)
);
```

`skills` and `gems` default to `EmptyTowerCSV(N)` — comma-separated zeros — when NULL.
Gems are encoded as `tier << 8 | (index_within_tier + 1)` in wire responses.

### Guild columns (`guilds` and `guild_characters`)

```sql
-- guilds
tower_mission_page  integer DEFAULT 1   -- Current Tenrouirai mission page
tower_rp            integer DEFAULT 0   -- Accumulated guild tower RP

-- guild_characters
tower_mission_1  integer   -- Member's score for mission slot 1
tower_mission_2  integer   -- Member's score for mission slot 2
tower_mission_3  integer   -- Member's score for mission slot 3
```

---

## Packet Overview

Ten packets implement the Tower system. All live in `network/mhfpacket/`. None have
`Build()` implemented (all return `NOT IMPLEMENTED`).

### `MsgMhfGetTowerInfo` — Client → Server → Client

Fetches character tower data. The `InfoType` field selects what data to return.

**Request** (`msg_mhf_get_tower_info.go`):
```
AckHandle uint32
InfoType  uint32   — 1=TR/TRP, 2=TSP+skills, 3=level(pre-G7), 4=history, 5=level(G7+)
Unk0      uint32   — unknown; never used by handler
Unk1      uint32   — unknown; never used by handler
```

**Response**: variable-length array of frames via `doAckEarthSucceed`.

| InfoType | Response per frame |
|----------|-------------------|
| 1 | `TR int32, TRP int32` |
| 2 | `TSP int32, Skills [64]int16` |
| 3, 5 | `Floors int32, Unk1 int32, Unk2 int32, Unk3 int32` — one frame per era (1 for G7, 2 for G8+) |
| 4 | `[5]int16 (history group 0), [5]int16 (history group 1)` |

**InfoTypes 3 and 5** use the same code path and both return `TowerInfoLevel` entries.
The distinction between them is not understood. The `wip/tower` branch treats them
identically.

**TowerInfoLevel Unk1/Unk2/Unk3**: three of the four level-entry fields are unknown.
They are hardcoded to `5` in `wip/tower` and `0` on `develop`. Whether they carry
max floor, session count, or display state is not known.

**InfoType 4 (history)**: returns two groups of 5 × int16. The `wip/tower` branch
hardcodes them as `{1, 2, 3, 4, 5}` / `{1, 2, 3, 4, 5}`. Their meaning (e.g. recent
clear times, floor high scores) is not reverse-engineered. On `develop` they return zeros.

**Current state on develop**: Implemented. Reads from `repo_tower.GetTowerData()`.
History data is zero-filled (semantics unknown). Level Unk1–Unk3 are zero-filled.

---

### `MsgMhfPostTowerInfo` — Client → Server → Client

Submits updated tower progress after a quest.

**Request** (`msg_mhf_post_tower_info.go`):
```
AckHandle uint32
InfoType  uint32   — 1 or 7 = progress update, 2 = skill purchase
Unk1      uint32   — unknown; logged in debug mode
Skill     int32    — skill index to level up (InfoType=2 only)
TR        int32    — new Tower Rank to set
TRP       int32    — TRP earned (added to existing)
Cost      int32    — TSP cost (InfoType=2) or TSP earned (InfoType=1,7)
Unk6      int32    — unknown; logged in debug mode
Unk7      int32    — unknown; logged in debug mode
Block1    int32    — floor count increment (InfoType=1,7)
Unk9      int64    — develop: reads as int64; wip/tower: reads as TimeTaken int32 + CID int32
```

**Field disambiguation — `Unk9` vs `TimeTaken + CID`**: the `wip/tower` branch splits the
final 8 bytes into `TimeTaken int32` (quest duration in seconds) and `CID int32`
(character ID). This interpretation appears more correct than a single int64 — the character
ID would make sense as a submission attribution field. `develop` keeps it as `Unk9 int64`
until confirmed. This should be verified with a packet capture.

**InfoType 7**: handled identically to InfoType 1. The difference between them is unknown
— it may relate to whether the run included Duremudira or was a normal floor clear.

**TSP rate note**: the handler comment in both branches says "This might give too much TSP?
No idea what the rate is supposed to be." The `Cost` field is used for both TSP earned
(on progress updates) and TSP spent (on skill purchases); the actual earn rate formula is
unknown.

**`block2` not written**: `UpdateProgress` only writes `block1`. The `block2` column (G7+
floor era) is never incremented by the handler. This is likely a bug — `block2` should be
written when the client sends a G7+ floor run.

**Current state on develop**: Implemented. Calls `towerRepo.UpdateSkills` (InfoType=2) and
`towerRepo.UpdateProgress` (InfoType=1,7). `Unk9`/`Unk1`/`Unk6`/`Unk7` are logged in
debug mode but not acted on.

---

### `MsgMhfGetTenrouirai` — Client → Server → Client

Fetches Tenrouirai (guild mission) data.

**Request** (`msg_mhf_get_tenrouirai.go`):
```
AckHandle    uint32
Unk0         uint8    — unknown; never used
DataType     uint8    — 1=mission defs, 2=rewards, 4=guild progress, 5=char scores, 6=guild RP
GuildID      uint32
MissionIndex uint8    — which mission to query scores for (DataType=5 only; 1-3)
Unk4         uint8    — unknown; never used
```

**DataType=1 response**: 33 frames, one per mission definition:
```
[uint8]  Block       — 1–3
[uint8]  Mission     — type (1–6, see table above)
[uint16] Goal        — score required
[uint16] Cost        — RP cost to unlock/advance
[uint8]  Skill1–6    — 6 skill requirement bytes (values: 80, 40, 40, 20, 40, 50)
```

**DataType=2 response (rewards)**: `TenrouiraiReward` struct is defined but never
populated. Returns an empty array. Response format:
```
[uint8]   Index
[uint16]  Item[0..4]     — 5 item IDs
[uint8]   Quantity[0..4] — 5 quantities
```
No captures of a populated reward response are known.

**DataType=4 response**: 1 frame:
```
[uint8]  Page       — current mission page (1-based)
[uint16] Mission1   — aggregated guild score for slot 1 (capped to goal)
[uint16] Mission2   — aggregated guild score for slot 2 (capped to goal)
[uint16] Mission3   — aggregated guild score for slot 3 (capped to goal)
```

**DataType=5 response**: N frames, one per guild member with a non-null score:
```
[int32]   Score
[14 bytes] HunterName (null-padded)
```

**DataType=6 response**: 1 frame:
```
[uint8]  Unk0    — always 0
[uint32] RP      — guild's accumulated tower RP
[uint32] Unk2    — unknown; always 0
```

**`TenrouiraiKeyScore`** (`Unk0 uint8, Unk1 int32`): defined and included in the
`Tenrouirai` struct but never written into or sent. Likely related to an unimplemented
DataType (possibly 3). Purpose unknown.

**Current state on develop**: DataTypes 1, 4, 5, 6 implemented. DataType 2 (rewards)
returns empty. `Unk0`, `Unk4`, and `TenrouiraiKeyScore` are unresolved.

---

### `MsgMhfPostTenrouirai` — Client → Server → Client

Submits Tenrouirai results or donates guild RP.

**Request** (`msg_mhf_post_tenrouirai.go`):
```
AckHandle uint32
Unk0      uint8
Op        uint8    — 1 = submit mission results, 2 = donate RP
GuildID   uint32
Unk1      uint8    — unknown

Op=1 fields:
  Floors    uint16   — floors climbed this run
  Antiques  uint16   — antiques collected
  Chests    uint16   — chests opened
  Cats      uint16   — Felynes saved
  TRP       uint16   — TRP obtained
  Slays     uint16   — monsters slain

Op=2 fields:
  DonatedRP  uint16  — RP to donate
  PreviousRP uint16  — prior RP total (from client; used for display only?)
  Unk2_0–3   uint16  — unknown; 4 reserved fields
```

**Critical gap — Op=1 does nothing**: the handler logs the fields in debug mode and
returns a success ACK, but **does not write any data to the database**. Mission scores
(`guild_characters.tower_mission_1/2/3`) are never updated from quest results. This means
Tenrouirai missions can never actually advance via normal gameplay — the `SUM` aggregation
in DataType=4 will always return zero.

To fix: the handler needs to determine which of the three active missions the current run
contributes to (based on mission type and the run's stats), then write to the appropriate
`tower_mission_N` column.

**Op=2 (RP donation)**: implemented. Deducts RP from character save data, updates
`guilds.tower_rp`, and advances the mission page when the cumulative donation threshold
is met. `Unk0`, `Unk1`, and `Unk2_0-3` are parsed but unused.

**Current state on develop**: Op=2 fully implemented. Op=1 is a no-op.

---

### `MsgMhfGetGemInfo` — Client → Server → Client

Fetches gem inventory or gem acquisition history.

**Request** (`msg_mhf_get_gem_info.go`):
```
AckHandle uint32
QueryType uint32   — 1=gem inventory, 2=gem history
Unk1      uint32   — unknown
Unk2–Unk6 int32    — unknown; 5 additional fields
```

**QueryType=1 response**: 30 frames (one per gem slot):
```
[uint16] Gem       — encoded as (tier << 8) | (index_within_tier + 1)
[uint16] Quantity
```

**QueryType=2 response**: N frames (gem history):
```
[uint16]   Gem
[uint16]   Message    — purpose unknown; likely a display string ID
[uint32]   Timestamp  — Unix timestamp
[14 bytes] Sender     — null-padded character name
```

**Current state on develop**: QueryType=1 implemented via `towerRepo.GetGems()`.
QueryType=2 returns empty (the `GemHistory` slice is never populated). `Unk1`–`Unk6`
are parsed but unused; purpose unknown.

---

### `MsgMhfPostGemInfo` — Client → Server → Client

Adds or transfers gems.

**Request** (`msg_mhf_post_gem_info.go`):
```
AckHandle uint32
Op        uint32   — 1=add gem, 2=transfer gem
Unk1      uint32   — unknown
Gem       int32    — gem ID encoded as (tier << 8) | (index+1)
Quantity  int32    — amount
CID       int32    — target character ID (Op=2 likely uses this)
Message   int32    — display message ID? purpose unknown
Unk6      int32    — unknown
```

**Op=1 (add gem)**: implemented. Decodes the gem index from the `Gem` field, increments
the quantity in the CSV, and saves. Note: the index computation `(pkt.Gem >> 8 * 5) + (pkt.Gem - pkt.Gem&0xFF00 - 1%5)` may have operator precedence issues — verify with captures.

**Op=2 (transfer gem)**: not implemented. Handler comment: *"no way im doing this for now"*.
The `CID` field likely identifies the recipient character. Format of the response (if any
acknowledgement is sent to the recipient) is unknown.

**Current state on develop**: Op=1 implemented via `towerService.AddGem()`. Op=2 stub.
`Unk1`, `Message`, `Unk6` purposes unknown.

---

### `MsgMhfPresentBox` — Client → Server → Client

Fetches or claims items from the Tower present box (a reward inbox for seibatsu/Tower
milestone awards). This packet's field names differ between `develop` and `wip/tower`:

**Request** — `wip/tower` naming (more accurate than develop's all-`Unk*` version):
```
AckHandle    uint32
Unk0         uint32
Operation    uint32   — 1=open list, 2=claim item, 3=close
PresentCount uint32   — number of PresentType entries that follow
Unk3         uint32   — unknown
Unk4         uint32   — unknown
Unk5         uint32   — unknown
Unk6         uint32   — unknown
PresentType  []uint32 — array of present type IDs (length = PresentCount)
```

On `develop`, `Operation` is `Unk1` and `PresentCount` is `Unk2` — the field is correctly
used to drive the `for` loop but the semantic name is lost.

**Response** for Op=1 and Op=2: N frames via `doAckEarthSucceed`, each:
```
[uint32] ItemClaimIndex    — unique claim ID
[int32]  PresentType       — echoes the request PresentType
[int32]  Unk2–Unk7         — 6 unknown fields (always 0 in captures)
[int32]  DistributionType  — 7201=item, 7202=N-Points, 7203=guild contribution
[int32]  ItemID
[int32]  Amount
```

**Critical gap — no claim tracking**: `ItemClaimIndex` is a sequential ID that the client
uses for "claimed" state, but the server has no DB table or flag for it. Every call
returns the same hardcoded items, so a player can claim the same rewards repeatedly.

**Op=3**: returns an empty buffer (close/dismiss).

**Current state on develop**: handler returns an empty item list (`data` slice is nil).
The `wip/tower` branch has a working hardcoded handler (7 dummy items per `PresentType`)
with the correct response structure. `Unk0`, `Unk3`–`Unk6` purposes unknown.

**Note**: `wip/tower`'s `MsgMhfPresentBox.Parse()` still contains `fmt.Printf` debug
print statements that must be removed before any merge.

---

### `MsgMhfGetNotice` — Client → Server → Client

Purpose unknown. Likely fetches in-lobby Tower notices or announcements.

**Request** (`msg_mhf_get_notice.go`):
```
AckHandle uint32
Unk0      uint32
Unk1      uint32
Unk2      int32
```

**Current state**: Stub on `develop` — returns `{0, 0, 0, 0}`. Response format unknown.

---

### `MsgMhfPostNotice` — Client → Server → Client

Purpose unknown. Likely submits a read-receipt or acknowledgement for a Tower notice.

**Request** (`msg_mhf_post_notice.go`):
```
AckHandle uint32
Unk0      uint32
Unk1      uint32
Unk2      int32
Unk3      int32
```

**Current state**: Stub on `develop` — returns `{0, 0, 0, 0}`. Response format unknown.

---

## What Is Already Working on Develop

- Character tower data (TR, TRP, TSP, skills, floor counts) is read and written via the
  full repository pattern.
- Skill levelling (InfoType=2) deducts TSP and increments the correct CSV index.
- Floor progress (InfoType=1,7) updates TR, TRP, TSP, and block1.
- All 33 Tenrouirai mission definitions are hardcoded and served correctly.
- Guild Tenrouirai progress (page, aggregated mission scores) is read and score-capped.
- Per-character Tenrouirai leaderboard (DataType=5) is read from DB.
- Guild tower RP donation (Op=2) deducts player RP, accumulates guild RP, and advances
  the mission page when the threshold is met.
- Gem inventory (QueryType=1) is read and returned correctly.
- Gem add (Op=1) updates the CSV at the correct index.

---

## What Needs RE or Implementation

### Functional bugs (affect gameplay today)

| Issue | Location | Notes |
|-------|----------|-------|
| `PostTenrouirai` Op=1 is a no-op | `handlers_tower.go` | Mission scores are never written; Tenrouirai cannot advance via normal play |
| `block2` never written | `repo_tower.go → UpdateProgress` | G7+ floor count not persisted; requires captures to confirm which InfoType sends it |
| `PresentBox` returns empty list | `handlers_tower.go` | No items are ever shown; `wip/tower` handler logic can be adapted |
| Present claim tracking absent | DB schema | No "claimed" flag; players can re-claim indefinitely once handler is populated |

### Unknown packet fields (need captures to resolve)

| Field | Packet | Notes |
|-------|--------|-------|
| `Unk9 int64` vs `TimeTaken int32 + CID int32` | `MsgMhfPostTowerInfo` | `wip/tower` splits this into two int32; likely correct — needs capture confirmation |
| `Unk1`, `Unk6`, `Unk7` | `MsgMhfPostTowerInfo` | Logged in debug but unused; may encode run metadata |
| `Unk0`, `Unk1` | `MsgMhfGetTowerInfo` | Never used in handler |
| `TowerInfoLevel.Unk1–Unk3` | `GetTowerInfo` InfoType=3,5 | 3 of 4 level-entry fields zero-filled; may be max floor, session count, display state |
| `TowerInfoHistory` 10 × int16 | `GetTowerInfo` InfoType=4 | Two groups of 5; semantics unknown (recent clear times? floor high scores?) |
| InfoType 3 vs 5 distinction | `MsgMhfGetTowerInfo` | Same code path; difference not understood |
| InfoType 7 vs 1 distinction | `MsgMhfPostTowerInfo` | Both update progress; difference not understood |
| `Unk0`, `Unk4` | `MsgMhfGetTenrouirai` | Always 0 in known captures |
| `TenrouiraiKeyScore` | `GetTenrouirai` | Struct defined, never sent; likely an unimplemented DataType |
| `Unk0`, `Unk1`, `Unk2_0–3` | `MsgMhfPostTenrouirai` | 6 parsed fields that are never used |
| `Unk1–Unk6` | `MsgMhfGetGemInfo` | 6 extra fields; may filter by tier, season, or character |
| `Message`, `Unk1`, `Unk6` | `MsgMhfPostGemInfo` | `Message` likely a display string ID for gem transfer notices |
| `Unk0`, `Unk3–Unk6` | `MsgMhfPresentBox` | 5 unknown request fields |
| All fields | `MsgMhfGetNotice`, `MsgMhfPostNotice` | Both packets entirely uncharacterized |

### Missing features (require further RE + design)

| Feature | Notes |
|---------|-------|
| Tenrouirai mission score submission | `PostTenrouirai` Op=1 needs to map run stats to the correct mission type and write `tower_mission_N` |
| Tenrouirai rewards (DataType=2) | `TenrouiraiReward` response format is known; item IDs and quantities are not |
| Gem transfer (PostGemInfo Op=2) | Recipient lookup via `CID`; likely requires a notification to the target session |
| Gem history (GetGemInfo QueryType=2) | Response structure is known; DB storage is not — would require a `gem_history` table |
| PresentBox claim tracking | Needs a `present_claims` table or a bitfield on the character |
| Notice system | Both Get/Post are stubs; may be Tower bulletin board or reward notifications |

---

## Relation to the Conquest War Doc

`docs/conquest-war.md` covers the `GetWeeklySeibatuRankingReward` handler which already
implements the Tower dure kill reward table (Op=5, ID 260001) and the Tower floor reward
table (Op=5, ID 260003). Those are not missing here — they live in `handlers_seibattle.go`
on the `feature/conquest` branch. When that branch is eventually integrated, ensure the
Tower floor reward data is preserved.
