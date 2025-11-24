package main

import (
	"flag"
	"fmt"
	"os"
)

// main is the entry point for the user check CLI tool.
//
// The tool provides commands to query connected users and server status
// by reading from the Erupe database. It's designed to be used while
// the server is running to monitor player activity.
//
// Usage:
//
//	usercheck <command> [options]
//	usercheck list                    # List all connected users
//	usercheck count                   # Count connected users
//	usercheck search -name "player"   # Search for a specific player
//	usercheck servers                 # Show server/channel status
//	usercheck history -name "player"  # Show player login history
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Erupe User Check - Tool to query connected users and server status\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s <command> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Available commands:\n")
		fmt.Fprintf(os.Stderr, "  list      List all currently connected users\n")
		fmt.Fprintf(os.Stderr, "  count     Show count of connected users per server\n")
		fmt.Fprintf(os.Stderr, "  search    Search for a specific connected user by name\n")
		fmt.Fprintf(os.Stderr, "  servers   Show server/channel status and player counts\n")
		fmt.Fprintf(os.Stderr, "  history   Show recent login history for a player\n")
		fmt.Fprintf(os.Stderr, "\nDatabase configuration:\n")
		fmt.Fprintf(os.Stderr, "  By default, reads from config.json in the project root.\n")
		fmt.Fprintf(os.Stderr, "  Use flags to override specific settings.\n\n")
		fmt.Fprintf(os.Stderr, "  -config    Path to config.json (auto-detected if not specified)\n")
		fmt.Fprintf(os.Stderr, "  -host      Database host (overrides config.json)\n")
		fmt.Fprintf(os.Stderr, "  -port      Database port (overrides config.json)\n")
		fmt.Fprintf(os.Stderr, "  -user      Database user (overrides config.json)\n")
		fmt.Fprintf(os.Stderr, "  -password  Database password (overrides config.json)\n")
		fmt.Fprintf(os.Stderr, "  -dbname    Database name (overrides config.json)\n")
		fmt.Fprintf(os.Stderr, "\nUse '%s <command> -h' for more information about a command.\n", os.Args[0])
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "list":
		runList(args)
	case "count":
		runCount(args)
	case "search":
		runSearch(args)
	case "servers":
		runServers(args)
	case "history":
		runHistory(args)
	case "-h", "--help", "help":
		flag.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}
