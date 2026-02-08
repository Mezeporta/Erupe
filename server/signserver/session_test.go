package signserver

import (
	"bytes"
	"database/sql"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/network"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
	mu       sync.Mutex
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53312}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestSessionStruct(t *testing.T) {
	logger := zap.NewNop()
	conn := newMockConn()

	s := &Session{
		logger:    logger,
		server:    nil,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn),
	}

	if s.logger != logger {
		t.Error("Session logger not set correctly")
	}
	if s.rawConn != conn {
		t.Error("Session rawConn not set correctly")
	}
	if s.cryptConn == nil {
		t.Error("Session cryptConn should not be nil")
	}
}

func TestSessionStructDefaults(t *testing.T) {
	s := &Session{}

	if s.logger != nil {
		t.Error("Default Session logger should be nil")
	}
	if s.server != nil {
		t.Error("Default Session server should be nil")
	}
	if s.rawConn != nil {
		t.Error("Default Session rawConn should be nil")
	}
	if s.cryptConn != nil {
		t.Error("Default Session cryptConn should be nil")
	}
}

func TestSessionMutex(t *testing.T) {
	s := &Session{}

	// Test that we can lock and unlock
	s.Lock()
	s.Unlock()

	// Test concurrent access
	done := make(chan bool)
	go func() {
		s.Lock()
		time.Sleep(10 * time.Millisecond)
		s.Unlock()
		done <- true
	}()

	// Small delay to ensure goroutine starts
	time.Sleep(5 * time.Millisecond)

	// This should block until the goroutine releases the lock
	s.Lock()
	s.Unlock()

	<-done
}

func TestHandlePacketUnknownRequest(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
	}

	conn := newMockConn()
	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn),
	}

	// Create a packet with unknown request type
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("UNKNOWN:100"))
	bf.WriteNullTerminatedBytes([]byte("data"))

	err := session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}
}

func TestHandlePacketEmptyRequest(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
	}

	conn := newMockConn()
	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn),
	}

	// Create a packet with empty request type (just null terminator)
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte(""))

	err := session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error for empty request: %v", err)
	}
}

func TestHandlePacketWithDevModeLogging(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			LogInboundMessages: true,
		},
	}

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
	}

	conn := newMockConn()
	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn),
	}

	// Create a packet with unknown request type
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("TEST:100"))

	err := session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() with dev mode returned error: %v", err)
	}
}

func TestHandlePacketRequestTypes(t *testing.T) {
	tests := []struct {
		name    string
		reqType string
	}{
		{"unknown", "UNKNOWN:100"},
		{"invalid", "INVALID"},
		{"empty_version", "TEST:"},
		{"no_version", "NOVERSION"},
		{"special_chars", "TEST@#$:100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			erupeConfig := &config.Config{DevMode: false}
			server := &Server{
				logger:      logger,
				erupeConfig: erupeConfig,
			}

			conn := newMockConn()
			session := &Session{
				logger:    logger,
				server:    server,
				rawConn:   conn,
				cryptConn: network.NewCryptConn(conn),
			}

			bf := byteframe.NewByteFrame()
			bf.WriteNullTerminatedBytes([]byte(tt.reqType))

			err := session.handlePacket(bf.Data())
			if err != nil {
				t.Errorf("handlePacket(%s) returned error: %v", tt.reqType, err)
			}
		})
	}
}

func TestMockConnImplementsNetConn(t *testing.T) {
	var _ net.Conn = (*mockConn)(nil)
}

func TestMockConnReadWrite(t *testing.T) {
	conn := newMockConn()

	// Write some data to the read buffer (simulating incoming data)
	testData := []byte("hello")
	conn.readBuf.Write(testData)

	// Read it back
	buf := make([]byte, len(testData))
	n, err := conn.Read(buf)
	if err != nil {
		t.Errorf("Read() error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Read() n = %d, want %d", n, len(testData))
	}
	if !bytes.Equal(buf, testData) {
		t.Errorf("Read() data = %v, want %v", buf, testData)
	}

	// Write data
	outData := []byte("world")
	n, err = conn.Write(outData)
	if err != nil {
		t.Errorf("Write() error: %v", err)
	}
	if n != len(outData) {
		t.Errorf("Write() n = %d, want %d", n, len(outData))
	}
	if !bytes.Equal(conn.writeBuf.Bytes(), outData) {
		t.Errorf("Write() buffer = %v, want %v", conn.writeBuf.Bytes(), outData)
	}
}

func TestMockConnClose(t *testing.T) {
	conn := newMockConn()

	err := conn.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	if !conn.closed {
		t.Error("conn.closed should be true after Close()")
	}

	// Read after close should return EOF
	buf := make([]byte, 10)
	_, err = conn.Read(buf)
	if err != io.EOF {
		t.Errorf("Read() after close should return EOF, got: %v", err)
	}

	// Write after close should return error
	_, err = conn.Write([]byte("test"))
	if err != io.ErrClosedPipe {
		t.Errorf("Write() after close should return ErrClosedPipe, got: %v", err)
	}
}

func TestMockConnAddresses(t *testing.T) {
	conn := newMockConn()

	local := conn.LocalAddr()
	if local == nil {
		t.Error("LocalAddr() should not be nil")
	}
	if local.String() != "127.0.0.1:53312" {
		t.Errorf("LocalAddr() = %s, want 127.0.0.1:53312", local.String())
	}

	remote := conn.RemoteAddr()
	if remote == nil {
		t.Error("RemoteAddr() should not be nil")
	}
	if remote.String() != "127.0.0.1:12345" {
		t.Errorf("RemoteAddr() = %s, want 127.0.0.1:12345", remote.String())
	}
}

func TestMockConnDeadlines(t *testing.T) {
	conn := newMockConn()
	deadline := time.Now().Add(time.Second)

	if err := conn.SetDeadline(deadline); err != nil {
		t.Errorf("SetDeadline() error: %v", err)
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		t.Errorf("SetReadDeadline() error: %v", err)
	}
	if err := conn.SetWriteDeadline(deadline); err != nil {
		t.Errorf("SetWriteDeadline() error: %v", err)
	}
}

func TestSessionWithCryptConn(t *testing.T) {
	conn := newMockConn()
	cryptConn := network.NewCryptConn(conn)

	if cryptConn == nil {
		t.Fatal("NewCryptConn() returned nil")
	}

	session := &Session{
		rawConn:   conn,
		cryptConn: cryptConn,
	}

	if session.cryptConn != cryptConn {
		t.Error("Session cryptConn not set correctly")
	}
}

// Note: Tests for DSGN:100, DLTSKEYSIGN:100, and DELETE:100 request types
// require a database connection. These are integration tests that should be
// run with a test database. The handlePacket method routes to these handlers
// which immediately access the database.

func TestSessionWorkWithDevModeLogging(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			LogInboundMessages: true,
		},
	}

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
	}

	// Use net.Pipe for bidirectional communication
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Close client side to cause read error
	clientConn.Close()

	// work() should exit gracefully on read error
	session.work()
}

func TestSessionWorkWithEmptyRead(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
	}

	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Close client side immediately to cause read failure
	clientConn.Close()

	// work() should handle the read error gracefully
	session.work()
}

// TestHandlePacketDSGNRequest tests the DSGN:100 path with a mocked database.
func TestHandlePacketDSGNRequest(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	// Use net.Pipe for bidirectional communication
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Create a DSGN:100 packet with username "testuser" and password "testpass"
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("testuser"))
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user not found, auto-create off
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	// Read the response in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	// Allow response to be sent
	time.Sleep(50 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDLTSKEYSIGN tests the DLTSKEYSIGN:100 path (falls through to DSGN:100)
func TestHandlePacketDLTSKEYSIGN(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Create a DLTSKEYSIGN:100 packet
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DLTSKEYSIGN:100"))
	bf.WriteNullTerminatedBytes([]byte("testuser"))
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user not found
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDELETE tests the DELETE:100 path
func TestHandlePacketDELETE(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Create a DELETE:100 packet
	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DELETE:100"))
	bf.WriteNullTerminatedBytes([]byte("login-token-abc"))
	bf.WriteUint32(123) // characterID
	bf.WriteUint32(456) // login_token_number

	// Mock DB: Token verification
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM sign_sessions WHERE token = \\$1").
		WithArgs("login-token-abc").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Check if new character
	mock.ExpectQuery("SELECT is_new_character FROM characters WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"is_new_character"}).AddRow(false))

	// Soft delete
	mock.ExpectExec("UPDATE characters SET deleted = true WHERE id = \\$1").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Read all response data in a goroutine (SendPacket writes header + encrypted data)
	done := make(chan []byte, 1)
	go func() {
		var all []byte
		buf := make([]byte, 4096)
		for {
			n, readErr := clientConn.Read(buf)
			if n > 0 {
				all = append(all, buf[:n]...)
			}
			if readErr != nil {
				break
			}
		}
		done <- all
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	// Close server side so the reader goroutine finishes
	serverConn.Close()

	select {
	case <-done:
		// Response received successfully
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for response")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDSGNWithAutoCreate tests DSGN:100 with auto-create account enabled
func TestHandlePacketDSGNWithAutoCreate(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			AutoCreateAccount: true,
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("newuser"))
	bf.WriteNullTerminatedBytes([]byte("newpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user not found
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	// Auto-create: insert user
	mock.ExpectExec("INSERT INTO users \\(username, password, return_expires\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs("newuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Auto-create: get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Auto-create: insert character
	mock.ExpectExec("INSERT INTO characters").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Now get new user ID for makeSignInResp
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// makeSignInResp calls getReturnExpiry
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(time.Now()))

	// getReturnExpiry: get return_expires
	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"return_expires"}).AddRow(time.Now().Add(time.Hour * 24 * 30)))

	// getReturnExpiry: update last_login
	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// getCharactersForUser
	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id ASC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hrp", "gr", "weapon_type", "last_login"}))

	// registerToken
	mock.ExpectExec("INSERT INTO sign_sessions \\(user_id, token\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// getLastCID
	mock.ExpectQuery("SELECT last_character FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_character"}).AddRow(0))

	// getUserRights
	mock.ExpectQuery("SELECT rights FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"rights"}).AddRow(2))

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDSGNWithValidPassword tests DSGN:100 with correct password
func TestHandlePacketDSGNWithValidPassword(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Generate a bcrypt hash for "testpass"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("existinguser"))
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user found with correct password
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("existinguser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(1, string(hashedPassword)))

	// makeSignInResp calls getReturnExpiry
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(time.Now()))

	// getReturnExpiry: get return_expires
	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"return_expires"}).AddRow(time.Now().Add(time.Hour * 24 * 30)))

	// getReturnExpiry: update last_login
	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// getCharactersForUser
	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id ASC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hrp", "gr", "weapon_type", "last_login"}))

	// registerToken
	mock.ExpectExec("INSERT INTO sign_sessions \\(user_id, token\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// getLastCID
	mock.ExpectQuery("SELECT last_character FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_character"}).AddRow(0))

	// getUserRights
	mock.ExpectQuery("SELECT rights FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"rights"}).AddRow(2))

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDSGNWrongPassword tests DSGN:100 with wrong password
func TestHandlePacketDSGNWrongPassword(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Generate a bcrypt hash for "correctpass"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("testuser"))
	bf.WriteNullTerminatedBytes([]byte("wrongpass")) // Wrong password
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user found but password will not match
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(1, string(hashedPassword)))

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDSGNWithDBError tests DSGN:100 with a database error
func TestHandlePacketDSGNWithDBError(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("testuser"))
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: generic error
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnError(sql.ErrConnDone)

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDSGNNewCharaRequest tests DSGN:100 with the '+' suffix for new character
func TestHandlePacketDSGNNewCharaRequest(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: false,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	// Generate a bcrypt hash for "testpass"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("testuser+")) // '+' suffix means new character request
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user found
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(1, string(hashedPassword)))

	// newUserChara: get user ID
	mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// newUserChara: check existing new chars
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM characters WHERE user_id = \\$1 AND is_new_character = true").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// newUserChara: insert character
	mock.ExpectExec("INSERT INTO characters").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// makeSignInResp calls
	mock.ExpectQuery("SELECT COALESCE\\(last_login, now\\(\\)\\) FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_login"}).AddRow(time.Now()))

	mock.ExpectQuery("SELECT return_expires FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"return_expires"}).AddRow(time.Now().Add(time.Hour * 24 * 30)))

	mock.ExpectExec("UPDATE users SET last_login=\\$1 WHERE id=\\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = \\$1 AND deleted = false ORDER BY id ASC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_female", "is_new_character", "name", "unk_desc_string", "hrp", "gr", "weapon_type", "last_login"}))

	mock.ExpectExec("INSERT INTO sign_sessions \\(user_id, token\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT last_character FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"last_character"}).AddRow(0))

	mock.ExpectQuery("SELECT rights FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"rights"}).AddRow(2))

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestHandlePacketDSGNWithDevModeOutboundLogging tests dev mode outbound logging
func TestHandlePacketDSGNWithDevModeOutboundLogging(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &config.Config{
		DevMode: true,
		DevModeOptions: config.DevModeOptions{
			LogOutboundMessages: true,
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	server := &Server{
		logger:      logger,
		erupeConfig: erupeConfig,
		db:          sqlxDB,
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	session := &Session{
		logger:    logger,
		server:    server,
		rawConn:   serverConn,
		cryptConn: network.NewCryptConn(serverConn),
	}

	bf := byteframe.NewByteFrame()
	bf.WriteNullTerminatedBytes([]byte("DSGN:100"))
	bf.WriteNullTerminatedBytes([]byte("testuser"))
	bf.WriteNullTerminatedBytes([]byte("testpass"))
	bf.WriteNullTerminatedBytes([]byte("unk"))

	// Mock DB: user not found, dev mode but no auto create
	mock.ExpectQuery("SELECT id, password FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			_, err := clientConn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	err = session.handlePacket(bf.Data())
	if err != nil {
		t.Errorf("handlePacket() returned error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	clientConn.Close()
	<-done

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
