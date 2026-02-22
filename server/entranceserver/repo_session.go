package entranceserver

import "github.com/jmoiron/sqlx"

// EntranceSessionRepository implements EntranceSessionRepo with PostgreSQL.
type EntranceSessionRepository struct {
	db *sqlx.DB
}

// NewEntranceSessionRepository creates a new EntranceSessionRepository.
func NewEntranceSessionRepository(db *sqlx.DB) *EntranceSessionRepository {
	return &EntranceSessionRepository{db: db}
}

func (r *EntranceSessionRepository) GetServerIDForCharacter(charID uint32) (uint16, error) {
	var sid uint16
	err := r.db.QueryRow("SELECT(SELECT server_id FROM sign_sessions WHERE char_id=$1) AS _", charID).Scan(&sid)
	if err != nil {
		return 0, err
	}
	return sid, nil
}
