package signserver

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// SignCharacterRepository implements SignCharacterRepo with PostgreSQL.
type SignCharacterRepository struct {
	db *sqlx.DB
}

// NewSignCharacterRepository creates a new SignCharacterRepository.
func NewSignCharacterRepository(db *sqlx.DB) *SignCharacterRepository {
	return &SignCharacterRepository{db: db}
}

func (r *SignCharacterRepository) CountNewCharacters(uid uint32) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM characters WHERE user_id = $1 AND is_new_character = true", uid).Scan(&count)
	return count, err
}

func (r *SignCharacterRepository) CreateCharacter(uid uint32, lastLogin uint32) error {
	_, err := r.db.Exec(`
		INSERT INTO characters (
			user_id, is_female, is_new_character, name, unk_desc_string,
			hr, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 0, 0, 0, $2)`,
		uid, lastLogin,
	)
	return err
}

func (r *SignCharacterRepository) GetForUser(uid uint32) ([]character, error) {
	characters := make([]character, 0)
	err := r.db.Select(&characters, "SELECT id, is_female, is_new_character, name, unk_desc_string, hr, gr, weapon_type, last_login FROM characters WHERE user_id = $1 AND deleted = false ORDER BY id", uid)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *SignCharacterRepository) IsNewCharacter(cid int) (bool, error) {
	var isNew bool
	err := r.db.QueryRow("SELECT is_new_character FROM characters WHERE id = $1", cid).Scan(&isNew)
	return isNew, err
}

func (r *SignCharacterRepository) HardDelete(cid int) error {
	_, err := r.db.Exec("DELETE FROM characters WHERE id = $1", cid)
	return err
}

func (r *SignCharacterRepository) SoftDelete(cid int) error {
	_, err := r.db.Exec("UPDATE characters SET deleted = true WHERE id = $1", cid)
	return err
}

// GetFriends returns friends for a character using parameterized queries
// (fixes the SQL injection vector from the original string-concatenated approach).
func (r *SignCharacterRepository) GetFriends(charID uint32) ([]members, error) {
	var friendsCSV string
	err := r.db.QueryRow("SELECT friends FROM characters WHERE id=$1", charID).Scan(&friendsCSV)
	if err != nil {
		return nil, err
	}
	if friendsCSV == "" {
		return nil, nil
	}

	friendsSlice := strings.Split(friendsCSV, ",")
	// Filter out empty strings
	ids := make([]string, 0, len(friendsSlice))
	for _, s := range friendsSlice {
		s = strings.TrimSpace(s)
		if s != "" {
			ids = append(ids, s)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}

	// Use parameterized ANY($1) instead of string-concatenated WHERE id=X OR id=Y
	friends := make([]members, 0)
	err = r.db.Select(&friends, "SELECT id, name FROM characters WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, err
	}
	return friends, nil
}

// GetGuildmates returns guildmates for a character.
func (r *SignCharacterRepository) GetGuildmates(charID uint32) ([]members, error) {
	var inGuild int
	err := r.db.QueryRow("SELECT count(*) FROM guild_characters WHERE character_id=$1", charID).Scan(&inGuild)
	if err != nil {
		return nil, err
	}
	if inGuild == 0 {
		return nil, nil
	}

	var guildID int
	err = r.db.QueryRow("SELECT guild_id FROM guild_characters WHERE character_id=$1", charID).Scan(&guildID)
	if err != nil {
		return nil, err
	}

	guildmates := make([]members, 0)
	err = r.db.Select(&guildmates, "SELECT character_id AS id, c.name FROM guild_characters gc JOIN characters c ON c.id = gc.character_id WHERE guild_id=$1 AND character_id!=$2", guildID, charID)
	if err != nil {
		return nil, err
	}
	return guildmates, nil
}
