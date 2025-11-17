# Gameplay Options

Gameplay modifiers and balance settings for Erupe.

## Configuration

```json
{
  "GameplayOptions": {
    "FeaturedWeapons": 1,
    "MaximumNP": 100000,
    "MaximumRP": 50000,
    "DisableLoginBoost": false,
    "DisableBoostTime": false,
    "BoostTimeDuration": 120,
    "GuildMealDuration": 60,
    "BonusQuestAllowance": 3,
    "DailyQuestAllowance": 1
  }
}
```

## Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `FeaturedWeapons` | number | `1` | Number of Active Feature weapons generated daily |
| `MaximumNP` | number | `100000` | Maximum Network Points (NP) a player can hold |
| `MaximumRP` | number | `50000` | Maximum Road Points (RP) a player can hold |
| `DisableLoginBoost` | boolean | `false` | Disable login boost system entirely |
| `DisableBoostTime` | boolean | `false` | Disable daily NetCafe boost time |
| `BoostTimeDuration` | number | `120` | NetCafe boost time duration in minutes |
| `GuildMealDuration` | number | `60` | Guild meal activation duration in minutes |
| `BonusQuestAllowance` | number | `3` | Daily Bonus Point Quest allowance |
| `DailyQuestAllowance` | number | `1` | Daily Quest allowance |

## Detailed Explanations

### Featured Weapons

Featured/Active Feature weapons are special weapon variants with unique properties. This setting controls how many are generated and available each day.

- Set to `0` to disable featured weapons
- Set to `1`-`3` for normal operation
- Higher values generate more variety

### Network Points (NP) and Road Points (RP)

NP and RP are in-game currencies/points used for various purchases and progression:

- **Network Points (NP)**: Used for purchasing items, materials, and services
- **Road Points (RP)**: Used for unlocking road/progression rewards

**Default Caps:**

- NP: `100,000`
- RP: `50,000`

You can increase these caps for more relaxed gameplay or decrease them to maintain balance.

### Boost Systems

Monster Hunter Frontier has several boost systems that increase rewards and experience:

#### Login Boost

Automatically granted when logging in. Disable with `DisableLoginBoost: true`.

#### NetCafe Boost Time

Daily time-limited boost that simulates NetCafe benefits:

```json
{
  "DisableBoostTime": false,
  "BoostTimeDuration": 120
}
```

- `DisableBoostTime: false` - Boost time is active
- `BoostTimeDuration: 120` - Lasts 120 minutes (2 hours)

### Guild Meals

Guild meals are buffs that guild members can activate:

```json
{
  "GuildMealDuration": 60
}
```

Duration in minutes after cooking before the meal expires.

### Quest Allowances

Daily limits for special quest types:

- **BonusQuestAllowance**: Number of Bonus Point Quests per day
- **DailyQuestAllowance**: Number of Daily Quests per day

Set to `0` to disable limits entirely.

## Examples

### Casual/Relaxed Server

```json
{
  "GameplayOptions": {
    "FeaturedWeapons": 3,
    "MaximumNP": 999999,
    "MaximumRP": 999999,
    "DisableLoginBoost": false,
    "DisableBoostTime": false,
    "BoostTimeDuration": 240,
    "GuildMealDuration": 120,
    "BonusQuestAllowance": 10,
    "DailyQuestAllowance": 5
  }
}
```

### Balanced Server (Default)

```json
{
  "GameplayOptions": {
    "FeaturedWeapons": 1,
    "MaximumNP": 100000,
    "MaximumRP": 50000,
    "DisableLoginBoost": false,
    "DisableBoostTime": false,
    "BoostTimeDuration": 120,
    "GuildMealDuration": 60,
    "BonusQuestAllowance": 3,
    "DailyQuestAllowance": 1
  }
}
```

### Hardcore/Challenge Server

```json
{
  "GameplayOptions": {
    "FeaturedWeapons": 0,
    "MaximumNP": 50000,
    "MaximumRP": 25000,
    "DisableLoginBoost": true,
    "DisableBoostTime": true,
    "BonusQuestAllowance": 1,
    "DailyQuestAllowance": 1
  }
}
```

## Related Documentation

- [Courses](courses.md) - Subscription course configuration
- [In-Game Commands](commands.md) - Player commands
