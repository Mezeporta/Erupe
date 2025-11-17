package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ErrorGroup represents a collection of similar errors grouped together.
//
// Errors can be grouped by message, caller, or logger to identify patterns
// and recurring issues in the logs.
type ErrorGroup struct {
	Message   string          // Primary message for this error group
	Count     int             // Total number of occurrences
	FirstSeen string          // Timestamp of first occurrence
	LastSeen  string          // Timestamp of last occurrence
	Examples  []*LogEntry     // Sample log entries (limited by the limit flag)
	Callers   map[string]int  // Map of caller locations to occurrence counts
}

// runErrors implements the errors command for extracting and analyzing errors.
//
// The errors command processes log files to find all errors and warnings, groups them
// by a specified criterion (message, caller, or logger), and presents statistics and
// examples for each group.
//
// Features:
//   - Groups errors by message (default), caller, or logger
//   - Shows total error and warning counts
//   - Displays first and last occurrence timestamps
//   - Optionally shows stack traces for detailed debugging
//   - Provides summary or detailed views
//   - Tracks which callers produced each error
//
// Options:
//   - f: Path to log file (default: "erupe.log")
//   - group: Group errors by "message", "caller", or "logger" (default: "message")
//   - stack: Show stack traces in detailed view
//   - limit: Maximum number of example entries per group (default: 10)
//   - summary: Show summary table only
//   - detailed: Show detailed information including examples and extra data
//
// Examples:
//   runErrors([]string{"-summary"})
//   runErrors([]string{"-detailed", "-stack"})
//   runErrors([]string{"-group", "caller", "-limit", "20"})
func runErrors(args []string) {
	fs := flag.NewFlagSet("errors", flag.ExitOnError)

	logFile := fs.String("f", "erupe.log", "Path to log file")
	groupBy := fs.String("group", "message", "Group errors by: message, caller, or logger")
	showStack := fs.Bool("stack", false, "Show stack traces")
	limit := fs.Int("limit", 10, "Limit number of examples per error group")
	summary := fs.Bool("summary", false, "Show summary only (grouped by error type)")
	detailed := fs.Bool("detailed", false, "Show detailed error information")

	fs.Parse(args)

	errorGroups := make(map[string]*ErrorGroup)
	var totalErrors int
	var totalWarnings int

	err := StreamLogFile(*logFile, func(entry *LogEntry) error {
		// Only process errors and warnings
		if entry.Level != "error" && entry.Level != "warn" {
			return nil
		}

		if entry.Level == "error" {
			totalErrors++
		} else {
			totalWarnings++
		}

		// Determine grouping key
		var key string
		switch *groupBy {
		case "message":
			key = entry.Message
		case "caller":
			key = entry.Caller
		case "logger":
			key = entry.Logger
		default:
			key = entry.Message
		}

		// Create or update error group
		group, exists := errorGroups[key]
		if !exists {
			group = &ErrorGroup{
				Message:   entry.Message,
				Callers:   make(map[string]int),
				Examples:  make([]*LogEntry, 0),
				FirstSeen: entry.Timestamp.Format("2006-01-02 15:04:05"),
			}
			errorGroups[key] = group
		}

		group.Count++
		group.LastSeen = entry.Timestamp.Format("2006-01-02 15:04:05")

		if entry.Caller != "" {
			group.Callers[entry.Caller]++
		}

		// Store example (limit to avoid memory issues)
		if len(group.Examples) < *limit {
			group.Examples = append(group.Examples, entry)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing log file: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("=== Error Analysis ===\n")
	fmt.Printf("Total Errors: %d\n", totalErrors)
	fmt.Printf("Total Warnings: %d\n", totalWarnings)
	fmt.Printf("Unique Error Groups: %d\n\n", len(errorGroups))

	if *summary {
		printErrorSummary(errorGroups)
	} else {
		printDetailedErrors(errorGroups, *showStack, *detailed)
	}
}

// printErrorSummary displays a tabular summary of error groups sorted by occurrence count.
//
// The summary table includes:
//   - Error message (truncated to 60 characters if longer)
//   - Total count of occurrences
//   - First seen timestamp
//   - Last seen timestamp
//
// Groups are sorted by count in descending order (most frequent first).
//
// Parameters:
//   - groups: Map of error groups to summarize
func printErrorSummary(groups map[string]*ErrorGroup) {
	// Sort by count
	type groupPair struct {
		key   string
		group *ErrorGroup
	}

	var pairs []groupPair
	for key, group := range groups {
		pairs = append(pairs, groupPair{key, group})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].group.Count > pairs[j].group.Count
	})

	fmt.Printf("%-60s | %-8s | %-19s | %-19s\n", "Error Message", "Count", "First Seen", "Last Seen")
	fmt.Println(strings.Repeat("-", 120))

	for _, pair := range pairs {
		msg := pair.group.Message
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		fmt.Printf("%-60s | %-8d | %-19s | %-19s\n",
			msg,
			pair.group.Count,
			pair.group.FirstSeen,
			pair.group.LastSeen)
	}
}

// printDetailedErrors displays comprehensive information about each error group.
//
// For each error group, displays:
//   - Group number and occurrence count
//   - Error message
//   - First and last seen timestamps
//   - Caller locations with counts
//   - Example occurrences with full details (if detailed=true)
//   - Stack traces (if showStack=true and available)
//
// Groups are sorted by occurrence count in descending order.
//
// Parameters:
//   - groups: Map of error groups to display
//   - showStack: Whether to include stack traces in the output
//   - detailed: Whether to show example occurrences and extra data
func printDetailedErrors(groups map[string]*ErrorGroup, showStack, detailed bool) {
	// Sort by count
	type groupPair struct {
		key   string
		group *ErrorGroup
	}

	var pairs []groupPair
	for key, group := range groups {
		pairs = append(pairs, groupPair{key, group})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].group.Count > pairs[j].group.Count
	})

	for idx, pair := range pairs {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("Error Group #%d (Count: %d)\n", idx+1, pair.group.Count)
		fmt.Printf("%s\n", strings.Repeat("=", 80))
		fmt.Printf("Message: %s\n", pair.group.Message)
		fmt.Printf("First Seen: %s\n", pair.group.FirstSeen)
		fmt.Printf("Last Seen: %s\n", pair.group.LastSeen)

		if len(pair.group.Callers) > 0 {
			fmt.Printf("\nCallers:\n")
			// Sort callers by count
			type callerPair struct {
				name  string
				count int
			}
			var callers []callerPair
			for name, count := range pair.group.Callers {
				callers = append(callers, callerPair{name, count})
			}
			sort.Slice(callers, func(i, j int) bool {
				return callers[i].count > callers[j].count
			})
			for _, c := range callers {
				fmt.Printf("  %s: %d times\n", c.name, c.count)
			}
		}

		if detailed && len(pair.group.Examples) > 0 {
			fmt.Printf("\nExample occurrences:\n")
			for i, example := range pair.group.Examples {
				fmt.Printf("\n  [Example %d] %s\n", i+1, example.Timestamp.Format("2006-01-02 15:04:05.000"))
				fmt.Printf("  Logger: %s\n", example.Logger)
				if example.Caller != "" {
					fmt.Printf("  Caller: %s\n", example.Caller)
				}
				if example.Error != "" {
					fmt.Printf("  Error: %s\n", example.Error)
				}

				// Print extra data
				if len(example.ExtraData) > 0 {
					fmt.Printf("  Extra Data:\n")
					for k, v := range example.ExtraData {
						fmt.Printf("    %s: %v\n", k, v)
					}
				}

				if showStack && example.StackTrace != "" {
					fmt.Printf("  Stack Trace:\n")
					lines := strings.Split(example.StackTrace, "\n")
					for _, line := range lines {
						fmt.Printf("    %s\n", line)
					}
				}
			}
		}
	}
}
