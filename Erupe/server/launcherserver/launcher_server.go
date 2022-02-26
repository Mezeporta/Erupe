package launcherserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Solenataris/Erupe/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	Logger                   *zap.Logger
	DB                       *sqlx.DB
	ErupeConfig              *config.Config
	UseOriginalLauncherFiles bool
}

// Server is the MHF launcher HTTP server.
type Server struct {
	sync.Mutex
	logger                   *zap.Logger
	erupeConfig              *config.Config
	db                       *sqlx.DB
	httpServer               *http.Server
	useOriginalLauncherFiles bool
	isShuttingDown           bool
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:                   config.Logger,
		erupeConfig:              config.ErupeConfig,
		db:                       config.DB,
		useOriginalLauncherFiles: config.UseOriginalLauncherFiles,
		httpServer:               &http.Server{},
	}
	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	// Set up the routes responsible for serving the launcher HTML, serverlist, unique name check, and JP auth.
	r := mux.NewRouter()

	// Universal serverlist.xml route
	s.setupServerlistRoutes(r)

	// Change the launcher HTML routes if we are using the custom launcher instead of the original.
	if s.useOriginalLauncherFiles {
		s.setupOriginalLauncherRotues(r)
	} else {
		s.setupCustomLauncherRotues(r)
	}

	s.httpServer.Addr = fmt.Sprintf(":%d", s.erupeConfig.Launcher.Port)
	s.httpServer.Handler = handlers.LoggingHandler(os.Stdout, r)

	serveError := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			// Send error if any.
			serveError <- err
		}
	}()

	// Get the error from calling ListenAndServe, otherwise assume it's good after 250 milliseconds.
	select {
	case err := <-serveError:
		return err
	case <-time.After(250 * time.Millisecond):
		return nil
	}
}

// Shutdown exits the server gracefully.
func (s *Server) Shutdown() {
	s.logger.Debug("Shutting down")

	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		// Just warn because we are shutting down the server anyway.
		s.logger.Warn("Got error on httpServer shutdown", zap.Error(err))
	}
}
