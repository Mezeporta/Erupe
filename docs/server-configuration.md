# Server Configuration

Configuration for Erupe's three-server architecture: Sign, Entrance, and Channel servers.

## Three-Server Architecture

Erupe uses a multi-server architecture that mirrors the original Monster Hunter Frontier server design:

```
Client → Sign Server (Auth) → Entrance Server (World List) → Channel Server (Gameplay)
```

1. **Sign Server**: Authentication and account management
2. **Entrance Server**: World/server selection and character list
3. **Channel Servers**: Actual gameplay sessions, quests, and player interactions

## Sign Server

Handles authentication and account management.

### Configuration

```json
{
  "Sign": {
    "Enabled": true,
    "Port": 53312
  }
}
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `Enabled` | boolean | `true` | Enable the legacy sign server |
| `Port` | number | `53312` | Port for sign server (client default: 53312) |

### Details

- Located in [server/signserver/](../server/signserver/)
- Creates sign sessions with tokens for channel server authentication
- Legacy TCP-based protocol
- Required unless using SignV2

## SignV2 Server

Modern HTTP-based sign server (alternative to legacy sign server).

### Configuration

```json
{
  "SignV2": {
    "Enabled": false,
    "Port": 8080
  }
}
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `Enabled` | boolean | `false` | Enable the modern HTTP-based sign server |
| `Port` | number | `8080` | Port for SignV2 server |

### Details

- Located in [server/signv2server/](../server/signv2server/)
- HTTP-based authentication (easier to proxy/load balance)
- Alternative to legacy sign server
- **Only enable one sign server at a time** (Sign OR SignV2, not both)

## Channel Server

Handles actual gameplay sessions, quests, and player interactions.

### Configuration

```json
{
  "Channel": {
    "Enabled": true
  }
}
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `Enabled` | boolean | `true` | Enable channel servers (required for gameplay) |

### Details

- Located in [server/channelserver/](../server/channelserver/)
- Most complex component - handles all gameplay logic
- Multiple instances can run simultaneously
- Ports configured in Entrance server entries
- Features:
  - Session management
  - Packet handling
  - Stage/room system
  - Quest system
  - Guild operations
  - Special events (Raviente, Diva Defense, etc.)

See [CLAUDE.md](../CLAUDE.md#channel-server-internal-architecture) for detailed architecture.

## Entrance Server

Manages world/server selection and character lists.

### Configuration

```json
{
  "Entrance": {
    "Enabled": true,
    "Port": 53310,
    "Entries": [
      {
        "Name": "Newbie",
        "Description": "",
        "IP": "",
        "Type": 3,
        "Recommended": 2,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54001, "MaxPlayers": 100 },
          { "Port": 54002, "MaxPlayers": 100 }
        ]
      }
    ]
  }
}
```

### Settings Reference

| Setting | Type | Description |
|---------|------|-------------|
| `Enabled` | boolean | Enable entrance server (required for server list) |
| `Port` | number | Entrance server port (default: 53310) |
| `Entries` | array | List of worlds/servers shown to players |

### Entrance Entries

Each entry represents a "world" in the server list.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | World name displayed to players (e.g., "Newbie", "Normal") |
| `Description` | string | World description (optional, usually empty) |
| `IP` | string | Override IP address (leave empty to use global `Host` setting) |
| `Type` | number | World type (see below) |
| `Recommended` | number | Recommendation badge: `0`=None, `2`=Recommended, `6`=Special |
| `AllowedClientFlags` | number | Client version flags (0 = all versions allowed) |
| `Channels` | array | List of channel servers in this world |

### World Types

| Type | Name | Purpose |
|------|------|---------|
| `1` | Normal | Standard gameplay world |
| `2` | Cities | Social/town areas |
| `3` | Newbie | For new players (typically recommended) |
| `4` | Tavern | Bar/tavern areas |
| `5` | Return | For returning players |
| `6` | MezFes | MezFes event world |

### Channel Configuration

Each world has multiple channels (like "servers" within a "world"):

```json
{
  "Channels": [
    { "Port": 54001, "MaxPlayers": 100 },
    { "Port": 54002, "MaxPlayers": 100 }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Port` | number | Channel server port (must be unique across all channels) |
| `MaxPlayers` | number | Maximum players allowed in this channel |
| `CurrentPlayers` | number | Current player count (auto-updated at runtime) |

## Complete Server Configurations

### Minimal Setup (Single World, Single Channel)

```json
{
  "Sign": {
    "Enabled": true,
    "Port": 53312
  },
  "SignV2": {
    "Enabled": false
  },
  "Channel": {
    "Enabled": true
  },
  "Entrance": {
    "Enabled": true,
    "Port": 53310,
    "Entries": [
      {
        "Name": "Main",
        "Type": 1,
        "Recommended": 0,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54001, "MaxPlayers": 100 }
        ]
      }
    ]
  }
}
```

### Standard Setup (Multiple Worlds)

```json
{
  "Sign": {
    "Enabled": true,
    "Port": 53312
  },
  "Channel": {
    "Enabled": true
  },
  "Entrance": {
    "Enabled": true,
    "Port": 53310,
    "Entries": [
      {
        "Name": "Newbie",
        "Description": "",
        "IP": "",
        "Type": 3,
        "Recommended": 2,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54001, "MaxPlayers": 100 },
          { "Port": 54002, "MaxPlayers": 100 }
        ]
      },
      {
        "Name": "Normal",
        "Description": "",
        "IP": "",
        "Type": 1,
        "Recommended": 0,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54003, "MaxPlayers": 100 },
          { "Port": 54004, "MaxPlayers": 100 }
        ]
      },
      {
        "Name": "Cities",
        "Description": "",
        "IP": "",
        "Type": 2,
        "Recommended": 0,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54005, "MaxPlayers": 100 }
        ]
      }
    ]
  }
}
```

### Large-Scale Setup

```json
{
  "Sign": {
    "Enabled": true,
    "Port": 53312
  },
  "Channel": {
    "Enabled": true
  },
  "Entrance": {
    "Enabled": true,
    "Port": 53310,
    "Entries": [
      {
        "Name": "Newbie",
        "Type": 3,
        "Recommended": 2,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54001, "MaxPlayers": 150 },
          { "Port": 54002, "MaxPlayers": 150 },
          { "Port": 54003, "MaxPlayers": 150 }
        ]
      },
      {
        "Name": "Normal",
        "Type": 1,
        "Recommended": 0,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54004, "MaxPlayers": 200 },
          { "Port": 54005, "MaxPlayers": 200 },
          { "Port": 54006, "MaxPlayers": 200 },
          { "Port": 54007, "MaxPlayers": 200 }
        ]
      },
      {
        "Name": "Cities",
        "Type": 2,
        "Recommended": 0,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54008, "MaxPlayers": 100 },
          { "Port": 54009, "MaxPlayers": 100 }
        ]
      },
      {
        "Name": "MezFes",
        "Type": 6,
        "Recommended": 6,
        "AllowedClientFlags": 0,
        "Channels": [
          { "Port": 54010, "MaxPlayers": 100 }
        ]
      }
    ]
  }
}
```

## Port Allocation

Default port assignments:

| Server | Port | Configurable |
|--------|------|--------------|
| Sign | 53312 | Yes |
| SignV2 | 8080 | Yes |
| Entrance | 53310 | Yes |
| Channels | 54001+ | Yes (per-channel) |

**Important:**

- All ports must be unique
- Firewall must allow inbound connections on these ports
- Client expects Sign on 53312 and Entrance on 53310 by default

## Firewall Configuration

### Linux (ufw)

```bash
# Allow sign server
sudo ufw allow 53312/tcp

# Allow entrance server
sudo ufw allow 53310/tcp

# Allow channel servers (range)
sudo ufw allow 54001:54010/tcp
```

### Linux (iptables)

```bash
# Sign server
sudo iptables -A INPUT -p tcp --dport 53312 -j ACCEPT

# Entrance server
sudo iptables -A INPUT -p tcp --dport 53310 -j ACCEPT

# Channel servers (range)
sudo iptables -A INPUT -p tcp --dport 54001:54010 -j ACCEPT
```

### Windows Firewall

```powershell
# Allow specific ports
New-NetFirewallRule -DisplayName "Erupe Sign" -Direction Inbound -Protocol TCP -LocalPort 53312 -Action Allow
New-NetFirewallRule -DisplayName "Erupe Entrance" -Direction Inbound -Protocol TCP -LocalPort 53310 -Action Allow
New-NetFirewallRule -DisplayName "Erupe Channels" -Direction Inbound -Protocol TCP -LocalPort 54001-54010 -Action Allow
```

## Load Balancing

For high-traffic servers, consider:

1. **Multiple Entrance Servers**: Run multiple entrance server instances behind a load balancer
2. **Distributed Channels**: Spread channel servers across multiple physical servers
3. **Database Connection Pooling**: Use PgBouncer for database connections
4. **SignV2 with Reverse Proxy**: Use nginx/HAProxy with SignV2 for better scaling

## Monitoring

Monitor server health:

```bash
# Check if servers are listening
netstat -tlnp | grep erupe

# Check open ports
ss -tlnp | grep -E '(53312|53310|54001)'

# Monitor connections per channel
watch -n 1 'netstat -an | grep ESTABLISHED | grep 54001 | wc -l'
```

## Troubleshooting

### Can't Connect to Sign Server

- Verify Sign server is enabled
- Check port 53312 is open
- Verify client is configured for correct IP/port

### World List Empty

- Verify Entrance server is enabled
- Check Entrance server port (53310)
- Ensure at least one Entry is configured

### Can't Enter World

- Verify Channel server is enabled
- Check channel ports are open
- Verify channel ports in Entrance entries match actual running servers

### Server Crashes on Startup

- Check all ports are unique
- Verify database connection (password not empty)
- Check logs for specific errors

## Related Documentation

- [Database](database.md) - Database configuration
- [Basic Settings](basic-settings.md) - Host and network settings
- [CLAUDE.md](../CLAUDE.md#architecture) - Detailed architecture overview
