# Quick Reference Guide

## Installation

```bash
cd tools/loganalyzer
go build -o loganalyzer
```

## Quick Commands

### View Statistics

```bash
./loganalyzer stats -f ../../erupe.log
./loganalyzer stats -f ../../erupe.log -detailed
```

### Filter Logs

```bash
# Errors only
./loganalyzer filter -f ../../erupe.log -level error

# Last hour
./loganalyzer filter -f ../../erupe.log -since 1h

# Last 50 entries
./loganalyzer filter -f ../../erupe.log -tail 50

# Search message
./loganalyzer filter -f ../../erupe.log -msg "connection reset"
```

### Analyze Errors

```bash
# Error summary
./loganalyzer errors -f ../../erupe.log -summary

# Detailed with stack traces
./loganalyzer errors -f ../../erupe.log -detailed -stack
```

### Track Connections

```bash
# Connection stats
./loganalyzer connections -f ../../erupe.log

# Player sessions
./loganalyzer connections -f ../../erupe.log -sessions

# Specific player
./loganalyzer connections -f ../../erupe.log -player "PlayerName" -sessions -v
```

### Follow Logs

```bash
# Like tail -f
./loganalyzer tail -f ../../erupe.log

# Only errors
./loganalyzer tail -f ../../erupe.log -level error
```

## Common Workflows

### Troubleshooting a crash

```bash
# 1. Check recent errors
./loganalyzer filter -f erupe.log -level error -tail 20

# 2. Analyze error patterns
./loganalyzer errors -f erupe.log -detailed -stack

# 3. Check what was happening before crash
./loganalyzer filter -f erupe.log -since "2025-11-12T23:00:00Z" -tail 100
```

### Player investigation

```bash
# 1. Find player sessions
./loganalyzer connections -f erupe.log -player "PlayerName" -sessions -v

# 2. Check errors for that player
./loganalyzer filter -f erupe.log -logger "*PlayerName*"
```

### Monitoring

```bash
# Real-time error monitoring
./loganalyzer tail -f erupe.log -level error

# Daily statistics
./loganalyzer stats -f erupe.log -detailed
```

## Tips

1. **Pipe to less for long output**: `./loganalyzer filter -f erupe.log | less -R`
2. **Save to file**: `./loganalyzer stats -f erupe.log > stats.txt`
3. **Combine with grep**: `./loganalyzer filter -f erupe.log -level error | grep "mail"`
4. **Use -count for quick checks**: `./loganalyzer filter -f erupe.log -level error -count`
5. **Time ranges**: `-since` accepts both absolute (RFC3339) and relative (1h, 30m) times

## Output Format

Default output is colorized:

- Errors: Red
- Warnings: Yellow
- Info: Green

Disable colors with `-color=false` for piping to files.
