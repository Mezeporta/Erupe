# Erupe Configuration Guide

This guide explains the important configuration sections in `config.json`. Use [config.example.json](../config.example.json) as a template.

## Table of Contents

- [Basic Settings](#basic-settings)
- [Development Mode](#development-mode)
- [Gameplay Options](#gameplay-options)
- [Logging](#logging)
- [Discord Integration](#discord-integration)
- [In-Game Commands](#in-game-commands)
- [Courses](#courses)
- [Database](#database)
- [Server Configuration](#server-configuration)

---

## Basic Settings

```json
{
  "Host": "127.0.0.1",
  "BinPath": "bin",
  "Language": "en",
  "DisableSoftCrash": false,
  "HideLoginNotice": true
}
```

| Setting | Type | Description |
|---------|------|-------------|
| `Host` | string | Server IP address. Leave empty to auto-detect. Use `"127.0.0.1"` for local hosting |
| `BinPath` | string | Path to binary game data (quests, scenarios, etc.) |
| `Language` | string | Server language. `"en"` for English, `"jp"` for Japanese |
| `DisableSoftCrash` | boolean | When `true`, server exits immediately on crash (useful for auto-restart scripts) |
| `HideLoginNotice` | boolean | Hide the Erupe welcome notice on login |
| `LoginNotices` | array | Custom MHFML-formatted login notices to display |

### Additional Settings

- **`PatchServerManifest`**: Override URL for patch manifest server
- **`PatchServerFile`**: Override URL for patch file server
- **`ScreenshotAPIURL`**: Destination for screenshots uploaded to BBS
- **`DeleteOnSaveCorruption`**: If `true`, corrupted save data will be flagged for deletion

---

## Development Mode

```json
{
  "DevMode": true,
  "DevModeOptions": {
    "AutoCreateAccount": true,
    "CleanDB": false,
    "MaxLauncherHR": false,
    "LogInboundMessages": false,
    "LogOutboundMessages": false,
    "MaxHexdumpLength": 256,
    "DivaEvent": 0,
    "FestaEvent": -1,
    "TournamentEvent": 0,
    "MezFesEvent": true,
    "MezFesAlt": false,
    "DisableTokenCheck": false,
    "QuestDebugTools": false,
    "SaveDumps": {
      "Enabled": true,
      "OutputDir": "savedata"
    }
  }
}
```

| Setting | Type | Description |
|---------|------|-------------|
| `DevMode` | boolean | Enables development mode (more verbose logging, development logger format) |
| `AutoCreateAccount` | boolean | **⚠️ SECURITY RISK**: Auto-create accounts on login (disable in production) |
| `CleanDB` | boolean | **⚠️ DESTRUCTIVE**: Wipes database on server start |
| `MaxLauncherHR` | boolean | Sets launcher HR to HR7 to join non-beginner worlds |
| `LogInboundMessages` | boolean | Log all packets received from clients (very verbose) |
| `LogOutboundMessages` | boolean | Log all packets sent to clients (very verbose) |
| `MaxHexdumpLength` | number | Maximum bytes to display in packet hexdumps |
| `DivaEvent` | number | Diva Defense event status (0 = off, higher = active) |
| `FestaEvent` | number | Hunter's Festa event status (-1 = off, higher = active) |
| `TournamentEvent` | number | VS Tournament event status (0 = off, higher = active) |
| `MezFesEvent` | boolean | Enable/disable MezFes event |
| `MezFesAlt` | boolean | Swap Volpakkun for Tokotoko in MezFes |
| `DisableTokenCheck` | boolean | **⚠️ SECURITY RISK**: Skip login token validation |
| `QuestDebugTools` | boolean | Enable quest debugging logs |
| `SaveDumps.Enabled` | boolean | Enable saving character data dumps |
| `SaveDumps.OutputDir` | string | Directory for save data dumps |

---

## Gameplay Options

```json
{
  "GameplayOptions": {
    "FeaturedWeapons": 1,
    "MaximumNP": 100000,
    "MaximumRP": 50000,
    "DisableLoginBoost": false,
    "DisableBoostTime": false,
    "BoostTimeDuration": 120,
    "ClanMealDuration": 3600,
    "BonusQuestAllowance": 3,
    "DailyQuestAllowance": 1
  }
}
```

| Setting | Type | Description |
|---------|------|-------------|
| `FeaturedWeapons` | number | Number of Active Feature weapons generated daily |
| `MaximumNP` | number | Maximum Network Points (NP) a player can hold |
| `MaximumRP` | number | Maximum Road Points (RP) a player can hold |
| `DisableLoginBoost` | boolean | Disable login boost system |
| `DisableBoostTime` | boolean | Disable daily NetCafe boost time |
| `BoostTimeDuration` | number | NetCafe boost time duration in minutes (default: 120) |
| `ClanMealDuration` | number | Clan meal activation duration in seconds (default: 3600) |
| `BonusQuestAllowance` | number | Daily Bonus Point Quest allowance |
| `DailyQuestAllowance` | number | Daily Quest allowance |

---

## Logging

```json
{
  "Logging": {
    "LogToFile": true,
    "LogFilePath": "logs/erupe.log",
    "LogMaxSize": 100,
    "LogMaxBackups": 3,
    "LogMaxAge": 28,
    "LogCompress": true
  }
}
```

Erupe uses [lumberjack](https://github.com/natefinch/lumberjack) for automatic log rotation and compression.

| Setting | Type | Description |
|---------|------|-------------|
| `LogToFile` | boolean | Enable file logging (logs to both console and file) |
| `LogFilePath` | string | Path to log file (directory will be created automatically) |
| `LogMaxSize` | number | Maximum log file size in MB before rotation (default: 100) |
| `LogMaxBackups` | number | Number of old log files to keep (default: 3) |
| `LogMaxAge` | number | Maximum days to retain old logs (default: 28) |
| `LogCompress` | boolean | Compress rotated log files with gzip |

**Log Format:**

- When `DevMode: true`: Console format (human-readable)
- When `DevMode: false`: JSON format (production)

**Log Analysis:**
Use the built-in log analyzer tool to analyze logs. See [Log Analysis Guide](../tools/loganalyzer/README.md) or [CLAUDE.md](../CLAUDE.md#log-analysis).

---

## Discord Integration

```json
{
  "Discord": {
    "Enabled": false,
    "BotToken": "",
    "RealtimeChannelID": ""
  }
}
```

Erupe includes an optional Discord bot that posts real-time server activity to a Discord channel.

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable Discord integration |
| `BotToken` | string | Discord bot token from Discord Developer Portal |
| `RealtimeChannelID` | string | Discord channel ID where activity messages will be posted |

### How Discord Integration Works

When enabled, the Discord bot:

1. **Connects on Server Startup**: The bot authenticates using the provided bot token
2. **Monitors Game Activity**: Listens for in-game chat messages and events
3. **Posts to Discord**: Sends formatted messages to the specified channel

**What Gets Posted:**

- Player chat messages (when sent to world/server chat)
- Player connection/disconnection events
- Quest completions
- Special event notifications

### Setup Instructions

1. **Create a Discord Bot:**
   - Go to [Discord Developer Portal](https://discord.com/developers/applications)
   - Create a new application
   - Go to the "Bot" section and create a bot
   - Copy the bot token

2. **Get Channel ID:**
   - Enable Developer Mode in Discord (User Settings → Advanced → Developer Mode)
   - Right-click the channel you want to use
   - Click "Copy ID"

3. **Add Bot to Server:**
   - Go to OAuth2 → URL Generator in the Developer Portal
   - Select scopes: `bot`
   - Select permissions: `Send Messages`, `Read Message History`
   - Use the generated URL to invite the bot to your server

4. **Configure Erupe:**

   ```json
   {
     "Discord": {
       "Enabled": true,
       "BotToken": "YOUR_BOT_TOKEN_HERE",
       "RealtimeChannelID": "YOUR_CHANNEL_ID_HERE"
     }
   }
   ```

**Implementation Details:**

- Bot code: [server/discordbot/discord_bot.go](../server/discordbot/discord_bot.go)
- Uses [discordgo](https://github.com/bwmarrin/discordgo) library
- Message normalization for Discord mentions and emojis
- Non-blocking message sending (errors are logged but don't crash the server)

---

## In-Game Commands

```json
{
  "Commands": [
    {
      "Name": "Rights",
      "Enabled": false,
      "Prefix": "!rights"
    },
    {
      "Name": "Raviente",
      "Enabled": true,
      "Prefix": "!ravi"
    },
    {
      "Name": "Teleport",
      "Enabled": false,
      "Prefix": "!tele"
    },
    {
      "Name": "Reload",
      "Enabled": true,
      "Prefix": "!reload"
    },
    {
      "Name": "KeyQuest",
      "Enabled": false,
      "Prefix": "!kqf"
    },
    {
      "Name": "Course",
      "Enabled": true,
      "Prefix": "!course"
    }
  ]
}
```

In-game chat commands allow players to perform various actions by typing commands in the chat.

### Available Commands

| Command | Prefix | Description | Usage |
|---------|--------|-------------|-------|
| **Rights** | `!rights` | Modify account rights/permissions | `!rights <number>` |
| **Raviente** | `!ravi` | Control Raviente event | `!ravi start`, `!ravi cm` (check multiplier) |
| **Teleport** | `!tele` | Teleport to locations | `!tele <location>` |
| **Reload** | `!reload` | Reload all players and objects in current stage | `!reload` |
| **KeyQuest** | `!kqf` | Get/set Key Quest flags | `!kqf get`, `!kqf set <hex>` |
| **Course** | `!course` | Enable/disable subscription courses | `!course <course_name>` |

### How Commands Work

1. **Player Types Command**: Player sends a chat message starting with the command prefix
2. **Command Parser**: Server checks if message matches any enabled command prefix
3. **Command Handler**: Executes the command logic
4. **Response**: Server sends feedback message to the player

**Implementation Details:**

- Commands are parsed in [handlers_cast_binary.go:90](../server/channelserver/handlers_cast_binary.go#L90)
- Command map is initialized from config on server startup
- Each command has its own handler function
- Disabled commands return a "command disabled" message

### Example Command Usage

**Reload Command:**

```text
Player types: !reload
Server: "Reloading all players and objects..."
Server: Removes all other players/objects from view
Server: Re-adds all players/objects with updated data
```

**Course Command:**

```text
Player types: !course premium
Server: "Premium Course enabled!"
Server: Updates player's account rights in database
Server: Updates player's rights in current session
```

**KeyQuest Command:**

```text
Player types: !kqf get
Server: "Your KQF is: 0123456789ABCDEF"

Player types: !kqf set 0000000000000000
Server: "KQF set successfully!"
```

### Security Considerations

- **Rights Command**: Can grant admin privileges - disable in production
- **KeyQuest Command**: Can unlock content - disable if not desired
- Commands are per-server configuration (can be different per channel server)

---

## Courses

```json
{
  "Courses": [
    {"Name": "HunterLife", "Enabled": true},
    {"Name": "Extra", "Enabled": true},
    {"Name": "Premium", "Enabled": true},
    {"Name": "Assist", "Enabled": false},
    {"Name": "N", "Enabled": false},
    {"Name": "Hiden", "Enabled": false},
    {"Name": "HunterSupport", "Enabled": false},
    {"Name": "NBoost", "Enabled": false},
    {"Name": "NetCafe", "Enabled": true},
    {"Name": "HLRenewing", "Enabled": true},
    {"Name": "EXRenewing", "Enabled": true}
  ]
}
```

Courses are subscription-based features in Monster Hunter Frontier (similar to premium subscriptions).

| Course | Description |
|--------|-------------|
| `HunterLife` | Hunter Life Course - Basic subscription benefits |
| `Extra` | Extra Course - Additional benefits |
| `Premium` | Premium Course - Premium features |
| `Assist` | Assist Course - Helper features |
| `N` | N Course |
| `Hiden` | Hiden Course - Secret/hidden features |
| `HunterSupport` | Hunter Support Course |
| `NBoost` | N Boost Course |
| `NetCafe` | NetCafe Course - Internet cafe benefits |
| `HLRenewing` | Hunter Life Renewing Course |
| `EXRenewing` | Extra Renewing Course |

**Note:** When `Enabled: true`, players can toggle the course on/off using the `!course` command (if enabled). When `Enabled: false`, the course is completely locked and cannot be activated.

---

## Database

```json
{
  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "postgres",
    "Password": "",
    "Database": "erupe"
  }
}
```

PostgreSQL database configuration.

| Setting | Type | Description |
|---------|------|-------------|
| `Host` | string | Database host (default: `localhost`) |
| `Port` | number | Database port (default: `5432`) |
| `User` | string | Database user |
| `Password` | string | Database password (**required**, must not be empty) |
| `Database` | string | Database name |

**Setup:**

1. Install PostgreSQL
2. Create database: `createdb erupe`
3. Apply schema: `psql -U postgres -d erupe -f schema.sql`
4. Apply patches in order: `psql -U postgres -d erupe -f patch-schema/01_patch.sql`

See [CLAUDE.md](../CLAUDE.md#database-operations) for more details.

---

## Server Configuration

### Sign Server (Authentication)

```json
{
  "Sign": {
    "Enabled": true,
    "Port": 53312
  }
}
```

Legacy sign server for authentication and account management.

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable the sign server |
| `Port` | number | Port number (default: 53312) |

### SignV2 Server (Modern Authentication)

```json
{
  "SignV2": {
    "Enabled": false,
    "Port": 8080
  }
}
```

Modern HTTP-based sign server (alternative to legacy sign server).

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable the modern sign server |
| `Port` | number | Port number (default: 8080) |

**Note:** Only enable one sign server at a time (Sign OR SignV2, not both).

### Channel Server

```json
{
  "Channel": {
    "Enabled": true
  }
}
```

Channel servers handle actual gameplay sessions, quests, and player interactions.

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable channel servers (required for gameplay) |

### Entrance Server

```json
{
  "Entrance": {
    "Enabled": true,
    "Port": 53310,
    "Entries": [
      {
        "Name": "Newbie",
        "Description": "",
        "IP": "",
        "Type": 3,
        "Recommended": 2,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54001, "MaxPlayers": 100 },
          { "Port": 54002, "MaxPlayers": 100 }
        ]
      }
    ]
  }
}
```

Entrance server manages world/server selection and character lists.

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable entrance server (required) |
| `Port` | number | Entrance server port (default: 53310) |

#### Entrance Entries

Each entry represents a "world" in the server list.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | World name displayed to players |
| `Description` | string | World description (optional) |
| `IP` | string | Override IP (leave empty to use Host setting) |
| `Type` | number | World type: `1`=Normal, `2`=Cities, `3`=Newbie, `4`=Tavern, `5`=Return, `6`=MezFes |
| `Recommended` | number | Recommendation status: `0`=None, `2`=Recommended, `6`=Special |
| `AllowedClientFlags` | number | Client version flags (0 = all allowed) |
| `Channels` | array | List of channel servers in this world |

#### Channel Configuration

| Field | Type | Description |
|-------|------|-------------|
| `Port` | number | Channel server port (must be unique) |
| `MaxPlayers` | number | Maximum players per channel |
| `CurrentPlayers` | number | Current player count (auto-updated, can be set to 0) |

**World Types Explained:**

- **Newbie (Type 3)**: For new players, typically recommended
- **Normal (Type 1)**: Standard gameplay world
- **Cities (Type 2)**: Social/town areas
- **Tavern (Type 4)**: Bar/tavern areas
- **Return (Type 5)**: For returning players
- **MezFes (Type 6)**: MezFes event world

---

## Complete Example

Here's a minimal production-ready configuration:

```json
{
  "Host": "",
  "BinPath": "bin",
  "Language": "en",
  "DisableSoftCrash": true,
  "HideLoginNotice": false,
  "DevMode": false,
  "DevModeOptions": {
    "AutoCreateAccount": false
  },
  "GameplayOptions": {
    "MaximumNP": 100000,
    "MaximumRP": 50000
  },
  "Logging": {
    "LogToFile": true,
    "LogFilePath": "logs/erupe.log",
    "LogMaxSize": 100,
    "LogMaxBackups": 7,
    "LogMaxAge": 30,
    "LogCompress": true
  },
  "Discord": {
    "Enabled": false
  },
  "Commands": [
    {"Name": "Reload", "Enabled": true, "Prefix": "!reload"},
    {"Name": "Course", "Enabled": true, "Prefix": "!course"}
  ],
  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "erupe",
    "Password": "CHANGE_ME",
    "Database": "erupe"
  },
  "Sign": {
    "Enabled": true,
    "Port": 53312
  },
  "Channel": {
    "Enabled": true
  },
  "Entrance": {
    "Enabled": true,
    "Port": 53310,
    "Entries": [
      {
        "Name": "Main",
        "Type": 1,
        "Channels": [
          { "Port": 54001, "MaxPlayers": 100 }
        ]
      }
    ]
  }
}
```

---

## Additional Resources

- [CLAUDE.md](../CLAUDE.md) - Development guide
- [config.example.json](../config.example.json) - Full example configuration
- [Log Analyzer Tool](../tools/loganalyzer/) - Log analysis utilities
