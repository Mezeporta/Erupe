# Erupe

[![Build and Test](https://github.com/Mezeporta/Erupe/actions/workflows/go-improved.yml/badge.svg)](https://github.com/Mezeporta/Erupe/actions/workflows/go-improved.yml)
[![CodeQL](https://github.com/Mezeporta/Erupe/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/Mezeporta/Erupe/actions/workflows/github-code-scanning/codeql)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Mezeporta/Erupe)](https://go.dev/)
[![Latest Release](https://img.shields.io/github/v/release/Mezeporta/Erupe)](https://github.com/Mezeporta/Erupe/releases/latest)

Erupe is a community-maintained server emulator for Monster Hunter Frontier written in Go. It is a complete reverse-engineered solution to self-host a Monster Hunter Frontier server, using no code from Capcom.

## Quick Start

Pick one of three installation methods, then continue to [Quest & Scenario Files](#quest--scenario-files).

### Option A: Docker (recommended)

Docker handles the database automatically. You only need to provide quest files and a config.

1. Clone the repository and enter the Docker directory:

   ```bash
   git clone https://github.com/Mezeporta/Erupe.git
   cd Erupe
   ```

2. Copy and edit the config (set your database password to match `docker-compose.yml`):

   ```bash
   cp config.example.json docker/config.json
   # Edit docker/config.json — set Database.Host to "db"
   ```

3. Download [quest/scenario files](#quest--scenario-files) and extract them to `docker/bin/`

4. Start everything:

   ```bash
   cd docker
   docker compose up
   ```

   pgAdmin is available at `http://localhost:5050` for database management.

   See [docker/README.md](./docker/README.md) for more details (local builds, troubleshooting).

### Option B: Pre-compiled Binary

1. Download the latest release for your platform from [GitHub Releases](https://github.com/Mezeporta/Erupe/releases/latest):
   - `erupe-ce` for Linux
   - `erupe.exe` for Windows

2. Set up PostgreSQL and create a database:

   ```bash
   wget https://github.com/Mezeporta/Erupe/releases/latest/download/SCHEMA.sql
   psql -U postgres -d erupe -f SCHEMA.sql
   ```

3. Apply any patch schemas from [schemas/patch-schema/](./schemas/patch-schema/) in numerical order:

   ```bash
   psql -U postgres -d erupe -f schemas/patch-schema/01_patch.sql
   # Repeat for each patch file
   ```

4. Copy and edit the config:

   ```bash
   cp config.example.json config.json
   # Edit config.json with your database credentials
   ```

5. Download [quest/scenario files](#quest--scenario-files) and extract them to `bin/`

6. Run: `./erupe-ce`

### Option C: From Source

Requires [Go 1.25+](https://go.dev/dl/) and [PostgreSQL](https://www.postgresql.org/download/).

1. Clone and build:

   ```bash
   git clone https://github.com/Mezeporta/Erupe.git
   cd Erupe
   go mod download
   go build -o erupe-ce
   ```

2. Set up the database (same as Option B, steps 2–3)

3. Copy and edit the config:

   ```bash
   cp config.example.json config.json
   ```

4. Download [quest/scenario files](#quest--scenario-files) and extract them to `bin/`

5. Run: `./erupe-ce`

## Quest & Scenario Files

**Download**: [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)

These files contain quest definitions and scenario data that the server sends to clients during gameplay. Extract the archive into your `bin/` directory (or `docker/bin/` for Docker installs). The path must match the `BinPath` setting in your config (default: `"bin"`).

**Without these files, quests will not load and the client will crash.**

## Client Setup

1. Obtain a Monster Hunter Frontier client (version G10 or later recommended)
2. Point the client to your server by editing `host.txt` or using a launcher to redirect to your server's IP
3. Launch `mhf.exe`, select your server, and create an account

If you have an **installed** copy of Monster Hunter Frontier on an old hard drive, **please** get in contact so we can archive it!

## Updating

### From Source

```bash
git pull origin main
go mod tidy
go build -o erupe-ce
```

**Check for new patch schemas** in [schemas/patch-schema/](./schemas/patch-schema/) after pulling — apply any you haven't run yet, in numerical order.

### Docker

```bash
cd docker
docker compose down
docker compose build
docker compose up
```

Apply any new patch schemas via pgAdmin or `psql` into the running container.

## Configuration

Edit `config.json` before starting the server. The essential settings are:

```json
{
  "Host": "127.0.0.1",
  "BinPath": "bin",
  "Language": "en",
  "ClientMode": "ZZ",
  "Database": {
    "Host": "localhost",
    "Port": 5432,
    "User": "postgres",
    "Password": "your_password",
    "Database": "erupe"
  }
}
```

| Setting | Description |
|---------|-------------|
| `Host` | Bind address. Use `127.0.0.1` for local, `0.0.0.0` for remote access |
| `ClientMode` | Target client version (`ZZ`, `G10`, `Forward4`, etc.) |
| `BinPath` | Path to quest/scenario files |
| `Language` | `"en"` or `"jp"` |

For the full configuration reference (gameplay multipliers, debug options, Discord integration, in-game commands, entrance/channel definitions), see [config.example.json](./config.example.json) and the [Erupe Wiki](https://github.com/Mezeporta/Erupe/wiki).

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

Multiple channel servers can run simultaneously, organized by world types: Newbie, Normal, Cities, Tavern, Return, and MezFes.

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

## Database Schemas

Erupe uses a structured schema system:

- **Initialization Schema**: Bootstraps database to version 9.1.0
- **Update Schemas**: Production-ready updates for new releases
- **Patch Schemas**: Development updates (subject to change)
- **Bundled Schemas**: Demo templates for shops, distributions, events, and gacha in [schemas/bundled-schema/](./schemas/bundled-schema/)

**Note**: Only use patch schemas if you're following active development. They get consolidated into update schemas on release.

## Development

### Branch Strategy

- **main**: Active development branch with the latest features and improvements
- **stable/v9.2.x**: Stable release branch for those seeking stability over cutting-edge features

### Running Tests

```bash
go test -v ./...           # Run all tests
go test -v -race ./...     # Check for race conditions (mandatory before merging)
```

## Troubleshooting

### Server won't start

- Verify PostgreSQL is running: `systemctl status postgresql` (Linux) or `pg_ctl status` (Windows)
- Check database credentials in `config.json`
- Ensure all required ports are available and not blocked by firewall

### Client can't connect

- Verify server is listening: `netstat -an | grep 53310`
- Check firewall rules allow traffic on ports 53310, 53312, and 54001+
- Ensure client's `host.txt` points to correct server IP
- For remote connections, set `"Host"` in config.json to `0.0.0.0` or your server's IP

### Database schema errors

- Ensure all patch files are applied in order
- Check PostgreSQL logs for detailed error messages
- Verify database user has sufficient privileges

### Quest files not loading

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

- **Quest/Scenario Files**: [Download (catbox)](https://files.catbox.moe/xf0l7w.7z)
- **Documentation**: [Erupe Wiki](https://github.com/Mezeporta/Erupe/wiki)
- **Discord Communities**:
  - [Mezeporta Square](https://discord.gg/DnwcpXM488)
  - [Mogapedia](https://discord.gg/f77VwBX5w7) (French Monster Hunter community, current Erupe maintainers)
  - [PewPewDojo](https://discord.gg/CFnzbhQ)
- **Community Tools**:
  - [Ferias](https://xl3lackout.github.io/MHFZ-Ferias-English-Project/) — Material and item database
  - [Damage Calculator](https://mh.fist.moe/damagecalc.html) — Online damage calculator
  - [Armor Set Searcher](https://github.com/matthe815/mhfz-ass/releases) — Armor set search application

## Changelog

View [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Authors

A list of authors can be found at [AUTHORS.md](AUTHORS.md).
