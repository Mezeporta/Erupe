package api

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type apiEventRepository struct {
	db *sqlx.DB
}

// NewAPIEventRepository creates an APIEventRepo backed by PostgreSQL.
func NewAPIEventRepository(db *sqlx.DB) APIEventRepo {
	return &apiEventRepository{db: db}
}

func (r *apiEventRepository) GetFeatureWeapon(ctx context.Context, startTime time.Time) (*FeatureWeaponRow, error) {
	var row FeatureWeaponRow
	err := r.db.GetContext(ctx, &row, `SELECT start_time, featured FROM feature_weapon WHERE start_time=$1`, startTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *apiEventRepository) GetActiveEvents(ctx context.Context, eventType string) ([]EventRow, error) {
	var rows []EventRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type=$1`, eventType)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
