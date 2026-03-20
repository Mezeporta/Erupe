package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// SessionRepository centralizes all database access for sign_sessions and servers tables.
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// ValidateLoginToken validates that the given token, session ID, and character ID
// correspond to a valid sign session. Returns an error if the token is invalid.
func (r *SessionRepository) ValidateLoginToken(token string, sessionID uint32, charID uint32) error {
	var t string
	return r.db.QueryRow("SELECT token FROM sign_sessions ss INNER JOIN public.users u on ss.user_id = u.id WHERE token=$1 AND ss.id=$2 AND u.id=(SELECT c.user_id FROM characters c WHERE c.id=$3)", token, sessionID, charID).Scan(&t)
}

// BindSession associates a sign session token with a server and character.
func (r *SessionRepository) BindSession(token string, serverID uint16, charID uint32) error {
	_, err := r.db.Exec("UPDATE sign_sessions SET server_id=$1, char_id=$2 WHERE token=$3", serverID, charID, token)
	return err
}

// ClearSession removes the server and character association from a sign session.
func (r *SessionRepository) ClearSession(token string) error {
	_, err := r.db.Exec("UPDATE sign_sessions SET server_id=NULL, char_id=NULL WHERE token=$1", token)
	return err
}

// UpdatePlayerCount updates the current player count for a server.
func (r *SessionRepository) UpdatePlayerCount(serverID uint16, count int) error {
	_, err := r.db.Exec("UPDATE servers SET current_players=$1 WHERE server_id=$2", count, serverID)
	return err
}
