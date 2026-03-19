# Unimplemented Handlers

Tracks channel server handlers that are empty stubs. Regenerate by searching the source:

```bash
# All unimplemented game features
grep -rn "// stub: unimplemented" server/channelserver/handlers_*.go

# All reserved protocol slots
grep -rn "// stub: reserved" server/channelserver/handlers_reserve.go
```

All empty handlers carry an inline comment — `// stub: unimplemented` for real game features,
`// stub: reserved` for protocol message IDs that MHF itself never uses.

---

## Unimplemented (68 handlers)

Grouped by handler file / game subsystem.

### Achievements (`handlers_achievement.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfResetAchievement` | Resets achievement progress for a character |
| `handleMsgMhfPaymentAchievement` | Achievement reward payout (currency/items) |
| `handleMsgMhfGetCaAchievementHist` | Fetch CA (cross-platform?) achievement history |
| `handleMsgMhfSetCaAchievement` | Set CA achievement state |

### Caravan (`handlers_caravan.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfPostRyoudama` | Post/submit Ryoudama (Caravan jewel) data |

### Cast Binary (`handlers_cast_binary.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgSysCastedBinary` | Relay of already-cast binary state (complement to `MsgSysCastBinary`) |

### Client Management (`handlers_clients.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfShutClient` | Server-initiated client disconnect |
| `handleMsgSysHideClient` | Hide a client session from the session list |

### Data / Auth (`handlers_data.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgSysAuthData` | Supplemental auth data exchange |

### Events (`handlers_event.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfGetRestrictionEvent` | Fetch event-based gameplay restrictions |

### Guild (`handlers_guild.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfUpdateForceGuildRank` | Force-set a guild's rank (admin/GM operation) |
| `handleMsgMhfUpdateGuild` | Update generic guild metadata |
| `handleMsgMhfUpdateGuildcard` | Update guild card display data |

### House / My Room (`handlers_house.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfResetTitle` | Reset a character's displayed title |

### Items (`handlers_items.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfStampcardPrize` | Claim a stamp card prize |

### Misc (`handlers_misc.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfKickExportForce` | Force-kick a character from an export/transfer |
| `handleMsgMhfRegistSpabiTime` | Register Spabi (practice area?) usage time |
| `handleMsgMhfDebugPostValue` | Debug value submission (client-side debug tool) |
| `handleMsgMhfGetDailyMissionMaster` | Fetch daily mission master data (template list) |
| `handleMsgMhfGetDailyMissionPersonal` | Fetch character's daily mission progress |
| `handleMsgMhfSetDailyMissionPersonal` | Save character's daily mission progress |
| `handleMsgMhfUseUdShopCoin` | Spend a UD Shop coin |

### Mutex (`handlers_mutex.go`)

All five mutex handlers are empty. MHF mutexes are distributed locks used for event coordination (Raviente, etc.). The server currently uses its own semaphore system instead.

| Handler | Notes |
|---------|-------|
| `handleMsgSysCreateMutex` | Create a named distributed mutex |
| `handleMsgSysCreateOpenMutex` | Create and immediately open a mutex |
| `handleMsgSysOpenMutex` | Acquire an existing mutex |
| `handleMsgSysCloseMutex` | Release a mutex |
| `handleMsgSysDeleteMutex` | Destroy a mutex |

### Object Sync (`handlers_object.go`)

Object sync is partially implemented (create, position, binary set/notify work). The following secondary operations are stubs:

| Handler | Notes |
|---------|-------|
| `handleMsgSysDeleteObject` | Delete a stage object |
| `handleMsgSysRotateObject` | Broadcast object rotation |
| `handleMsgSysDuplicateObject` | Duplicate a stage object |
| `handleMsgSysGetObjectBinary` | Fetch binary state of a specific object |
| `handleMsgSysGetObjectOwner` | Get the owning session of an object |
| `handleMsgSysUpdateObjectBinary` | Update object binary state |
| `handleMsgSysCleanupObject` | Clean up stale object state |
| `handleMsgSysAddObject` | Add an object to the stage |
| `handleMsgSysDelObject` | Remove an object from the stage |
| `handleMsgSysDispObject` | Display/show a previously hidden object |
| `handleMsgSysHideObject` | Hide an object from other clients |

### Quests (`handlers_quest.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfEnterTournamentQuest` | Enter a tournament-mode quest |

### Register (`handlers_register.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgSysNotifyRegister` | Notify server of a client-side registration event |

### Rewards (`handlers_reward.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfUseRewardSong` | Use/activate a reward song buff |
| `handleMsgMhfAddRewardSongCount` | Increment reward song usage counter |
| `handleMsgMhfAcceptReadReward` | Claim a reward for reading an in-game notice |

### Session (`handlers_session.go`)

Some of these may be intentionally no-ops (e.g. `MsgSysAck` is a client-to-server confirmation that needs no reply). Others are genuine feature gaps.

| Handler | Notes |
|---------|-------|
| `handleMsgHead` | Raw packet header handler (likely intentional no-op) |
| `handleMsgSysAck` | Client acknowledgement — no server reply expected |
| `handleMsgSysSetStatus` | Set session/character status flags |
| `handleMsgSysEcho` | Protocol echo/ping (no reply needed) |
| `handleMsgSysUpdateRight` | Update client rights/entitlements |
| `handleMsgSysAuthQuery` | Auth capability query |
| `handleMsgSysAuthTerminal` | Terminate auth session |
| `handleMsgSysTransBinary` | Transfer binary data between clients |
| `handleMsgSysCollectBinary` | Collect binary data from clients |
| `handleMsgSysGetState` | Get session state snapshot |
| `handleMsgSysSerialize` | Serialize session data |
| `handleMsgSysEnumlobby` | Enumerate available lobbies |
| `handleMsgSysEnumuser` | Enumerate users in context |
| `handleMsgSysInfokyserver` | Fetch key server info |
| `handleMsgCaExchangeItem` | CA (cross-platform) item exchange |
| `handleMsgMhfServerCommand` | Server-push command to client |
| `handleMsgMhfSetLoginwindow` | Configure client login window state |
| `handleMsgMhfGetCaUniqueID` | Fetch CA unique identifier |

### Stage (`handlers_stage.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgSysStageDestruct` | Destroy/teardown a stage |
| `handleMsgSysLeaveStage` | Client leaving a stage (complement to `handleMsgSysEnterStage`) |

### Tactics / UD (`handlers_tactics.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgMhfSetUdTacticsFollower` | Set UD (Unlimited?) Tactics follower data |
| `handleMsgMhfGetUdTacticsLog` | Fetch UD Tactics combat log |

### Users (`handlers_users.go`)

| Handler | Notes |
|---------|-------|
| `handleMsgSysInsertUser` | Register a new user session |
| `handleMsgSysDeleteUser` | Remove a user session |
| `handleMsgSysNotifyUserBinary` | Notify clients of a user binary state change |

---

## Reserved Protocol Slots (56 handlers)

`handlers_reserve.go` contains 56 empty handlers for message IDs that are reserved in the MHF protocol but never sent by any known client version. These are **intentionally empty** and are not missing features.

Two reserve IDs (`188`, `18B`) have partial implementations returning hardcoded responses — they appear in an unknown subsystem and are documented with inline comments in the source.

Full list: `handleMsgSysReserve01–07`, `0C–0E`, `4A–4F`, `55–57`, `5C`, `5E–5F`, `71–7C`, `7E`, `180`, `18E–18F`, `19B–19F`, `1A4`, `1A6–1AF`, `192–194`, `handleMsgMhfReserve10F`.
