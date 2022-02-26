package signserver

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerDBAccount(username string, password string) error {
	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, string(passwordHash))
	if err != nil {
		return err
	}

	var id int
	err = s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return err
	}

	// Create a base new character.
	_, err = s.db.Exec(`
		INSERT INTO characters (
			user_id, is_female, is_new_character, small_gr_level, gr_override_mode, name, unk_desc_string,
			gr_override_level, gr_override_unk0, gr_override_unk1, exp, weapon, last_login)
		VALUES($1, False, True, 0, True, '', '', 0, 0, 0, 0, 0, $2)`,
		id,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

type character struct {
	ID              uint32 `db:"id"`
	IsFemale        bool   `db:"is_female"`
	IsNewCharacter  bool   `db:"is_new_character"`
	SmallGRLevel    uint8  `db:"small_gr_level"`
	GROverrideMode  bool   `db:"gr_override_mode"`
	Name            string `db:"name"`
	UnkDescString   string `db:"unk_desc_string"`
	GROverrideLevel uint16 `db:"gr_override_level"`
	GROverrideUnk0  uint8  `db:"gr_override_unk0"`
	GROverrideUnk1  uint8  `db:"gr_override_unk1"`
	Exp             uint16 `db:"exp"`
	Weapon          uint16 `db:"weapon"`
	LastLogin       uint32 `db:"last_login"`
}

func (s *Server) getCharactersForUser(uid int) ([]character, error) {
	characters := []character{}
	err := s.db.Select(&characters, "SELECT id, is_female, is_new_character, small_gr_level, gr_override_mode, name, unk_desc_string, gr_override_level, gr_override_unk0, gr_override_unk1, exp, weapon, last_login FROM characters WHERE user_id = $1", uid)
	if err != nil {
		return nil, err
	}
	return characters, nil
}