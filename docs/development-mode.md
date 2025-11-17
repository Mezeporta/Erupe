# Development Mode

Development mode configuration for testing and debugging Erupe.

## Configuration

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

## Settings Reference

### DevMode

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `DevMode` | boolean | `false` | Enables development mode (more verbose logging, development logger format) |

When `DevMode` is enabled:

- Logging uses console format (human-readable) instead of JSON
- More detailed stack traces on errors
- Development-friendly output

### DevModeOptions

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `AutoCreateAccount` | boolean | `false` | **⚠️ SECURITY RISK**: Auto-create accounts on login (disable in production) |
| `CleanDB` | boolean | `false` | **⚠️ DESTRUCTIVE**: Wipes database on server start (deletes all users, characters, guilds) |
| `MaxLauncherHR` | boolean | `false` | Sets launcher HR to HR7 to join non-beginner worlds |
| `LogInboundMessages` | boolean | `false` | Log all packets received from clients (very verbose) |
| `LogOutboundMessages` | boolean | `false` | Log all packets sent to clients (very verbose) |
| `MaxHexdumpLength` | number | `256` | Maximum bytes to display in packet hexdumps |
| `DivaEvent` | number | `0` | Diva Defense event status (0 = off, higher = active) |
| `FestaEvent` | number | `-1` | Hunter's Festa event status (-1 = off, higher = active) |
| `TournamentEvent` | number | `0` | VS Tournament event status (0 = off, higher = active) |
| `MezFesEvent` | boolean | `false` | Enable/disable MezFes event |
| `MezFesAlt` | boolean | `false` | Swap Volpakkun for Tokotoko in MezFes |
| `DisableTokenCheck` | boolean | `false` | **⚠️ SECURITY RISK**: Skip login token validation |
| `QuestDebugTools` | boolean | `false` | Enable quest debugging logs |

### Save Dumps

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `SaveDumps.Enabled` | boolean | `false` | Enable saving character data dumps for analysis |
| `SaveDumps.OutputDir` | string | `"savedata"` | Directory for save data dumps |

## Security Warnings

### AutoCreateAccount

**Never enable in production!** This setting allows anyone to create an account by simply trying to log in with any username. This is convenient for development but a major security risk for public servers.

### CleanDB

**Extremely destructive!** This setting wipes all user data from the database on every server restart. Only use for rapid testing cycles in isolated development environments.

### DisableTokenCheck

**Security vulnerability!** Bypasses login token validation, allowing unauthorized access. Only use in isolated development environments.

## Packet Logging

When debugging network issues, enable packet logging:

```json
{
  "DevModeOptions": {
    "LogInboundMessages": true,
    "LogOutboundMessages": true,
    "MaxHexdumpLength": 512
  }
}
```

**Warning:** This generates **massive** log files very quickly. Only enable when actively debugging specific packet issues.

## Event Testing

Test special events by enabling them:

```json
{
  "DevModeOptions": {
    "DivaEvent": 1,
    "FestaEvent": 0,
    "TournamentEvent": 1,
    "MezFesEvent": true,
    "MezFesAlt": false
  }
}
```

## Examples

### Safe Development Configuration

```json
{
  "DevMode": true,
  "DevModeOptions": {
    "AutoCreateAccount": true,
    "CleanDB": false,
    "MaxLauncherHR": true,
    "QuestDebugTools": true,
    "SaveDumps": {
      "Enabled": true,
      "OutputDir": "savedata"
    }
  }
}
```

### Production Configuration

```json
{
  "DevMode": false,
  "DevModeOptions": {
    "AutoCreateAccount": false,
    "CleanDB": false,
    "DisableTokenCheck": false
  }
}
```

### Packet Debugging Configuration

```json
{
  "DevMode": true,
  "DevModeOptions": {
    "LogInboundMessages": true,
    "LogOutboundMessages": true,
    "MaxHexdumpLength": 1024
  }
}
```

## Related Documentation

- [Logging](logging.md) - Logging configuration
- [Basic Settings](basic-settings.md) - Basic server settings
