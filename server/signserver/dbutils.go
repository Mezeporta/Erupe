package signserver

import (
	"database/sql"
	"errors"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/common/token"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) newUserChara(uid uint32) error {
	var numNewChars int
	err := s.db.QueryRow("SELECT COUNT(*) FROM characters WHERE user_id = $1 AND is_new_character = true", uid).Scan(&numNewChars)
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
			hr, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 0, 0, 0, $2)`,
		uid,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) registerDBAccount(username string, password string) (uint32, error) {
	var uid uint32
	s.logger.Info("Creating user", zap.String("User", username))

	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	err = s.db.QueryRow("INSERT INTO users (username, password, return_expires) VALUES ($1, $2, $3) RETURNING id", username, string(passwordHash), time.Now().Add(time.Hour*24*30)).Scan(&uid)
	if err != nil {
		return 0, err
	}

	return uid, nil
}

type character struct {
	ID             uint32 `db:"id"`
	IsFemale       bool   `db:"is_female"`
	IsNewCharacter bool   `db:"is_new_character"`
	Name           string `db:"name"`
	UnkDescString  string `db:"unk_desc_string"`
	HR             uint16 `db:"hr"`
	GR             uint16 `db:"gr"`
	WeaponType     uint16 `db:"weapon_type"`
	LastLogin      uint32 `db:"last_login"`
}

func (s *Server) getCharactersForUser(uid uint32) ([]character, error) {
	characters := make([]character, 0)
	err := s.db.Select(&characters, "SELECT id, is_female, is_new_character, name, unk_desc_string, hr, gr, weapon_type, last_login FROM characters WHERE user_id = $1 AND deleted = false ORDER BY id", uid)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func (s *Server) getReturnExpiry(uid uint32) time.Time {
	var returnExpiry, lastLogin time.Time
	_ = s.db.Get(&lastLogin, "SELECT COALESCE(last_login, now()) FROM users WHERE id=$1", uid)
	if time.Now().Add((time.Hour * 24) * -90).After(lastLogin) {
		returnExpiry = time.Now().Add(time.Hour * 24 * 30)
		_, _ = s.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
	} else {
		err := s.db.Get(&returnExpiry, "SELECT return_expires FROM users WHERE id=$1", uid)
		if err != nil {
			returnExpiry = time.Now()
			_, _ = s.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
		}
	}
	_, _ = s.db.Exec("UPDATE users SET last_login=$1 WHERE id=$2", time.Now(), uid)
	return returnExpiry
}

func (s *Server) getLastCID(uid uint32) uint32 {
	var lastPlayed uint32
	_ = s.db.QueryRow("SELECT last_character FROM users WHERE id=$1", uid).Scan(&lastPlayed)
	return lastPlayed
}

func (s *Server) getUserRights(uid uint32) uint32 {
	var rights uint32
	if uid != 0 {
		_ = s.db.QueryRow("SELECT rights FROM users WHERE id=$1", uid).Scan(&rights)
		_, rights = mhfcourse.GetCourseStruct(rights)
	}
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
		_ = s.db.QueryRow("SELECT friends FROM characters WHERE id=$1", char.ID).Scan(&friendsCSV)
		friendsSlice := strings.Split(friendsCSV, ",")
		friendQuery := "SELECT id, name FROM characters WHERE id="
		for i := 0; i < len(friendsSlice); i++ {
			friendQuery += friendsSlice[i]
			if i+1 != len(friendsSlice) {
				friendQuery += " OR id="
			}
		}
		charFriends := make([]members, 0)
		err := s.db.Select(&charFriends, friendQuery)
		if err != nil {
			continue
		}
		for i := range charFriends {
			charFriends[i].CID = char.ID
		}
		friends = append(friends, charFriends...)
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
	return guildmates
}

func (s *Server) deleteCharacter(cid int, token string, tokenID uint32) error {
	if !s.validateToken(token, tokenID) {
		return errors.New("invalid token")
	}
	var isNew bool
	err := s.db.QueryRow("SELECT is_new_character FROM characters WHERE id = $1", cid).Scan(&isNew)
	if err != nil {
		return err
	}
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

func (s *Server) registerUidToken(uid uint32) (uint32, string, error) {
	_token := token.Generate(16)
	var tid uint32
	err := s.db.QueryRow(`INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2) RETURNING id`, uid, _token).Scan(&tid)
	return tid, _token, err
}

func (s *Server) registerPsnToken(psn string) (uint32, string, error) {
	_token := token.Generate(16)
	var tid uint32
	err := s.db.QueryRow(`INSERT INTO sign_sessions (psn_id, token) VALUES ($1, $2) RETURNING id`, psn, _token).Scan(&tid)
	return tid, _token, err
}

func (s *Server) validateToken(token string, tokenID uint32) bool {
	query := `SELECT count(*) FROM sign_sessions WHERE token = $1`
	if tokenID > 0 {
		query += ` AND id = $2`
	}
	var exists int
	err := s.db.QueryRow(query, token, tokenID).Scan(&exists)
	if err != nil || exists == 0 {
		return false
	}
	return true
}

func (s *Server) validateLogin(user string, pass string) (uint32, RespID) {
	var uid uint32
	var passDB string
	err := s.db.QueryRow(`SELECT id, password FROM users WHERE username = $1`, user).Scan(&uid, &passDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Info("User not found", zap.String("User", user))
			if s.erupeConfig.AutoCreateAccount {
				uid, err = s.registerDBAccount(user, pass)
				if err == nil {
					return uid, SIGN_SUCCESS
				} else {
					return 0, SIGN_EABORT
				}
			}
			return 0, SIGN_EAUTH
		}
		return 0, SIGN_EABORT
	} else {
		if bcrypt.CompareHashAndPassword([]byte(passDB), []byte(pass)) == nil {
			var bans int
			err = s.db.QueryRow(`SELECT count(*) FROM bans WHERE user_id=$1 AND expires IS NULL`, uid).Scan(&bans)
			if err == nil && bans > 0 {
				return uid, SIGN_EELIMINATE
			}
			err = s.db.QueryRow(`SELECT count(*) FROM bans WHERE user_id=$1 AND expires > now()`, uid).Scan(&bans)
			if err == nil && bans > 0 {
				return uid, SIGN_ESUSPEND
			}
			return uid, SIGN_SUCCESS
		}
		return 0, SIGN_EPASS
	}
}
