package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// runList lists all currently connected users across all channels.
//
// Queries the sign_sessions table joined with characters and servers
// to show all active sessions with character details.
//
// Options:
//   - Database connection flags (host, port, user, password, dbname)
//   - v: Verbose output with additional details
func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)
	verbose := fs.Bool("v", false, "Verbose output with additional details")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	db, err := connectDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	query := `
		SELECT
			ss.char_id,
			COALESCE(c.name, '') as char_name,
			COALESCE(ss.server_id, 0) as server_id,
			COALESCE(s.world_name, 'Unknown') as world_name,
			COALESCE(c.user_id, 0) as user_id,
			COALESCE(u.username, '') as username,
			c.last_login,
			COALESCE(c.hrp, 0) as hr,
			COALESCE(c.gr, 0) as gr
		FROM sign_sessions ss
		LEFT JOIN characters c ON ss.char_id = c.id
		LEFT JOIN servers s ON ss.server_id = s.server_id
		LEFT JOIN users u ON c.user_id = u.id
		WHERE ss.char_id IS NOT NULL AND ss.server_id IS NOT NULL
		ORDER BY s.world_name, c.name
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = rows.Close() }()

	var users []ConnectedUser
	for rows.Next() {
		var u ConnectedUser
		err := rows.Scan(&u.CharID, &u.CharName, &u.ServerID, &u.ServerName, &u.UserID, &u.Username, &u.LastLogin, &u.HR, &u.GR)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
			continue
		}
		users = append(users, u)
	}

	if len(users) == 0 {
		fmt.Println("No users currently connected.")
		return
	}

	fmt.Printf("=== Connected Users (%d total) ===\n\n", len(users))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if *verbose {
		_, _ = fmt.Fprintln(w, "CHAR ID\tNAME\tSERVER\tHR\tGR\tUSERNAME\tLAST LOGIN")
		_, _ = fmt.Fprintln(w, "-------\t----\t------\t--\t--\t--------\t----------")
		for _, u := range users {
			lastLogin := "N/A"
			if u.LastLogin.Valid {
				lastLogin = u.LastLogin.Time.Format("2006-01-02 15:04:05")
			}
			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%s\t%s\n",
				u.CharID, u.CharName, u.ServerName, u.HR, u.GR, u.Username, lastLogin)
		}
	} else {
		_, _ = fmt.Fprintln(w, "CHAR ID\tNAME\tSERVER\tHR\tGR")
		_, _ = fmt.Fprintln(w, "-------\t----\t------\t--\t--")
		for _, u := range users {
			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\n",
				u.CharID, u.CharName, u.ServerName, u.HR, u.GR)
		}
	}
	_ = w.Flush()
}

// runCount shows the count of connected users per server/channel.
//
// Options:
//   - Database connection flags (host, port, user, password, dbname)
func runCount(args []string) {
	fs := flag.NewFlagSet("count", flag.ExitOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	db, err := connectDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	// Get total from sign_sessions
	var totalConnected int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sign_sessions
		WHERE char_id IS NOT NULL AND server_id IS NOT NULL
	`).Scan(&totalConnected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying total: %v\n", err)
		os.Exit(1)
	}

	// Get count per server
	query := `
		SELECT
			s.server_id,
			s.world_name,
			s.current_players,
			s.land
		FROM servers s
		ORDER BY s.world_name, s.land
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying servers: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = rows.Close() }()

	fmt.Printf("=== Connected Users Summary ===\n\n")
	fmt.Printf("Total Connected: %d\n\n", totalConnected)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "SERVER ID\tWORLD\tLAND\tPLAYERS")
	_, _ = fmt.Fprintln(w, "---------\t-----\t----\t-------")

	for rows.Next() {
		var serverID uint16
		var worldName string
		var currentPlayers, land int
		if err := rows.Scan(&serverID, &worldName, &currentPlayers, &land); err != nil {
			continue
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", serverID, worldName, land, currentPlayers)
	}
	_ = w.Flush()
}

// runSearch searches for a specific connected user by name.
//
// Options:
//   - name: Player name to search for (partial match, case-insensitive)
//   - Database connection flags (host, port, user, password, dbname)
func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)
	name := fs.String("name", "", "Player name to search for (partial match)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *name == "" {
		fmt.Fprintf(os.Stderr, "Error: -name flag is required\n")
		fs.Usage()
		os.Exit(1)
	}

	db, err := connectDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	query := `
		SELECT
			ss.char_id,
			COALESCE(c.name, '') as char_name,
			COALESCE(ss.server_id, 0) as server_id,
			COALESCE(s.world_name, 'Unknown') as world_name,
			COALESCE(c.user_id, 0) as user_id,
			COALESCE(u.username, '') as username,
			c.last_login,
			COALESCE(c.hrp, 0) as hr,
			COALESCE(c.gr, 0) as gr
		FROM sign_sessions ss
		LEFT JOIN characters c ON ss.char_id = c.id
		LEFT JOIN servers s ON ss.server_id = s.server_id
		LEFT JOIN users u ON c.user_id = u.id
		WHERE ss.char_id IS NOT NULL
			AND ss.server_id IS NOT NULL
			AND LOWER(c.name) LIKE LOWER($1)
		ORDER BY c.name
	`

	rows, err := db.Query(query, "%"+*name+"%")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = rows.Close() }()

	var users []ConnectedUser
	for rows.Next() {
		var u ConnectedUser
		err := rows.Scan(&u.CharID, &u.CharName, &u.ServerID, &u.ServerName, &u.UserID, &u.Username, &u.LastLogin, &u.HR, &u.GR)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	if len(users) == 0 {
		fmt.Printf("No connected users found matching '%s'\n", *name)
		return
	}

	fmt.Printf("=== Search Results for '%s' (%d found) ===\n\n", *name, len(users))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "CHAR ID\tNAME\tSERVER\tHR\tGR\tUSERNAME")
	_, _ = fmt.Fprintln(w, "-------\t----\t------\t--\t--\t--------")
	for _, u := range users {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%s\n",
			u.CharID, u.CharName, u.ServerName, u.HR, u.GR, u.Username)
	}
	_ = w.Flush()
}

// runServers shows server/channel status and player counts.
//
// Options:
//   - Database connection flags (host, port, user, password, dbname)
func runServers(args []string) {
	fs := flag.NewFlagSet("servers", flag.ExitOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	db, err := connectDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	query := `
		SELECT
			server_id,
			world_name,
			world_description,
			land,
			current_players,
			season
		FROM servers
		ORDER BY world_name, land
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying servers: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = rows.Close() }()

	var servers []ServerStatus
	var totalPlayers int
	for rows.Next() {
		var s ServerStatus
		if err := rows.Scan(&s.ServerID, &s.WorldName, &s.WorldDesc, &s.Land, &s.CurrentPlayers, &s.Season); err != nil {
			continue
		}
		servers = append(servers, s)
		totalPlayers += s.CurrentPlayers
	}

	if len(servers) == 0 {
		fmt.Println("No servers found. Is the Erupe server running?")
		return
	}

	fmt.Printf("=== Server Status ===\n\n")
	fmt.Printf("Total Servers: %d\n", len(servers))
	fmt.Printf("Total Players: %d\n\n", totalPlayers)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tWORLD\tLAND\tPLAYERS\tSEASON\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "--\t-----\t----\t-------\t------\t-----------")

	seasonNames := map[int]string{0: "Green", 1: "Orange", 2: "Blue"}
	for _, s := range servers {
		seasonName := seasonNames[s.Season]
		if seasonName == "" {
			seasonName = fmt.Sprintf("%d", s.Season)
		}
		// Truncate description if too long
		desc := s.WorldDesc
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%s\t%s\n",
			s.ServerID, s.WorldName, s.Land, s.CurrentPlayers, seasonName, desc)
	}
	_ = w.Flush()
}

// runHistory shows recent login history for a player.
//
// Options:
//   - name: Player name to search for (partial match, case-insensitive)
//   - limit: Maximum number of entries to show (default: 20)
//   - all: Show all characters, not just with recent logins
//   - Database connection flags (host, port, user, password, dbname)
func runHistory(args []string) {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)
	name := fs.String("name", "", "Player name to search for (partial match)")
	limit := fs.Int("limit", 20, "Maximum number of entries to show")
	showAll := fs.Bool("all", false, "Show all characters (including those without recent logins)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *name == "" {
		fmt.Fprintf(os.Stderr, "Error: -name flag is required\n")
		fs.Usage()
		os.Exit(1)
	}

	db, err := connectDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	whereClause := "WHERE LOWER(c.name) LIKE LOWER($1)"
	if !*showAll {
		whereClause += " AND c.last_login IS NOT NULL"
	}

	query := fmt.Sprintf(`
		SELECT
			c.id as char_id,
			c.name as char_name,
			c.last_login,
			COALESCE(c.hrp, 0) as hr,
			COALESCE(c.gr, 0) as gr,
			COALESCE(u.username, '') as username
		FROM characters c
		LEFT JOIN users u ON c.user_id = u.id
		%s
		ORDER BY c.last_login DESC NULLS LAST
		LIMIT $2
	`, whereClause)

	rows, err := db.Query(query, "%"+*name+"%", *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = rows.Close() }()

	var history []LoginHistory
	for rows.Next() {
		var h LoginHistory
		if err := rows.Scan(&h.CharID, &h.CharName, &h.LastLogin, &h.HR, &h.GR, &h.Username); err != nil {
			continue
		}
		history = append(history, h)
	}

	if len(history) == 0 {
		fmt.Printf("No characters found matching '%s'\n", *name)
		return
	}

	fmt.Printf("=== Login History for '%s' (%d entries) ===\n\n", *name, len(history))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "CHAR ID\tNAME\tHR\tGR\tUSERNAME\tLAST LOGIN\tONLINE")
	_, _ = fmt.Fprintln(w, "-------\t----\t--\t--\t--------\t----------\t------")

	for _, h := range history {
		lastLogin := "Never"
		online := "No"
		if h.LastLogin.Valid {
			lastLogin = h.LastLogin.Time.Format("2006-01-02 15:04:05")
			// Check if logged in within last hour (rough online indicator)
			if time.Since(h.LastLogin.Time) < time.Hour {
				online = "Possibly"
			}
		}

		// Check if actually online by checking sign_sessions
		var isOnline bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM sign_sessions WHERE char_id = $1 AND server_id IS NOT NULL)", h.CharID).Scan(&isOnline); err == nil && isOnline {
			online = "Yes"
		}

		_, _ = fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%s\t%s\t%s\n",
			h.CharID, h.CharName, h.HR, h.GR, h.Username, lastLogin, online)
	}
	_ = w.Flush()

	// Show legend
	fmt.Println()
	fmt.Println("ONLINE column: Yes = Currently connected, Possibly = Last login within 1 hour, No = Not connected")
}
