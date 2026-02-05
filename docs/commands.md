# In-Game Commands

In-game chat commands for players and administrators.

> **Looking for player documentation?** See [Player Commands Reference](player-commands.md) for a user-friendly guide to using commands in-game.

## Configuration

```json
{
  "Commands": {
    "Enabled": true,
    "CommandPrefix": "!",
    "Reload": { "Enabled": true, "Prefix": "reload" },
    "Raviente": { "Enabled": true, "Prefix": "ravi" },
    "Course": { "Enabled": true, "Prefix": "course" },
    "Rights": { "Enabled": false, "Prefix": "rights" },
    "Teleport": { "Enabled": false, "Prefix": "tele" },
    "KeyQuest": { "Enabled": false, "Prefix": "kqf" },
    "Ban": { "Enabled": false, "Prefix": "ban" },
    "Help": { "Enabled": true, "Prefix": "help" },
    "PSN": { "Enabled": true, "Prefix": "psn" },
    "Discord": { "Enabled": true, "Prefix": "discord" },
    "Timer": { "Enabled": true, "Prefix": "timer" },
    "Playtime": { "Enabled": true, "Prefix": "playtime" }
  }
}
```

## How Commands Work

1. **Player Types Command**: Player sends a chat message starting with the command prefix
2. **Command Parser**: Server checks if message matches any enabled command prefix
3. **Command Handler**: Executes the command logic
4. **Response**: Server sends feedback message to the player

Commands are parsed in [handlers_cast_binary.go:90](../server/channelserver/handlers_cast_binary.go#L90).

## Available Commands

### Reload

**Prefix:** `!reload`
**Recommended:** Enabled
**Usage:** `!reload`

Reloads all players and objects in the current stage:

1. Removes all other players/objects from view
2. Waits 500ms
3. Re-adds all players/objects with updated data

**Use Cases:**

- Visual glitches (players appearing in wrong positions)
- Objects not displaying correctly
- Stage synchronization issues

**Example:**

```text
Player types: !reload
Server: "Reloading all players and objects..."
[View refreshes]
```

### Raviente

**Prefix:** `!ravi`
**Recommended:** Enabled
**Usage:** `!ravi <subcommand>`

Control Raviente raid event.

**Subcommands:**

- `!ravi start` - Start the Raviente event
- `!ravi cm` or `!ravi checkmultiplier` - Check current damage multiplier

**Examples:**

```text
Player types: !ravi start
Server: "Raviente event started!"

Player types: !ravi cm
Server: "Current Raviente multiplier: 2.5x"
```

### Course

**Prefix:** `!course`
**Recommended:** Enabled
**Usage:** `!course <course_name>`

Enable or disable subscription courses for your character.

**Course Names:**

- `hunterlife` or `hl` - Hunter Life Course
- `extra` or `ex` - Extra Course
- `premium` - Premium Course
- `assist` - Assist Course
- `n` - N Course
- `hiden` - Hiden Course
- `huntersupport` or `hs` - Hunter Support Course
- `nboost` - N Boost Course
- `netcafe` or `nc` - NetCafe Course
- `hlrenewing` - Hunter Life Renewing Course
- `exrenewing` - Extra Renewing Course

**Note:** Only courses with `Enabled: true` in the [Courses](courses.md) configuration can be toggled.

**Examples:**

```text
Player types: !course premium
Server: "Premium Course enabled!"
[Player's account rights updated]

Player types: !course premium
Server: "Premium Course disabled!"
[Toggled off]

Player types: !course hiden
Server: "Hiden Course is locked on this server"
[Course not enabled in config]
```

### KeyQuest (KQF)

**Prefix:** `!kqf`
**Recommended:** Disabled (unless intentional)
**Usage:** `!kqf get` or `!kqf set <hex>`

Get or set Key Quest flags (unlocks).

**Subcommands:**

- `!kqf get` - Display current KQF value
- `!kqf set <16-char hex>` - Set KQF value

**Examples:**

```text
Player types: !kqf get
Server: "Your KQF is: 0123456789ABCDEF"

Player types: !kqf set 0000000000000000
Server: "KQF set successfully!"
[Quest unlocks updated]

Player types: !kqf set invalid
Server: "Usage: !kqf set <16 hex characters>"
```

**Warning:** This allows players to unlock content. Disable unless you want players to have this power.

### Rights

**Prefix:** `!rights`
**Recommended:** Disabled
**Usage:** `!rights <number>`

Modify account rights/permissions (bitfield of enabled courses and permissions).

**Example:**

```text
Player types: !rights 255
Server: "Account rights set to: 255"
[All courses/permissions enabled]

Player types: !rights abc
Server: "Usage: !rights <number>"
```

**⚠️ SECURITY WARNING:** This allows players to grant themselves admin privileges and all courses. **Only enable in development or for trusted administrators.**

### Teleport

**Prefix:** `!tele`
**Recommended:** Disabled
**Usage:** `!tele <x> <y>`

Teleport to specific stage coordinates.

**Examples:**

```text
Player types: !tele 500 -200
Server: "Teleporting to 500 -200"
[Character moves to coordinates]
```

**⚠️ SECURITY WARNING:** Allows bypassing normal movement. Only enable for administrators.

---

### Help

**Prefix:** `!help`
**Recommended:** Enabled
**Usage:** `!help`

Display a list of all enabled commands on the server.

**Example:**

```text
Player types: !help
Server: "Available commands: !reload, !ravi, !course, !help..."
```

---

### Timer

**Prefix:** `!timer`
**Recommended:** Enabled
**Usage:** `!timer`

Toggle the quest timer display on/off. Preference is saved to the player's account.

**Example:**

```text
Player types: !timer
Server: "Quest timer disabled"

Player types: !timer
Server: "Quest timer enabled"
```

---

### Playtime

**Prefix:** `!playtime`
**Recommended:** Enabled
**Usage:** `!playtime`

Display total character playtime.

**Example:**

```text
Player types: !playtime
Server: "Playtime: 142h 35m 12s"
```

---

### PSN

**Prefix:** `!psn`
**Recommended:** Enabled
**Usage:** `!psn <psn_id>`

Link a PlayStation Network ID to the player's account.

**Example:**

```text
Player types: !psn MyPSNUsername
Server: "PSN ID set to: MyPSNUsername"
```

**Use Cases:**
- Cross-platform account linking
- Server-specific PSN integrations

---

### Discord

**Prefix:** `!discord`
**Recommended:** Enabled
**Usage:** `!discord`

Generate a temporary token for Discord account linking.

**Example:**

```text
Player types: !discord
Server: "Discord token: abc123-xyz789"
[Player uses token with Discord bot to link accounts]
```

**Note:** Requires [Discord Integration](discord-integration.md) to be configured.

---

### Ban

**Prefix:** `!ban`
**Recommended:** Disabled
**Usage:** `!ban <character_id> [duration]`

Ban a player from the server. Supports temporary and permanent bans.

**Duration Format:** `<number><unit>` where unit is:
- `s` - seconds
- `m` - minutes
- `h` - hours
- `d` - days
- `mo` - months
- `y` - years

**Examples:**

```text
Admin types: !ban ABC123
Server: "User ABC123 permanently banned"

Admin types: !ban ABC123 7d
Server: "User ABC123 banned for 7 days"

Admin types: !ban ABC123 2h
Server: "User ABC123 banned for 2 hours"
```

**⚠️ SECURITY WARNING:** Only enable for trusted administrators. Requires operator permissions.

## Command Configuration

Each command has three properties:

```json
{
  "Name": "CommandName",
  "Enabled": true,
  "Prefix": "!prefix"
}
```

| Property | Type | Description |
|----------|------|-------------|
| `Name` | string | Command identifier (must match internal name) |
| `Enabled` | boolean | Whether the command is active |
| `Prefix` | string | Chat prefix that triggers the command |

### Customizing Prefixes

You can change command prefixes:

```json
{
  "Name": "Reload",
  "Enabled": true,
  "Prefix": "/reload"
}
```

Now players would type `/reload` instead of `!reload`.

### Disabling Commands

Set `Enabled: false` to disable:

```json
{
  "Name": "Rights",
  "Enabled": false,
  "Prefix": "!rights"
}
```

When disabled, typing the command returns:

```text
Server: "The Rights command is disabled on this server"
```

## Security Considerations

### High-Risk Commands (Disable in Production)

- **Rights**: Grants admin privileges and all courses
- **KeyQuest**: Unlocks restricted content
- **Ban**: Can permanently remove players from server
- **Teleport**: Bypasses normal movement restrictions

### Medium-Risk Commands (Use with Caution)

- **Course**: Allows players to toggle courses without payment
  - Mitigate by limiting which courses are `Enabled` in [Courses](courses.md)

### Safe Commands

- **Reload**: Only affects visual state, no persistent changes
- **Raviente**: Only affects raid event, no account changes
- **Help**: Read-only, shows available commands
- **Timer**: User preference only, no gameplay impact
- **Playtime**: Read-only display
- **PSN**: Account linking, low risk
- **Discord**: Token generation, requires Discord bot to be useful

## Implementation Details

Commands are initialized on server startup from the config:

```go
// handlers_cast_binary.go:40
func init() {
    commands = make(map[string]config.Command)

    for _, cmd := range config.ErupeConfig.Commands {
        commands[cmd.Name] = cmd
        if cmd.Enabled {
            logger.Info(fmt.Sprintf("Command %s: Enabled, prefix: %s", cmd.Name, cmd.Prefix))
        }
    }
}
```

Each command handler checks if the command is enabled before executing.

## Examples

### Minimal Safe Configuration

```json
{
  "Commands": {
    "Enabled": true,
    "CommandPrefix": "!",
    "Reload": { "Enabled": true, "Prefix": "reload" },
    "Raviente": { "Enabled": true, "Prefix": "ravi" },
    "Help": { "Enabled": true, "Prefix": "help" }
  }
}
```

### Community Server Configuration

```json
{
  "Commands": {
    "Enabled": true,
    "CommandPrefix": "!",
    "Reload": { "Enabled": true, "Prefix": "reload" },
    "Raviente": { "Enabled": true, "Prefix": "ravi" },
    "Course": { "Enabled": true, "Prefix": "course" },
    "Help": { "Enabled": true, "Prefix": "help" },
    "Timer": { "Enabled": true, "Prefix": "timer" },
    "Playtime": { "Enabled": true, "Prefix": "playtime" },
    "Discord": { "Enabled": true, "Prefix": "discord" }
  }
}
```

### Full Development Configuration

```json
{
  "Commands": {
    "Enabled": true,
    "CommandPrefix": "!",
    "Reload": { "Enabled": true, "Prefix": "reload" },
    "Raviente": { "Enabled": true, "Prefix": "ravi" },
    "Course": { "Enabled": true, "Prefix": "course" },
    "Rights": { "Enabled": true, "Prefix": "rights" },
    "Teleport": { "Enabled": true, "Prefix": "tele" },
    "KeyQuest": { "Enabled": true, "Prefix": "kqf" },
    "Ban": { "Enabled": true, "Prefix": "ban" },
    "Help": { "Enabled": true, "Prefix": "help" },
    "PSN": { "Enabled": true, "Prefix": "psn" },
    "Discord": { "Enabled": true, "Prefix": "discord" },
    "Timer": { "Enabled": true, "Prefix": "timer" },
    "Playtime": { "Enabled": true, "Prefix": "playtime" }
  }
}
```

### Custom Prefixes (Slash Commands)

```json
{
  "Commands": {
    "Enabled": true,
    "CommandPrefix": "/",
    "Reload": { "Enabled": true, "Prefix": "refresh" },
    "Raviente": { "Enabled": true, "Prefix": "ravi" },
    "Course": { "Enabled": true, "Prefix": "sub" }
  }
}
```

## Related Documentation

- [Player Commands Reference](player-commands.md) - User-friendly guide for players
- [Courses](courses.md) - Course configuration for the `!course` command
- [Discord Integration](discord-integration.md) - Commands may post to Discord
- [Gameplay Options](gameplay-options.md) - Gameplay balance settings
