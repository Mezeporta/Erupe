package api

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	_config "erupe-ce/config"
	"erupe-ce/common/gametime"
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

const (
	NotificationDefault = iota
	NotificationNew
)

type LauncherResponse struct {
	Banners  []_config.APISignBanner  `json:"banners"`
	Messages []_config.APISignMessage `json:"messages"`
	Links    []_config.APISignLink    `json:"links"`
}

type User struct {
	TokenID uint32 `json:"tokenId"`
	Token   string `json:"token"`
	Rights  uint32 `json:"rights"`
}

type Character struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	IsFemale  bool   `json:"isFemale" db:"is_female"`
	Weapon    uint32 `json:"weapon" db:"weapon_type"`
	HR        uint32 `json:"hr" db:"hr"`
	GR        uint32 `json:"gr"`
	LastLogin int32  `json:"lastLogin" db:"last_login"`
}

type MezFes struct {
	ID           uint32   `json:"id"`
	Start        uint32   `json:"start"`
	End          uint32   `json:"end"`
	SoloTickets  uint32   `json:"soloTickets"`
	GroupTickets uint32   `json:"groupTickets"`
	Stalls       []uint32 `json:"stalls"`
}

type AuthData struct {
	CurrentTS     uint32      `json:"currentTs"`
	ExpiryTS      uint32      `json:"expiryTs"`
	EntranceCount uint32      `json:"entranceCount"`
	Notices       []string    `json:"notices"`
	User          User        `json:"user"`
	Characters    []Character `json:"characters"`
	MezFes        *MezFes     `json:"mezFes"`
	PatchServer   string      `json:"patchServer"`
}

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
	if s.erupeConfig.DebugOptions.MaxLauncherHR {
		for i := range resp.Characters {
			resp.Characters[i].HR = 7
		}
	}
	stalls := []uint32{10, 3, 6, 9, 4, 8, 5, 7}
	if s.erupeConfig.GameplayOptions.MezFesSwitchMinigame {
		stalls[4] = 2
	}
	resp.MezFes = &MezFes{
		ID:           uint32(gametime.WeekStart().Unix()),
		Start:        uint32(gametime.WeekStart().Add(-time.Duration(s.erupeConfig.GameplayOptions.MezFesDuration) * time.Second).Unix()),
		End:          uint32(gametime.WeekNext().Unix()),
		SoloTickets:  s.erupeConfig.GameplayOptions.MezFesSoloTickets,
		GroupTickets: s.erupeConfig.GameplayOptions.MezFesGroupTickets,
		Stalls:       stalls,
	}
	if !s.erupeConfig.HideLoginNotice {
		resp.Notices = append(resp.Notices, strings.Join(s.erupeConfig.LoginNotices[:], "<PAGE>"))
	}
	return resp
}

func (s *APIServer) Launcher(w http.ResponseWriter, r *http.Request) {
	var respData LauncherResponse
	respData.Banners = s.erupeConfig.API.Banners
	respData.Messages = s.erupeConfig.API.Messages
	respData.Links = s.erupeConfig.API.Links
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(respData)
}

func (s *APIServer) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	var (
		userID     uint32
		userRights uint32
		password   string
	)
	err := s.db.QueryRow("SELECT id, password, rights FROM users WHERE username = $1", reqData.Username).Scan(&userID, &password, &userRights)
	if err == sql.ErrNoRows {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("username-error"))
		return
	} else if err != nil {
		s.logger.Warn("SQL query error", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(password), []byte(reqData.Password)) != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("password-error"))
		return
	}

	userTokenID, userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Warn("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	characters, err := s.getCharactersForUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	if characters == nil {
		characters = []Character{}
	}
	respData := s.newAuthData(userID, userRights, userTokenID, userToken, characters)
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(respData)
}

func (s *APIServer) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	if reqData.Username == "" || reqData.Password == "" {
		w.WriteHeader(400)
		return
	}
	s.logger.Info("Creating account", zap.String("username", reqData.Username))
	userID, userRights, err := s.createNewUser(ctx, reqData.Username, reqData.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("username-exists-error"))
			return
		}
		s.logger.Error("Error checking user", zap.Error(err), zap.String("username", reqData.Username))
		w.WriteHeader(500)
		return
	}

	userTokenID, userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	respData := s.newAuthData(userID, userRights, userTokenID, userToken, []Character{})
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(respData)
}

func (s *APIServer) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		return
	}

	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	character, err := s.createCharacter(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to create character", zap.Error(err), zap.String("token", reqData.Token))
		w.WriteHeader(500)
		return
	}
	if s.erupeConfig.DebugOptions.MaxLauncherHR {
		character.HR = 7
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(character)
}

func (s *APIServer) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token  string `json:"token"`
		CharID uint32 `json:"charId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	if err := s.deleteCharacter(ctx, userID, reqData.CharID); err != nil {
		s.logger.Error("Failed to delete character", zap.Error(err), zap.String("token", reqData.Token), zap.Uint32("charID", reqData.CharID))
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct{}{})
}

func (s *APIServer) ExportSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token  string `json:"token"`
		CharID uint32 `json:"charId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	character, err := s.exportSave(ctx, userID, reqData.CharID)
	if err != nil {
		s.logger.Error("Failed to export save", zap.Error(err), zap.String("token", reqData.Token), zap.Uint32("charID", reqData.CharID))
		w.WriteHeader(500)
		return
	}
	save := ExportData{
		Character: character,
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(save)
}
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
	result, err := verifyPath(path, safePath)

	if err != nil {
		fmt.Println("Error " + err.Error())
	} else {
		fmt.Println("Canonical: " + result)

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
	verified, err := verifyPath(path, safePath)
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
