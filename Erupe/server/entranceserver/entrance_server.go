package entranceserver

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Solenataris/Erupe/config"
	"github.com/Solenataris/Erupe/network"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Server is a MHF entrance server.
type Server struct {
	sync.Mutex
	logger         *zap.Logger
	erupeConfig    *config.Config
	db             *sqlx.DB
	listener       net.Listener
	isShuttingDown bool
}

// Config struct allows configuring the server.
type Config struct {
	Logger      *zap.Logger
	DB          *sqlx.DB
	ErupeConfig *config.Config
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

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.erupeConfig.Entrance.Port))
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

//acceptClients handles accepting new clients in a loop.
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
				continue
			}
		}

		// Start a new goroutine for the connection so that we don't block other incoming connections.
		go s.handleEntranceServerConnection(conn)
	}
}

func (s *Server) handleEntranceServerConnection(conn net.Conn) {
	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	n, err := io.ReadFull(conn, nullInit)
	if err != nil {
		s.logger.Warn("Failed to read 8 NULL init", zap.Error(err))
		return
	} else if n != len(nullInit) {
		s.logger.Warn("io.ReadFull couldn't read the full 8 byte init.")
		return
	}

	// Create a new encrypted connection handler and read a packet from it.
	cc := network.NewCryptConn(conn)
	pkt, err := cc.ReadPacket()
	if err != nil {
		s.logger.Warn("Error reading packet", zap.Error(err))
		return
	}

	s.logger.Debug("Got entrance server command:\n", zap.String("raw", hex.Dump(pkt)))

	data := makeSv2Resp(s.erupeConfig.Entrance.Entries, s)
	if len(pkt) > 5 {
		data = append(data, makeUsrResp(pkt)...)
	}
	cc.SendPacket(data)
	// Close because we only need to send the response once.
	// Any further requests from the client will come from a new connection.
	conn.Close()
}
