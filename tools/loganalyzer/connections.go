package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// PlayerSession represents a single player's connection session to the server.
//
// A session is identified by the combination of channel and IP:port, tracking
// all activities from when a player connects until they disconnect.
type PlayerSession struct {
	Name       string    // Player name
	IPPort     string    // Client IP address and port (e.g., "192.168.1.1:12345")
	Channel    string    // Server channel (e.g., "channel-4")
	FirstSeen  time.Time // Timestamp of first activity
	LastSeen   time.Time // Timestamp of last activity
	Activities []string  // List of player activities
	Stages     []string  // List of stage changes
	Objects    []string  // List of objects broadcast by this player
	Errors     int       // Number of errors encountered during session
	SaveCount  int       // Number of save operations performed
}

// ConnectionStats aggregates statistics about player connections across all sessions.
//
// This structure tracks high-level metrics useful for understanding server usage
// patterns, peak times, and common connection issues.
type ConnectionStats struct {
	TotalConnections    int             // Total number of player sessions
	UniqueIPs           map[string]int  // IP addresses to connection count
	UniquePlayers       map[string]bool // Set of unique player names
	ConnectionsPerDay   map[string]int  // Date to connection count
	ChannelDistribution map[string]int  // Channel to connection count
	DisconnectReasons   map[string]int  // Disconnect reason to count
}

// runConnections implements the connections command for analyzing player connection patterns.
//
// The connections command tracks player sessions from connection to disconnection, providing
// both aggregate statistics and individual session details. It can identify patterns in
// player activity, track connection issues, and analyze channel usage.
//
// Features:
//   - Tracks individual player sessions with timestamps and activities
//   - Aggregates connection statistics (total, unique players, IPs)
//   - Shows channel distribution and peak connection times
//   - Analyzes disconnect reasons
//   - Supports filtering by player name
//   - Provides verbose session details including objects and stage changes
//
// Options:
//   - f: Path to log file (default: "logs/erupe.log")
//   - player: Filter sessions by player name (case-insensitive substring match)
//   - sessions: Show individual player sessions
//   - stats: Show connection statistics (default: true)
//   - v: Verbose output including objects and stage changes
//
// Examples:
//
//	runConnections([]string{"-stats"})
//	runConnections([]string{"-sessions", "-v"})
//	runConnections([]string{"-player", "Sarah", "-sessions"})
func runConnections(args []string) {
	fs := flag.NewFlagSet("connections", flag.ExitOnError)

	logFile := fs.String("f", "logs/erupe.log", "Path to log file")
	player := fs.String("player", "", "Filter by player name")
	showSessions := fs.Bool("sessions", false, "Show individual player sessions")
	showStats := fs.Bool("stats", true, "Show connection statistics")
	verbose := fs.Bool("v", false, "Verbose output")

	fs.Parse(args)

	stats := &ConnectionStats{
		UniqueIPs:           make(map[string]int),
		UniquePlayers:       make(map[string]bool),
		ConnectionsPerDay:   make(map[string]int),
		ChannelDistribution: make(map[string]int),
		DisconnectReasons:   make(map[string]int),
	}

	sessions := make(map[string]*PlayerSession) // key: channel-IP:port

	err := StreamLogFile(*logFile, func(entry *LogEntry) error {
		// Track player activities
		if strings.Contains(entry.Message, "Sending existing stage objects to") {
			// Extract player name
			parts := strings.Split(entry.Message, " to ")
			if len(parts) == 2 {
				playerName := strings.TrimSpace(parts[1])

				// Extract IP:port and channel from logger
				sessionKey := extractSessionKey(entry.Logger)
				if sessionKey != "" {
					session, exists := sessions[sessionKey]
					if !exists {
						session = &PlayerSession{
							Name:       playerName,
							IPPort:     extractIPPort(entry.Logger),
							Channel:    extractChannel(entry.Logger),
							FirstSeen:  entry.Timestamp,
							Activities: make([]string, 0),
							Stages:     make([]string, 0),
							Objects:    make([]string, 0),
						}
						sessions[sessionKey] = session

						stats.TotalConnections++
						stats.UniquePlayers[playerName] = true

						if session.IPPort != "" {
							ip := strings.Split(session.IPPort, ":")[0]
							stats.UniqueIPs[ip]++
						}

						if session.Channel != "" {
							stats.ChannelDistribution[session.Channel]++
						}

						day := entry.Timestamp.Format("2006-01-02")
						stats.ConnectionsPerDay[day]++
					}

					session.LastSeen = entry.Timestamp
					session.Activities = append(session.Activities, entry.Message)
				}
			}
		}

		// Track broadcasts
		if strings.Contains(entry.Message, "Broadcasting new object:") {
			sessionKey := extractSessionKey(entry.Logger)
			if session, exists := sessions[sessionKey]; exists {
				parts := strings.Split(entry.Message, "Broadcasting new object: ")
				if len(parts) == 2 {
					session.Objects = append(session.Objects, parts[1])
				}
			}
		}

		// Track stage changes
		if strings.Contains(entry.Message, "Sending notification to old stage clients") {
			sessionKey := extractSessionKey(entry.Logger)
			if session, exists := sessions[sessionKey]; exists {
				session.Stages = append(session.Stages, "Stage changed")
			}
		}

		// Track save operations
		if strings.Contains(entry.Message, "Wrote recompressed savedata back to DB") {
			sessionKey := extractSessionKey(entry.Logger)
			if session, exists := sessions[sessionKey]; exists {
				session.SaveCount++
			}
		}

		// Track disconnections
		if strings.Contains(entry.Message, "Error on ReadPacket, exiting recv loop") ||
			strings.Contains(entry.Message, "Error reading packet") {
			sessionKey := extractSessionKey(entry.Logger)
			if session, exists := sessions[sessionKey]; exists {
				session.Errors++
			}

			// Extract disconnect reason
			if entry.Error != "" {
				reason := entry.Error
				if strings.Contains(reason, "connection reset by peer") {
					reason = "connection reset by peer"
				} else if strings.Contains(reason, "timeout") {
					reason = "timeout"
				}
				stats.DisconnectReasons[reason]++
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing log file: %v\n", err)
		os.Exit(1)
	}

	// Filter by player if specified
	if *player != "" {
		filteredSessions := make(map[string]*PlayerSession)
		for key, session := range sessions {
			if strings.Contains(strings.ToLower(session.Name), strings.ToLower(*player)) {
				filteredSessions[key] = session
			}
		}
		sessions = filteredSessions
	}

	// Display results
	if *showStats {
		printConnectionStats(stats)
	}

	if *showSessions {
		printPlayerSessions(sessions, *verbose)
	}
}

// printConnectionStats displays aggregate connection statistics in a formatted report.
//
// The report includes:
//   - Total connections and unique player/IP counts
//   - Channel distribution showing which channels are most popular
//   - Connections per day to identify peak usage days
//   - Disconnect reasons to identify common connection issues
//   - Top IP addresses by connection count
//
// All sections are sorted for easy analysis (channels alphabetically,
// days chronologically, others by frequency).
//
// Parameters:
//   - stats: ConnectionStats structure containing aggregated data
func printConnectionStats(stats *ConnectionStats) {
	fmt.Printf("=== Connection Statistics ===\n\n")
	fmt.Printf("Total Connections: %d\n", stats.TotalConnections)
	fmt.Printf("Unique Players: %d\n", len(stats.UniquePlayers))
	fmt.Printf("Unique IP Addresses: %d\n", len(stats.UniqueIPs))

	if len(stats.ChannelDistribution) > 0 {
		fmt.Printf("\n--- Channel Distribution ---\n")
		// Sort channels
		type channelPair struct {
			name  string
			count int
		}
		var channels []channelPair
		for name, count := range stats.ChannelDistribution {
			channels = append(channels, channelPair{name, count})
		}
		sort.Slice(channels, func(i, j int) bool {
			return channels[i].name < channels[j].name
		})
		for _, ch := range channels {
			fmt.Printf("  %s: %d connections\n", ch.name, ch.count)
		}
	}

	if len(stats.ConnectionsPerDay) > 0 {
		fmt.Printf("\n--- Connections Per Day ---\n")
		// Sort by date
		type dayPair struct {
			date  string
			count int
		}
		var days []dayPair
		for date, count := range stats.ConnectionsPerDay {
			days = append(days, dayPair{date, count})
		}
		sort.Slice(days, func(i, j int) bool {
			return days[i].date < days[j].date
		})
		for _, day := range days {
			fmt.Printf("  %s: %d connections\n", day.date, day.count)
		}
	}

	if len(stats.DisconnectReasons) > 0 {
		fmt.Printf("\n--- Disconnect Reasons ---\n")
		// Sort by count
		type reasonPair struct {
			reason string
			count  int
		}
		var reasons []reasonPair
		for reason, count := range stats.DisconnectReasons {
			reasons = append(reasons, reasonPair{reason, count})
		}
		sort.Slice(reasons, func(i, j int) bool {
			return reasons[i].count > reasons[j].count
		})
		for _, r := range reasons {
			fmt.Printf("  %s: %d times\n", r.reason, r.count)
		}
	}

	if len(stats.UniqueIPs) > 0 {
		fmt.Printf("\n--- Top IP Addresses ---\n")
		type ipPair struct {
			ip    string
			count int
		}
		var ips []ipPair
		for ip, count := range stats.UniqueIPs {
			ips = append(ips, ipPair{ip, count})
		}
		sort.Slice(ips, func(i, j int) bool {
			return ips[i].count > ips[j].count
		})
		// Show top 10
		limit := 10
		if len(ips) < limit {
			limit = len(ips)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("  %s: %d connections\n", ips[i].ip, ips[i].count)
		}
	}
}

// printPlayerSessions displays detailed information about individual player sessions.
//
// For each session, displays:
//   - Player name, channel, and IP:port
//   - Connection duration (first seen to last seen)
//   - Number of save operations and errors
//   - Objects and stage changes (if verbose=true)
//
// Sessions are sorted chronologically by first seen time.
//
// Parameters:
//   - sessions: Map of session keys to PlayerSession data
//   - verbose: Whether to show detailed activity information
func printPlayerSessions(sessions map[string]*PlayerSession, verbose bool) {
	// Sort sessions by first seen
	type sessionPair struct {
		key     string
		session *PlayerSession
	}
	var pairs []sessionPair
	for key, session := range sessions {
		pairs = append(pairs, sessionPair{key, session})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].session.FirstSeen.Before(pairs[j].session.FirstSeen)
	})

	fmt.Printf("\n\n=== Player Sessions ===\n")
	fmt.Printf("Total Sessions: %d\n\n", len(sessions))

	for idx, pair := range pairs {
		session := pair.session
		duration := session.LastSeen.Sub(session.FirstSeen)

		fmt.Printf("%s\n", strings.Repeat("-", 80))
		fmt.Printf("Session #%d: %s\n", idx+1, session.Name)
		fmt.Printf("%s\n", strings.Repeat("-", 80))
		fmt.Printf("Channel: %s\n", session.Channel)
		fmt.Printf("IP:Port: %s\n", session.IPPort)
		fmt.Printf("First Seen: %s\n", session.FirstSeen.Format("2006-01-02 15:04:05"))
		fmt.Printf("Last Seen: %s\n", session.LastSeen.Format("2006-01-02 15:04:05"))
		fmt.Printf("Duration: %s\n", formatDuration(duration))
		fmt.Printf("Save Operations: %d\n", session.SaveCount)
		fmt.Printf("Errors: %d\n", session.Errors)

		if verbose {
			if len(session.Objects) > 0 {
				fmt.Printf("\nObjects: %s\n", strings.Join(session.Objects, ", "))
			}

			if len(session.Stages) > 0 {
				fmt.Printf("Stage Changes: %d\n", len(session.Stages))
			}
		}
		fmt.Println()
	}
}

// extractSessionKey extracts a unique session identifier from a logger string.
//
// Logger format: "main.channel-X.IP:port"
// Returns: "channel-X.IP:port"
//
// This key uniquely identifies a player session by combining the channel
// and the client's IP:port combination.
//
// Parameters:
//   - logger: The logger field from a log entry
//
// Returns:
//   - A session key string, or empty string if the format is invalid
func extractSessionKey(logger string) string {
	// Logger format: "main.channel-X.IP:port"
	parts := strings.Split(logger, ".")
	if len(parts) >= 3 {
		return strings.Join(parts[1:], ".")
	}
	return ""
}

// extractIPPort extracts the client IP address and port from a logger string.
//
// Logger format: "main.channel-X.A.B.C.D:port" where A.B.C.D is the IPv4 address
// Returns: "A.B.C.D:port"
//
// Parameters:
//   - logger: The logger field from a log entry
//
// Returns:
//   - The IP:port string, or empty string if extraction fails
func extractIPPort(logger string) string {
	parts := strings.Split(logger, ".")
	if len(parts) >= 4 {
		// Last part might be IP:port
		lastPart := parts[len(parts)-1]
		if strings.Contains(lastPart, ":") {
			// Reconstruct IP:port (handle IPv4)
			if len(parts) >= 4 {
				ip := strings.Join(parts[len(parts)-4:len(parts)-1], ".")
				port := lastPart
				return ip + ":" + port
			}
		}
	}
	return ""
}

// extractChannel extracts the channel name from a logger string.
//
// Logger format: "main.channel-X.IP:port"
// Returns: "channel-X"
//
// Parameters:
//   - logger: The logger field from a log entry
//
// Returns:
//   - The channel name (e.g., "channel-4"), or empty string if not found
func extractChannel(logger string) string {
	if strings.Contains(logger, "channel-") {
		parts := strings.Split(logger, "channel-")
		if len(parts) >= 2 {
			channelPart := strings.Split(parts[1], ".")[0]
			return "channel-" + channelPart
		}
	}
	return ""
}

// formatDuration formats a time duration into a human-readable string.
//
// The format varies based on duration:
//   - Less than 1 minute: "N seconds"
//   - Less than 1 hour: "N.N minutes"
//   - 1 hour or more: "N.N hours"
//
// Parameters:
//   - d: The duration to format
//
// Returns:
//   - A human-readable string representation of the duration
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
}
