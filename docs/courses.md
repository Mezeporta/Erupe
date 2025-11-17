# Courses

Subscription course configuration for Monster Hunter Frontier.

## Configuration

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

## What Are Courses?

Courses are subscription-based premium features in Monster Hunter Frontier, similar to premium subscriptions or season passes. Each course grants different benefits, bonuses, and access to content.

## Course List

| Course | Name | Description |
|--------|------|-------------|
| `HunterLife` | Hunter Life Course | Basic subscription with fundamental benefits |
| `Extra` | Extra Course | Additional premium features and bonuses |
| `Premium` | Premium Course | Premium-tier benefits and exclusive content |
| `Assist` | Assist Course | Helper features for newer players |
| `N` | N Course | Special N-series benefits |
| `Hiden` | Hiden Course | Secret/hidden features and content |
| `HunterSupport` | Hunter Support Course | Support features for hunters |
| `NBoost` | N Boost Course | N-series boost benefits |
| `NetCafe` | NetCafe Course | Internet cafe benefits (bonus rewards, boost time) |
| `HLRenewing` | Hunter Life Renewing | Renewed Hunter Life benefits |
| `EXRenewing` | Extra Renewing | Renewed Extra Course benefits |

## How Courses Work

### Enabled vs Disabled

- **`Enabled: true`**: Players can toggle the course on/off using the `!course` command
- **`Enabled: false`**: Course is locked and cannot be activated by players

### Account Rights

Courses are stored as a bitfield in the database (`users.rights` column). Each course has a unique bit position:

```text
Rights Bitfield:
Bit 0:  HunterLife
Bit 1:  Extra
Bit 2:  Premium
... (and so on)
```

### Using the !course Command

When the [!course command](commands.md#course) is enabled, players can toggle courses:

```text
Player types: !course premium
Server: "Premium Course enabled!"

Player types: !course premium
Server: "Premium Course disabled!"
```

**Important:** Players can only toggle courses that are `Enabled: true` in the configuration.

## Configuration Examples

### Open Server (All Courses Available)

```json
{
  "Courses": [
    {"Name": "HunterLife", "Enabled": true},
    {"Name": "Extra", "Enabled": true},
    {"Name": "Premium", "Enabled": true},
    {"Name": "Assist", "Enabled": true},
    {"Name": "N", "Enabled": true},
    {"Name": "Hiden", "Enabled": true},
    {"Name": "HunterSupport", "Enabled": true},
    {"Name": "NBoost", "Enabled": true},
    {"Name": "NetCafe", "Enabled": true},
    {"Name": "HLRenewing", "Enabled": true},
    {"Name": "EXRenewing", "Enabled": true}
  ]
}
```

### Restricted Server (Core Courses Only)

```json
{
  "Courses": [
    {"Name": "HunterLife", "Enabled": true},
    {"Name": "Extra", "Enabled": true},
    {"Name": "Premium", "Enabled": false},
    {"Name": "Assist", "Enabled": false},
    {"Name": "N", "Enabled": false},
    {"Name": "Hiden", "Enabled": false},
    {"Name": "HunterSupport", "Enabled": false},
    {"Name": "NBoost", "Enabled": false},
    {"Name": "NetCafe", "Enabled": true},
    {"Name": "HLRenewing", "Enabled": false},
    {"Name": "EXRenewing", "Enabled": false}
  ]
}
```

### Free-to-Play Server (NetCafe Only)

```json
{
  "Courses": [
    {"Name": "HunterLife", "Enabled": false},
    {"Name": "Extra", "Enabled": false},
    {"Name": "Premium", "Enabled": false},
    {"Name": "Assist", "Enabled": false},
    {"Name": "N", "Enabled": false},
    {"Name": "Hiden", "Enabled": false},
    {"Name": "HunterSupport", "Enabled": false},
    {"Name": "NBoost", "Enabled": false},
    {"Name": "NetCafe", "Enabled": true},
    {"Name": "HLRenewing", "Enabled": false},
    {"Name": "EXRenewing", "Enabled": false}
  ]
}
```

## Course Benefits

While Erupe emulates the course system, the exact benefits depend on game client implementation. Common benefits include:

### HunterLife

- Increased reward multipliers
- Extra item box space
- Access to HunterLife-exclusive quests

### Extra

- Additional bonus rewards
- Extra carve/gather attempts
- Special decorations and items

### Premium

- All HunterLife + Extra benefits
- Premium-exclusive content
- Enhanced reward rates

### NetCafe

- Boost time periods (see [Gameplay Options](gameplay-options.md))
- Increased reward rates
- NetCafe-exclusive bonuses

## Setting Courses Manually

### Via Database

Directly modify the `users.rights` column:

```sql
-- Enable all courses for a user
UPDATE users SET rights = 2047 WHERE username = 'player';

-- Enable only HunterLife (bit 0 = 2^0 = 1)
UPDATE users SET rights = 1 WHERE username = 'player';

-- Enable HunterLife + Extra + NetCafe (bits 0, 1, 8 = 1 + 2 + 256 = 259)
UPDATE users SET rights = 259 WHERE username = 'player';
```

### Via !rights Command

If the [!rights command](commands.md#rights) is enabled:

```text
!rights 2047
```

Enables all courses (2047 = all 11 bits set).

**Bitfield Calculator:**

- HunterLife: 2^0 = 1
- Extra: 2^1 = 2
- Premium: 2^2 = 4
- NetCafe: 2^8 = 256
- Total: Add the values together

## Implementation Details

Course checking is implemented throughout the codebase:

```go
// Check if player has a specific course
if mhfcourse.CourseExists(courseID, session.courses) {
    // Grant course benefits
}
```

Course data is loaded from the database on character login and cached in the session.

## Related Documentation

- [In-Game Commands](commands.md) - Using `!course` command
- [Gameplay Options](gameplay-options.md) - NetCafe boost time configuration
