package setup

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"

	"erupe-ce/server/migrations"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

//go:embed wizard.html
var wizardHTML embed.FS

// wizardServer holds state for the setup wizard HTTP handlers.
type wizardServer struct {
	logger *zap.Logger
	done   chan struct{} // closed when setup is complete
}

func (ws *wizardServer) handleIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := wizardHTML.ReadFile("wizard.html")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (ws *wizardServer) handleDetectIP(w http.ResponseWriter, _ *http.Request) {
	ip, err := detectOutboundIP()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ip": ip})
}

func (ws *wizardServer) handleClientModes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"modes": clientModes()})
}

func (ws *wizardServer) handleCheckQuests(w http.ResponseWriter, _ *http.Request) {
	status := checkQuestFiles("")
	writeJSON(w, http.StatusOK, status)
}

func (ws *wizardServer) handlePresets(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"presets": availablePresets()})
}

// testDBRequest is the JSON body for POST /api/setup/test-db.
type testDBRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
}

func (ws *wizardServer) handleTestDB(w http.ResponseWriter, r *http.Request) {
	var req testDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	status, err := testDBConnection(req.Host, req.Port, req.User, req.Password, req.DBName)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":  err.Error(),
			"status": status,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status})
}

// initDBRequest is the JSON body for POST /api/setup/init-db.
type initDBRequest struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DBName       string `json:"dbName"`
	CreateDB     bool   `json:"createDB"`
	ApplySchema  bool   `json:"applySchema"`
	ApplyBundled bool   `json:"applyBundled"`
}

func (ws *wizardServer) handleInitDB(w http.ResponseWriter, r *http.Request) {
	var req initDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	var log []string
	addLog := func(msg string) {
		log = append(log, msg)
		ws.logger.Info(msg)
	}

	if req.CreateDB {
		addLog(fmt.Sprintf("Creating database '%s'...", req.DBName))
		if err := createDatabase(req.Host, req.Port, req.User, req.Password, req.DBName); err != nil {
			addLog(fmt.Sprintf("ERROR: %s", err))
			writeJSON(w, http.StatusOK, map[string]interface{}{"success": false, "log": log})
			return
		}
		addLog("Database created successfully")
	}

	if req.ApplySchema || req.ApplyBundled {
		connStr := fmt.Sprintf(
			"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
			req.Host, req.Port, req.User, req.Password, req.DBName,
		)
		db, err := sqlx.Open("postgres", connStr)
		if err != nil {
			addLog(fmt.Sprintf("ERROR connecting to database: %s", err))
			writeJSON(w, http.StatusOK, map[string]interface{}{"success": false, "log": log})
			return
		}
		defer func() { _ = db.Close() }()

		if req.ApplySchema {
			addLog("Running database migrations...")
			applied, err := migrations.Migrate(db, ws.logger)
			if err != nil {
				addLog(fmt.Sprintf("ERROR: %s", err))
				writeJSON(w, http.StatusOK, map[string]interface{}{"success": false, "log": log})
				return
			}
			addLog(fmt.Sprintf("Schema migrations applied (%d migration(s))", applied))
		}

		if req.ApplyBundled {
			addLog("Applying bundled data (shops, events, gacha)...")
			applied, err := migrations.ApplySeedData(db, ws.logger)
			if err != nil {
				addLog(fmt.Sprintf("ERROR: %s", err))
				writeJSON(w, http.StatusOK, map[string]interface{}{"success": false, "log": log})
				return
			}
			addLog(fmt.Sprintf("Bundled data applied (%d files)", applied))
		}
	}

	addLog("Database initialization complete!")
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "log": log})
}

func (ws *wizardServer) handleFinish(w http.ResponseWriter, r *http.Request) {
	var req FinishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	config := buildDefaultConfig(req)
	if err := writeConfig(config); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ws.logger.Info("config.json written successfully")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	// Signal completion — this will cause the HTTP server to shut down.
	close(ws.done)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
