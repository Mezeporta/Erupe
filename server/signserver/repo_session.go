package signserver

import "github.com/jmoiron/sqlx"

// SignSessionRepository implements SignSessionRepo with PostgreSQL.
type SignSessionRepository struct {
	db *sqlx.DB
}

// NewSignSessionRepository creates a new SignSessionRepository.
func NewSignSessionRepository(db *sqlx.DB) *SignSessionRepository {
	return &SignSessionRepository{db: db}
}

func (r *SignSessionRepository) RegisterUID(uid uint32, token string) (uint32, error) {
	var tid uint32
	err := r.db.QueryRow(`INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2) RETURNING id`, uid, token).Scan(&tid)
	return tid, err
}

func (r *SignSessionRepository) RegisterPSN(psnID, token string) (uint32, error) {
	var tid uint32
	err := r.db.QueryRow(`INSERT INTO sign_sessions (psn_id, token) VALUES ($1, $2) RETURNING id`, psnID, token).Scan(&tid)
	return tid, err
}

func (r *SignSessionRepository) Validate(token string, tokenID uint32) (bool, error) {
	query := `SELECT count(*) FROM sign_sessions WHERE token = $1`
	if tokenID > 0 {
		query += ` AND id = $2`
	}
	var exists int
	err := r.db.QueryRow(query, token, tokenID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *SignSessionRepository) GetPSNIDByToken(token string) (string, error) {
	var psnID string
	err := r.db.QueryRow(`SELECT psn_id FROM sign_sessions WHERE token = $1`, token).Scan(&psnID)
	return psnID, err
}
