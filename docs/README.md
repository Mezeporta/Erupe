# Erupe Configuration Documentation

Comprehensive configuration documentation for Erupe, the Monster Hunter Frontier server emulator.

## Quick Start

1. Copy [config.example.json](../config.example.json) to `config.json`
2. Read [Basic Settings](basic-settings.md) and [Server Configuration](server-configuration.md)
3. Set up [Database](database.md)
4. Adjust [Gameplay Options](gameplay-options.md) to your preference
5. Start the server!

## Documentation Index

### Essential Configuration

- **[Basic Settings](basic-settings.md)** - Host, language, and basic server options
- **[Server Configuration](server-configuration.md)** - Sign, Entrance, and Channel server setup
- **[Database](database.md)** - PostgreSQL configuration and schema management

### Development

- **[Development Mode](development-mode.md)** - Debug options, packet logging, and testing tools
- **[Logging](logging.md)** - File logging, rotation, and log analysis

### Gameplay & Features

- **[Gameplay Options](gameplay-options.md)** - NP/RP caps, boost times, quest allowances
- **[Courses](courses.md)** - Subscription course configuration
- **[In-Game Commands](commands.md)** - Chat commands for players and admins

### Optional Features

- **[Discord Integration](discord-integration.md)** - Real-time Discord bot for server activity

## Configuration File Structure

The main configuration file is `config.json` with this structure:

```json
{
  "Host": "127.0.0.1",
  "BinPath": "bin",
  "Language": "en",

  "DevMode": false,
  "DevModeOptions": { ... },

  "GameplayOptions": { ... },
  "Logging": { ... },
  "Discord": { ... },
  "Commands": [ ... ],
  "Courses": [ ... ],

  "Database": { ... },

  "Sign": { ... },
  "SignV2": { ... },
  "Channel": { ... },
  "Entrance": { ... }
}
```

## Configuration Sections

| Section | Documentation | Purpose |
|---------|---------------|---------|
| Basic Settings | [basic-settings.md](basic-settings.md) | Host, paths, language, login notices |
| DevMode | [development-mode.md](development-mode.md) | Development and debugging options |
| GameplayOptions | [gameplay-options.md](gameplay-options.md) | Gameplay balance and modifiers |
| Logging | [logging.md](logging.md) | File logging and rotation |
| Discord | [discord-integration.md](discord-integration.md) | Discord bot integration |
| Commands | [commands.md](commands.md) | In-game chat commands |
| Courses | [courses.md](courses.md) | Subscription courses |
| Database | [database.md](database.md) | PostgreSQL connection |
| Sign/SignV2 | [server-configuration.md](server-configuration.md#sign-server) | Authentication servers |
| Channel | [server-configuration.md](server-configuration.md#channel-server) | Gameplay server |
| Entrance | [server-configuration.md](server-configuration.md#entrance-server) | World list server |

## Common Configuration Scenarios

### Local Development Server

Perfect for testing and development:

```json
{
  "Host": "127.0.0.1",
  "DevMode": true,
  "DevModeOptions": {
    "AutoCreateAccount": true,
    "MaxLauncherHR": true
  },
  "Database": {
    "Host": "localhost",
    "User": "postgres",
    "Password": "dev",
    "Database": "erupe_dev"
  }
}
```

See: [development-mode.md](development-mode.md)

### Production Server (Minimal)

Minimal production-ready configuration:

```json
{
  "Host": "",
  "DevMode": false,
  "DisableSoftCrash": true,
  "Logging": {
    "LogToFile": true,
    "LogMaxBackups": 7,
    "LogMaxAge": 30
  },
  "Database": {
    "Host": "localhost",
    "User": "erupe",
    "Password": "SECURE_PASSWORD_HERE",
    "Database": "erupe"
  }
}
```

See: [basic-settings.md](basic-settings.md), [logging.md](logging.md)

### Community Server

Feature-rich community server:

```json
{
  "Host": "",
  "DevMode": false,
  "HideLoginNotice": false,
  "LoginNotices": ["Welcome to our server!"],
  "GameplayOptions": {
    "MaximumNP": 999999,
    "BoostTimeDuration": 240
  },
  "Discord": {
    "Enabled": true,
    "BotToken": "YOUR_TOKEN",
    "RealtimeChannelID": "YOUR_CHANNEL_ID"
  },
  "Commands": [
    {"Name": "Reload", "Enabled": true, "Prefix": "!reload"},
    {"Name": "Course", "Enabled": true, "Prefix": "!course"}
  ]
}
```

See: [gameplay-options.md](gameplay-options.md), [discord-integration.md](discord-integration.md), [commands.md](commands.md)

## Security Checklist

Before running a public server, verify:

- [ ] `DevMode: false` - Disable development mode
- [ ] `AutoCreateAccount: false` - Require manual account creation
- [ ] `DisableTokenCheck: false` - Enable token validation
- [ ] `CleanDB: false` - Don't wipe database on startup
- [ ] Strong database password set
- [ ] `Rights` command disabled (or carefully controlled)
- [ ] `KeyQuest` command disabled (unless intentional)
- [ ] Firewall configured for only necessary ports
- [ ] Database not exposed publicly
- [ ] Logging enabled for monitoring

See: [development-mode.md#security-warnings](development-mode.md#security-warnings)

## Performance Tuning

For large servers:

1. **Increase player limits**: Adjust `MaxPlayers` in channel configuration
2. **Add more channels**: Distribute load across multiple channel servers
3. **Optimize database**: Use connection pooling, increase shared buffers
4. **Increase log rotation**: Larger `LogMaxSize` and `LogMaxBackups`
5. **Monitor resources**: Use log analyzer to track errors and performance

See: [server-configuration.md](server-configuration.md), [database.md#performance-tuning](database.md#performance-tuning)

## Configuration Validation

Erupe validates configuration on startup. Common errors:

| Error | Cause | Fix |
|-------|-------|-----|
| "Database password is blank" | Empty password field | Set a password in config |
| "Invalid host address" | Malformed Host value | Use valid IP or leave empty for auto-detect |
| "Discord failed" | Invalid Discord config | Check bot token and channel ID |
| Port already in use | Port conflict | Change port or stop conflicting service |

## Environment-Specific Configuration

### Docker

When running in Docker, use service names for hosts:

```json
{
  "Database": {
    "Host": "db",
    "Port": 5432
  }
}
```

See: [database.md#docker-setup](database.md#docker-setup)

### Cloud Hosting

For cloud deployments:

- Use environment variables for secrets (requires code modification)
- Enable `DisableSoftCrash: true` for auto-restart
- Use absolute paths for logs (`/var/log/erupe/erupe.log`)
- Consider external database (RDS, Cloud SQL)

## Additional Resources

- [CLAUDE.md](../CLAUDE.md) - Development guide and architecture
- [config.example.json](../config.example.json) - Full example configuration
- [Log Analyzer](../tools/loganalyzer/) - Log analysis tools
- [GitHub Issues](https://github.com/Andoryuuta/Erupe/issues) - Report bugs and request features

## Getting Help

If you need help with configuration:

1. Check the relevant documentation page above
2. Review [config.example.json](../config.example.json) for examples
3. Check server logs for specific errors
4. Search [GitHub Issues](https://github.com/Andoryuuta/Erupe/issues)
5. Ask in the Erupe Discord community

## Contributing

Found an error or want to improve these docs?

1. Fork the repository
2. Edit the documentation in `docs/`
3. Submit a pull request

See [CLAUDE.md](../CLAUDE.md#commit-and-pr-guidelines) for contribution guidelines.
