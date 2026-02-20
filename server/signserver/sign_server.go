package signserver

import (
	"fmt"
	"io"
	"net"
	"sync"

	"erupe-ce/config"
	"erupe-ce/network"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	Logger      *zap.Logger
	DB          *sqlx.DB
	ErupeConfig *_config.Config
}

// Server is a MHF sign server.
type Server struct {
	sync.Mutex
	logger         *zap.Logger
	erupeConfig    *_config.Config
	db             *sqlx.DB
	listener       net.Listener
	isShuttingDown bool
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:      config.Logger,
		erupeConfig: config.ErupeConfig,
		db:          config.DB,
	}
	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.erupeConfig.Sign.Port))
	if err != nil {
		return err
	}
	s.listener = l

	go s.acceptClients()

	return nil
}

// Shutdown exits the server gracefully.
func (s *Server) Shutdown() {
	s.logger.Debug("Shutting down...")

	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	// This will cause the acceptor goroutine to error and exit gracefully.
	_ = s.listener.Close()
}

func (s *Server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if we are shutting down and exit gracefully if so.
			s.Lock()
			shutdown := s.isShuttingDown
			s.Unlock()

			if shutdown {
				break
			} else {
				panic(err)
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	s.logger.Debug("New connection", zap.String("RemoteAddr", conn.RemoteAddr().String()))
	defer func() { _ = conn.Close() }()

	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	_, err := io.ReadFull(conn, nullInit)
	if err != nil {
		s.logger.Error("Error initializing connection", zap.Error(err))
		return
	}

	// Create a new session.
	session := &Session{
		logger:    s.logger,
		server:    s,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn, s.erupeConfig.RealClientMode, s.logger),
	}

	// Do the session's work.
	session.work()
}
