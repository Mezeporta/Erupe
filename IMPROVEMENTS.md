# Erupe Improvement Recommendations

This document outlines prioritized improvements identified through codebase analysis.

---

## Cherry-Pick from Main Branch

The `main` branch is 589 commits ahead of `9.2.0-clean` but is unstable for players. The following commits should be cherry-picked (and fixed if necessary) for 9.3.0.

### Tier 1: Critical Stability Fixes (Cherry-pick immediately)

| Commit | Description | Files Changed | Risk |
|--------|-------------|---------------|------|
| `e1a461e` | fix(stage): fix deadlock preventing stage change | handlers_stage.go, sys_session.go | Low |
| `060635e` | fix(stage): fix race condition with stages | handlers_stage.go | Low |
| `1c32be9` | fix(session): race condition | sys_session.go | Low |
| `73e874f` | fix: array bound crashes on clans | Multiple | Low |
| `5028355` | prevent nil pointer in MhfGetGuildManageRight | handlers_guild.go | Low |
| `ba1eea8` | prevent save error crashes | handlers.go, handlers_character.go | Low |
| `60e86c7` | mitigate LoadDecoMyset crashing on older versions | handlers | Low |

**Command:**
```bash
git cherry-pick e1a461e 060635e 1c32be9 73e874f 5028355 ba1eea8 60e86c7
```

### Tier 2: Security Updates (Cherry-pick after Tier 1)

| Commit | Description | Risk |
|--------|-------------|------|
| `c13d6e6` | Bump golang.org/x/net from 0.33.0 to 0.38.0 | Low |
| `da43ad0` | Bump golang.org/x/crypto from 0.31.0 to 0.35.0 | Low |
| `0bf39b9` | Bump golang.org/x/net from 0.23.0 to 0.33.0 | Low |
| `c715578` | Bump golang.org/x/crypto from 0.15.0 to 0.17.0 | Low |

**Note:** May need to cherry-pick in order or resolve conflicts.

### Tier 3: Important Bug Fixes (Review before cherry-pick)

| Commit | Description | Files | Notes |
|--------|-------------|-------|-------|
| `d1dfc3f` | packet queue fix proposal | 6 files | Review carefully - touches core networking |
| `76858bb` | bypass full Stage check if reserve slot | handlers_stage.go | Simple fix |
| `c539905` | implement SysWaitStageBinary timeout | handlers_stage.go | Simple fix |
| `7459ded` | fix guild poogie outfit unlock | handlers | Simple fix |
| `8a55c5f` | fix inflated festa rewards | handlers | Review impact |
| `7d760bd` | fix EntranceServer clan member list limits | entranceserver | Simple fix |

### Tier 4: Version Compatibility Fixes

| Commit | Description | Versions Affected |
|--------|-------------|-------------------|
| `8d1c6a7` | S6 compatibility fix | Season 6.0 |
| `d26ae45` | fix G1 compatibility | G1 |
| `3d0114c` | fix MhfAcquireCafeItem cost in G1-G5.2 | G1-G5.2 |
| `8c219be` | fix InfoGuild response on <G10 | Pre-G10 |
| `183f886` | fix InfoFesta response on S6.0 | S6.0 |
| `1c4370b` | fix EnumerateFestaMember prior to Z2 | Pre-Z2 |

### Tier 5: Warehouse & Save System Fixes (Test thoroughly)

These commits fix critical player data issues but require careful testing:

| Commit | Description | Risk |
|--------|-------------|------|
| `9f19358` | fix Warehouse serialisation across versions | Medium - test all versions |
| `caf4deb` | fix Warehouse Equipment dereference | Medium |
| `e80a03d` | fix Warehouse Item functions | Medium |
| `b969c53` | fix Warehouse packet parsing | Medium |
| `717d785` | fix possible warehouse error | Low |

**Warning:** Save system changes (`36065ce`, `afc554f`, `18592c5`) are experimental and may have caused the instability on main. Test in isolation first.

### Tier 6: Features to Consider

| Commit | Feature | Dependencies | Notes |
|--------|---------|--------------|-------|
| `4eed6a9` | playtime chat command | None | Safe to cherry-pick |
| `0caaeac` | ngword filter | stringsupport | Useful for moderation |
| `1ab6940` | extra Distribution fields | Schema patch 23 | Requires DB migration |
| `2c58968` | emulate retail semaphore logic | None | May improve stability |

### Schema Patches Required

Main branch has 28 patch schema files. Cherry-picked commits may require these:

| Patch | Required For |
|-------|--------------|
| `23-rework-distributions-2.sql` | Distribution fields (`1ab6940`) |
| `24-fix-weekly-stamps.sql` | Weekly stamp fixes |
| `25-fix-rasta-id.sql` | Rasta ID fixes |
| `26-fix-mail.sql` | Mail fixes |
| `27-fix-character-defaults.sql` | Stage deadlock fix (`e1a461e`) |

### Commits to AVOID

These commits caused or may cause instability:

| Commit | Reason |
|--------|--------|
| `edd357f` | concatenate packets during send - later reverted |
| `ae32951` | packet concatenation - caused issues |
| `36065ce`, `afc554f`, `18592c5` | Save system changes - incomplete/experimental |
| Large feature branches | Event cycling, Discord improvements - too complex for point cherry-pick |

### Cherry-Pick Strategy

1. **Create feature branch:** `git checkout -b cherry-pick-stability`
2. **Cherry-pick Tier 1** (critical fixes) one by one, testing after each
3. **Run tests:** `go test -race ./...`
4. **Cherry-pick Tier 2** (security)
5. **Test with local client** before proceeding to Tier 3+
6. **Document any conflicts** and resolutions
7. **Apply required schema patches** to test database

### Verification Checklist

After cherry-picking, verify:
- [ ] Server starts without errors
- [ ] Player can log in
- [ ] Stage changes work (test quest entry/exit)
- [ ] No race conditions: `go test -race ./...`
- [ ] Guild operations work
- [ ] Warehouse access works
- [ ] Save/load works correctly

---

## Critical Priority

### 1. Test Coverage

**Current state:** 7.5% coverage on core channelserver (12,351 lines of code)

**Recommendations:**
- Add tests for packet handlers - 400+ handlers with minimal coverage
- Focus on critical files:
  - `server/channelserver/handlers_quest.go`
  - `server/channelserver/handlers_guild.go`
  - `server/channelserver/sys_session.go`
  - `server/channelserver/sys_stage.go`
- Create table-driven tests for the handler table
- Add fuzzing tests for packet parsing in `common/byteframe/`
- Target: 40%+ coverage on channelserver

### 2. Update Dependencies

Outdated packages in `go.mod` with potential security implications:

| Package | Current | Latest |
|---------|---------|--------|
| `go.uber.org/zap` | 1.18.1 | 1.27.0+ |
| `github.com/spf13/viper` | 1.8.1 | 1.18.0+ |
| `golang.org/x/crypto` | 0.1.0 | latest |
| `github.com/lib/pq` | 1.10.4 | latest |

**Action:**
```bash
go get -u ./...
go mod tidy
go test -v ./...
```

### 3. Add Context-Based Cancellation

`server/channelserver/sys_session.go` spawns goroutines without `context.Context`, preventing graceful shutdown and causing potential goroutine leaks.

**Changes needed:**
- Add `context.Context` to `Session.Start()`
- Pass context to `sendLoop()` and `recvLoop()`
- Implement cancellation on session close
- Add timeout contexts for database operations

---

## Important Priority

### 4. Fix Error Handling

**Issues found:**
- 61 instances of `panic()` or `Fatal()` that crash the entire server
- Ignored errors in `main.go` lines 29, 32, 195-196:
  ```go
  _ = db.MustExec("DELETE FROM guild_characters")  // Error ignored
  ```
- Typos in error messages (e.g., "netcate" instead of "netcafe" in `handlers_cafe.go`)

**Action:**
```bash
# Find all panics to review
grep -rn "panic(" server/

# Find ignored errors
grep -rn "_ = " server/ | grep -E "(Exec|Query)"
```

### 5. Refactor Large Files

Files exceeding maintainability guidelines:

| File | Lines |
|------|-------|
| `handlers_guild.go` | 1,986 |
| `handlers.go` | 1,835 |
| `handlers_shop_gacha.go` | 679 |
| `handlers_house.go` | 589 |

**Recommendations:**
- Split large handler files by functionality
- Move massive hex strings in `handlers_tactics.go` and `handlers_quest.go` to separate data files or compressed format
- Extract repeated patterns into utility functions

### 6. Enhance CI/CD Pipeline

**Current gaps:**
- No code coverage threshold enforcement
- No security scanning
- No database migration testing

**Add to `.github/workflows/`:**
- Coverage threshold (fail build if coverage drops below 30%)
- `gosec` for security scanning
- Integration tests with test database
- `go mod audit` for vulnerability scanning (Go 1.22+)

---

## Nice to Have

### 7. Logging Cleanup

**Issues:**
- 17 remaining `fmt.Print`/`println` calls should use zap
- `handlers_cast_binary.go` creates a new logger on every handler call (inefficient)

**Action:**
```bash
# Find printf calls that should use zap
grep -rn "fmt.Print" server/
grep -rn "println" server/
```

### 8. Configuration Improvements

**Hardcoded values to extract:**

| Value | Location | Suggested Config Key |
|-------|----------|---------------------|
| `maxDecoMysets = 40` | `handlers_house.go` | `GameplayOptions.MaxDecoMysets` |
| `decoMysetSize = 78` | `handlers_house.go` | `GameplayOptions.DecoMysetSize` |
| Session timeout (30s) | `sys_session.go:132` | `Channel.SessionTimeout` |
| Packet queue buffer (20) | `sys_session.go` | `Channel.PacketQueueSize` |

**Recommendation:** Create `config/constants.go` for game constants or add to `ErupeConfig`.

### 9. Resolve Technical Debt

14 TODO/FIXME comments in core code:

| File | Issue |
|------|-------|
| `signserver/session.go` | Token expiration not implemented |
| `handlers.go` | Off-by-one error in log key index |
| `handlers_guild.go` | Multiple incomplete features |
| `handlers_stage.go` | Unknown packet behavior |
| `crypto/crypto_test.go` | Failing test case needs debugging |

---

## Quick Wins

### Immediate Actions

```bash
# 1. Update dependencies
go get -u ./... && go mod tidy

# 2. Run security check
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# 3. Find all panics
grep -rn "panic(" server/ --include="*.go"

# 4. Find ignored errors
grep -rn "_ = " server/ --include="*.go" | grep -v "_test.go"

# 5. Check for race conditions
go test -race ./...

# 6. Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Low-Effort High-Impact

1. Fix error message typo in `handlers_cafe.go` ("netcate" -> "netcafe")
2. Add `defer rows.Close()` where missing after database queries
3. Replace `fmt.Print` calls with zap logger
4. Add missing error checks for `db.Exec()` calls

---

## Metrics to Track

| Metric | Current | Target |
|--------|---------|--------|
| Test coverage (channelserver) | 7.5% | 40%+ |
| Test coverage (overall) | 21.4% | 50%+ |
| Panic/Fatal calls | 61 | 0 (in handlers) |
| Ignored errors | ~20 | 0 |
| TODO/FIXME comments | 14 | 0 |
| Outdated dependencies | 4+ | 0 |

---

## Implementation Order

1. **Week 1:** Update dependencies, fix critical error handling
2. **Week 2:** Add context cancellation to session lifecycle
3. **Week 3-4:** Expand test coverage for core handlers
4. **Week 5:** Refactor large files, extract constants
5. **Ongoing:** Resolve TODO comments, improve documentation

---

## Release Milestones (9.3.0)

The following milestones are organized for the upcoming 9.3.0 release.

### Milestone 1: Security & Stability

**Token Lifecycle Management**
- [ ] Implement automatic token cleanup after inactivity (`signserver/session.go:133`)
- [ ] Add configurable token expiration time
- [ ] Add rate limiting on sign-in attempts
- [ ] Document security implications of `DisableTokenCheck` option

**Graceful Error Handling**
- [ ] Replace 9 `panic()` calls in `handlers_guild_scout.go` with proper error returns
- [ ] Replace `panic()` in `handlers_tower.go:43` (GetOwnTowerLevelV3) with stub response
- [ ] Convert fatal errors to recoverable errors where possible

**Database Connection Resilience**
- [ ] Configure connection pooling in `main.go:182`:
  ```go
  db.SetMaxOpenConns(25)
  db.SetMaxIdleConns(5)
  db.SetConnMaxLifetime(5 * time.Minute)
  db.SetConnMaxIdleTime(2 * time.Minute)
  ```
- [ ] Add connection health monitoring
- [ ] Implement reconnection logic on connection loss

---

### Milestone 2: Database Performance

**Add Missing Indexes**
- [ ] `CREATE INDEX idx_characters_user_id ON characters(user_id)`
- [ ] `CREATE INDEX idx_guild_characters_guild_id ON guild_characters(guild_id)`
- [ ] `CREATE INDEX idx_mail_sender_id ON mail(sender_id)`
- [ ] `CREATE INDEX idx_user_binary_character_id ON user_binary(character_id)`
- [ ] `CREATE INDEX idx_gacha_entries_gacha_id ON gacha_entries(gacha_id)`
- [ ] `CREATE INDEX idx_distribution_items_dist_id ON distribution_items(distribution_id)`

**Fix N+1 Query Patterns**
- [ ] `handlers_guild.go:1419-1444` - Batch alliance member queries into single UNION query
- [ ] `signserver/dbutils.go:135-162` - Rewrite friend/guildmate queries as JOINs
- [ ] `handlers_distitem.go:34-46` - Replace subquery with JOIN + GROUP BY
- [ ] `handlers_cafe.go:29-88` - Combine 4 single-field queries into one

**Implement Caching Layer**
- [ ] Create `server/channelserver/cache/cache.go` with `sync.RWMutex`-protected maps
- [ ] Cache gacha shop data at server startup (`handlers_shop_gacha.go:112`)
- [ ] Cache normal shop items
- [ ] Add cache invalidation on admin updates
- [ ] Cache guild information during session lifetime

---

### Milestone 3: Feature Completeness

**Guild System**
- [ ] Implement daily RP reset (`handlers_guild.go:740`)
- [ ] Enable guild alliance applications (`handlers_guild.go:1281`)
- [ ] Add guild message board cleanup (`handlers_guild.go:1888`)
- [ ] Record guild user counts to database (`handlers_guild.go:1946`)
- [ ] Implement monthly reward tracker (`handlers_guild.go:1967`)
- [ ] Handle alliance application deletion (`handlers_guild_alliance.go:154`)

**Daily/Recurring Systems**
- [ ] Implement gacha daily reset at noon (`handlers_shop_gacha.go:513`)
- [ ] Add achievement rank notifications (`handlers_achievement.go:122`)

**Daily Mission System** (currently empty handlers)
- [ ] Implement `handleMsgMhfGetDailyMissionMaster()`
- [ ] Implement `handleMsgMhfGetDailyMissionPersonal()`
- [ ] Implement `handleMsgMhfSetDailyMissionPersonal()`

**Tournament System** (`handlers_tournament.go`)
- [ ] Implement `handleMsgMhfEntryTournament()` (line 58)
- [ ] Implement `handleMsgMhfAcquireTournament()` (line 60)
- [ ] Complete tournament info handler with real data (line 14-25)

**Tower System** (`handlers_tower.go`)
- [ ] Fix `GetOwnTowerLevelV3` panic (line 43)
- [ ] Handle tenrou/irai hex decode errors gracefully (line 75)

**Seibattle System**
- [ ] Implement `handleMsgMhfGetSeibattle()` (`handlers.go:1708-1711`)
- [ ] Implement `handleMsgMhfPostSeibattle()`
- [ ] Add configuration toggle for Seibattle feature

---

### Milestone 4: Operational Excellence

**Health Checks & Monitoring**
- [ ] Add `/health` HTTP endpoint for container orchestration
- [ ] Add `/ready` readiness probe
- [ ] Add `/live` liveness probe
- [ ] Implement basic Prometheus metrics:
  - `erupe_active_sessions` gauge
  - `erupe_active_stages` gauge
  - `erupe_packet_processed_total` counter
  - `erupe_db_query_duration_seconds` histogram

**Logging Improvements**
- [ ] Replace all `fmt.Print`/`println` calls with zap (17 instances)
- [ ] Fix logger creation in `handlers_cast_binary.go` (create once, reuse)
- [ ] Add correlation IDs for request tracing
- [ ] Add structured context fields (player ID, stage ID, guild ID)

**Configuration Management**
- [ ] Create `config/constants.go` for game constants
- [ ] Make session timeout configurable
- [ ] Make packet queue buffer size configurable
- [ ] Add feature flags for incomplete systems (Tournament, Seibattle)

---

### Milestone 5: Discord Bot Enhancements

**Current state:** Output-only with minimal features

**New Features**
- [ ] Player login/logout notifications
- [ ] Quest completion announcements
- [ ] Achievement unlock notifications
- [ ] Guild activity feed (joins, leaves, rank changes)
- [ ] Administrative commands:
  - `/status` - Server status
  - `/players` - Online player count
  - `/kick` - Kick player (admin only)
  - `/announce` - Server-wide announcement
- [ ] Two-way chat bridge (Discord ↔ in-game)

---

### Milestone 6: Multi-Version Support

**Client Version Handling**
- [ ] Audit handlers for missing client version checks
- [ ] Document version-specific packet format differences
- [ ] Create version compatibility matrix
- [ ] Add version-specific tower system handling
- [ ] Test S6.0 through ZZ compatibility systematically

---

### Milestone 7: Schema Management

**Patch Schema Infrastructure**
- [ ] Create numbered patch files in `patch-schema/`:
  - `01_add_indexes.sql` - Performance indexes
  - `02_token_expiry.sql` - Token cleanup support
  - `03_daily_mission.sql` - Daily mission tables
- [ ] Add schema version tracking table
- [ ] Create migration runner script
- [ ] Document patch application process

**Schema Cleanup**
- [ ] Add PRIMARY KEY to `shop_items_bought`
- [ ] Add PRIMARY KEY to `cafe_accepted`
- [ ] Add foreign key constraints to child tables
- [ ] Remove or document unused tables (`achievement`, `titles`, `feature_weapon`)

---

### Milestone 8: Packet Implementation

**High-Value Packets** (393 files with "NOT IMPLEMENTED")

Priority implementations:
- [ ] `msg_mhf_create_joint.go` - Joint quest creation
- [ ] `msg_mhf_mercenary_huntdata.go` - Mercenary hunt data
- [ ] `msg_mhf_save_deco_myset.go` - Decoration preset saving
- [ ] `msg_mhf_get_ud_ranking.go` - User-defined quest rankings
- [ ] `msg_mhf_load_hunter_navi.go` - Hunter Navi system
- [ ] `msg_mhf_answer_guild_scout.go` - Guild scouting responses
- [ ] `msg_mhf_acquire_guild_tresure.go` - Guild treasure acquisition
- [ ] `msg_mhf_payment_achievement.go` - Payment achievements
- [ ] `msg_mhf_stampcard_prize.go` - Stamp card prizes

---

## Release Checklist

Before 9.3.0 release:

- [ ] All Milestone 1 items completed (Security & Stability)
- [ ] Critical database indexes added (Milestone 2)
- [ ] N+1 queries fixed (Milestone 2)
- [ ] Guild system TODOs resolved (Milestone 3)
- [ ] Health check endpoints added (Milestone 4)
- [ ] Schema patches created and tested (Milestone 7)
- [ ] Test coverage increased to 30%+
- [ ] All tests passing with race detector
- [ ] Dependencies updated
- [ ] CHANGELOG.md updated
- [ ] Documentation reviewed

---

## Metrics to Track

| Metric | Current | 9.3.0 Target |
|--------|---------|--------------|
| Test coverage (channelserver) | 7.5% | 40%+ |
| Test coverage (overall) | 21.4% | 50%+ |
| Panic/Fatal calls | 61 | <10 (critical paths only) |
| Ignored errors | ~20 | 0 |
| TODO/FIXME comments | 18 | <5 |
| Outdated dependencies | 4+ | 0 |
| N+1 query patterns | 4 | 0 |
| Missing critical indexes | 6 | 0 |
| Unimplemented packets | 393 | 380 (13 high-value done) |

---

*Generated: 2026-02-01*
*Updated: 2026-02-01 - Added release milestones*
