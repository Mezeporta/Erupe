# Erupe

Erupe is a community-maintained server emulator for Monster Hunter Frontier written in Go. It is a complete reverse-engineered solution to self-host a Monster Hunter Frontier server, using no code from Capcom.

## About This Fork

This fork is maintained by the Mogapedia community, focusing on stability, documentation, and quality improvements to Erupe. Our goal is to provide a reliable, well-documented server platform for Monster Hunter Frontier enthusiasts.

### Branch Strategy

- **main**: Active development branch with the latest features and improvements
- **9.2.0-clean**: Stable release branch providing a clean, functional 9.2.0 baseline for those seeking stability over cutting-edge features

The 9.2.0-clean branch was created to address stability concerns after the 9.2.0 release, providing a solid foundation for future development while the main branch continues to evolve.

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
   psql -U your_user -d your_database -f schemas/patch-schema/01_patch.sql
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

3. Apply any new schema patches from [schemas/patch-schema](./schemas/patch-schema) that you haven't run yet

4. Rebuild and restart:

   ```bash
   go build
   ./erupe-ce
   ```

### Docker Installation

For quick setup and development (not recommended for production), see [docker/README.md](./docker/README.md).

## Configuration

Edit `config.json` to configure your server. Key settings include:

### Core Settings

```json
{
  "Host": "127.0.0.1",           // Server binding address
  "BinPath": "bin",              // Path to quest/scenario binaries
  "Language": "en",              // "en" or "jp"
  "ClientMode": "ZZ"             // Target client version
}
```

### Database

```json
{
  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "postgres",
    "Password": "your_password",
    "Database": "erupe"
  }
}
```

### Server Ports

```json
{
  "Sign": {
    "Enabled": true,
    "Port": 53312              // Authentication server
  },
  "Entrance": {
    "Enabled": true,
    "Port": 53310              // World selection server
  }
}
```

Channel servers are configured under `Entrance.Entries[].Channels[]` with individual ports (default: 54001+).

### Development Options

```json
{
  "DebugOptions": {
    "LogInboundMessages": false,   // Log incoming packets
    "LogOutboundMessages": false,  // Log outgoing packets
    "MaxHexdumpLength": 256       // Max bytes for hexdump logs
  }
}
```

### Gameplay Options

```json
{
  "GameplayOptions": {
    "MaximumNP": 100000,           // Max Netcafe Points
    "MaximumRP": 50000,            // Max Road Points
    "BoostTimeDuration": 7200,     // Login boost duration (seconds)
    "BonusQuestAllowance": 3,      // Daily bonus quests
    "DailyQuestAllowance": 1       // Daily quest limit
  }
}
```

### In-game Commands

Configure available commands and their prefixes:

```json
{
  "CommandPrefix": "!",
  "Commands": [
    {"Name": "Raviente", "Enabled": true, "Prefix": "ravi"},
    {"Name": "Reload", "Enabled": true, "Prefix": "reload"},
    {"Name": "Course", "Enabled": true, "Prefix": "course"}
  ]
}
```

For a complete configuration example, see [config.example.json](./config.example.json).

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

## Development

### Running Tests

```bash
# Run all tests
go test -v ./...

# Check for race conditions
go test -v -race ./...
```

## Troubleshooting

### Common Issues

#### Server won't start

- Verify PostgreSQL is running: `systemctl status postgresql` (Linux) or `pg_ctl status` (Windows)
- Check database credentials in `config.json`
- Ensure all required ports are available and not blocked by firewall

#### Client can't connect

- Verify server is listening: `netstat -an | grep 53310`
- Check firewall rules allow traffic on ports 53310, 53312, and 54001+
- Ensure client's `host.txt` points to correct server IP
- For remote connections, set `"Host"` in config.json to `0.0.0.0` or your server's IP

#### Database schema errors

- Ensure all patch files are applied in order
- Check PostgreSQL logs for detailed error messages
- Verify database user has sufficient privileges

#### Quest files not loading

- Confirm `BinPath` in config.json points to extracted quest/scenario files
- Verify binary files match your `ClientMode` setting
- Check file permissions

### Debug Logging

Enable detailed logging in `config.json`:

```json
{
  "DebugOptions": {
    "LogInboundMessages": true,
    "LogOutboundMessages": true
  }
}
```

## Resources

- **Binary Files**: [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- **Discord Communities**:
  - [Mezeporta Square Discord](https://discord.gg/DnwcpXM488)
  - [Mogapedia's Discord](https://discord.gg/f77VwBX5w7)
  - [PewPewDojo Discord](https://discord.gg/CFnzbhQ)
- **Documentation**: [Erupe Wiki](https://github.com/Mezeporta/Erupe/wiki)
- **FAQ**: [Community FAQ Pastebin](https://pastebin.com/QqAwZSTC)

## Changelog

View [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Authors

A list of authors can be found at [AUTHORS.md](AUTHORS.md).
