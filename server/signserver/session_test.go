package signserver

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/network"

	"go.uber.org/zap"
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

// Note: Tests for handleDSGNRequest require a database connection.
// The function immediately queries the database for user authentication.
// These tests should be implemented as integration tests with a test database
// or using sqlmock for database mocking.
