package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// LogStats aggregates comprehensive statistics about log file contents.
//
// This structure tracks various metrics including temporal patterns, log levels,
// message types, and server operations to provide insights into server behavior
// and activity patterns.
type LogStats struct {
	TotalEntries      int            // Total number of log entries
	EntriesByLevel    map[string]int // Log level to count
	EntriesByLogger   map[string]int // Logger name to count
	EntriesByDay      map[string]int // Date string to count
	EntriesByHour     map[int]int    // Hour (0-23) to count
	TopMessages       map[string]int // Message text to count
	FirstEntry        time.Time      // Timestamp of first entry
	LastEntry         time.Time      // Timestamp of last entry
	SaveOperations    int            // Count of save operations
	ObjectBroadcasts  int            // Count of object broadcasts
	StageChanges      int            // Count of stage changes
	TerminalLogs      int            // Count of terminal log entries
	UniqueCallers     map[string]bool // Set of unique caller locations
}

// runStats implements the stats command for generating comprehensive log statistics.
//
// The stats command processes the entire log file to collect statistics about:
//   - Overall log volume and time span
//   - Distribution of log levels (info, warn, error, etc.)
//   - Server operation counts (saves, broadcasts, stage changes)
//   - Temporal patterns (activity by day and hour)
//   - Top loggers and message types
//   - Unique code locations generating logs
//
// This provides a high-level overview of server activity and can help identify
// patterns, peak usage times, and potential issues.
//
// Options:
//   - f: Path to log file (default: "erupe.log")
//   - top: Number of top items to show in detailed view (default: 10)
//   - detailed: Show detailed statistics including temporal patterns and top messages
//
// Examples:
//   runStats([]string{})  // Basic statistics
//   runStats([]string{"-detailed"})  // Full statistics with temporal analysis
//   runStats([]string{"-detailed", "-top", "20"})  // Show top 20 items
func runStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)

	logFile := fs.String("f", "erupe.log", "Path to log file")
	topN := fs.Int("top", 10, "Show top N messages/loggers")
	detailed := fs.Bool("detailed", false, "Show detailed statistics")

	fs.Parse(args)

	stats := &LogStats{
		EntriesByLevel:  make(map[string]int),
		EntriesByLogger: make(map[string]int),
		EntriesByDay:    make(map[string]int),
		EntriesByHour:   make(map[int]int),
		TopMessages:     make(map[string]int),
		UniqueCallers:   make(map[string]bool),
	}

	err := StreamLogFile(*logFile, func(entry *LogEntry) error {
		stats.TotalEntries++

		// Track first and last entry
		if stats.FirstEntry.IsZero() || entry.Timestamp.Before(stats.FirstEntry) {
			stats.FirstEntry = entry.Timestamp
		}
		if entry.Timestamp.After(stats.LastEntry) {
			stats.LastEntry = entry.Timestamp
		}

		// Count by level
		stats.EntriesByLevel[entry.Level]++

		// Count by logger
		stats.EntriesByLogger[entry.Logger]++

		// Count by day
		if !entry.Timestamp.IsZero() {
			day := entry.Timestamp.Format("2006-01-02")
			stats.EntriesByDay[day]++

			// Count by hour of day
			hour := entry.Timestamp.Hour()
			stats.EntriesByHour[hour]++
		}

		// Count message types
		msg := entry.Message
		if len(msg) > 80 {
			msg = msg[:80] + "..."
		}
		stats.TopMessages[msg]++

		// Track unique callers
		if entry.Caller != "" {
			stats.UniqueCallers[entry.Caller] = true
		}

		// Count specific operations
		if strings.Contains(entry.Message, "Wrote recompressed savedata back to DB") {
			stats.SaveOperations++
		}
		if strings.Contains(entry.Message, "Broadcasting new object") {
			stats.ObjectBroadcasts++
		}
		if strings.Contains(entry.Message, "Sending notification to old stage clients") {
			stats.StageChanges++
		}
		if entry.Message == "SysTerminalLog" {
			stats.TerminalLogs++
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing log file: %v\n", err)
		os.Exit(1)
	}

	printStats(stats, *topN, *detailed)
}

// printStats displays formatted statistics output.
//
// In basic mode, shows:
//   - Total entries, time range, and average rate
//   - Distribution by log level
//   - Operation counts
//
// In detailed mode, additionally shows:
//   - Top N loggers by volume
//   - Entries by day
//   - Activity distribution by hour of day (with bar chart)
//   - Top N message types
//
// Parameters:
//   - stats: LogStats structure containing collected statistics
//   - topN: Number of top items to display in detailed view
//   - detailed: Whether to show detailed statistics
func printStats(stats *LogStats, topN int, detailed bool) {
	fmt.Printf("=== Erupe Log Statistics ===\n\n")

	// Basic stats
	fmt.Printf("Total Log Entries: %s\n", formatNumber(stats.TotalEntries))
	if !stats.FirstEntry.IsZero() && !stats.LastEntry.IsZero() {
		duration := stats.LastEntry.Sub(stats.FirstEntry)
		fmt.Printf("Time Range: %s to %s\n",
			stats.FirstEntry.Format("2006-01-02 15:04:05"),
			stats.LastEntry.Format("2006-01-02 15:04:05"))
		fmt.Printf("Total Duration: %s\n", formatDuration(duration))

		if duration.Hours() > 0 {
			entriesPerHour := float64(stats.TotalEntries) / duration.Hours()
			fmt.Printf("Average Entries/Hour: %.1f\n", entriesPerHour)
		}
	}
	fmt.Println()

	// Log levels
	fmt.Printf("--- Entries by Log Level ---\n")
	levels := []string{"info", "warn", "error", "fatal", "panic", "unknown"}
	for _, level := range levels {
		if count, ok := stats.EntriesByLevel[level]; ok {
			percentage := float64(count) / float64(stats.TotalEntries) * 100
			fmt.Printf("  %-8s: %s (%.1f%%)\n", strings.ToUpper(level), formatNumber(count), percentage)
		}
	}
	fmt.Println()

	// Operation counts
	fmt.Printf("--- Operation Counts ---\n")
	fmt.Printf("  Save Operations: %s\n", formatNumber(stats.SaveOperations))
	fmt.Printf("  Object Broadcasts: %s\n", formatNumber(stats.ObjectBroadcasts))
	fmt.Printf("  Stage Changes: %s\n", formatNumber(stats.StageChanges))
	fmt.Printf("  Terminal Logs: %s\n", formatNumber(stats.TerminalLogs))
	fmt.Printf("  Unique Callers: %s\n", formatNumber(len(stats.UniqueCallers)))
	fmt.Println()

	if detailed {
		// Top loggers
		if len(stats.EntriesByLogger) > 0 {
			fmt.Printf("--- Top %d Loggers ---\n", topN)
			printTopMap(stats.EntriesByLogger, topN, stats.TotalEntries)
			fmt.Println()
		}

		// Entries by day
		if len(stats.EntriesByDay) > 0 {
			fmt.Printf("--- Entries by Day ---\n")
			printDayMap(stats.EntriesByDay)
			fmt.Println()
		}

		// Entries by hour
		if len(stats.EntriesByHour) > 0 {
			fmt.Printf("--- Activity by Hour of Day ---\n")
			printHourDistribution(stats.EntriesByHour, stats.TotalEntries)
			fmt.Println()
		}

		// Top messages
		if len(stats.TopMessages) > 0 {
			fmt.Printf("--- Top %d Messages ---\n", topN)
			printTopMap(stats.TopMessages, topN, stats.TotalEntries)
			fmt.Println()
		}
	}
}

// printTopMap displays the top N items from a map sorted by count.
//
// The output includes:
//   - Rank number (1, 2, 3, ...)
//   - Item key (truncated to 60 characters if longer)
//   - Count with thousand separators
//   - Percentage of total
//
// Parameters:
//   - m: Map of items to counts
//   - topN: Maximum number of items to display
//   - total: Total count for calculating percentages
func printTopMap(m map[string]int, topN, total int) {
	type pair struct {
		key   string
		count int
	}

	var pairs []pair
	for k, v := range m {
		pairs = append(pairs, pair{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	if len(pairs) > topN {
		pairs = pairs[:topN]
	}

	for i, p := range pairs {
		percentage := float64(p.count) / float64(total) * 100
		key := p.key
		if len(key) > 60 {
			key = key[:57] + "..."
		}
		fmt.Printf("  %2d. %-60s: %s (%.1f%%)\n", i+1, key, formatNumber(p.count), percentage)
	}
}

// printDayMap displays entries grouped by day in chronological order.
//
// Output format: "YYYY-MM-DD: count"
// Days are sorted chronologically from earliest to latest.
//
// Parameters:
//   - m: Map of date strings (YYYY-MM-DD format) to counts
func printDayMap(m map[string]int) {
	type pair struct {
		day   string
		count int
	}

	var pairs []pair
	for k, v := range m {
		pairs = append(pairs, pair{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].day < pairs[j].day
	})

	for _, p := range pairs {
		fmt.Printf("  %s: %s\n", p.day, formatNumber(p.count))
	}
}

// printHourDistribution displays log activity by hour of day with a bar chart.
//
// For each hour (0-23), shows:
//   - Hour range (e.g., "14:00 - 14:59")
//   - ASCII bar chart visualization (█ characters proportional to percentage)
//   - Count with thousand separators
//   - Percentage of total
//
// Hours with no activity are skipped.
//
// Parameters:
//   - m: Map of hours (0-23) to entry counts
//   - total: Total number of entries for percentage calculation
func printHourDistribution(m map[int]int, total int) {
	for hour := 0; hour < 24; hour++ {
		count := m[hour]
		if count == 0 {
			continue
		}
		percentage := float64(count) / float64(total) * 100
		bar := strings.Repeat("█", int(percentage))
		fmt.Printf("  %02d:00 - %02d:59: %-20s %s (%.1f%%)\n",
			hour, hour, bar, formatNumber(count), percentage)
	}
}

// formatNumber formats an integer with thousand separators for readability.
//
// Examples:
//   - 123 -> "123"
//   - 1234 -> "1,234"
//   - 1234567 -> "1,234,567"
//
// Parameters:
//   - n: The integer to format
//
// Returns:
//   - A string with comma separators
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n/1000)%1000, n%1000)
}
