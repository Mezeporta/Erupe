// Package main provides a comprehensive suite of tools for analyzing Erupe server logs.
//
// The log analyzer supports both JSON-formatted logs and tab-delimited timestamp logs,
// providing commands for filtering, error analysis, connection tracking, statistics
// generation, and real-time log following.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LogEntry represents a parsed log entry from either JSON or timestamp-based format.
//
// The parser supports two log formats:
//  1. JSON format: {"level":"info","ts":1762989571.547817,"logger":"main","msg":"Starting"}
//  2. Timestamp format: 2025-11-12T23:19:31.546Z	INFO	commands	Command Help: Enabled
type LogEntry struct {
	Raw        string                 // Original log line
	Level      string                 // Log level: info, warn, error, fatal
	Timestamp  time.Time              // Parsed timestamp
	Logger     string                 // Logger name
	Caller     string                 // Caller file:line
	Message    string                 // Log message
	Error      string                 // Error message (if present)
	StackTrace string                 // Stack trace (if present)
	ExtraData  map[string]interface{} // Additional fields
	IsJSON     bool                   // True if parsed from JSON format
}

// ParseLogFile reads and parses an entire log file into memory.
//
// This function loads all log entries into memory and is suitable for smaller log files
// or when random access to entries is needed. For large files or streaming operations,
// use StreamLogFile instead.
//
// The function automatically handles both JSON and timestamp-based log formats,
// skips empty lines and "nohup: ignoring input" messages, and uses a large buffer
// (1MB) to handle long lines like stack traces.
//
// Parameters:
//   - filename: Path to the log file to parse
//
// Returns:
//   - A slice of LogEntry pointers containing all parsed entries
//   - An error if the file cannot be opened or read
//
// Example:
//
//	entries, err := ParseLogFile("erupe.log")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Parsed %d entries\n", len(entries))
func ParseLogFile(filename string) ([]*LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var entries []*LogEntry
	scanner := bufio.NewScanner(file)

	// Increase buffer size for long lines (like stack traces)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "nohup: ignoring input" {
			continue
		}

		entry := ParseLogLine(line)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return entries, nil
}

// ParseLogLine parses a single log line into a LogEntry.
//
// This function attempts to parse the line in the following order:
//  1. JSON format: Lines starting with '{' are parsed as JSON objects
//  2. Timestamp format: Tab-delimited lines with RFC3339 timestamps
//  3. Unknown format: Lines that don't match either format are marked as "unknown" level
//
// For JSON logs, all standard fields (level, ts, logger, caller, msg, error, stacktrace)
// are extracted, and any additional fields are stored in ExtraData.
//
// Parameters:
//   - line: A single line from the log file
//
// Returns:
//   - A LogEntry pointer containing the parsed data, or nil if the line is invalid
//
// Example:
//
//	entry := ParseLogLine(`{"level":"info","ts":1762989571.547817,"msg":"Starting"}`)
//	fmt.Println(entry.Level, entry.Message)
func ParseLogLine(line string) *LogEntry {
	entry := &LogEntry{
		Raw:       line,
		ExtraData: make(map[string]interface{}),
	}

	// Try parsing as JSON first
	if strings.HasPrefix(line, "{") {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
			entry.IsJSON = true

			// Extract standard fields
			if level, ok := jsonData["level"].(string); ok {
				entry.Level = level
			}

			if ts, ok := jsonData["ts"].(float64); ok {
				entry.Timestamp = time.Unix(int64(ts), int64((ts-float64(int64(ts)))*1e9))
			}

			if logger, ok := jsonData["logger"].(string); ok {
				entry.Logger = logger
			}

			if caller, ok := jsonData["caller"].(string); ok {
				entry.Caller = caller
			}

			if msg, ok := jsonData["msg"].(string); ok {
				entry.Message = msg
			}

			if errMsg, ok := jsonData["error"].(string); ok {
				entry.Error = errMsg
			}

			if stackTrace, ok := jsonData["stacktrace"].(string); ok {
				entry.StackTrace = stackTrace
			}

			// Store any extra fields
			for k, v := range jsonData {
				if k != "level" && k != "ts" && k != "logger" && k != "caller" &&
					k != "msg" && k != "error" && k != "stacktrace" {
					entry.ExtraData[k] = v
				}
			}

			return entry
		}
	}

	// Try parsing as timestamp-based log (2025-11-12T23:19:31.546Z INFO commands ...)
	parts := strings.SplitN(line, "\t", 4)
	if len(parts) >= 3 {
		// Parse timestamp
		if ts, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
			entry.Timestamp = ts
			entry.Level = strings.ToLower(parts[1])
			entry.Logger = parts[2]
			if len(parts) == 4 {
				entry.Message = parts[3]
			}
			return entry
		}
	}

	// If we can't parse it, return a basic entry
	entry.Level = "unknown"
	entry.Message = line
	return entry
}

// StreamLogFile reads a log file line by line and calls the callback for each entry.
//
// This function is memory-efficient and suitable for processing large log files as it
// processes entries one at a time without loading the entire file into memory. The
// callback function is called for each successfully parsed log entry.
//
// The function uses a 1MB buffer to handle long lines such as those containing stack traces.
// Empty lines and "nohup: ignoring input" messages are automatically skipped.
//
// If the callback returns an error, processing stops immediately and that error is returned.
//
// Parameters:
//   - filename: Path to the log file to process
//   - callback: Function to call for each parsed LogEntry
//
// Returns:
//   - An error if the file cannot be opened, read, or if the callback returns an error
//
// Example:
//
//	err := StreamLogFile("erupe.log", func(entry *LogEntry) error {
//	    if entry.Level == "error" {
//	        fmt.Println(entry.Message)
//	    }
//	    return nil
//	})
func StreamLogFile(filename string, callback func(*LogEntry) error) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Increase buffer size for long lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "nohup: ignoring input" {
			continue
		}

		entry := ParseLogLine(line)
		if entry != nil {
			if err := callback(entry); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

// FormatLogEntry formats a log entry for human-readable display.
//
// The output format is: "TIMESTAMP LEVEL [LOGGER] MESSAGE key=value ..."
//
// Timestamps are formatted as "2006-01-02 15:04:05.000". Log levels can be colorized
// for terminal display:
//   - Errors (error, fatal, panic): Red
//   - Warnings: Yellow
//   - Info: Green
//
// If the entry contains an error message, it's appended as error="message".
// Any extra fields in ExtraData are appended as key=value pairs.
//
// Parameters:
//   - entry: The LogEntry to format
//   - colorize: Whether to add ANSI color codes for terminal display
//
// Returns:
//   - A formatted string representation of the log entry
//
// Example:
//
//	formatted := FormatLogEntry(entry, true)
//	fmt.Println(formatted)
//	// Output: 2025-11-12 23:19:31.546 INFO [main] Starting Erupe
func FormatLogEntry(entry *LogEntry, colorize bool) string {
	var sb strings.Builder

	// Format timestamp
	if !entry.Timestamp.IsZero() {
		sb.WriteString(entry.Timestamp.Format("2006-01-02 15:04:05.000"))
		sb.WriteString(" ")
	}

	// Format level with colors
	levelStr := strings.ToUpper(entry.Level)
	if colorize {
		switch entry.Level {
		case "error", "fatal", "panic":
			levelStr = fmt.Sprintf("\033[31m%s\033[0m", levelStr) // Red
		case "warn":
			levelStr = fmt.Sprintf("\033[33m%s\033[0m", levelStr) // Yellow
		case "info":
			levelStr = fmt.Sprintf("\033[32m%s\033[0m", levelStr) // Green
		}
	}
	sb.WriteString(fmt.Sprintf("%-5s ", levelStr))

	// Format logger
	if entry.Logger != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", entry.Logger))
	}

	// Format message
	sb.WriteString(entry.Message)

	// Add error if present
	if entry.Error != "" {
		sb.WriteString(fmt.Sprintf(" error=%q", entry.Error))
	}

	// Add extra data
	if len(entry.ExtraData) > 0 {
		sb.WriteString(" ")
		first := true
		for k, v := range entry.ExtraData {
			if !first {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	return sb.String()
}
