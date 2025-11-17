# In-Game Commands

In-game chat commands for players and administrators.

## Configuration

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
**Recommended:** Disabled (not yet implemented)
**Usage:** `!tele <location>`

Teleport to specific locations.

**Status:** Command structure exists but handler not fully implemented in the codebase.

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

### Medium-Risk Commands (Use with Caution)

- **Course**: Allows players to toggle courses without payment
  - Mitigate by limiting which courses are `Enabled` in [Courses](courses.md)

### Safe Commands

- **Reload**: Only affects visual state, no persistent changes
- **Raviente**: Only affects raid event, no account changes

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
  "Commands": [
    {"Name": "Reload", "Enabled": true, "Prefix": "!reload"},
    {"Name": "Raviente", "Enabled": true, "Prefix": "!ravi"}
  ]
}
```

### Full Development Configuration

```json
{
  "Commands": [
    {"Name": "Rights", "Enabled": true, "Prefix": "!rights"},
    {"Name": "Raviente", "Enabled": true, "Prefix": "!ravi"},
    {"Name": "Teleport", "Enabled": true, "Prefix": "!tele"},
    {"Name": "Reload", "Enabled": true, "Prefix": "!reload"},
    {"Name": "KeyQuest", "Enabled": true, "Prefix": "!kqf"},
    {"Name": "Course", "Enabled": true, "Prefix": "!course"}
  ]
}
```

### Custom Prefixes

```json
{
  "Commands": [
    {"Name": "Reload", "Enabled": true, "Prefix": "/refresh"},
    {"Name": "Raviente", "Enabled": true, "Prefix": "/ravi"},
    {"Name": "Course", "Enabled": true, "Prefix": "/sub"}
  ]
}
```

## Related Documentation

- [Courses](courses.md) - Course configuration for the `!course` command
- [Discord Integration](discord-integration.md) - Commands may post to Discord
- [Gameplay Options](gameplay-options.md) - Gameplay balance settings
