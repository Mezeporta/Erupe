package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"
)

// runTail implements the tail command for following log files in real-time.
//
// The tail command mimics the Unix `tail -f` command, displaying the last N lines
// of a log file and then continuously monitoring the file for new entries. This is
// useful for real-time monitoring of server activity.
//
// The command operates in two phases:
//  1. Initial display: Shows the last N matching entries from the file
//  2. Follow mode: Continuously monitors for new lines and displays them as they appear
//
// Both phases support filtering by log level and colorized output.
//
// Options:
//   - f: Path to log file (default: "logs/erupe.log")
//   - n: Number of initial lines to show (default: 10)
//   - follow: Whether to continue following the file (default: true)
//   - level: Filter by log level (info, warn, error, fatal)
//   - color: Colorize output (default: true)
//
// The follow mode polls the file every 100ms for new content. Use Ctrl+C to stop.
//
// Examples:
//
//	runTail([]string{})  // Show last 10 lines and follow
//	runTail([]string{"-n", "50"})  // Show last 50 lines and follow
//	runTail([]string{"-level", "error"})  // Only show errors
//	runTail([]string{"-follow=false", "-n", "20"})  // Just show last 20 lines, don't follow
func runTail(args []string) {
	fs := flag.NewFlagSet("tail", flag.ExitOnError)

	logFile := fs.String("f", "logs/erupe.log", "Path to log file")
	lines := fs.Int("n", 10, "Number of initial lines to show")
	follow := fs.Bool("follow", true, "Follow the log file (like tail -f)")
	level := fs.String("level", "", "Filter by log level")
	colorize := fs.Bool("color", true, "Colorize output")

	fs.Parse(args)

	// First, show last N lines
	if *lines > 0 {
		entries, err := ParseLogFile(*logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading log file: %v\n", err)
			os.Exit(1)
		}

		// Filter by level if specified
		var filtered []*LogEntry
		for _, entry := range entries {
			if *level == "" || entry.Level == *level {
				filtered = append(filtered, entry)
			}
		}

		// Show last N lines
		start := len(filtered) - *lines
		if start < 0 {
			start = 0
		}

		for i := start; i < len(filtered); i++ {
			fmt.Println(FormatLogEntry(filtered[i], *colorize))
		}
	}

	// If follow is enabled, watch for new lines
	if *follow {
		file, err := os.Open(*logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		// Seek to end of file
		file.Seek(0, 2)

		reader := bufio.NewReader(file)

		fmt.Fprintln(os.Stderr, "Following log file... (Ctrl+C to stop)")

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				// No more data, wait a bit and try again
				time.Sleep(100 * time.Millisecond)
				continue
			}

			entry := ParseLogLine(line)
			if entry != nil {
				// Filter by level if specified
				if *level == "" || entry.Level == *level {
					fmt.Println(FormatLogEntry(entry, *colorize))
				}
			}
		}
	}
}
