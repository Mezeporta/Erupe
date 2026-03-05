package api

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// APIUserRepository implements APIUserRepo with PostgreSQL.
type APIUserRepository struct {
	db *sqlx.DB
}

// NewAPIUserRepository creates a new APIUserRepository.
func NewAPIUserRepository(db *sqlx.DB) *APIUserRepository {
	return &APIUserRepository{db: db}
}

func (r *APIUserRepository) Register(ctx context.Context, username, passwordHash string, returnExpires time.Time) (uint32, uint32, error) {
	var (
		id     uint32
		rights uint32
	)
	err := r.db.QueryRowContext(
		ctx, `
		INSERT INTO users (username, password, return_expires)
		VALUES ($1, $2, $3)
		RETURNING id, rights
		`,
		username, passwordHash, returnExpires,
	).Scan(&id, &rights)
	return id, rights, err
}

func (r *APIUserRepository) GetCredentials(ctx context.Context, username string) (uint32, string, uint32, error) {
	var (
		id           uint32
		passwordHash string
		rights       uint32
	)
	err := r.db.QueryRowContext(ctx, "SELECT id, password, rights FROM users WHERE username = $1", username).Scan(&id, &passwordHash, &rights)
	return id, passwordHash, rights, err
}

func (r *APIUserRepository) GetLastLogin(uid uint32) (time.Time, error) {
	var lastLogin time.Time
	err := r.db.Get(&lastLogin, "SELECT COALESCE(last_login, now()) FROM users WHERE id=$1", uid)
	return lastLogin, err
}

func (r *APIUserRepository) GetReturnExpiry(uid uint32) (time.Time, error) {
	var returnExpiry time.Time
	err := r.db.Get(&returnExpiry, "SELECT return_expires FROM users WHERE id=$1", uid)
	return returnExpiry, err
}

func (r *APIUserRepository) UpdateReturnExpiry(uid uint32, expiry time.Time) error {
	_, err := r.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", expiry, uid)
	return err
}

func (r *APIUserRepository) UpdateLastLogin(uid uint32, loginTime time.Time) error {
	_, err := r.db.Exec("UPDATE users SET last_login=$1 WHERE id=$2", loginTime, uid)
	return err
}
