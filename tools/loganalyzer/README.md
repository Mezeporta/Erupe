# Erupe Log Analyzer

A comprehensive suite of Go tools to analyze Erupe server logs (`erupe.log`).

## Features

- **Filter logs** by level, logger, message content, or time range
- **Analyze errors** with grouping, statistics, and stack trace display
- **Track connections** and player sessions with detailed statistics
- **Generate statistics** about log activity, operations, and patterns
- **Tail logs** in real-time like `tail -f`

## Installation

```bash
cd tools/loganalyzer
go build -o loganalyzer
```

This will create a `loganalyzer` binary in the current directory.

## Usage

The tool provides multiple commands, each with its own options:

```bash
./loganalyzer <command> [options]
```

### Commands

#### 1. `filter` - Filter logs by various criteria

Filter logs by level, logger, message content, or time range.

**Examples:**

```bash
# Show only errors
./loganalyzer filter -f ../../erupe.log -level error

# Show warnings from the last hour
./loganalyzer filter -f ../../erupe.log -level warn -since 1h

# Filter by logger (supports wildcards)
./loganalyzer filter -f ../../erupe.log -logger "channel-4*"

# Search for specific message content
./loganalyzer filter -f ../../erupe.log -msg "connection reset"

# Show only last 50 entries
./loganalyzer filter -f ../../erupe.log -tail 50

# Count matching entries without displaying them
./loganalyzer filter -f ../../erupe.log -level error -count
```

**Options:**

- `-f` - Path to log file (default: `erupe.log`)
- `-level` - Filter by log level (info, warn, error, fatal)
- `-logger` - Filter by logger name (supports wildcards with *)
- `-msg` - Filter by message content (case-insensitive)
- `-since` - Show logs since this time (RFC3339 or duration like '1h', '30m')
- `-until` - Show logs until this time (RFC3339)
- `-color` - Colorize output (default: true)
- `-count` - Only show count of matching entries
- `-tail` - Show last N entries

#### 2. `errors` - Analyze errors and warnings

Extract and analyze errors with grouping by message, caller, or logger.

**Examples:**

```bash
# Show error summary grouped by message
./loganalyzer errors -f ../../erupe.log -summary

# Show detailed error information with examples
./loganalyzer errors -f ../../erupe.log -detailed

# Show errors with stack traces
./loganalyzer errors -f ../../erupe.log -stack -detailed

# Group errors by caller instead of message
./loganalyzer errors -f ../../erupe.log -group caller -summary

# Show more examples per error group
./loganalyzer errors -f ../../erupe.log -detailed -limit 20
```

**Options:**

- `-f` - Path to log file (default: `erupe.log`)
- `-group` - Group errors by: message, caller, or logger (default: message)
- `-stack` - Show stack traces
- `-limit` - Limit number of examples per error group (default: 10)
- `-summary` - Show summary only (grouped by error type)
- `-detailed` - Show detailed error information

#### 3. `connections` - Analyze player connections and sessions

Track connection events, player sessions, and connection statistics.

**Examples:**

```bash
# Show connection statistics
./loganalyzer connections -f ../../erupe.log

# Show individual player sessions
./loganalyzer connections -f ../../erupe.log -sessions

# Show detailed session information
./loganalyzer connections -f ../../erupe.log -sessions -v

# Filter by player name
./loganalyzer connections -f ../../erupe.log -player "Sarah" -sessions

# Show only statistics without sessions
./loganalyzer connections -f ../../erupe.log -stats -sessions=false
```

**Options:**

- `-f` - Path to log file (default: `erupe.log`)
- `-player` - Filter by player name
- `-sessions` - Show individual player sessions
- `-stats` - Show connection statistics (default: true)
- `-v` - Verbose output

**Statistics provided:**

- Total connections
- Unique players and IP addresses
- Channel distribution
- Connections per day
- Top IP addresses
- Disconnect reasons

#### 4. `stats` - Generate comprehensive statistics

Analyze overall log statistics, activity patterns, and operation counts.

**Examples:**

```bash
# Show basic statistics
./loganalyzer stats -f ../../erupe.log

# Show detailed statistics including top loggers and messages
./loganalyzer stats -f ../../erupe.log -detailed

# Show top 20 instead of default 10
./loganalyzer stats -f ../../erupe.log -detailed -top 20
```

**Options:**

- `-f` - Path to log file (default: `erupe.log`)
- `-top` - Show top N messages/loggers (default: 10)
- `-detailed` - Show detailed statistics

**Statistics provided:**

- Total log entries and time range
- Entries by log level
- Operation counts (saves, broadcasts, stage changes)
- Top loggers and messages
- Activity by day and hour
- Unique callers

#### 5. `tail` - Follow logs in real-time

Watch log file for new entries, similar to `tail -f`.

**Examples:**

```bash
# Follow log file showing last 10 lines first
./loganalyzer tail -f ../../erupe.log

# Show last 50 lines and follow
./loganalyzer tail -f ../../erupe.log -n 50

# Follow only errors
./loganalyzer tail -f ../../erupe.log -level error

# Don't follow, just show last 20 lines
./loganalyzer tail -f ../../erupe.log -n 20 -follow=false
```

**Options:**

- `-f` - Path to log file (default: `erupe.log`)
- `-n` - Number of initial lines to show (default: 10)
- `-follow` - Follow the log file (default: true)
- `-level` - Filter by log level
- `-color` - Colorize output (default: true)

## Common Use Cases

### Finding the cause of a server crash

```bash
# Look for errors around a specific time
./loganalyzer filter -f erupe.log -level error -since "2025-11-12T23:00:00Z"

# Analyze all errors with stack traces
./loganalyzer errors -f erupe.log -stack -detailed
```

### Analyzing player activity

```bash
# See which players connected today
./loganalyzer connections -f erupe.log -sessions -v

# Find all activity for a specific player
./loganalyzer connections -f erupe.log -player "Sarah" -sessions -v
```

### Monitoring server health

```bash
# Real-time monitoring of errors
./loganalyzer tail -f erupe.log -level error

# Check overall statistics
./loganalyzer stats -f erupe.log -detailed

# Analyze connection patterns
./loganalyzer connections -f erupe.log -stats
```

### Investigating specific issues

```bash
# Find all connection reset errors
./loganalyzer filter -f erupe.log -msg "connection reset"

# Analyze database errors
./loganalyzer errors -f erupe.log -group caller | grep -i database

# Check activity during peak hours
./loganalyzer stats -f erupe.log -detailed
```

## Log Format Support

The tool supports both log formats found in Erupe logs:

1. **JSON format** (structured logs):

   ```json
   {"level":"info","ts":1762989571.547817,"logger":"main","caller":"Erupe/main.go:57","msg":"Starting Erupe"}
   ```

2. **Timestamp format** (simple logs):

   ```text
   2025-11-12T23:19:31.546Z INFO commands Command Help: Enabled
   ```

## Performance

The tool uses streaming parsing to handle large log files efficiently:

- Memory-efficient streaming for filter and stats commands
- Fast pattern matching for message filtering
- Handles log files with millions of entries

## Output

By default, output is colorized for better readability:

- **Errors** are displayed in red
- **Warnings** are displayed in yellow
- **Info** messages are displayed in green

Colorization can be disabled with `-color=false` for piping to files or other tools.

## Tips

1. Use `-count` with filter to quickly see how many entries match without displaying them all
2. Combine `filter` with `grep` for more complex searches: `./loganalyzer filter -f erupe.log | grep pattern`
3. Use `-tail` to limit output when exploring logs interactively
4. The `-since` option accepts both absolute timestamps and relative durations (1h, 30m, 24h)
5. Use `-summary` with errors command for a quick overview before diving into details

## Building from Source

```bash
cd tools/loganalyzer
go build -o loganalyzer
```

Or to install it system-wide:

```bash
go install
```

## Contributing

Feel free to add new commands or improve existing ones. The codebase is modular:

- `parser.go` - Log parsing logic
- `filter.go` - Filter command
- `errors.go` - Error analysis command
- `connections.go` - Connection tracking command
- `stats.go` - Statistics generation
- `tail.go` - Real-time log following
- `main.go` - Command routing

## License

Part of the Erupe project.
