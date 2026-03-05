package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"go.uber.org/zap"
)

//go:embed dashboard.html
var dashboardHTML string

var dashboardTmpl = template.Must(template.New("dashboard").Parse(dashboardHTML))

// DashboardStats is the JSON payload returned by GET /api/dashboard/stats.
type DashboardStats struct {
	Uptime          string        `json:"uptime"`
	ServerVersion   string        `json:"serverVersion"`
	ClientMode      string        `json:"clientMode"`
	OnlinePlayers   int           `json:"onlinePlayers"`
	TotalAccounts   int           `json:"totalAccounts"`
	TotalCharacters int           `json:"totalCharacters"`
	Channels        []ChannelInfo `json:"channels"`
	DatabaseOK      bool          `json:"databaseOK"`
}

// ChannelInfo describes a single channel server entry from the servers table.
type ChannelInfo struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Players int    `json:"players"`
}

// Dashboard serves the embedded HTML dashboard page at /dashboard.
func (s *APIServer) Dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTmpl.Execute(w, nil); err != nil {
		s.logger.Error("Failed to render dashboard", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// DashboardStatsJSON serves GET /api/dashboard/stats with live server statistics.
func (s *APIServer) DashboardStatsJSON(w http.ResponseWriter, r *http.Request) {
	stats := DashboardStats{
		ServerVersion: "Erupe-CE",
		ClientMode:    s.erupeConfig.ClientMode,
	}

	// Compute uptime.
	if !s.startTime.IsZero() {
		stats.Uptime = formatDuration(time.Since(s.startTime))
	} else {
		stats.Uptime = "unknown"
	}

	// Check database connectivity.
	if s.db != nil {
		if err := s.db.Ping(); err != nil {
			s.logger.Warn("Dashboard: database ping failed", zap.Error(err))
			stats.DatabaseOK = false
		} else {
			stats.DatabaseOK = true
		}
	}

	// Query total accounts.
	if s.db != nil {
		if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalAccounts); err != nil {
			s.logger.Warn("Dashboard: failed to count users", zap.Error(err))
		}
	}

	// Query total characters.
	if s.db != nil {
		if err := s.db.QueryRow("SELECT COUNT(*) FROM characters").Scan(&stats.TotalCharacters); err != nil {
			s.logger.Warn("Dashboard: failed to count characters", zap.Error(err))
		}
	}

	// Query channel info from servers table.
	if s.db != nil {
		rows, err := s.db.Query("SELECT server_id, current_players, world_name, land FROM servers ORDER BY server_id")
		if err != nil {
			s.logger.Warn("Dashboard: failed to query servers", zap.Error(err))
		} else {
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				var serverID, players, land int
				var worldName *string
				if err := rows.Scan(&serverID, &players, &worldName, &land); err != nil {
					s.logger.Warn("Dashboard: failed to scan server row", zap.Error(err))
					continue
				}
				name := "Channel"
				if worldName != nil {
					name = *worldName
				}
				ch := ChannelInfo{
					Name:    name,
					Port:    54000 + serverID,
					Players: players,
				}
				stats.Channels = append(stats.Channels, ch)
				stats.OnlinePlayers += players
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Error("Dashboard: failed to encode stats", zap.Error(err))
	}
}

// formatDuration produces a human-readable duration string like "2d 5h 32m 10s".
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
