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
# List connected users
./usercheck list -password "dbpass"

# Verbose output with last login times
./usercheck list -v -password "dbpass"

# Search for a player
./usercheck search -name "Hunter" -password "dbpass"

# Show server status
./usercheck servers -password "dbpass"

# Player login history
./usercheck history -name "Hunter" -password "dbpass"
```

### Database Options

| Flag | Env Variable | Default |
|------|--------------|---------|
| `-host` | `ERUPE_DB_HOST` | localhost |
| `-port` | - | 5432 |
| `-user` | `ERUPE_DB_USER` | postgres |
| `-password` | `ERUPE_DB_PASSWORD` | (required) |
| `-dbname` | `ERUPE_DB_NAME` | erupe |
