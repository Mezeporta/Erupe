package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"erupe-ce/common/gametime"
	"erupe-ce/common/mhfcourse"
	cfg "erupe-ce/config"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Notification type constants for launcher messages.
const (
	// NotificationDefault represents a standard notification.
	NotificationDefault = iota
	// NotificationNew represents a new/unread notification.
	NotificationNew
)

// LauncherResponse is the JSON payload returned by the /launcher endpoint,
// containing banners, messages, and links for the game launcher UI.
type LauncherResponse struct {
	Banners  []cfg.APISignBanner  `json:"banners"`
	Messages []cfg.APISignMessage `json:"messages"`
	Links    []cfg.APISignLink    `json:"links"`
}

// User represents an authenticated user's session credentials and permissions.
type User struct {
	TokenID uint32 `json:"tokenId"`
	Token   string `json:"token"`
	Rights  uint32 `json:"rights"`
}

// Character represents a player character's summary data as returned by the API.
type Character struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	IsFemale  bool   `json:"isFemale" db:"is_female"`
	Weapon    uint32 `json:"weapon" db:"weapon_type"`
	HR        uint32 `json:"hr" db:"hr"`
	GR        uint32 `json:"gr"`
	LastLogin int32  `json:"lastLogin" db:"last_login"`
	Returning bool   `json:"returning"`
}

// CourseInfo describes an active subscription course for the authenticated user.
type CourseInfo struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}

// MezFes represents the current Mezeporta Festival event schedule and ticket configuration.
type MezFes struct {
	ID           uint32   `json:"id"`
	Start        uint32   `json:"start"`
	End          uint32   `json:"end"`
	SoloTickets  uint32   `json:"soloTickets"`
	GroupTickets uint32   `json:"groupTickets"`
	Stalls       []uint32 `json:"stalls"`
}

// AuthData is the JSON payload returned after successful login or registration,
// containing session info, character list, event data, and server notices.
type AuthData struct {
	CurrentTS     uint32       `json:"currentTs"`
	ExpiryTS      uint32       `json:"expiryTs"`
	EntranceCount uint32       `json:"entranceCount"`
	Notices       []string     `json:"notices"`
	User          User         `json:"user"`
	Characters    []Character  `json:"characters"`
	Courses       []CourseInfo `json:"courses"`
	MezFes        *MezFes      `json:"mezFes"`
	PatchServer   string       `json:"patchServer"`
}

// ExportData wraps a character's full database row for save export.
type ExportData struct {
	Character map[string]interface{} `json:"character"`
}

func (s *APIServer) newAuthData(userID uint32, userRights uint32, userTokenID uint32, userToken string, characters []Character) AuthData {
	resp := AuthData{
		CurrentTS:     uint32(gametime.Adjusted().Unix()),
		ExpiryTS:      uint32(s.getReturnExpiry(userID).Unix()),
		EntranceCount: 1,
		User: User{
			Rights:  userRights,
			TokenID: userTokenID,
			Token:   userToken,
		},
		Characters:  characters,
		PatchServer: s.erupeConfig.API.PatchServer,
		Notices:     []string{},
	}
	// Compute returning status per character
	ninetyDaysAgo := time.Now().Add(-90 * 24 * time.Hour)
	for i := range resp.Characters {
		resp.Characters[i].Returning = time.Unix(int64(resp.Characters[i].LastLogin), 0).Before(ninetyDaysAgo)
	}
	// Derive active courses from user rights
	courses, _ := mhfcourse.GetCourseStruct(userRights, s.erupeConfig.DefaultCourses)
	resp.Courses = make([]CourseInfo, 0, len(courses))
	for _, c := range courses {
		name := ""
		if aliases := c.Aliases(); len(aliases) > 0 {
			name = aliases[0]
		}
		resp.Courses = append(resp.Courses, CourseInfo{ID: c.ID, Name: name})
	}
	if s.erupeConfig.DebugOptions.MaxLauncherHR {
		for i := range resp.Characters {
			resp.Characters[i].HR = 7
		}
	}
	resp.MezFes = s.buildMezFes()
	if !s.erupeConfig.HideLoginNotice {
		resp.Notices = append(resp.Notices, strings.Join(s.erupeConfig.LoginNotices[:], "<PAGE>"))
	}
	return resp
}

// VersionResponse is the JSON payload returned by the /version endpoint.
type VersionResponse struct {
	ClientMode string `json:"clientMode"`
	Name       string `json:"name"`
}

// Version handles GET /version and returns the server name and client mode.
func (s *APIServer) Version(w http.ResponseWriter, r *http.Request) {
	resp := VersionResponse{
		ClientMode: s.erupeConfig.ClientMode,
		Name:       "Erupe-CE",
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Launcher handles GET /launcher and returns banners, messages, and links for the launcher UI.
func (s *APIServer) Launcher(w http.ResponseWriter, r *http.Request) {
	var respData LauncherResponse
	respData.Banners = s.erupeConfig.API.Banners
	respData.Messages = s.erupeConfig.API.Messages
	respData.Links = s.erupeConfig.API.Links
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(respData)
}

// Login handles POST /login, authenticating a user by username and password
// and returning a session token with character data.
func (s *APIServer) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}
	userID, password, userRights, err := s.userRepo.GetCredentials(ctx, reqData.Username)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusBadRequest, "invalid_username", "Username not found")
		return
	} else if err != nil {
		s.logger.Warn("SQL query error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(password), []byte(reqData.Password)) != nil {
		writeError(w, http.StatusBadRequest, "invalid_password", "Incorrect password")
		return
	}

	userTokenID, userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Warn("Error registering login token", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	characters, err := s.getCharactersForUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if characters == nil {
		characters = []Character{}
	}
	respData := s.newAuthData(userID, userRights, userTokenID, userToken, characters)
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(respData)
}

// Register handles POST /register, creating a new user account and returning
// a session token.
func (s *APIServer) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}
	if reqData.Username == "" || reqData.Password == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "Username and password required")
		return
	}
	s.logger.Info("Creating account", zap.String("username", reqData.Username))
	userID, userRights, err := s.createNewUser(ctx, reqData.Username, reqData.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
			writeError(w, http.StatusBadRequest, "username_exists", "Username already taken")
			return
		}
		s.logger.Error("Error checking user", zap.Error(err), zap.String("username", reqData.Username))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	userTokenID, userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error registering login token", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	respData := s.newAuthData(userID, userRights, userTokenID, userToken, []Character{})
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(respData)
}

// CreateCharacter handles POST /characters (v2) or POST /character/create (legacy),
// creating a new character slot for the authenticated user.
func (s *APIServer) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		// Legacy path: read token from body
		var reqData struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			s.logger.Error("JSON decode error", zap.Error(err))
			writeError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
			return
		}
		var err error
		userID, err = s.userIDFromToken(ctx, reqData.Token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
			return
		}
	}
	character, err := s.createCharacter(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to create character", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if s.erupeConfig.DebugOptions.MaxLauncherHR {
		character.HR = 7
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(character)
}

// DeleteCharacter handles POST /characters/{id}/delete (v2) or POST /character/delete (legacy),
// soft-deleting an existing character or removing an unfinished one.
func (s *APIServer) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := UserIDFromContext(ctx)
	var charID uint32
	if ok {
		// V2 path: user ID from middleware, char ID from URL
		vars := mux.Vars(r)
		if idStr, exists := vars["id"]; exists {
			if _, err := fmt.Sscanf(idStr, "%d", &charID); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_request", "Invalid character ID")
				return
			}
		}
	} else {
		// Legacy path: read token and charId from body
		var reqData struct {
			Token  string `json:"token"`
			CharID uint32 `json:"charId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			s.logger.Error("JSON decode error", zap.Error(err))
			writeError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
			return
		}
		var err error
		userID, err = s.userIDFromToken(ctx, reqData.Token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
			return
		}
		charID = reqData.CharID
	}
	if err := s.deleteCharacter(ctx, userID, charID); err != nil {
		s.logger.Error("Failed to delete character", zap.Error(err), zap.Uint32("charID", charID))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct{}{})
}

// ExportSave handles GET /characters/{id}/export (v2) or POST /character/export (legacy),
// returning the full character database row as JSON for backup purposes.
func (s *APIServer) ExportSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := UserIDFromContext(ctx)
	var charID uint32
	if ok {
		// V2 path: user ID from middleware, char ID from URL
		vars := mux.Vars(r)
		if idStr, exists := vars["id"]; exists {
			if _, err := fmt.Sscanf(idStr, "%d", &charID); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_request", "Invalid character ID")
				return
			}
		}
	} else {
		// Legacy path: read token and charId from body
		var reqData struct {
			Token  string `json:"token"`
			CharID uint32 `json:"charId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			s.logger.Error("JSON decode error", zap.Error(err))
			writeError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
			return
		}
		var err error
		userID, err = s.userIDFromToken(ctx, reqData.Token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
			return
		}
		charID = reqData.CharID
	}
	character, err := s.exportSave(ctx, userID, charID)
	if err != nil {
		s.logger.Error("Failed to export save", zap.Error(err), zap.Uint32("charID", charID))
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	save := ExportData{
		Character: character,
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(save)
}

// ScreenShotGet handles GET /api/ss/bbs/{id}, serving a previously uploaded
// screenshot image by its token ID.
func (s *APIServer) ScreenShotGet(w http.ResponseWriter, r *http.Request) {
	// Get the 'id' parameter from the URL
	token := mux.Vars(r)["id"]
	var tokenPattern = regexp.MustCompile(`[A-Za-z0-9]+`)

	if !tokenPattern.MatchString(token) || token == "" {
		http.Error(w, "Not Valid Token", http.StatusBadRequest)

	}
	// Open the image file
	safePath := s.erupeConfig.Screenshots.OutputDir
	path := filepath.Join(safePath, fmt.Sprintf("%s.jpg", token))
	result, err := verifyPath(path, safePath, s.logger)

	if err != nil {
		s.logger.Warn("Screenshot path verification failed", zap.Error(err))
	} else {
		s.logger.Debug("Screenshot canonical path", zap.String("path", result))

		file, err := os.Open(result)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		defer func() { _ = file.Close() }()
		// Set content type header to image/jpeg
		w.Header().Set("Content-Type", "image/jpeg")
		// Copy the image content to the response writer
		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, "Unable to send image", http.StatusInternalServerError)
			return
		}
	}
}

// ScreenShot handles POST /api/ss/bbs/upload.php, accepting a JPEG image
// upload from the game client and saving it to the configured output directory.
func (s *APIServer) ScreenShot(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		XMLName xml.Name `xml:"result"`
		Code    string   `xml:"code"`
	}

	writeResult := func(code string) {
		w.Header().Set("Content-Type", "text/xml")
		xmlData, err := xml.Marshal(Result{Code: code})
		if err != nil {
			http.Error(w, "Unable to marshal XML", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(xmlData)
	}

	if !s.erupeConfig.Screenshots.Enabled {
		writeResult("400")
		return
	}
	if r.Method != http.MethodPost {
		writeResult("405")
		return
	}

	var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	token := r.FormValue("token")
	if !tokenPattern.MatchString(token) {
		writeResult("401")
		return
	}

	file, _, err := r.FormFile("img")
	if err != nil {
		writeResult("400")
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		writeResult("400")
		return
	}

	safePath := s.erupeConfig.Screenshots.OutputDir
	path := filepath.Join(safePath, fmt.Sprintf("%s.jpg", token))
	verified, err := verifyPath(path, safePath, s.logger)
	if err != nil {
		writeResult("500")
		return
	}

	if err := os.MkdirAll(safePath, os.ModePerm); err != nil {
		s.logger.Error("Error writing screenshot, could not create folder", zap.Error(err))
		writeResult("500")
		return
	}

	outputFile, err := os.Create(verified)
	if err != nil {
		s.logger.Error("Error writing screenshot, could not create file", zap.Error(err))
		writeResult("500")
		return
	}
	defer func() { _ = outputFile.Close() }()

	if err := jpeg.Encode(outputFile, img, &jpeg.Options{Quality: s.erupeConfig.Screenshots.UploadQuality}); err != nil {
		s.logger.Error("Error writing screenshot, could not write file", zap.Error(err))
		writeResult("500")
		return
	}

	writeResult("200")
}

func (s *APIServer) buildMezFes() *MezFes {
	stalls := []uint32{10, 3, 6, 9, 4, 8, 5, 7}
	if s.erupeConfig.GameplayOptions.MezFesSwitchMinigame {
		stalls[4] = 2
	}
	return &MezFes{
		ID:           uint32(gametime.WeekStart().Unix()),
		Start:        uint32(gametime.WeekStart().Add(-time.Duration(s.erupeConfig.GameplayOptions.MezFesDuration) * time.Second).Unix()),
		End:          uint32(gametime.WeekNext().Unix()),
		SoloTickets:  s.erupeConfig.GameplayOptions.MezFesSoloTickets,
		GroupTickets: s.erupeConfig.GameplayOptions.MezFesGroupTickets,
		Stalls:       stalls,
	}
}

// ServerStatusResponse is the JSON payload returned by GET /v2/server/status.
type ServerStatusResponse struct {
	MezFes         *MezFes            `json:"mezFes"`
	FeaturedWeapon *FeatureWeaponInfo `json:"featuredWeapon"`
	Events         EventStatus        `json:"events"`
}

// FeatureWeaponInfo describes the currently featured weapons.
type FeatureWeaponInfo struct {
	StartTime      uint32 `json:"startTime"`
	ActiveFeatures uint32 `json:"activeFeatures"`
}

// EventStatus indicates which recurring events are currently active.
type EventStatus struct {
	FestaActive bool `json:"festaActive"`
	DivaActive  bool `json:"divaActive"`
}

// ServerStatus handles GET /v2/server/status, returning MezFes schedule,
// featured weapon, and event activity status.
func (s *APIServer) ServerStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := ServerStatusResponse{
		MezFes: s.buildMezFes(),
	}

	if s.eventRepo != nil {
		weekStart := gametime.WeekStart()
		fw, err := s.eventRepo.GetFeatureWeapon(ctx, weekStart)
		if err != nil {
			s.logger.Warn("Failed to query feature weapon", zap.Error(err))
		} else if fw != nil {
			resp.FeaturedWeapon = &FeatureWeaponInfo{
				StartTime:      uint32(fw.StartTime.Unix()),
				ActiveFeatures: fw.ActiveFeatures,
			}
		}

		festaEvents, err := s.eventRepo.GetActiveEvents(ctx, "festa")
		if err != nil {
			s.logger.Warn("Failed to query festa events", zap.Error(err))
		} else {
			resp.Events.FestaActive = len(festaEvents) > 0
		}

		divaEvents, err := s.eventRepo.GetActiveEvents(ctx, "diva")
		if err != nil {
			s.logger.Warn("Failed to query diva events", zap.Error(err))
		} else {
			resp.Events.DivaActive = len(divaEvents) > 0
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Health handles GET /health, returning the server's health status.
// It pings the database to verify connectivity.
func (s *APIServer) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database not configured",
		})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// ImportSave handles POST /v2/characters/{id}/import.
// The request body must contain a one-time import_token (granted by an admin
// via saveutil) plus a character export blob in the same format as ExportSave.
func (s *APIServer) ImportSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := UserIDFromContext(ctx)

	var charID uint32
	if _, err := fmt.Sscanf(mux.Vars(r)["id"], "%d", &charID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid character ID")
		return
	}

	var req struct {
		ImportToken string                 `json:"import_token"`
		Character   map[string]interface{} `json:"character"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}
	if req.ImportToken == "" {
		writeError(w, http.StatusBadRequest, "missing_token", "import_token is required")
		return
	}

	blobs, err := saveBlobsFromMap(req.Character)
	if err != nil {
		s.logger.Warn("ImportSave: failed to extract blobs", zap.Error(err), zap.Uint32("charID", charID))
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid save data: "+err.Error())
		return
	}

	// Compute savedata hash server-side.
	if len(blobs.Savedata) > 0 {
		decompressed, err := nullcomp.Decompress(blobs.Savedata)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "savedata decompression failed")
			return
		}
		h := sha256.Sum256(decompressed)
		blobs.SavedataHash = h[:]
	}

	if err := s.charRepo.ImportSave(ctx, charID, userID, req.ImportToken, blobs); err != nil {
		s.logger.Warn("ImportSave: failed", zap.Error(err), zap.Uint32("charID", charID))
		writeError(w, http.StatusForbidden, "import_denied", "Import token invalid, expired, or character not owned by user")
		return
	}

	s.logger.Info("ImportSave: save imported successfully", zap.Uint32("charID", charID), zap.Uint32("userID", userID))
	w.WriteHeader(http.StatusOK)
}

// saveBlobsFromMap extracts save blob columns from an export character map.
// Values must be base64-encoded strings (as produced by json.Marshal on []byte).
func saveBlobsFromMap(m map[string]interface{}) (SaveBlobs, error) {
	var b SaveBlobs
	var err error
	b.Savedata, err = extractBlob(m, "savedata")
	if err != nil {
		return b, err
	}
	b.Decomyset, err = extractBlob(m, "decomyset")
	if err != nil {
		return b, err
	}
	b.Hunternavi, err = extractBlob(m, "hunternavi")
	if err != nil {
		return b, err
	}
	b.Otomoairou, err = extractBlob(m, "otomoairou")
	if err != nil {
		return b, err
	}
	b.Partner, err = extractBlob(m, "partner")
	if err != nil {
		return b, err
	}
	b.Platebox, err = extractBlob(m, "platebox")
	if err != nil {
		return b, err
	}
	b.Platedata, err = extractBlob(m, "platedata")
	if err != nil {
		return b, err
	}
	b.Platemyset, err = extractBlob(m, "platemyset")
	if err != nil {
		return b, err
	}
	b.Rengokudata, err = extractBlob(m, "rengokudata")
	if err != nil {
		return b, err
	}
	b.Savemercenary, err = extractBlob(m, "savemercenary")
	if err != nil {
		return b, err
	}
	b.GachaItems, err = extractBlob(m, "gacha_items")
	if err != nil {
		return b, err
	}
	b.HouseInfo, err = extractBlob(m, "house_info")
	if err != nil {
		return b, err
	}
	b.LoginBoost, err = extractBlob(m, "login_boost")
	if err != nil {
		return b, err
	}
	b.SkinHist, err = extractBlob(m, "skin_hist")
	if err != nil {
		return b, err
	}
	b.Scenariodata, err = extractBlob(m, "scenariodata")
	if err != nil {
		return b, err
	}
	b.Savefavoritequest, err = extractBlob(m, "savefavoritequest")
	if err != nil {
		return b, err
	}
	b.Mezfes, err = extractBlob(m, "mezfes")
	if err != nil {
		return b, err
	}
	return b, nil
}

// extractBlob decodes a single base64-encoded blob from a character export map.
// Returns nil (not an error) if the key is absent or its value is JSON null.
func extractBlob(m map[string]interface{}, key string) ([]byte, error) {
	v, ok := m[key]
	if !ok || v == nil {
		return nil, nil
	}
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("field %q: expected base64 string, got %T", key, v)
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("field %q: base64 decode: %w", key, err)
	}
	return b, nil
}
