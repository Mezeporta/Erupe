# Logging Configuration

File logging and log rotation configuration for Erupe.

## Configuration

```json
{
  "Logging": {
    "LogToFile": true,
    "LogFilePath": "logs/erupe.log",
    "LogMaxSize": 100,
    "LogMaxBackups": 3,
    "LogMaxAge": 28,
    "LogCompress": true
  }
}
```

## Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `LogToFile` | boolean | `true` | Enable file logging (logs to both console and file) |
| `LogFilePath` | string | `"logs/erupe.log"` | Path to log file (directory will be created automatically) |
| `LogMaxSize` | number | `100` | Maximum log file size in MB before rotation |
| `LogMaxBackups` | number | `3` | Number of old log files to keep |
| `LogMaxAge` | number | `28` | Maximum days to retain old logs |
| `LogCompress` | boolean | `true` | Compress rotated log files with gzip |

## How It Works

Erupe uses [lumberjack](https://github.com/natefinch/lumberjack) for automatic log rotation and compression.

### Log Rotation

When the current log file reaches `LogMaxSize` MB:

1. Current log is closed and renamed to `erupe.log.YYYY-MM-DD-HH-MM-SS`
2. If `LogCompress: true`, the old log is compressed to `.gz` format
3. A new `erupe.log` file is created
4. Old logs beyond `LogMaxBackups` count are deleted
5. Logs older than `LogMaxAge` days are deleted

### Example Log Files

```text
logs/
├── erupe.log              (current, 45 MB)
├── erupe.log.2025-11-17-14-23-45.gz (100 MB compressed)
├── erupe.log.2025-11-16-08-15-32.gz (100 MB compressed)
└── erupe.log.2025-11-15-19-42-18.gz (100 MB compressed)
```

## Log Format

Log format depends on `DevMode` setting:

### Development Mode (DevMode: true)

Console format (human-readable):

```text
2025-11-18T10:30:45.123Z    INFO    channelserver    Player connected    {"charID": 12345, "ip": "127.0.0.1"}
2025-11-18T10:30:46.456Z    ERROR   channelserver    Failed to load data {"error": "database timeout"}
```

### Production Mode (DevMode: false)

JSON format (machine-parsable):

```json
{"level":"info","ts":"2025-11-18T10:30:45.123Z","logger":"channelserver","msg":"Player connected","charID":12345,"ip":"127.0.0.1"}
{"level":"error","ts":"2025-11-18T10:30:46.456Z","logger":"channelserver","msg":"Failed to load data","error":"database timeout"}
```

## Log Analysis

Erupe includes a built-in log analyzer tool in `tools/loganalyzer/`:

```bash
# Filter by log level
./loganalyzer filter -f ../../logs/erupe.log -level error

# Analyze errors with stack traces
./loganalyzer errors -f ../../logs/erupe.log -stack -detailed

# Track player connections
./loganalyzer connections -f ../../logs/erupe.log -sessions

# Real-time monitoring
./loganalyzer tail -f ../../logs/erupe.log -level error

# Generate statistics
./loganalyzer stats -f ../../logs/erupe.log -detailed
```

See [CLAUDE.md](../CLAUDE.md#log-analysis) for more details.

## Examples

### Minimal Logging (Development)

```json
{
  "Logging": {
    "LogToFile": false
  }
}
```

Only logs to console, no file logging.

### Standard Production Logging

```json
{
  "Logging": {
    "LogToFile": true,
    "LogFilePath": "logs/erupe.log",
    "LogMaxSize": 100,
    "LogMaxBackups": 7,
    "LogMaxAge": 30,
    "LogCompress": true
  }
}
```

Keeps up to 7 log files, 30 days maximum, compressed.

### High-Volume Server

```json
{
  "Logging": {
    "LogToFile": true,
    "LogFilePath": "/var/log/erupe/erupe.log",
    "LogMaxSize": 500,
    "LogMaxBackups": 14,
    "LogMaxAge": 60,
    "LogCompress": true
  }
}
```

Larger log files (500 MB), more backups (14), longer retention (60 days).

### Debug/Testing (No Rotation)

```json
{
  "Logging": {
    "LogToFile": true,
    "LogFilePath": "logs/debug.log",
    "LogMaxSize": 1000,
    "LogMaxBackups": 0,
    "LogMaxAge": 0,
    "LogCompress": false
  }
}
```

Single large log file, no rotation, no compression. Useful for debugging sessions.

## Disk Space Considerations

Calculate approximate disk usage:

```text
Total Disk Usage = (LogMaxSize × LogMaxBackups) × CompressionRatio
```

**Compression ratios:**

- Text logs: ~10:1 (100 MB → 10 MB compressed)
- JSON logs: ~8:1 (100 MB → 12.5 MB compressed)

**Example:**

```text
LogMaxSize: 100 MB
LogMaxBackups: 7
Compression: enabled (~10:1 ratio)

Total: (100 MB × 7) / 10 = 70 MB (approximately)
```

## Best Practices

1. **Enable compression** - Saves significant disk space
2. **Set reasonable MaxSize** - 100-200 MB works well for most servers
3. **Adjust retention** - Keep logs for at least 7 days, preferably 30
4. **Use absolute paths in production** - `/var/log/erupe/erupe.log` instead of `logs/erupe.log`
5. **Monitor disk space** - Set up alerts if disk usage exceeds 80%
6. **Use JSON format in production** - Easier to parse with log analysis tools

## Related Documentation

- [Development Mode](development-mode.md) - DevMode affects log format
- [Basic Settings](basic-settings.md) - Basic server configuration
- [CLAUDE.md](../CLAUDE.md#log-analysis) - Log analyzer tool usage
