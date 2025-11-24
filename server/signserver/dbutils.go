package signserver

import (
	"erupe-ce/common/mhfcourse"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) newUserChara(username string) error {
	var id int
	err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return err
	}

	var numNewChars int
	err = s.db.QueryRow("SELECT COUNT(*) FROM characters WHERE user_id = $1 AND is_new_character = true", id).Scan(&numNewChars)
	if err != nil {
		return err
	}

	// prevent users with an uninitialised character from creating more
	if numNewChars >= 1 {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO characters (
			user_id, is_female, is_new_character, name, unk_desc_string,
			hrp, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 0, 0, 0, $2)`,
		id,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) registerDBAccount(username string, password string) error {
	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO users (username, password, return_expires) VALUES ($1, $2, $3)", username, string(passwordHash), time.Now().Add(time.Hour*24*30))
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
			user_id, is_female, is_new_character, name, unk_desc_string,
			hrp, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 0, 0, 0, $2)`,
		id,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

type character struct {
	ID             uint32 `db:"id"`
	IsFemale       bool   `db:"is_female"`
	IsNewCharacter bool   `db:"is_new_character"`
	Name           string `db:"name"`
	UnkDescString  string `db:"unk_desc_string"`
	HRP            uint16 `db:"hrp"`
	GR             uint16 `db:"gr"`
	WeaponType     uint16 `db:"weapon_type"`
	LastLogin      uint32 `db:"last_login"`
}

func (s *Server) getCharactersForUser(uid int) ([]character, error) {
	characters := make([]character, 0)
	err := s.db.Select(&characters, "SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = $1 AND deleted = false ORDER BY id ASC", uid)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func (s *Server) getReturnExpiry(uid int) time.Time {
	var returnExpiry, lastLogin time.Time
	s.db.Get(&lastLogin, "SELECT COALESCE(last_login, now()) FROM users WHERE id=$1", uid)
	if time.Now().Add((time.Hour * 24) * -90).After(lastLogin) {
		returnExpiry = time.Now().Add(time.Hour * 24 * 30)
		s.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
	} else {
		err := s.db.Get(&returnExpiry, "SELECT return_expires FROM users WHERE id=$1", uid)
		if err != nil {
			returnExpiry = time.Now()
			s.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
		}
	}
	s.db.Exec("UPDATE users SET last_login=$1 WHERE id=$2", time.Now(), uid)
	return returnExpiry
}

func (s *Server) getLastCID(uid int) uint32 {
	var lastPlayed uint32
	_ = s.db.QueryRow("SELECT last_character FROM users WHERE id=$1", uid).Scan(&lastPlayed)
	return lastPlayed
}

func (s *Server) getUserRights(uid int) uint32 {
	rights := uint32(2)
	_ = s.db.QueryRow("SELECT rights FROM users WHERE id=$1", uid).Scan(&rights)
	_, rights = mhfcourse.GetCourseStruct(rights)
	return rights
}

type members struct {
	CID  uint32 // Local character ID
	ID   uint32 `db:"id"`
	Name string `db:"name"`
}

func (s *Server) getFriendsForCharacters(chars []character) []members {
	friends := make([]members, 0)
	for _, char := range chars {
		friendsCSV := ""
		err := s.db.QueryRow("SELECT friends FROM characters WHERE id=$1", char.ID).Scan(&friendsCSV)
		friendsSlice := strings.Split(friendsCSV, ",")
		friendQuery := "SELECT id, name FROM characters WHERE id="
		for i := 0; i < len(friendsSlice); i++ {
			friendQuery += friendsSlice[i]
			if i+1 != len(friendsSlice) {
				friendQuery += " OR id="
			}
		}
		charFriends := make([]members, 0)
		err = s.db.Select(&charFriends, friendQuery)
		if err != nil {
			continue
		}
		for i := range charFriends {
			charFriends[i].CID = char.ID
		}
		friends = append(friends, charFriends...)
	}
	if len(friends) > 255 { // Uint8
		friends = friends[:255]
	}
	return friends
}

func (s *Server) getGuildmatesForCharacters(chars []character) []members {
	guildmates := make([]members, 0)
	for _, char := range chars {
		var inGuild int
		_ = s.db.QueryRow("SELECT count(*) FROM guild_characters WHERE character_id=$1", char.ID).Scan(&inGuild)
		if inGuild > 0 {
			var guildID int
			err := s.db.QueryRow("SELECT guild_id FROM guild_characters WHERE character_id=$1", char.ID).Scan(&guildID)
			if err != nil {
				continue
			}
			charGuildmates := make([]members, 0)
			err = s.db.Select(&charGuildmates, "SELECT character_id AS id, c.name FROM guild_characters gc JOIN characters c ON c.id = gc.character_id WHERE guild_id=$1 AND character_id!=$2", guildID, char.ID)
			if err != nil {
				continue
			}
			for i := range charGuildmates {
				charGuildmates[i].CID = char.ID
			}
			guildmates = append(guildmates, charGuildmates...)
		}
	}
	if len(guildmates) > 255 { // Uint8
		guildmates = guildmates[:255]
	}
	return guildmates
}

func (s *Server) deleteCharacter(cid int, token string) error {
	var verify int
	err := s.db.QueryRow("SELECT count(*) FROM sign_sessions WHERE token = $1", token).Scan(&verify)
	if err != nil {
		return err // Invalid token
	}
	var isNew bool
	err = s.db.QueryRow("SELECT is_new_character FROM characters WHERE id = $1", cid).Scan(&isNew)
	if isNew {
		_, err = s.db.Exec("DELETE FROM characters WHERE id = $1", cid)
	} else {
		_, err = s.db.Exec("UPDATE characters SET deleted = true WHERE id = $1", cid)
	}
	if err != nil {
		return err
	}
	return nil
}

// Unused
func (s *Server) checkToken(uid int) (bool, error) {
	var exists int
	err := s.db.QueryRow("SELECT count(*) FROM sign_sessions WHERE user_id = $1", uid).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists > 0 {
		return true, nil
	}
	return false, nil
}

func (s *Server) registerToken(uid int, token string) error {
	_, err := s.db.Exec("INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2)", uid, token)
	if err != nil {
		return err
	}
	return nil
}
