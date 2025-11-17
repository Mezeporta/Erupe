package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// runFilter implements the filter command for filtering log entries by various criteria.
//
// The filter command supports the following filters:
//   - level: Filter by log level (info, warn, error, fatal)
//   - logger: Filter by logger name (supports wildcards with *)
//   - msg: Filter by message content (case-insensitive substring match)
//   - since: Show logs since this time (RFC3339 format or duration like "1h", "30m")
//   - until: Show logs until this time (RFC3339 format)
//   - tail: Show only the last N matching entries
//   - count: Show only the count of matching entries instead of the entries themselves
//   - color: Enable/disable colorized output (default: true)
//
// All filters are combined with AND logic.
//
// Examples:
//   runFilter([]string{"-level", "error"})
//   runFilter([]string{"-since", "1h", "-logger", "channel-4*"})
//   runFilter([]string{"-msg", "connection reset", "-count"})
func runFilter(args []string) {
	fs := flag.NewFlagSet("filter", flag.ExitOnError)

	logFile := fs.String("f", "erupe.log", "Path to log file")
	level := fs.String("level", "", "Filter by log level (info, warn, error, fatal)")
	logger := fs.String("logger", "", "Filter by logger name (supports wildcards)")
	message := fs.String("msg", "", "Filter by message content (case-insensitive)")
	sinceStr := fs.String("since", "", "Show logs since this time (RFC3339 or duration like '1h')")
	untilStr := fs.String("until", "", "Show logs until this time (RFC3339)")
	colorize := fs.Bool("color", true, "Colorize output")
	count := fs.Bool("count", false, "Only show count of matching entries")
	tail := fs.Int("tail", 0, "Show last N entries")

	fs.Parse(args)

	// Parse time filters
	var since, until time.Time
	var err error

	if *sinceStr != "" {
		// Try parsing as duration first
		if duration, err := time.ParseDuration(*sinceStr); err == nil {
			since = time.Now().Add(-duration)
		} else if since, err = time.Parse(time.RFC3339, *sinceStr); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid since time format: %s\n", *sinceStr)
			os.Exit(1)
		}
	}

	if *untilStr != "" {
		if until, err = time.Parse(time.RFC3339, *untilStr); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid until time format: %s\n", *untilStr)
			os.Exit(1)
		}
	}

	// Collect matching entries
	var matches []*LogEntry
	var totalCount int

	err = StreamLogFile(*logFile, func(entry *LogEntry) error {
		totalCount++

		// Apply filters
		if *level != "" && !strings.EqualFold(entry.Level, *level) {
			return nil
		}

		if *logger != "" && !matchWildcard(entry.Logger, *logger) {
			return nil
		}

		if *message != "" && !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(*message)) {
			return nil
		}

		if !since.IsZero() && entry.Timestamp.Before(since) {
			return nil
		}

		if !until.IsZero() && entry.Timestamp.After(until) {
			return nil
		}

		matches = append(matches, entry)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing log file: %v\n", err)
		os.Exit(1)
	}

	// Handle tail option
	if *tail > 0 && len(matches) > *tail {
		matches = matches[len(matches)-*tail:]
	}

	if *count {
		fmt.Printf("Total entries: %d\n", totalCount)
		fmt.Printf("Matching entries: %d\n", len(matches))
	} else {
		for _, entry := range matches {
			fmt.Println(FormatLogEntry(entry, *colorize))
		}
		if len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "\n%d of %d entries matched\n", len(matches), totalCount)
		}
	}
}

// matchWildcard performs simple wildcard matching where * matches any sequence of characters.
//
// The function supports the following patterns:
//   - "*" matches everything
//   - "foo*" matches strings starting with "foo"
//   - "*foo" matches strings ending with "foo"
//   - "*foo*" matches strings containing "foo"
//   - "foo*bar" matches strings starting with "foo" and ending with "bar"
//
// Matching is case-insensitive. If the pattern contains no wildcards, it performs
// a simple case-insensitive substring match.
//
// Parameters:
//   - s: The string to match against
//   - pattern: The pattern with optional wildcards
//
// Returns:
//   - true if the string matches the pattern, false otherwise
//
// Examples:
//   matchWildcard("channel-4", "channel-*") // returns true
//   matchWildcard("main.channel-4.error", "*channel-4*") // returns true
//   matchWildcard("test", "foo*") // returns false
func matchWildcard(s, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") {
		return strings.Contains(strings.ToLower(s), strings.ToLower(pattern))
	}

	parts := strings.Split(pattern, "*")
	s = strings.ToLower(s)

	pos := 0
	for i, part := range parts {
		part = strings.ToLower(part)
		if part == "" {
			continue
		}

		idx := strings.Index(s[pos:], part)
		if idx == -1 {
			return false
		}

		// First part must match from beginning
		if i == 0 && idx != 0 {
			return false
		}

		pos += idx + len(part)
	}

	// Last part must match to end
	if !strings.HasSuffix(pattern, "*") {
		lastPart := strings.ToLower(parts[len(parts)-1])
		return strings.HasSuffix(s, lastPart)
	}

	return true
}
