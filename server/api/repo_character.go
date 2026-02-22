package api

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// APICharacterRepository implements APICharacterRepo with PostgreSQL.
type APICharacterRepository struct {
	db *sqlx.DB
}

// NewAPICharacterRepository creates a new APICharacterRepository.
func NewAPICharacterRepository(db *sqlx.DB) *APICharacterRepository {
	return &APICharacterRepository{db: db}
}

func (r *APICharacterRepository) GetNewCharacter(ctx context.Context, userID uint32) (Character, error) {
	var character Character
	err := r.db.GetContext(ctx, &character,
		"SELECT id, name, is_female, weapon_type, hr, gr, last_login FROM characters WHERE is_new_character = true AND user_id = $1 LIMIT 1",
		userID,
	)
	return character, err
}

func (r *APICharacterRepository) CountForUser(ctx context.Context, userID uint32) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM characters WHERE user_id = $1", userID).Scan(&count)
	return count, err
}

func (r *APICharacterRepository) Create(ctx context.Context, userID uint32, lastLogin uint32) (Character, error) {
	var character Character
	err := r.db.GetContext(ctx, &character, `
		INSERT INTO characters (
			user_id, is_female, is_new_character, name, unk_desc_string,
			hr, gr, weapon_type, last_login
		)
		VALUES ($1, false, true, '', '', 0, 0, 0, $2)
		RETURNING id, name, is_female, weapon_type, hr, gr, last_login`,
		userID, lastLogin,
	)
	return character, err
}

func (r *APICharacterRepository) IsNew(charID uint32) (bool, error) {
	var isNew bool
	err := r.db.QueryRow("SELECT is_new_character FROM characters WHERE id = $1", charID).Scan(&isNew)
	return isNew, err
}

func (r *APICharacterRepository) HardDelete(charID uint32) error {
	_, err := r.db.Exec("DELETE FROM characters WHERE id = $1", charID)
	return err
}

func (r *APICharacterRepository) SoftDelete(charID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET deleted = true WHERE id = $1", charID)
	return err
}

func (r *APICharacterRepository) GetForUser(ctx context.Context, userID uint32) ([]Character, error) {
	var characters []Character
	err := r.db.SelectContext(
		ctx, &characters, `
		SELECT id, name, is_female, weapon_type, hr, gr, last_login
		FROM characters
		WHERE user_id = $1 AND deleted = false AND is_new_character = false ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *APICharacterRepository) ExportSave(ctx context.Context, userID, charID uint32) (map[string]interface{}, error) {
	row := r.db.QueryRowxContext(ctx, "SELECT * FROM characters WHERE id=$1 AND user_id=$2", charID, userID)
	result := make(map[string]interface{})
	err := row.MapScan(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
