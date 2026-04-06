package api

import (
	"context"
	"errors"
	"time"

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
	if err != nil {
		return character, err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO user_binary (id) VALUES ($1)`, character.ID)
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

func (r *APICharacterRepository) GrantImportToken(ctx context.Context, charID, userID uint32, token string, expiry time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE characters SET savedata_import_token=$1, savedata_import_token_expiry=$2
         WHERE id=$3 AND user_id=$4 AND deleted=false`,
		token, expiry, charID, userID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("character not found or not owned by user")
	}
	return nil
}

func (r *APICharacterRepository) RevokeImportToken(ctx context.Context, charID, userID uint32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET savedata_import_token=NULL, savedata_import_token_expiry=NULL
         WHERE id=$1 AND user_id=$2`,
		charID, userID,
	)
	return err
}

func (r *APICharacterRepository) ImportSave(ctx context.Context, charID, userID uint32, token string, blobs SaveBlobs) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Validate token ownership and expiry, then clear it — all in one UPDATE.
	res, err := tx.ExecContext(ctx,
		`UPDATE characters
         SET savedata_import_token=NULL, savedata_import_token_expiry=NULL
         WHERE id=$1 AND user_id=$2
           AND savedata_import_token=$3
           AND savedata_import_token_expiry > now()`,
		charID, userID, token,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("import token invalid, expired, or character not owned by user")
	}

	// Write all save blobs.
	_, err = tx.ExecContext(ctx,
		`UPDATE characters SET
            savedata=$1, savedata_hash=$2, decomyset=$3, hunternavi=$4,
            otomoairou=$5, partner=$6, platebox=$7, platedata=$8,
            platemyset=$9, rengokudata=$10, savemercenary=$11, gacha_items=$12,
            house_info=$13, login_boost=$14, skin_hist=$15, scenariodata=$16,
            savefavoritequest=$17, mezfes=$18
         WHERE id=$19`,
		blobs.Savedata, blobs.SavedataHash, blobs.Decomyset, blobs.Hunternavi,
		blobs.Otomoairou, blobs.Partner, blobs.Platebox, blobs.Platedata,
		blobs.Platemyset, blobs.Rengokudata, blobs.Savemercenary, blobs.GachaItems,
		blobs.HouseInfo, blobs.LoginBoost, blobs.SkinHist, blobs.Scenariodata,
		blobs.Savefavoritequest, blobs.Mezfes,
		charID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}
