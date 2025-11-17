package main

import (
	"flag"
	"fmt"
	"os"
)

// main is the entry point for the log analyzer CLI tool.
//
// The tool provides five main commands:
//   - filter: Filter logs by level, logger, message content, or time range
//   - errors: Extract and analyze errors with grouping and stack traces
//   - connections: Track player connections and sessions with statistics
//   - stats: Generate comprehensive statistics about log activity
//   - tail: Follow logs in real-time (like tail -f)
//
// Usage:
//   loganalyzer <command> [options]
//   loganalyzer filter -level error -since 1h
//   loganalyzer errors -summary
//   loganalyzer stats -detailed
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Erupe Log Analyzer - Suite of tools to analyze erupe.log files\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s <command> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Available commands:\n")
		fmt.Fprintf(os.Stderr, "  filter      Filter logs by level, logger, or time range\n")
		fmt.Fprintf(os.Stderr, "  errors      Extract and analyze errors with stack traces\n")
		fmt.Fprintf(os.Stderr, "  connections Analyze connection events and player sessions\n")
		fmt.Fprintf(os.Stderr, "  stats       Generate statistics summary\n")
		fmt.Fprintf(os.Stderr, "  tail        Follow log file in real-time (like tail -f)\n")
		fmt.Fprintf(os.Stderr, "\nUse '%s <command> -h' for more information about a command.\n", os.Args[0])
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "filter":
		runFilter(args)
	case "errors":
		runErrors(args)
	case "connections":
		runConnections(args)
	case "stats":
		runStats(args)
	case "tail":
		runTail(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}
