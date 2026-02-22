package api

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// APISessionRepository implements APISessionRepo with PostgreSQL.
type APISessionRepository struct {
	db *sqlx.DB
}

// NewAPISessionRepository creates a new APISessionRepository.
func NewAPISessionRepository(db *sqlx.DB) *APISessionRepository {
	return &APISessionRepository{db: db}
}

func (r *APISessionRepository) CreateToken(ctx context.Context, uid uint32, token string) (uint32, error) {
	var tid uint32
	err := r.db.QueryRowContext(ctx, "INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2) RETURNING id", uid, token).Scan(&tid)
	return tid, err
}

func (r *APISessionRepository) GetUserIDByToken(ctx context.Context, token string) (uint32, error) {
	var userID uint32
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM sign_sessions WHERE token = $1", token).Scan(&userID)
	return userID, err
}
