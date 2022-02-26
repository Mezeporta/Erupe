package signserver

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Solenataris/Erupe/config"
	"github.com/Solenataris/Erupe/network"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	Logger      *zap.Logger
	DB          *sqlx.DB
	ErupeConfig *config.Config
}

// Server is a MHF sign server.
type Server struct {
	sync.Mutex
	logger         *zap.Logger
	erupeConfig    *config.Config
	sid            int
	sessions       map[int]*Session
	db             *sqlx.DB
	listener       net.Listener
	isShuttingDown bool
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:      config.Logger,
		erupeConfig: config.ErupeConfig,
		sid:         0,
		sessions:    make(map[int]*Session),
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
	s.logger.Debug("Shutting down")

	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	// This will cause the acceptor goroutine to error and exit gracefully.
	s.listener.Close()
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

		go s.handleConnection(s.sid, conn)
		s.sid++
	}
}

func (s *Server) handleConnection(sid int, conn net.Conn) {
	s.logger.Info("Got connection to sign server", zap.String("remoteaddr", conn.RemoteAddr().String()))

	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	_, err := io.ReadFull(conn, nullInit)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Create a new session.
	session := &Session{
		logger:    s.logger,
		server:    s,
		rawConn:   &conn,
		cryptConn: network.NewCryptConn(conn),
	}

	// Add the session to the server's sessions map.
	s.Lock()
	s.sessions[sid] = session
	s.Unlock()

	// Do the session's work.
	session.work()
}
