package signserver

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// spyConn implements network.Conn and records plaintext packets sent.
type spyConn struct {
	mu       sync.Mutex
	sent     [][]byte // plaintext packets captured from SendPacket
	readData []byte   // data to return from ReadPacket (unused in handler tests)
}

func (s *spyConn) ReadPacket() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.readData) == 0 {
		return nil, fmt.Errorf("no data")
	}
	data := s.readData
	s.readData = nil
	return data, nil
}

func (s *spyConn) SendPacket(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	s.sent = append(s.sent, cp)
	return nil
}

// lastSent returns the last packet sent, or nil if none.
func (s *spyConn) lastSent() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sent) == 0 {
		return nil
	}
	return s.sent[len(s.sent)-1]
}

// newHandlerSession creates a Session with a spyConn for handler tests.
func newHandlerSession(userRepo SignUserRepo, charRepo SignCharacterRepo, sessionRepo SignSessionRepo, erupeConfig *cfg.Config) (*Session, *spyConn) {
	logger := zap.NewNop()
	mc := newMockConn() // still needed for rawConn (used by makeSignResponse for RemoteAddr)
	spy := &spyConn{}
	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		userRepo:    userRepo,
		charRepo:    charRepo,
		sessionRepo: sessionRepo,
	}
	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   mc,
		cryptConn: spy,
	}
	return session, spy
}

// defaultConfig returns a minimal config suitable for most handler tests.
func defaultConfig() *cfg.Config {
	return &cfg.Config{
		RealClientMode: cfg.ZZ,
	}
}

// hashPassword creates a bcrypt hash for testing.
func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal("failed to hash password:", err)
	}
	return string(hash)
}

// --- sendCode ---

func TestSendCode(t *testing.T) {
	codes := []RespID{
		SIGN_SUCCESS,
		SIGN_EABORT,
		SIGN_ECOGLINK,
		SIGN_EPSI,
		SIGN_EMBID,
		SIGN_EELIMINATE,
		SIGN_ESUSPEND,
	}

	for _, code := range codes {
		t.Run(fmt.Sprintf("code_%d", code), func(t *testing.T) {
			session, spy := newHandlerSession(nil, nil, nil, defaultConfig())
			session.sendCode(code)

			pkt := spy.lastSent()
			if pkt == nil {
				t.Fatal("sendCode() sent no packet")
			}
			if len(pkt) != 1 {
				t.Fatalf("sendCode() packet len = %d, want 1", len(pkt))
			}
			if RespID(pkt[0]) != code {
				t.Errorf("sendCode() = %d, want %d", pkt[0], code)
			}
		})
	}
}

// --- authenticate ---

func TestAuthenticate_Success(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
	}
	charRepo := &mockSignCharacterRepo{
		characters: []character{{ID: 1, Name: "TestChar"}},
	}
	sessionRepo := &mockSignSessionRepo{
		registerUIDTokenID: 100,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.authenticate("testuser", pass)

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() sent no packet")
	}
	if len(pkt) < 1 {
		t.Fatal("authenticate() packet too short")
	}
	if RespID(pkt[0]) != SIGN_SUCCESS {
		t.Errorf("authenticate() first byte = %d, want SIGN_SUCCESS(%d)", pkt[0], SIGN_SUCCESS)
	}
}

func TestAuthenticate_NewCharaRequest(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
	}
	charRepo := &mockSignCharacterRepo{
		newCharCount: 0,
		characters:   []character{{ID: 1, Name: "TestChar"}},
	}
	sessionRepo := &mockSignSessionRepo{
		registerUIDTokenID: 100,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.authenticate("testuser+", pass)

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() sent no packet for new chara request")
	}
	if !charRepo.createCalled {
		t.Error("authenticate() with '+' suffix should call CreateCharacter")
	}
}

func TestAuthenticate_LoginFailed(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: sql.ErrNoRows,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.authenticate("unknownuser", "pass")

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() sent no packet for failed login")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EAUTH {
		t.Errorf("authenticate() failed login = %v, want [%d]", pkt, SIGN_EAUTH)
	}
}

func TestAuthenticate_WithDebugLogging(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: sql.ErrNoRows,
	}
	config := defaultConfig()
	config.DebugOptions.LogOutboundMessages = true

	session, spy := newHandlerSession(userRepo, nil, nil, config)
	session.authenticate("user", "pass")

	if spy.lastSent() == nil {
		t.Fatal("authenticate() with debug logging sent no packet")
	}
}

// --- handleDSGN ---

func TestHandleDSGN(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: sql.ErrNoRows,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("testuser"))
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	session.handleDSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	if spy.lastSent() == nil {
		t.Fatal("handleDSGN() sent no packet")
	}
}

// --- handleWIIUSGN ---

func TestHandleWIIUSGN_Success(t *testing.T) {
	userRepo := &mockSignUserRepo{
		wiiuUID: 10,
	}
	charRepo := &mockSignCharacterRepo{
		characters: []character{{ID: 1, Name: "WiiUChar"}},
	}
	sessionRepo := &mockSignSessionRepo{
		registerUIDTokenID: 200,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.client = WIIU

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(make([]byte, 1))
	key := make([]byte, 64)
	copy(key, []byte("wiiu-test-key"))
	bf.WriteBytes(key)

	session.handleWIIUSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handleWIIUSGN() sent no packet")
	}
	if RespID(pkt[0]) != SIGN_SUCCESS {
		t.Errorf("handleWIIUSGN() = %d, want SIGN_SUCCESS(%d)", pkt[0], SIGN_SUCCESS)
	}
}

func TestHandleWIIUSGN_UnlinkedKey(t *testing.T) {
	userRepo := &mockSignUserRepo{
		wiiuErr: sql.ErrNoRows,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.client = WIIU

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(make([]byte, 1))
	bf.WriteBytes(make([]byte, 64))

	session.handleWIIUSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handleWIIUSGN() sent no packet for unlinked key")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ECOGLINK {
		t.Errorf("handleWIIUSGN() unlinked = %v, want [%d]", pkt, SIGN_ECOGLINK)
	}
}

func TestHandleWIIUSGN_DBError(t *testing.T) {
	userRepo := &mockSignUserRepo{
		wiiuErr: errMockDB,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.client = WIIU

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(make([]byte, 1))
	bf.WriteBytes(make([]byte, 64))

	session.handleWIIUSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handleWIIUSGN() sent no packet for DB error")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("handleWIIUSGN() DB error = %v, want [%d]", pkt, SIGN_EABORT)
	}
}

// --- handlePSSGN ---

func TestHandlePSSGN_PS4_Success(t *testing.T) {
	userRepo := &mockSignUserRepo{
		psnUID: 20,
	}
	charRepo := &mockSignCharacterRepo{
		characters: []character{{ID: 1, Name: "PS4Char"}},
	}
	sessionRepo := &mockSignSessionRepo{
		registerUIDTokenID: 300,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.client = PS4

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("ps4_user_psn"))

	session.handlePSSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSSGN(PS4) sent no packet")
	}
	if RespID(pkt[0]) != SIGN_SUCCESS {
		t.Errorf("handlePSSGN(PS4) = %d, want SIGN_SUCCESS(%d)", pkt[0], SIGN_SUCCESS)
	}
}

func TestHandlePSSGN_PS3_Success(t *testing.T) {
	userRepo := &mockSignUserRepo{
		psnUID: 21,
	}
	charRepo := &mockSignCharacterRepo{
		characters: []character{{ID: 1, Name: "PS3Char"}},
	}
	sessionRepo := &mockSignSessionRepo{
		registerUIDTokenID: 301,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.client = PS3

	// PS3: needs ≥128 bytes remaining after current position
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("0000000255"))
	bf.WriteBytes([]byte("! "))     // 2 bytes
	bf.WriteBytes(make([]byte, 82)) // 82 bytes padding
	bf.WriteNullTerminatedBytes([]byte("ps3_user"))
	for len(bf.Data()) < 140 {
		bf.WriteUint8(0)
	}

	session.handlePSSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSSGN(PS3) sent no packet")
	}
}

func TestHandlePSSGN_MalformedShortBuffer(t *testing.T) {
	session, spy := newHandlerSession(nil, nil, nil, defaultConfig())
	session.client = PS3

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(make([]byte, 10))

	session.handlePSSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSSGN() short buffer sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("handlePSSGN() short buffer = %v, want [%d]", pkt, SIGN_EABORT)
	}
}

func TestHandlePSSGN_UnknownPSN(t *testing.T) {
	userRepo := &mockSignUserRepo{
		psnErr: sql.ErrNoRows,
	}
	charRepo := &mockSignCharacterRepo{}
	sessionRepo := &mockSignSessionRepo{
		registerPSNTokenID: 400,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.client = PS4

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("unknown_psn"))

	session.handlePSSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSSGN() unknown PSN sent no packet")
	}
	// Unknown PSN calls makeSignResponse(0) which should produce SIGN_SUCCESS
	if RespID(pkt[0]) != SIGN_SUCCESS {
		t.Errorf("handlePSSGN() unknown PSN first byte = %d, want SIGN_SUCCESS(%d)", pkt[0], SIGN_SUCCESS)
	}
	if session.psn != "unknown_psn" {
		t.Errorf("session.psn = %q, want %q", session.psn, "unknown_psn")
	}
}

func TestHandlePSSGN_DBError(t *testing.T) {
	userRepo := &mockSignUserRepo{
		psnErr: errMockDB,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.client = PS4

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("some_psn"))

	session.handlePSSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSSGN() DB error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("handlePSSGN() DB error = %v, want [%d]", pkt, SIGN_EABORT)
	}
}

// --- handlePSNLink ---

func TestHandlePSNLink_Success(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:          42,
		credPassword:     hash,
		psnCount:         0,
		psnIDForUsername: "",
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByToken: "linked_psn_id",
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() success sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_SUCCESS {
		t.Errorf("handlePSNLink() success = %v, want [%d]", pkt, SIGN_SUCCESS)
	}
	if !userRepo.setPSNIDCalled {
		t.Error("handlePSNLink() should call SetPSNID on success")
	}
}

func TestHandlePSNLink_InvalidCredentials(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: sql.ErrNoRows,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("baduser\nbadpass"))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() invalid creds sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ECOGLINK {
		t.Errorf("handlePSNLink() invalid creds = %v, want [%d]", pkt, SIGN_ECOGLINK)
	}
}

func TestHandlePSNLink_TokenLookupError(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByTokenErr: errMockDB,
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("badtoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() token error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ECOGLINK {
		t.Errorf("handlePSNLink() token error = %v, want [%d]", pkt, SIGN_ECOGLINK)
	}
}

func TestHandlePSNLink_PSNAlreadyLinked(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
		psnCount:     1,
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByToken: "already_linked_psn",
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() PSN already linked sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EPSI {
		t.Errorf("handlePSNLink() PSN already linked = %v, want [%d]", pkt, SIGN_EPSI)
	}
}

func TestHandlePSNLink_AccountAlreadyLinked(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:          42,
		credPassword:     hash,
		psnCount:         0,
		psnIDForUsername: "existing_psn",
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByToken: "new_psn",
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() account already linked sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EMBID {
		t.Errorf("handlePSNLink() account already linked = %v, want [%d]", pkt, SIGN_EMBID)
	}
}

func TestHandlePSNLink_SetPSNIDError(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:          42,
		credPassword:     hash,
		psnCount:         0,
		psnIDForUsername: "",
		setPSNIDErr:      errMockDB,
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByToken: "psn_to_link",
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() SetPSNID error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ECOGLINK {
		t.Errorf("handlePSNLink() SetPSNID error = %v, want [%d]", pkt, SIGN_ECOGLINK)
	}
}

func TestHandlePSNLink_CountByPSNIDError(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
		psnCountErr:  errMockDB,
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByToken: "psn_id",
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() CountByPSNID error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ECOGLINK {
		t.Errorf("handlePSNLink() CountByPSNID error = %v, want [%d]", pkt, SIGN_ECOGLINK)
	}
}

func TestHandlePSNLink_GetPSNIDForUsernameError(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:             42,
		credPassword:        hash,
		psnCount:            0,
		psnIDForUsernameErr: errMockDB,
	}
	sessionRepo := &mockSignSessionRepo{
		psnIDByToken: "psn_id",
	}

	session, spy := newHandlerSession(userRepo, nil, sessionRepo, defaultConfig())

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("client_id"))
	bf.WriteNullTerminatedBytes([]byte("testuser\n" + pass))
	bf.WriteNullTerminatedBytes([]byte("sometoken"))

	session.handlePSNLink(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSNLink() GetPSNIDForUsername error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ECOGLINK {
		t.Errorf("handlePSNLink() GetPSNIDForUsername error = %v, want [%d]", pkt, SIGN_ECOGLINK)
	}
}

// --- VITA client path ---

func TestHandlePSSGN_VITA_MalformedShortBuffer(t *testing.T) {
	session, spy := newHandlerSession(nil, nil, nil, defaultConfig())
	session.client = VITA

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(make([]byte, 50))

	session.handlePSSGN(byteframe.NewByteFrameFromBytes(bf.Data()))

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("handlePSSGN(VITA) short buffer sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("handlePSSGN(VITA) short buffer = %v, want [%d]", pkt, SIGN_EABORT)
	}
}

// --- authenticate error paths ---

func TestAuthenticate_WrongPassword(t *testing.T) {
	hash := hashPassword(t, "correctpassword")

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.authenticate("testuser", "wrongpassword")

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() wrong password sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EPASS {
		t.Errorf("authenticate() wrong password = %v, want [%d]", pkt, SIGN_EPASS)
	}
}

func TestAuthenticate_PermanentBan(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:       42,
		credPassword:  hash,
		permanentBans: 1,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.authenticate("banneduser", pass)

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() permanent ban sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EELIMINATE {
		t.Errorf("authenticate() permanent ban = %v, want [%d]", pkt, SIGN_EELIMINATE)
	}
}

func TestAuthenticate_ActiveBan(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
		activeBans:   1,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.authenticate("suspendeduser", pass)

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() active ban sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_ESUSPEND {
		t.Errorf("authenticate() active ban = %v, want [%d]", pkt, SIGN_ESUSPEND)
	}
}

func TestAuthenticate_DBError(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr: errMockDB,
	}

	session, spy := newHandlerSession(userRepo, nil, nil, defaultConfig())
	session.authenticate("user", "pass")

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() DB error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("authenticate() DB error = %v, want [%d]", pkt, SIGN_EABORT)
	}
}

func TestAuthenticate_AutoCreateError(t *testing.T) {
	userRepo := &mockSignUserRepo{
		credErr:     sql.ErrNoRows,
		registerErr: errMockDB,
	}
	config := defaultConfig()
	config.AutoCreateAccount = true

	session, spy := newHandlerSession(userRepo, nil, nil, config)
	session.authenticate("newuser", "pass")

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() auto-create error sent no packet")
	}
	if len(pkt) != 1 || RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("authenticate() auto-create error = %v, want [%d]", pkt, SIGN_EABORT)
	}
}

func TestAuthenticate_RegisterTokenError(t *testing.T) {
	pass := "hunter2"
	hash := hashPassword(t, pass)

	userRepo := &mockSignUserRepo{
		credUID:      42,
		credPassword: hash,
		lastLogin:    time.Now(),
		returnExpiry: time.Now().Add(time.Hour * 24 * 30),
	}
	charRepo := &mockSignCharacterRepo{
		characters: []character{{ID: 1, Name: "Char"}},
	}
	sessionRepo := &mockSignSessionRepo{
		registerUIDErr: errMockDB,
	}

	session, spy := newHandlerSession(userRepo, charRepo, sessionRepo, defaultConfig())
	session.authenticate("user", pass)

	pkt := spy.lastSent()
	if pkt == nil {
		t.Fatal("authenticate() token error sent no packet")
	}
	// When registerUidToken fails, makeSignResponse returns SIGN_EABORT as first byte
	if RespID(pkt[0]) != SIGN_EABORT {
		t.Errorf("authenticate() token error first byte = %d, want SIGN_EABORT(%d)", pkt[0], SIGN_EABORT)
	}
}
