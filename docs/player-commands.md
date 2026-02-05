# Player Commands Reference

This guide documents all chat commands available to players on Erupe servers.

## Using Commands

Type commands in the chat window. All commands start with `!` (configurable by server operators).

**Example:** Type `!course hunterlife` to toggle the Hunter Life course.

---

## General Commands

### !reload

Reloads all players in your current area.

**Usage:** `!reload`

**What it does:** Forces a refresh of all player data in your Land/Stage. Useful if you see visual glitches or players appear invisible.

**Response:** "Reloading players..."

---

### !course

Toggle subscription courses on your account.

**Usage:** `!course <course_name>`

**Available courses** (server-dependent):
| Course Name | Description |
|-------------|-------------|
| `hunterlife` | Hunter Life Course - Various hunting QoL benefits |
| `extra` | Extra Course - Additional features |
| `premium` | Premium Course - Premium benefits |
| `assist` | Assist Course - Helper features |
| `nboost` | N Boost - Point multipliers |
| `frontier` | Frontier Course |
| `legend` | Legend Course |

**Examples:**
```
!course hunterlife    → Toggles Hunter Life Course
!course premium       → Toggles Premium Course
```

**Response:**
- "Hunter Life Course enabled" (if turning on)
- "Hunter Life Course disabled" (if turning off)
- "Hunter Life Course is locked" (if not available on this server)

**Note:** Available courses depend on server configuration. Ask your server operator which courses are enabled.

---

### !help

Display available commands.

**Usage:** `!help`

**What it does:** Shows a list of all enabled commands on the server.

**Note:** Implementation varies by server version.

---

### !timer

Toggle the quest timer display.

**Usage:** `!timer`

**What it does:** Shows or hides the quest countdown timer during hunts. Your preference is saved to your account.

**Response:**
- "Quest timer enabled"
- "Quest timer disabled"

---

### !playtime

Show your total playtime.

**Usage:** `!playtime`

**What it does:** Displays how long you've played on your current character.

**Response:** "Playtime: 142h 35m 12s"

---

## Account Linking Commands

### !psn

Link a PlayStation Network ID to your account.

**Usage:** `!psn <your_psn_id>`

**Example:** `!psn MyPSNUsername`

**What it does:** Associates your PSN ID with your Erupe account. Used for cross-platform features or server integrations.

**Response:** "PSN ID set to: MyPSNUsername"

---

### !discord

Generate a Discord account linking token.

**Usage:** `!discord`

**What it does:** Creates a temporary token you can use to link your Discord account via the server's Discord bot.

**Response:** "Discord token: abc123-xyz789"

**How to use:**
1. Type `!discord` in-game
2. Copy the token shown
3. Use the token with the server's Discord bot (usually via a `/link` command in Discord)

---

## Raviente Siege Commands

These commands are used during the Great Slaying (Raviente) siege event.

### !ravi

**Usage:** `!ravi <subcommand>`

**Subcommands:**

| Command | Alias | Description |
|---------|-------|-------------|
| `!ravi start` | - | Start the Great Slaying event |
| `!ravi checkmultiplier` | `!ravi cm` | Check current damage multiplier |
| `!ravi sendres` | `!ravi sr` | Send resurrection support to other players |
| `!ravi sendsed` | `!ravi ss` | Send sedation support |
| `!ravi reqsed` | `!ravi rs` | Request sedation support from others |

**Examples:**
```
!ravi start    → Initiates the Raviente siege (if not already running)
!ravi cm       → Shows "Current multiplier: 2.5x"
!ravi sr       → Sends resurrection support
```

**Notes:**
- Commands only work when players are participating in a Raviente siege
- `!ravi start` fails if a siege is already in progress
- Support commands require resources/items to use

---

## Special Chat Features

### @dice

Roll a random number.

**Usage:** Type `@dice` in chat

**What it does:** Rolls a random number between 1 and 100, visible to all players in your area.

**Note:** This is not a `!` command - just type `@dice` in regular chat.

---

## Admin Commands

These commands are disabled by default and require operator permissions. Server administrators may enable them for staff use.

### !tele

Teleport to coordinates.

**Usage:** `!tele <x> <y>`

**Example:** `!tele 500 -200`

**What it does:** Instantly moves your character to the specified X/Y coordinates on the current map.

---

### !rights

Set account permission level.

**Usage:** `!rights <value>`

**Example:** `!rights 31`

**What it does:** Modifies the rights/permissions integer on your account. Used for granting course access or special permissions.

---

### !kqf

Manage Key Quest Flags (HR progression).

**Usage:**
- `!kqf get` - View current KQF value
- `!kqf set <hex_value>` - Set KQF to specific value

**Example:** `!kqf set 00000000FFFFFFFF`

**What it does:** Directly modifies Hunter Rank key quest completion flags. The value must be exactly 16 hexadecimal characters.

**Note:** Changes take effect after switching worlds/lands.

---

### !ban

Ban a player from the server.

**Usage:** `!ban <character_id> [duration]`

**Parameters:**
- `character_id` - The 6-character ID of the player to ban
- `duration` (optional) - How long to ban. Omit for permanent ban.

**Duration format:** `<number><unit>`
| Unit | Meaning |
|------|---------|
| `s` | Seconds |
| `m` | Minutes |
| `h` | Hours |
| `d` | Days |
| `mo` | Months |
| `y` | Years |

**Examples:**
```
!ban ABC123          → Permanent ban
!ban ABC123 7d       → 7-day ban
!ban ABC123 2h       → 2-hour ban
!ban ABC123 1mo      → 1-month ban
```

---

## Troubleshooting

### Command not working?

1. **Check spelling** - Commands are case-sensitive on some servers
2. **Check if enabled** - Server operators can disable specific commands
3. **Check permissions** - Admin commands require operator status
4. **Try !help** - See which commands are available on your server

### "Command is disabled" message

The server operator has disabled this command. Contact your server admin if you believe it should be available.

### Command does nothing

- Make sure you're typing in the correct chat channel
- Some commands only work in specific situations (e.g., `!ravi` during sieges)
- Check for typos in parameters

---

## For Server Operators

For detailed configuration options, see [In-Game Commands Configuration](commands.md).

Commands are configured in `config.json` under the `Commands` section:

```json
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
```

Set `"Enabled": false` to disable any command, or change `"Prefix"` to use a different trigger word.
