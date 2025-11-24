# usercheck

CLI tool to query connected users and server status from the Erupe database.

## Build

```bash
go build -o usercheck
```

## Usage

```bash
./usercheck <command> [options]
```

By default, the tool reads database credentials from `config.json` in the project root.

### Commands

| Command | Description |
|---------|-------------|
| `list` | List all currently connected users |
| `count` | Show count of connected users per server |
| `search` | Search for a connected user by name |
| `servers` | Show server/channel status |
| `history` | Show login history for a player |

### Examples

```bash
# List connected users (uses config.json)
./usercheck list

# Verbose output with last login times
./usercheck list -v

# Search for a player
./usercheck search -name "Hunter"

# Show server status
./usercheck servers

# Player login history
./usercheck history -name "Hunter"

# Use a specific config file
./usercheck list -config /path/to/config.json

# Override database password
./usercheck list -password "different_password"
```

### Configuration Priority

1. CLI flags (highest priority)
2. Environment variables (`ERUPE_DB_*`)
3. `config.json` file
4. Default values (lowest priority)

### Database Flags

| Flag | Env Variable | Description |
|------|--------------|-------------|
| `-config` | - | Path to config.json |
| `-host` | `ERUPE_DB_HOST` | Database host |
| `-port` | - | Database port |
| `-user` | `ERUPE_DB_USER` | Database user |
| `-password` | `ERUPE_DB_PASSWORD` | Database password |
| `-dbname` | `ERUPE_DB_NAME` | Database name |
