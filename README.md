# Erupe Community Edition

Erupe is a community-maintained server emulator for Monster Hunter Frontier written in Go. It is a complete reverse-engineered solution to self-host a Monster Hunter Frontier server, using no code from Capcom.

> [!IMPORTANT]
> The purpose of this branch is to have a clean transition from a functional 9.2.0 release, to a future 9.3.0 version.
> Over the last 2 years after the release of 9.2.0, many commits introduced broken features.

## Features

- **Multi-version Support**: Compatible with all Monster Hunter Frontier versions from Season 6.0 to ZZ
- **Multi-platform**: Supports PC, PlayStation 3, PlayStation Vita, and Wii U (up to Z2)
- **Complete Server Emulation**: Entry/Sign server, Channel server, and Launcher server
- **Gameplay Customization**: Configurable multipliers for experience, currency, and materials
- **Event Systems**: Support for Raviente, MezFes, Diva, Festa, and Tournament events
- **Discord Integration**: Optional real-time Discord bot integration
- **In-game Commands**: Extensible command system with configurable prefixes
- **Developer Tools**: Comprehensive logging, packet debugging, and save data dumps

## Architecture

Erupe consists of three main server components:

- **Sign Server** (Port 53312): Handles authentication and account management
- **Entrance Server** (Port 53310): Manages world/server selection
- **Channel Servers** (Ports 54001+): Handle game sessions, quests, and player interactions

Multiple channel servers can run simultaneously, organized by world types:

- **Newbie**: For new players
- **Normal**: Standard gameplay
- **Cities**: City-focused instances
- **Tavern**: Special tavern area
- **Return**: For returning players
- **MezFes**: Festival events

## Client Compatibility

### Platforms

- PC
- PlayStation 3
- PlayStation Vita
- Wii U (Up to Z2)

### Versions

- **G10-ZZ** (ClientMode): Extensively tested with great functionality
- **G3-Z2** (Wii U): Tested with good functionality
- **Forward.4**: Basic functionality
- **Season 6.0**: Limited functionality (oldest supported version)

If you have an **installed** copy of Monster Hunter Frontier on an old hard drive, **please** get in contact so we can archive it!

## Requirements

- [Go 1.25+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- Monster Hunter Frontier client (see [Client Setup](#client-setup))
- Quest and scenario binary files (see [Resources](#resources))

## Installation

### Quick Start (Pre-compiled Binary)

If you only want to run Erupe, download a [pre-compiled binary](https://github.com/ZeruLight/Erupe/releases/latest):

- `erupe-ce` for Linux
- `erupe.exe` for Windows

Then proceed to [Configuration](#configuration).

### Building from Source

#### First-time Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/ZeruLight/Erupe.git
   cd Erupe
   ```

2. Create a PostgreSQL database and install the base schema:

   ```bash
   # Download and apply the base schema
   wget https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql
   psql -U your_user -d your_database -f SCHEMA.sql
   ```

3. Apply schema patches in order:

   ```bash
   psql -U your_user -d your_database -f patch-schema/01_patch.sql
   # Repeat for each patch file in numerical order
   ```

4. Copy and configure the config file:

   ```bash
   cp config.example.json config.json
   # Edit config.json with your settings (see Configuration section)
   ```

5. Install dependencies and build:

   ```bash
   go mod download
   go build
   ```

6. Run the server:

   ```bash
   ./erupe-ce
   ```

   Or run directly without building:

   ```bash
   go run .
   ```

#### Updating an Existing Installation

1. Pull the latest changes:

   ```bash
   git pull origin main
   ```

2. Update dependencies:

   ```bash
   go mod tidy
   ```

3. Apply any new schema patches from [patch-schema](./patch-schema) that you haven't run yet

4. Rebuild and restart:

   ```bash
   go build
   ./erupe-ce
   ```

### Docker Installation

For quick setup and development (not recommended for production), see [docker/README.md](./docker/README.md).

## Configuration

Erupe is configured via `config.json`. Copy [config.example.json](./config.example.json) to get started.

### Quick Configuration Overview

```json
{
  "Host": "127.0.0.1",           // Server IP (leave empty for auto-detect)
  "BinPath": "bin",              // Path to quest/scenario files
  "Language": "en",              // "en" or "jp"

  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "postgres",
    "Password": "your_password",
    "Database": "erupe"
  },

  "Sign": {"Enabled": true, "Port": 53312},
  "Entrance": {"Enabled": true, "Port": 53310}
}
```

### Detailed Configuration Documentation

For comprehensive configuration details, see the [docs/](./docs/) directory:

- **[Configuration Overview](docs/README.md)** - Start here for quick setup guide
- **[Basic Settings](docs/basic-settings.md)** - Host, language, paths, login notices
- **[Server Configuration](docs/server-configuration.md)** - Sign, Entrance, and Channel servers
- **[Database Setup](docs/database.md)** - PostgreSQL configuration and schema management
- **[Gameplay Options](docs/gameplay-options.md)** - NP/RP caps, boost times, quest allowances
- **[Development Mode](docs/development-mode.md)** - Debug options and testing tools
- **[Logging](docs/logging.md)** - File logging, rotation, and analysis
- **[In-Game Commands](docs/commands.md)** - Chat commands configuration
- **[Courses](docs/courses.md)** - Subscription courses
- **[Discord Integration](docs/discord-integration.md)** - Optional Discord bot setup

**Example configurations:**

- [Local Development](docs/README.md#local-development-server) - Auto-account creation, debug logging
- [Production Server](docs/README.md#production-server-minimal) - Secure production settings
- [Community Server](docs/README.md#community-server) - Feature-rich with Discord integration

## Client Setup

1. Download and install a Monster Hunter Frontier client (version G10 or later recommended)
2. Download [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
3. Extract the binary files to the `bin` directory in your Erupe installation
4. Configure your client to point to your Erupe server IP/hostname
5. Modify the client's `host.txt` or use a launcher to redirect to your server

## Database Schemas

Erupe uses a structured schema system:

- **Initialization Schema**: Bootstraps database to version 9.1.0
- **Update Schemas**: Production-ready updates for new releases
- **Patch Schemas**: Development updates (subject to change)
- **Bundled Schemas**: Demo templates for shops, distributions, events, and gacha in [schemas/bundled-schema/](./schemas/bundled-schema/)

**Note**: Only use patch schemas if you're following active development. They get consolidated into update schemas on release.

For detailed database setup instructions, see [Database Configuration](docs/database.md).

## Development

For development guidelines, code architecture, and contribution instructions, see [CLAUDE.md](CLAUDE.md).

### Running Tests

```bash
# Run all tests
go test -v ./...

# Check for race conditions
go test -v -race ./...
```

### Log Analysis

Erupe includes a log analyzer tool in [tools/loganalyzer/](tools/loganalyzer/) for filtering, analyzing errors, tracking connections, and generating statistics. See [Logging Documentation](docs/logging.md) for details.

## Troubleshooting

### Quick Fixes

| Issue | Solution |
|-------|----------|
| Server won't start | Check PostgreSQL is running and database password is set |
| Client can't connect | Verify ports 53310, 53312, 54001+ are open in firewall |
| Database errors | Apply all schema patches in order (see [Database Setup](docs/database.md)) |
| Quest files not loading | Verify `BinPath` points to extracted binary files |

### Detailed Troubleshooting

For comprehensive troubleshooting guides, see:

- [Database Troubleshooting](docs/database.md#troubleshooting) - Connection, authentication, and schema issues
- [Server Configuration](docs/server-configuration.md#troubleshooting) - Port conflicts, connectivity issues
- [Development Mode](docs/development-mode.md#packet-logging) - Enable debug logging for packet analysis

## Documentation

- **[Configuration Documentation](docs/README.md)** - Complete configuration guide
- **[CLAUDE.md](CLAUDE.md)** - Development guide and architecture overview
- **[config.example.json](config.example.json)** - Full configuration example

## Resources

- **Binary Files**: [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- **Discord Communities**:
  - [Mezeporta Square Discord](https://discord.gg/DnwcpXM488)
  - [Mogapedia's Discord](https://discord.gg/f77VwBX5w7) - Maintainers of this branch
  - [PewPewDojo Discord](https://discord.gg/CFnzbhQ) - General community
- **Additional Resources**:
  - [Erupe Wiki](https://github.com/Mezeporta/Erupe/wiki)
  - [Community FAQ](https://pastebin.com/QqAwZSTC)

## Changelog

View [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Authors

A list of contributors can be found at AUTHORS.md (if available) or in the git commit history.
