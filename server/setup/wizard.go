package setup

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/lib/pq"
)

// clientModes returns all supported client version strings.
func clientModes() []string {
	return []string{
		"S1.0", "S1.5", "S2.0", "S2.5", "S3.0", "S3.5", "S4.0", "S5.0", "S5.5", "S6.0", "S7.0",
		"S8.0", "S8.5", "S9.0", "S10", "FW.1", "FW.2", "FW.3", "FW.4", "FW.5", "G1", "G2", "G3",
		"G3.1", "G3.2", "GG", "G5", "G5.1", "G5.2", "G6", "G6.1", "G7", "G8", "G8.1", "G9", "G9.1",
		"G10", "G10.1", "Z1", "Z2", "ZZ",
	}
}

// FinishRequest holds the user's configuration choices from the wizard.
type FinishRequest struct {
	DBHost            string `json:"dbHost"`
	DBPort            int    `json:"dbPort"`
	DBUser            string `json:"dbUser"`
	DBPassword        string `json:"dbPassword"`
	DBName            string `json:"dbName"`
	Host              string `json:"host"`
	Language          string `json:"language"`
	ClientMode        string `json:"clientMode"`
	AutoCreateAccount bool   `json:"autoCreateAccount"`
	Preset            string `json:"preset"`
}

// buildDefaultConfig produces a minimal config map with only user-provided values.
// All other settings are filled by Viper's registered defaults at load time.
func buildDefaultConfig(req FinishRequest) map[string]interface{} {
	lang := req.Language
	if lang == "" {
		lang = "jp"
	}
	cfg := map[string]interface{}{
		"Host":              req.Host,
		"Language":          lang,
		"ClientMode":        req.ClientMode,
		"AutoCreateAccount": req.AutoCreateAccount,
		"Database": map[string]interface{}{
			"Host":     req.DBHost,
			"Port":     req.DBPort,
			"User":     req.DBUser,
			"Password": req.DBPassword,
			"Database": req.DBName,
		},
	}

	// Apply preset overrides. The "community" preset uses Viper defaults and
	// adds nothing to the config file.
	if overrides, ok := presetConfigs()[req.Preset]; ok {
		for k, v := range overrides {
			cfg[k] = v
		}
	}

	return cfg
}

// presetConfigs returns config overrides keyed by preset ID.
// The "community" preset is intentionally absent — it relies entirely on
// Viper defaults.
func presetConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"solo": {
			"GameplayOptions": map[string]interface{}{
				"HRPMultiplier":       3.0,
				"SRPMultiplier":       3.0,
				"GRPMultiplier":       3.0,
				"GSRPMultiplier":      3.0,
				"ZennyMultiplier":     2.0,
				"GZennyMultiplier":    2.0,
				"MaterialMultiplier":  2.0,
				"GMaterialMultiplier": 2.0,
				"ExtraCarves":         2,
				"GExtraCarves":        2,
			},
			"Entrance": map[string]interface{}{
				"Entries": []map[string]interface{}{
					{
						"Name": "Solo",
						"Type": 1,
						"Channels": []map[string]interface{}{
							{"Port": 54001, "MaxPlayers": 100},
						},
					},
				},
			},
		},
		"small": {
			"Entrance": map[string]interface{}{
				"Entries": []map[string]interface{}{
					{
						"Name": "World 1",
						"Type": 1,
						"Channels": []map[string]interface{}{
							{"Port": 54001, "MaxPlayers": 100},
							{"Port": 54002, "MaxPlayers": 100},
						},
					},
				},
			},
		},
		"rebalanced": {
			"GameplayOptions": map[string]interface{}{
				"HRPMultiplier":  2.0,
				"SRPMultiplier":  2.0,
				"GRPMultiplier":  2.0,
				"GSRPMultiplier": 2.0,
				"ExtraCarves":    1,
				"GExtraCarves":   1,
			},
			"Entrance": map[string]interface{}{
				"Entries": []map[string]interface{}{
					{
						"Name": "Normal",
						"Type": 1,
						"Channels": []map[string]interface{}{
							{"Port": 54001, "MaxPlayers": 100},
							{"Port": 54002, "MaxPlayers": 100},
						},
					},
					{
						"Name": "Cities",
						"Type": 2,
						"Channels": []map[string]interface{}{
							{"Port": 54003, "MaxPlayers": 100},
							{"Port": 54004, "MaxPlayers": 100},
						},
					},
				},
			},
		},
	}
}

// QuestStatus holds the result of a quest files check.
type QuestStatus struct {
	QuestsFound bool `json:"questsFound"`
	QuestCount  int  `json:"questCount"`
}

// checkQuestFiles checks if quest files exist in the bin/quests/ directory.
func checkQuestFiles(binPath string) QuestStatus {
	if binPath == "" {
		binPath = "bin"
	}
	questDir := filepath.Join(binPath, "quests")
	entries, err := os.ReadDir(questDir)
	if err != nil {
		return QuestStatus{QuestsFound: false, QuestCount: 0}
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return QuestStatus{QuestsFound: count > 0, QuestCount: count}
}

// PresetInfo describes a gameplay preset for the wizard UI.
type PresetInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Channels    int    `json:"channels"`
}

// availablePresets returns the list of gameplay presets shown in the wizard.
func availablePresets() []PresetInfo {
	return []PresetInfo{
		{
			ID:          "solo",
			Name:        "Solo / Testing",
			Description: "Single channel, boosted XP rates (3x), relaxed grind. Ideal for solo play or development testing.",
			Channels:    1,
		},
		{
			ID:          "small",
			Name:        "Small Group (2-8 players)",
			Description: "Two channels with vanilla rates. Good for friends playing together.",
			Channels:    2,
		},
		{
			ID:          "community",
			Name:        "Community Server",
			Description: "Full 8-channel topology with vanilla rates. Ready for a public community.",
			Channels:    8,
		},
		{
			ID:          "rebalanced",
			Name:        "Rebalanced",
			Description: "Community-tuned rates: 2x GRP, 2x HRP, extra carves. Addresses G-Rank grind without trivializing content.",
			Channels:    4,
		},
	}
}

// writeConfig writes the config map to config.json with pretty formatting.
func writeConfig(config map[string]interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	if err := os.WriteFile("config.json", data, 0600); err != nil {
		return fmt.Errorf("writing config.json: %w", err)
	}
	return nil
}

// detectOutboundIP returns the preferred outbound IPv4 address.
func detectOutboundIP() (string, error) {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("detecting outbound IP: %w", err)
	}
	defer func() { _ = conn.Close() }()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.To4().String(), nil
}

// testDBConnection tests connectivity to the PostgreSQL server and checks
// whether the target database and its tables exist.
func testDBConnection(host string, port int, user, password, dbName string) (*DBStatus, error) {
	status := &DBStatus{}

	// Connect to the 'postgres' maintenance DB to check if target DB exists.
	adminConn := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='postgres' sslmode=disable",
		host, port, user, password,
	)
	adminDB, err := sql.Open("postgres", adminConn)
	if err != nil {
		return nil, fmt.Errorf("connecting to PostgreSQL: %w", err)
	}
	defer func() { _ = adminDB.Close() }()

	if err := adminDB.Ping(); err != nil {
		return nil, fmt.Errorf("cannot reach PostgreSQL: %w", err)
	}
	status.ServerReachable = true

	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return status, fmt.Errorf("checking database existence: %w", err)
	}
	status.DatabaseExists = exists

	if !exists {
		return status, nil
	}

	// Connect to the target DB to check for tables.
	targetConn := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
		host, port, user, password, dbName,
	)
	targetDB, err := sql.Open("postgres", targetConn)
	if err != nil {
		return status, nil
	}
	defer func() { _ = targetDB.Close() }()

	var tableCount int
	err = targetDB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		return status, nil
	}
	status.TablesExist = tableCount > 0
	status.TableCount = tableCount

	return status, nil
}

// DBStatus holds the result of a database connectivity check.
type DBStatus struct {
	ServerReachable bool `json:"serverReachable"`
	DatabaseExists  bool `json:"databaseExists"`
	TablesExist     bool `json:"tablesExist"`
	TableCount      int  `json:"tableCount"`
}

// createDatabase creates the target database by connecting to the 'postgres' maintenance DB.
func createDatabase(host string, port int, user, password, dbName string) error {
	adminConn := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='postgres' sslmode=disable",
		host, port, user, password,
	)
	db, err := sql.Open("postgres", adminConn)
	if err != nil {
		return fmt.Errorf("connecting to PostgreSQL: %w", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName))
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	return nil
}
