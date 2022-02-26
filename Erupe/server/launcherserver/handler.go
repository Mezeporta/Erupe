package launcherserver

import (
	"net/http"
)

// ServerHandler is a handler function akin to http.Handler's ServeHTTP,
// but has an additional *Server argument.
type ServerHandler func(*Server, http.ResponseWriter, *http.Request)

// ServerHandlerFunc is a small type that implements http.Handler and
// wraps a calling ServerHandler with a *Server argument.
type ServerHandlerFunc struct {
	server *Server
	f      ServerHandler
}

func (shf ServerHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shf.f(shf.server, w, r)
}
