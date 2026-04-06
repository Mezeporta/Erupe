# Official Hunting Tournament (公式狩猟大会)

Documents the tournament system implementation status, known protocol details, and remaining
reverse-engineering gaps.

The `feature/hunting-tournament` branch (origin) is **not mergeable** — it duplicates handlers
that already exist in `handlers_tournament.go`. Its useful findings are incorporated below.

---

## Game Context

The 公式狩猟大会 (Official Hunting Tournament) was a recurring competitive event numbered
sequentially from the first (late 2007) through at least the 150th before service ended in
December 2019. It ran during the **登録祭** (Registration Festival) — week 1 of each 3-week
Mezeporta Festival (狩人祭) cycle.

### Competition Cups (杯)

| Cup | Group | Type | Description |
|-----|-------|------|-------------|
| **個人 G級韋駄天杯** (Solo speed hunt) | 16 | 7 | Time-attack solo vs. a designated monster. Results ranked per weapon class (EventSubType 0–13+ map to weapon categories). |
| **猟団対抗韋駄天杯** (Guild speed hunt) | 17 | 7 | Same time-attack concept, up to 4 hunters from the same guild. Guild rankings determine 魂 (souls) payouts to Mezeporta Festival. EventSubType -1 = all weapon classes combined. |
| **巨大魚杯** (Giant fish cup) | 6 | 6 | Fish size competition. Three designated species; largest catch wins. EventSubType maps to fish species. |

### Tournament Schedule

The tournament ran inside each 登録祭 week, and had four phases:

| Phase | State byte | Duration |
|-------|-----------|---------|
| Before start | 0 | Until `StartTime` |
| Registration + hunting | 1 | `StartTime` → `EntryEnd` (~3 days, Fri 14:00 to Mon 14:00) |
| Scoring / ranking | 2 | `EntryEnd` → `RankingEnd` (~+8.9 days) |
| Reward distribution | 3 | `RankingEnd` → `RewardEnd` (+7 days) |

The four Unix timestamps (`StartTime`, `EntryEnd`, `RankingEnd`, `RewardEnd`) are all included in
the `EnumerateRanking` response alongside the current state byte.

### Rewards

| Placement | Reward |
|-----------|--------|
| All participants | カフの素 (Skill Cuff base materials), ネコ珠の素 (Cat Gem base) |
| Top 500 | 匠チケット + ハーフチケット白 |
| Top 100 | 猟団ポイント (Guild points) |
| Top 3 (speed hunt) | 公式のしるし【金/銀/銅】(Official Mark Gold/Silver/Bronze) |
| Top 3 (fish cup) | 魚杯のしるし【金/銀/銅】(Fish Cup Mark Gold/Silver/Bronze) |
| 1st place (from tournament 76+) | 王者のメダル (King's Medal) — crafts exclusive weapons |
| Guild rank 1–10 | 50,000 魂 to faction + 5,000 to guild (Mezeporta Festival souls) |
| Guild rank 11–30 | 20,000 魂 to faction + 2,000 to guild |

---

## Implementation Status in `develop`

The tournament is **substantially implemented** in `handlers_tournament.go` and `repo_tournament.go`
with a full repository pattern and DB schema (`server/migrations/sql/0015_tournament.sql`).

### What Works

| Handler | File | Status |
|---------|------|--------|
| `handleMsgMhfEnumerateRanking` | `handlers_tournament.go` | Full — DB-backed, state machine, cups + sub-events |
| `handleMsgMhfEnumerateOrder` | `handlers_tournament.go` | Partial — returns leaderboard entries, but ranked by submission time (see gaps) |
| `handleMsgMhfInfoTournament` | `handlers_tournament.go` | Partial — type 0 (listing) and type 1 (registration check) work; type 2 (reward structures) returns empty |
| `handleMsgMhfEntryTournament` | `handlers_tournament.go` | Full — registers character, returns `entryID` |
| `handleMsgMhfEnterTournamentQuest` | `handlers_tournament.go` | Partial — records the submission, but clear time is not stored (see gaps) |
| `handleMsgMhfAcquireTournament` | `handlers_tournament.go` | Stub — returns empty reward list |

### Database Schema

Five tables in `0015_tournament.sql`:

```
tournaments          — schedule: id, name, start_time, entry_end, ranking_end, reward_end
tournament_cups      — per-tournament cup categories (cup_group, cup_type, name, description)
tournament_sub_events — shared event definitions (cup_group, event_sub_type, quest_file_id, name)
tournament_entries   — per-character registration (char_id, tournament_id, UNIQUE)
tournament_results   — per-submission record (char_id, tournament_id, event_id, quest_slot, stage_handle, submitted_at)
```

Note: `tournament_results` records *when* a submission arrived but not the actual quest clear time.
The leaderboard in `GetLeaderboard` therefore ranks by `submitted_at ASC` (first to submit = rank 1)
which is incorrect — the real server ranked by quest clear time.

---

## Known Gaps (RE Required)

### 1. Ranking by Quest Clear Time

**Impact**: High — the leaderboard is fundamentally wrong.

`handleMsgMhfEnterTournamentQuest` receives `TournamentID`, `EntryHandle`, `Unk2` (likely
`EventID`), `QuestSlot`, and `StageHandle`. None of these fields carry the actual clear time
directly. The clear time likely arrives via a separate packet (possibly `MsgMhfEndQuest` or a
dedicated score submission packet) that is not yet identified. Until it is, ranking by submission
order is a best-effort placeholder.

### 2. Guild Leaderboard Filtering by `ClanID`

**Impact**: Medium — guild cup leaderboard shows all entries instead of filtering by clan.

`MsgMhfEnumerateOrder` sends both `EventID` and `ClanID` (field names confirmed by the
`feature/hunting-tournament` branch). The current `GetLeaderboard` implementation queries only
by `event_id` and ignores `ClanID`. The guild cup (cup_group 17) leaderboard is presumably
filtered to show only that clan's members, or possibly compared against other clans. The exact
filtering semantics are unknown.

### 3. `AcquireTournament` Reward Delivery

**Impact**: High — players cannot receive any tournament rewards.

`handleMsgMhfAcquireTournament` returns an empty `TournamentReward` list. The
`TournamentReward` struct has three `uint16` fields (`Unk0`, `Unk1`, `Unk2`) that are entirely
unknown. It is unclear whether these carry item IDs, quantities, and flags, or whether the reward
delivery uses a different mechanism (e.g. mail). The 王者のメダル and 公式のしるし item IDs are
also unknown.

### 4. `InfoTournament` Type 2 (Reward Structures)

**Impact**: Medium — in-game reward preview is empty.

Query type 2 returns `TournamentInfo21` and `TournamentInfo22` lists — these likely describe
the per-placement reward tiers shown in the UI before a player claims their prize. All fields in
both structs are unknown (`Unk0`–`Unk4`).

### 5. `TournamentInfo0` Unknown Fields

**Impact**: Low — mostly display metadata.

The `TournamentInfo0` struct (used in `InfoTournament` type 0) has several unknown fields:
`MaxPlayers`, `CurrentPlayers`, `TextColor`, `Unk1`–`Unk6`, `MinHR`, `MaxHR`, plus two
unknown strings. Currently all written as zero/empty. The HR min/max likely gate tournament
access by hunter rank; `TextColor` likely styles the tournament name in the UI.

### 6. Guild Cup Souls → Mezeporta Festival Attribution

**Impact**: Medium — guild cup placement does not feed into Festa soul pool.

The guild speed hunt cup (cup_group 17) awarded 魂 to the guild's Mezeporta Festival account
based on placement. `handleMsgMhfAcquireTournament` currently delivers no rewards at all, let
alone Festa souls. Even once reward delivery is implemented, the soul injection into the Festa
system (via `FestaRepo.SubmitSouls` or similar) needs to be wired up.

---

## What the `feature/hunting-tournament` Branch Adds

The branch is not mergeable because it adds `handleMsgMhfEnumerateRanking` and
`handleMsgMhfEnumerateOrder` to `handlers_festa.go`, creating duplicate definitions that already
exist in `handlers_tournament.go`. However it contains several useful findings:

**`ClanID` field name on `MsgMhfEnumerateOrder`**
The two unknown fields (`Unk0`, `Unk1`) are identified as `EventID` and `ClanID`. `EventID` was
already used correctly in develop; `ClanID` is the new insight (currently ignored).

**Phase timing constants**
The branch's `generateTournamentTimestamps` debug modes confirm the timestamp offsets:
- `StartTime` → `EntryEnd`: +259,200 s (3 days)
- `EntryEnd` → `RankingEnd`: +766,800 s (~8.9 days)
- `RankingEnd` → `RewardEnd`: +604,800 s (7 days)

These match the real-server cadence and are already reflected in `TournamentDefaults.sql`.

**Festa timing correction (unrelated side effect)**
The branch also modifies `generateFestaTimestamps` in two ways that are not related to the
tournament but should be evaluated independently:
- `RestartAfter` threshold: 2,977,200 s → 3,024,000 s (34.45 days → 35 days)
- New event start time: midnight+24h → midnight+35h (i.e. 11:00 the following morning)

These changes appear to better match the real server schedule but have no test coverage. They
should be assessed against packet captures before merging.

---

## Seed Data Reference

`server/migrations/seed/TournamentDefaults.sql` pre-populates:

- 1 tournament (tournament #150, "第150回公式狩猟大会") with correct phase durations
- 18 sub-events:
  - cup_group 16 (individual speed hunt): EventSubType 0–13 against Brachydios, quest_file_id 60691
  - cup_group 17 (guild speed hunt): EventSubType -1, quest_file_id 60690
  - cup_group 6 (fish): キレアジ (EventSubType 234), ハリマグロ (237), カクサンデメキン (239)
- 3 cups: 個人 巨大魚杯 (id 569), 猟団 G級韋駄天杯 (id 570), 個人 G級韋駄天杯 (id 571)

The cup descriptions contain hardcoded dates ("2019年11月22日") from the original live event.
These should be templated or made dynamic when reward delivery is implemented.
