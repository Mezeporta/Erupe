package entranceserver

import "github.com/jmoiron/sqlx"

// EntranceServerRepository implements EntranceServerRepo with PostgreSQL.
type EntranceServerRepository struct {
	db *sqlx.DB
}

// NewEntranceServerRepository creates a new EntranceServerRepository.
func NewEntranceServerRepository(db *sqlx.DB) *EntranceServerRepository {
	return &EntranceServerRepository{db: db}
}

func (r *EntranceServerRepository) GetCurrentPlayers(serverID int) (uint16, error) {
	var currentPlayers uint16
	err := r.db.QueryRow("SELECT current_players FROM servers WHERE server_id=$1", serverID).Scan(&currentPlayers)
	if err != nil {
		return 0, err
	}
	return currentPlayers, nil
}
