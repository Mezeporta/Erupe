package api

import (
	"context"
	"database/sql"
	"erupe-ce/common/token"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *APIServer) createNewUser(ctx context.Context, username string, password string) (uint32, uint32, error) {
	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, 0, err
	}

	var (
		id     uint32
		rights uint32
	)
	err = s.db.QueryRowContext(
		ctx, `
		INSERT INTO users (username, password, return_expires)
		VALUES ($1, $2, $3)
		RETURNING id, rights
		`,
		username, string(passwordHash), time.Now().Add(time.Hour*24*30),
	).Scan(&id, &rights)
	return id, rights, err
}

func (s *APIServer) createLoginToken(ctx context.Context, uid uint32) (uint32, string, error) {
	loginToken := token.Generate(16)
	var tid uint32
	err := s.db.QueryRowContext(ctx, "INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2) RETURNING id", uid, loginToken).Scan(&tid)
	if err != nil {
		return 0, "", err
	}
	return tid, loginToken, nil
}

func (s *APIServer) userIDFromToken(ctx context.Context, token string) (uint32, error) {
	var userID uint32
	err := s.db.QueryRowContext(ctx, "SELECT user_id FROM sign_sessions WHERE token = $1", token).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid login token")
	} else if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *APIServer) createCharacter(ctx context.Context, userID uint32) (Character, error) {
	var character Character
	err := s.db.GetContext(ctx, &character,
		"SELECT id, name, is_female, weapon_type, hr, gr, last_login FROM characters WHERE is_new_character = true AND user_id = $1 LIMIT 1",
		userID,
	)
	if err == sql.ErrNoRows {
		var count int
		_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM characters WHERE user_id = $1", userID).Scan(&count)
		if count >= 16 {
			return character, fmt.Errorf("cannot have more than 16 characters")
		}
		err = s.db.GetContext(ctx, &character, `
			INSERT INTO characters (
				user_id, is_female, is_new_character, name, unk_desc_string,
				hr, gr, weapon_type, last_login
			)
			VALUES ($1, false, true, '', '', 0, 0, 0, $2)
			RETURNING id, name, is_female, weapon_type, hr, gr, last_login`,
			userID, uint32(time.Now().Unix()),
		)
	}
	return character, err
}

func (s *APIServer) deleteCharacter(ctx context.Context, userID uint32, charID uint32) error {
	var isNew bool
	err := s.db.QueryRow("SELECT is_new_character FROM characters WHERE id = $1", charID).Scan(&isNew)
	if err != nil {
		return err
	}
	if isNew {
		_, err = s.db.Exec("DELETE FROM characters WHERE id = $1", charID)
	} else {
		_, err = s.db.Exec("UPDATE characters SET deleted = true WHERE id = $1", charID)
	}
	return err
}

func (s *APIServer) getCharactersForUser(ctx context.Context, uid uint32) ([]Character, error) {
	var characters []Character
	err := s.db.SelectContext(
		ctx, &characters, `
		SELECT id, name, is_female, weapon_type, hr, gr, last_login
		FROM characters
		WHERE user_id = $1 AND deleted = false AND is_new_character = false ORDER BY id ASC`,
		uid,
	)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func (s *APIServer) getReturnExpiry(uid uint32) time.Time {
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

func (s *APIServer) exportSave(ctx context.Context, uid uint32, cid uint32) (map[string]interface{}, error) {
	row := s.db.QueryRowxContext(ctx, "SELECT * FROM characters WHERE id=$1 AND user_id=$2", cid, uid)
	result := make(map[string]interface{})
	err := row.MapScan(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
