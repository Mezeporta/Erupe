package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// newTestRouter builds the same mux router as Start() without starting an HTTP server.
func newTestRouter(s *APIServer) *mux.Router {
	r := mux.NewRouter()

	// Legacy routes
	r.HandleFunc("/launcher", s.Launcher)
	r.HandleFunc("/login", s.Login)
	r.HandleFunc("/register", s.Register)
	r.HandleFunc("/character/create", s.CreateCharacter)
	r.HandleFunc("/character/delete", s.DeleteCharacter)
	r.HandleFunc("/character/export", s.ExportSave)
	r.HandleFunc("/health", s.Health)
	r.HandleFunc("/version", s.Version)

	// V2 routes
	v2 := r.PathPrefix("/v2").Subrouter()
	v2.HandleFunc("/login", s.Login).Methods("POST")
	v2.HandleFunc("/register", s.Register).Methods("POST")
	v2.HandleFunc("/launcher", s.Launcher).Methods("GET")
	v2.HandleFunc("/version", s.Version).Methods("GET")
	v2.HandleFunc("/health", s.Health).Methods("GET")

	// V2 authenticated routes
	v2Auth := v2.PathPrefix("").Subrouter()
	v2Auth.Use(s.AuthMiddleware)
	v2Auth.HandleFunc("/characters", s.CreateCharacter).Methods("POST")
	v2Auth.HandleFunc("/characters/{id}/delete", s.DeleteCharacter).Methods("POST")
	v2Auth.HandleFunc("/characters/{id}/export", s.ExportSave).Methods("GET")

	v2.HandleFunc("/server/status", s.ServerStatus).Methods("GET")

	return r
}

func TestV2LoginRoute(t *testing.T) {
	logger := NewTestLogger(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)

	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		userRepo: &mockAPIUserRepo{
			credentialsID:       1,
			credentialsPassword: string(hash),
			credentialsRights:   30,
			lastLogin:           time.Now(),
			returnExpiry:        time.Now().Add(time.Hour * 24 * 30),
		},
		sessionRepo: &mockAPISessionRepo{createTokenID: 1},
		charRepo:    &mockAPICharacterRepo{characters: []Character{}},
	}

	router := newTestRouter(server)

	body, _ := json.Marshal(map[string]string{"username": "user", "password": "pass"})
	req := httptest.NewRequest("POST", "/v2/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /v2/login: status = %d, want 200", rec.Code)
	}
}

func TestV2LoginRejectsGET(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("GET", "/v2/login", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// gorilla/mux subrouters return 404 for method mismatches (not 405)
	if rec.Code == http.StatusOK {
		t.Error("GET /v2/login should not succeed (POST only)")
	}
}

func TestV2HealthRoute(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		db:          nil,
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("GET", "/v2/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// 503 because db is nil, but it should route correctly
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("GET /v2/health: status = %d, want 503", rec.Code)
	}
}

func TestV2VersionRoute(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("GET", "/v2/version", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /v2/version: status = %d, want 200", rec.Code)
	}
}

func TestV2CharactersRequiresAuth(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		sessionRepo: &mockAPISessionRepo{
			userIDErr: sql.ErrNoRows,
		},
	}

	router := newTestRouter(server)

	// No auth header
	req := httptest.NewRequest("POST", "/v2/characters", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST /v2/characters (no auth): status = %d, want 401", rec.Code)
	}
}

func TestV2CharactersWithAuth(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		sessionRepo: &mockAPISessionRepo{userID: 1},
		charRepo: &mockAPICharacterRepo{
			newCharacter: Character{ID: 5, Name: "NewChar"},
		},
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("POST", "/v2/characters", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /v2/characters (with auth): status = %d, want 200", rec.Code)
	}
}

func TestV2DeleteCharacterRoute(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		sessionRepo: &mockAPISessionRepo{userID: 1},
		charRepo:    &mockAPICharacterRepo{isNewResult: true},
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("POST", "/v2/characters/5/delete", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /v2/characters/5/delete: status = %d, want 200", rec.Code)
	}
}

func TestV2ExportCharacterRoute(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		sessionRepo: &mockAPISessionRepo{userID: 1},
		charRepo: &mockAPICharacterRepo{
			exportResult: map[string]interface{}{"name": "Hunter"},
		},
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("GET", "/v2/characters/1/export", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /v2/characters/1/export: status = %d, want 200", rec.Code)
	}
}

func TestV2ServerStatusRoute_NoEventRepo(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("GET", "/v2/server/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /v2/server/status: status = %d, want 200", rec.Code)
	}

	var resp ServerStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.MezFes == nil {
		t.Error("MezFes should not be nil")
	}
	if resp.FeaturedWeapon != nil {
		t.Error("FeaturedWeapon should be nil without event repo")
	}
	if resp.Events.FestaActive || resp.Events.DivaActive {
		t.Error("events should be inactive without event repo")
	}
}

func TestV2ServerStatusRoute_WithEvents(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		eventRepo: &mockAPIEventRepo{
			featureWeapon: &FeatureWeaponRow{
				StartTime:      time.Now(),
				ActiveFeatures: 0xFF,
			},
			events: []EventRow{{ID: 1, StartTime: time.Now().Unix()}},
		},
	}

	router := newTestRouter(server)

	req := httptest.NewRequest("GET", "/v2/server/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp ServerStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.FeaturedWeapon == nil {
		t.Fatal("FeaturedWeapon should not be nil")
	}
	if resp.FeaturedWeapon.ActiveFeatures != 0xFF {
		t.Errorf("ActiveFeatures = %d, want 255", resp.FeaturedWeapon.ActiveFeatures)
	}
	// Both festa and diva use the same mock events slice, so both are active
	if !resp.Events.FestaActive {
		t.Error("FestaActive should be true")
	}
	if !resp.Events.DivaActive {
		t.Error("DivaActive should be true")
	}
}

func TestLegacyRoutesStillWork(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
	}

	router := newTestRouter(server)

	// Legacy /version should work with GET
	req := httptest.NewRequest("GET", "/version", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /version (legacy): status = %d, want 200", rec.Code)
	}
}
